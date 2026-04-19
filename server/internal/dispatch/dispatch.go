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

// DailyCounter returns how many items a binding has already printed within
// the rolling 24h window. Implementations are typically backed by the inbox
// repository using a count query.
type DailyCounter interface {
	CountPrintedInLast24h(ctx context.Context, bindingID string, since time.Time) (int, error)
}

// Service dispatches items out of the inbox into print jobs.
type Service struct {
	inbox     *inbox.Service
	plugins   PluginRuntime
	printer   PrinterJobCreator
	workspace WorkspaceRepository
	counter   DailyCounter
	clock     Clock
}

// NewService builds a dispatcher.
func NewService(
	inboxService *inbox.Service,
	pluginRuntime PluginRuntime,
	printerCreator PrinterJobCreator,
	workspaceRepo WorkspaceRepository,
	counter DailyCounter,
	clock Clock,
) *Service {
	return &Service{
		inbox:     inboxService,
		plugins:   pluginRuntime,
		printer:   printerCreator,
		workspace: workspaceRepo,
		counter:   counter,
		clock:     clock,
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
//
// The defaultDeviceID is used as a fallback for items that don't record their
// own device (for example, items ingested before a schedule had a device set).
func (s *Service) FlushBinding(ctx context.Context, bindingID string, defaultDeviceID string) (FlushResult, error) {
	result := FlushResult{}
	if strings.TrimSpace(bindingID) == "" {
		return result, errors.New("binding id is required")
	}

	binding, _, err := s.plugins.GetBindingByID(ctx, bindingID)
	if err != nil {
		return result, err
	}
	installation, _, err := s.plugins.GetInstallation(ctx, binding.PluginInstallationID)
	if err != nil {
		return result, err
	}

	perRun := binding.MaxPrintsPerRun
	if perRun <= 0 {
		perRun = DefaultBatchSize
	}

	perDay := binding.MaxPrintsPerDay
	if perDay <= 0 {
		perDay = DefaultDailyCap
	}

	printedToday, err := s.counter.CountPrintedInLast24h(ctx, bindingID, s.clock.Now().Add(-24*time.Hour))
	if err != nil {
		return result, err
	}

	remainingDay := perDay - printedToday
	if remainingDay <= 0 {
		return result, nil
	}
	budget := perRun
	if remainingDay < budget {
		budget = remainingDay
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
	cutoff := s.clock.Now().Add(-15 * time.Minute)
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
		binding, _, bindingErr := s.plugins.GetBindingByID(ctx, bindingID)
		if bindingErr != nil {
			continue
		}
		installation, _, installationErr := s.plugins.GetInstallation(ctx, binding.PluginInstallationID)
		if installationErr != nil {
			continue
		}
		result, flushErr := s.dispatchItems(ctx, binding, installation, batch, "")
		if flushErr != nil {
			continue
		}
		aggregate.Printed += result.Printed
		aggregate.Failed += result.Failed
		aggregate.Skipped += result.Skipped
		aggregate.PrintJobIDs = append(aggregate.PrintJobIDs, result.PrintJobIDs...)
	}
	return aggregate, nil
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

// pickDeviceID chooses the device id for a given item, preferring the item's
// own binding if set and falling back to the caller-provided default.
func pickDeviceID(item inbox.Item, defaultDeviceID string) string {
	if item.DeviceID != nil && strings.TrimSpace(*item.DeviceID) != "" {
		return *item.DeviceID
	}
	return defaultDeviceID
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
