// Package dispatch takes pending plugin items out of the inbox and turns them
// into printer jobs. It enforces per-binding rate limits, tracks per-item
// retry budgets, and wires the inbox state machine to the printer service.
package dispatch

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/inbox"
	"github.com/ruhuang/ink/server/internal/plugins"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/workspace"
)

// PrinterJobCreator is the subset of printer.Service the dispatcher needs.
type PrinterJobCreator interface {
	CreatePrintJobForUser(ctx context.Context, userID string, input printer.CreateJobInput) (workspace.PrintJob, error)
}

// PluginRuntime exposes the plugin lookups the dispatcher uses when turning
// items back into print jobs.
type PluginRuntime interface {
	GetBindingByID(ctx context.Context, bindingID string) (plugins.Binding, map[string]string, error)
	GetInstallation(ctx context.Context, installationID string) (plugins.Installation, plugins.Manifest, error)
}

// WorkspaceRepository mirrors the workspace lookup the scheduler already uses.
type WorkspaceRepository interface {
	FindByUserID(ctx context.Context, userID string) (*workspace.State, error)
}

// Clock returns wall time.
type Clock interface {
	Now() time.Time
}

// DefaultBatchSize is the maximum number of items the dispatcher will pull
// from the inbox per FlushBinding invocation when the caller does not set a
// per-binding limit.
const DefaultBatchSize = 20

// DefaultDailyCap limits printing to avoid thermal paper runaway when a
// binding has not configured its own MaxPrintsPerDay.
const DefaultDailyCap = 50

// DefaultRetryBackoff is the minimum delay between failed dispatch attempts.
const DefaultRetryBackoff = 15 * time.Minute

// DailyCounter returns how many items a binding has already printed within
// the rolling 24h window. Implementations are typically backed by the inbox
// repository using a count query.
type DailyCounter interface {
	CountPrintedInLast24h(ctx context.Context, bindingID string, since time.Time) (int, error)
}

// Service dispatches items out of the inbox into print jobs.
type Service struct {
	inbox        *inbox.Service
	plugins      PluginRuntime
	printer      PrinterJobCreator
	workspace    WorkspaceRepository
	counter      DailyCounter
	clock        Clock
	retryBackoff time.Duration
}

// NewService builds a dispatcher.
func NewService(
	inboxService *inbox.Service,
	pluginRuntime PluginRuntime,
	printerCreator PrinterJobCreator,
	workspaceRepo WorkspaceRepository,
	counter DailyCounter,
	clock Clock,
	retryBackoff time.Duration,
) *Service {
	if retryBackoff <= 0 {
		retryBackoff = DefaultRetryBackoff
	}
	return &Service{
		inbox:        inboxService,
		plugins:      pluginRuntime,
		printer:      printerCreator,
		workspace:    workspaceRepo,
		counter:      counter,
		clock:        clock,
		retryBackoff: retryBackoff,
	}
}

// FlushResult reports what happened during a flush.
type FlushResult struct {
	Printed     int
	Failed      int
	Skipped     int
	PrintJobIDs []string
}

// FlushBinding drains pending items for a binding into the printer up to the
// configured rate limits. It is safe to call from both schedule processors
// and the manual trigger endpoint.
func (s *Service) FlushBinding(ctx context.Context, bindingID string, defaultDeviceID string) (FlushResult, error) {
	return s.dispatchPending(ctx, bindingID, defaultDeviceID, 0)
}

// DrainPending flushes pending items across bindings so backlog created by a
// previous over-budget fetch can continue draining even when the source goes
// quiet. The limit caps the total number of items processed in one pass.
func (s *Service) DrainPending(ctx context.Context, limit int) (FlushResult, error) {
	if limit <= 0 {
		limit = DefaultBatchSize
	}

	bindingIDs, err := s.inbox.ListPendingBindingIDs(ctx, limit)
	if err != nil {
		return FlushResult{}, err
	}

	aggregate := FlushResult{}
	remaining := limit
	for _, bindingID := range bindingIDs {
		if remaining <= 0 {
			break
		}
		result, err := s.dispatchPending(ctx, bindingID, "", remaining)
		if err != nil {
			continue
		}
		mergeFlushResult(&aggregate, result)
		processed := result.Printed + result.Failed + result.Skipped
		if processed > 0 {
			remaining -= processed
		}
	}
	return aggregate, nil
}

// dispatchBudget returns the number of items the dispatcher is allowed to
// print for the given binding right now, respecting both the per-run and
// per-day caps. A non-positive return means the binding is saturated and
// callers should skip.
func (s *Service) dispatchBudget(ctx context.Context, binding plugins.Binding) (int, error) {
	perRun := binding.MaxPrintsPerRun
	if perRun <= 0 {
		perRun = DefaultBatchSize
	}
	perDay := binding.MaxPrintsPerDay
	if perDay <= 0 {
		perDay = DefaultDailyCap
	}
	printedToday, err := s.counter.CountPrintedInLast24h(ctx, binding.ID, s.clock.Now().Add(-24*time.Hour))
	if err != nil {
		return 0, err
	}
	remainingDay := perDay - printedToday
	if remainingDay <= 0 {
		return 0, nil
	}
	if remainingDay < perRun {
		return remainingDay, nil
	}
	return perRun, nil
}

func (s *Service) resolveDispatchContext(ctx context.Context, bindingID string) (plugins.Binding, plugins.Installation, error) {
	binding, _, err := s.plugins.GetBindingByID(ctx, bindingID)
	if err != nil {
		return plugins.Binding{}, plugins.Installation{}, err
	}
	installation, _, err := s.plugins.GetInstallation(ctx, binding.PluginInstallationID)
	if err != nil {
		return plugins.Binding{}, plugins.Installation{}, err
	}
	return binding, installation, nil
}

func (s *Service) dispatchPending(ctx context.Context, bindingID string, defaultDeviceID string, limit int) (FlushResult, error) {
	result := FlushResult{}
	if strings.TrimSpace(bindingID) == "" {
		return result, errors.New("binding id is required")
	}

	binding, installation, err := s.resolveDispatchContext(ctx, bindingID)
	if err != nil {
		return result, err
	}

	budget, err := s.dispatchBudget(ctx, binding)
	if err != nil {
		return result, err
	}
	if limit > 0 && budget > limit {
		budget = limit
	}
	if budget <= 0 {
		return result, nil
	}

	items, err := s.inbox.ListPendingByBinding(ctx, bindingID, budget)
	if err != nil {
		return result, err
	}

	return s.dispatchItems(ctx, binding, installation, items, defaultDeviceID)
}

// RetryFailed pulls retryable failed items across all bindings and attempts
// to flush them. It is intended to be called periodically from a background
// runner.
func (s *Service) RetryFailed(ctx context.Context, limit int) (FlushResult, error) {
	if limit <= 0 {
		limit = DefaultBatchSize
	}
	cutoff := s.clock.Now().Add(-s.retryBackoff)
	items, err := s.inbox.ListRetryable(ctx, cutoff, limit)
	if err != nil {
		return FlushResult{}, err
	}

	byBinding := map[string][]inbox.Item{}
	for _, item := range items {
		byBinding[item.PluginBindingID] = append(byBinding[item.PluginBindingID], item)
	}

	aggregate := FlushResult{}
	for bindingID, batch := range byBinding {
		result, ok := s.retryBindingBatch(ctx, bindingID, batch)
		if !ok {
			continue
		}
		mergeFlushResult(&aggregate, result)
	}
	return aggregate, nil
}

// retryBindingBatch flushes one binding's retry batch respecting the per-run
// and per-day caps. Returns (result, true) on success; (_, false) signals the
// binding should be skipped (missing binding/installation or budget lookup
// failure). Skipped-because-budget-exhausted is reported inside the result.
func (s *Service) retryBindingBatch(ctx context.Context, bindingID string, batch []inbox.Item) (FlushResult, bool) {
	binding, installation, err := s.resolveDispatchContext(ctx, bindingID)
	if err != nil {
		return FlushResult{}, false
	}
	budget, err := s.dispatchBudget(ctx, binding)
	if err != nil {
		return FlushResult{}, false
	}
	if budget <= 0 {
		return FlushResult{Skipped: len(batch)}, true
	}
	skipped := 0
	if len(batch) > budget {
		skipped = len(batch) - budget
		batch = batch[:budget]
	}
	result, err := s.dispatchItems(ctx, binding, installation, batch, "")
	if err != nil {
		return FlushResult{}, false
	}
	result.Skipped += skipped
	return result, true
}

func mergeFlushResult(dst *FlushResult, src FlushResult) {
	dst.Printed += src.Printed
	dst.Failed += src.Failed
	dst.Skipped += src.Skipped
	dst.PrintJobIDs = append(dst.PrintJobIDs, src.PrintJobIDs...)
}

// resolveSendConfirmation loads the caller's workspace preferences and
// returns whether print jobs should wait for user confirmation before
// submission. Missing workspace or lookup errors fall back to the platform
// default, which is true.
func (s *Service) resolveSendConfirmation(ctx context.Context, userID string) bool {
	if s.workspace == nil {
		return workspace.EmptyState().Preferences.SendConfirmationEnabled
	}
	state, err := s.workspace.FindByUserID(ctx, userID)
	if err != nil {
		return true
	}
	if state == nil {
		return workspace.EmptyState().Preferences.SendConfirmationEnabled
	}
	return workspace.NormalizeState(*state).Preferences.SendConfirmationEnabled
}

// pickDeviceID chooses the device id for a given item, preferring the caller's
// current default and only falling back to the stored inbox snapshot when the
// caller has no fresher routing information.
func pickDeviceID(item inbox.Item, defaultDeviceID string) string {
	if trimmed := strings.TrimSpace(defaultDeviceID); trimmed != "" {
		return trimmed
	}
	if item.DeviceID != nil && strings.TrimSpace(*item.DeviceID) != "" {
		return strings.TrimSpace(*item.DeviceID)
	}
	return ""
}

// dispatchOne attempts to print one item. It returns the outcome so the
// caller can aggregate results. Any inbox status transition is performed
// here so dispatchItems stays a thin loop.
type dispatchOutcome int

const (
	outcomePrinted dispatchOutcome = iota
	outcomeFailed
	outcomeSkipped
)

func (s *Service) dispatchOne(
	ctx context.Context,
	binding plugins.Binding,
	installation plugins.Installation,
	item inbox.Item,
	defaultDeviceID string,
	sendConfirmation bool,
) (dispatchOutcome, string) {
	if item.AttemptCount >= inbox.MaxDispatchAttempts {
		return outcomeSkipped, ""
	}

	rendered, err := printer.RenderBlocksToText(item.Blocks)
	if err != nil {
		_ = s.inbox.MarkFailed(ctx, item, fmt.Sprintf("render: %s", err.Error()))
		return outcomeFailed, ""
	}

	source := strings.TrimSpace(item.SourceLabel)
	if source == "" {
		source = installation.DisplayName
	}

	deviceID := pickDeviceID(item, defaultDeviceID)
	if strings.TrimSpace(deviceID) == "" {
		_ = s.inbox.MarkFailed(ctx, item, "no device bound for item")
		return outcomeFailed, ""
	}

	job, err := s.printer.CreatePrintJobForUser(ctx, binding.UserID, printer.CreateJobInput{
		Title:             item.Title,
		Source:            source,
		Content:           rendered,
		PrinterBindingID:  deviceID,
		SubmitImmediately: !sendConfirmation,
	})
	if err != nil {
		_ = s.inbox.MarkFailed(ctx, item, err.Error())
		return outcomeFailed, ""
	}
	if err := s.inbox.MarkPrinted(ctx, item, job.ID); err != nil {
		return outcomeFailed, ""
	}
	return outcomePrinted, job.ID
}

func (s *Service) dispatchItems(ctx context.Context, binding plugins.Binding, installation plugins.Installation, items []inbox.Item, defaultDeviceID string) (FlushResult, error) {
	result := FlushResult{}
	if len(items) == 0 {
		return result, nil
	}

	sendConfirmation := s.resolveSendConfirmation(ctx, binding.UserID)

	for _, item := range items {
		outcome, jobID := s.dispatchOne(ctx, binding, installation, item, defaultDeviceID, sendConfirmation)
		switch outcome {
		case outcomePrinted:
			result.Printed++
			result.PrintJobIDs = append(result.PrintJobIDs, jobID)
		case outcomeFailed:
			result.Failed++
		case outcomeSkipped:
			result.Skipped++
		}
	}

	return result, nil
}
