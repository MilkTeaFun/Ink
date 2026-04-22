// Package inbox is the transparent buffer that sits between a plugin's fetch
// output and the printer. Items produced by a plugin are persisted here
// (idempotent on (plugin_binding_id, external_id)) so that downstream
// dispatchers can flush them into print jobs at their own pace.
//
// The inbox is deliberately internal: end users never see a "plugin inbox"
// page. It only exposes enough surface area for the dispatcher and janitor.
package inbox

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/plugins"
)

// ItemStatus is the lifecycle state of an inbox item.
type ItemStatus string

const (
	// StatusPending is the initial state when an item has been ingested but
	// has not yet been invalidated or expired.
	StatusPending ItemStatus = "pending"
	// StatusPrinted is kept for legacy rows created before delivery tracking
	// moved to print_schedule_deliveries.
	StatusPrinted ItemStatus = "printed"
	// StatusInvalid indicates the item failed block validation on ingest
	// and will never be dispatched.
	StatusInvalid ItemStatus = "invalid"
	// StatusFailed indicates the dispatcher tried to create a print job
	// but the attempt errored. Items in this state may be retried until
	// MaxDispatchAttempts is reached.
	StatusFailed ItemStatus = "failed"
)

// MaxDispatchAttempts caps the number of times an item can fail before the
// dispatcher stops trying. Items beyond this count stay in StatusFailed for
// manual inspection.
const MaxDispatchAttempts = 3

// Item mirrors one row of the plugin_items table.
type Item struct {
	ID                   string
	UserID               string
	PluginInstallationID string
	PluginBindingID      string
	DeviceID             *string
	ExternalID           string
	Title                string
	SourceLabel          string
	PublishedAt          *time.Time
	Blocks               []plugins.ContentBlock
	Status               ItemStatus
	AttemptCount         int
	LastError            *string
	PrintJobID           *string
	FetchedAt            time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Repository is the storage contract the inbox service depends on.
type Repository interface {
	InsertItem(ctx context.Context, item Item) (bool, error)
	FindInboxItemByID(ctx context.Context, itemID string) (*Item, error)
	ListPendingByBinding(ctx context.Context, bindingID string, limit int) ([]Item, error)
	ListPendingBindingIDs(ctx context.Context, limit int) ([]string, error)
	ListRetryable(ctx context.Context, olderThan time.Time, limit int) ([]Item, error)
	UpdateStatus(ctx context.Context, item Item) error
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// IDGenerator mints new IDs prefixed by kind.
type IDGenerator interface {
	New(prefix string) (string, error)
}

// Clock returns the current wall time.
type Clock interface {
	Now() time.Time
}

// IngestInput is the payload the schedule / manual executor hands to the
// inbox after a plugin fetch returns.
type IngestInput struct {
	UserID               string
	PluginInstallationID string
	PluginBindingID      string
	DeviceID             string
	SourceLabelFallback  string
	Items                []plugins.Item
}

// IngestResult reports how ingestion proceeded across all items.
type IngestResult struct {
	Inserted   int
	Duplicates int
	Invalid    int
	ItemIDs    []string
}

// Service is the public API of the inbox package.
type Service struct {
	repo  Repository
	ids   IDGenerator
	clock Clock
}

// NewService constructs a Service.
func NewService(repo Repository, ids IDGenerator, clock Clock) *Service {
	return &Service{repo: repo, ids: ids, clock: clock}
}

// validateRawItem returns the status (pending or invalid) and last-error
// message for a single plugin item given its parsed fields.
func validateRawItem(externalID string, title string, blocks []plugins.ContentBlock) (ItemStatus, *string) {
	if externalID == "" {
		msg := "externalId is required"
		return StatusInvalid, &msg
	}
	if title == "" {
		msg := "title is required"
		return StatusInvalid, &msg
	}
	if err := plugins.ValidateBlocks(blocks); err != nil {
		msg := err.Error()
		return StatusInvalid, &msg
	}
	return StatusPending, nil
}

// buildItem converts a raw plugin item and the surrounding ingest context
// into an inbox.Item ready for persistence.
func buildItem(itemID string, input IngestInput, raw plugins.Item, now time.Time) Item {
	externalID := strings.TrimSpace(raw.ExternalID)
	title := strings.TrimSpace(raw.Title)
	sourceLabel := strings.TrimSpace(raw.SourceLabel)
	if sourceLabel == "" {
		sourceLabel = input.SourceLabelFallback
	}

	status, lastError := validateRawItem(externalID, title, raw.Blocks)

	// Invalid items still need a stable external id so repeated bad plugin
	// payloads dedupe instead of inserting a fresh row on every fetch.
	effectiveExternalID := externalID
	if effectiveExternalID == "" {
		effectiveExternalID = invalidExternalID(input, raw)
	}

	var deviceID *string
	if trimmed := strings.TrimSpace(input.DeviceID); trimmed != "" {
		deviceID = &trimmed
	}

	return Item{
		ID:                   itemID,
		UserID:               input.UserID,
		PluginInstallationID: input.PluginInstallationID,
		PluginBindingID:      input.PluginBindingID,
		DeviceID:             deviceID,
		ExternalID:           effectiveExternalID,
		Title:                title,
		SourceLabel:          sourceLabel,
		PublishedAt:          raw.PublishedAt,
		Blocks:               raw.Blocks,
		Status:               status,
		AttemptCount:         0,
		LastError:            lastError,
		FetchedAt:            now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func invalidExternalID(input IngestInput, raw plugins.Item) string {
	blocksJSON, err := json.Marshal(raw.Blocks)
	if err != nil {
		blocksJSON = []byte("[]")
	}
	publishedAt := ""
	if raw.PublishedAt != nil {
		publishedAt = raw.PublishedAt.UTC().Format(time.RFC3339Nano)
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(input.PluginBindingID),
		strings.TrimSpace(raw.Title),
		strings.TrimSpace(raw.SourceLabel),
		publishedAt,
		string(blocksJSON),
	}, "\n")))
	return fmt.Sprintf("invalid:%x", sum)
}

// Ingest persists a batch of plugin items. Each item is validated and either
// inserted with status=pending, skipped as a duplicate, or stored as invalid.
// The call is safe to retry: duplicates are detected via the (binding,
// external_id) unique index.
func (s *Service) Ingest(ctx context.Context, input IngestInput) (IngestResult, error) {
	if strings.TrimSpace(input.PluginBindingID) == "" {
		return IngestResult{}, errors.New("binding id is required")
	}

	now := s.clock.Now()
	result := IngestResult{}
	for _, raw := range input.Items {
		itemID, err := s.ids.New("item")
		if err != nil {
			return result, err
		}
		item := buildItem(itemID, input, raw, now)

		inserted, err := s.repo.InsertItem(ctx, item)
		if err != nil {
			return result, err
		}
		if !inserted {
			result.Duplicates++
			continue
		}
		if item.Status == StatusInvalid {
			result.Invalid++
		} else {
			result.Inserted++
		}
		result.ItemIDs = append(result.ItemIDs, item.ID)
	}

	return result, nil
}

// ListPendingByBinding returns pending items for a binding.
func (s *Service) ListPendingByBinding(ctx context.Context, bindingID string, limit int) ([]Item, error) {
	return s.repo.ListPendingByBinding(ctx, bindingID, limit)
}

// ListPendingBindingIDs returns binding ids that currently have pending items.
func (s *Service) ListPendingBindingIDs(ctx context.Context, limit int) ([]string, error) {
	return s.repo.ListPendingBindingIDs(ctx, limit)
}

// MarkPrinted moves an item to StatusPrinted, associating it with a print job.
func (s *Service) MarkPrinted(ctx context.Context, item Item, printJobID string) error {
	item.Status = StatusPrinted
	job := printJobID
	item.PrintJobID = &job
	item.AttemptCount++
	item.LastError = nil
	item.UpdatedAt = s.clock.Now()
	return s.repo.UpdateStatus(ctx, item)
}

// MarkFailed records a dispatch failure. Once AttemptCount reaches
// MaxDispatchAttempts the item remains in StatusFailed and will not be
// retried by the background runner.
func (s *Service) MarkFailed(ctx context.Context, item Item, reason string) error {
	item.Status = StatusFailed
	item.AttemptCount++
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		trimmed = "dispatch failed"
	}
	item.LastError = &trimmed
	item.UpdatedAt = s.clock.Now()
	return s.repo.UpdateStatus(ctx, item)
}

// ListRetryable returns failed items whose last attempt is older than the
// given cutoff and whose attempt count has not exceeded the retry budget.
func (s *Service) ListRetryable(ctx context.Context, olderThan time.Time, limit int) ([]Item, error) {
	return s.repo.ListRetryable(ctx, olderThan, limit)
}

// PurgeOlderThan removes collected items older than cutoff.
func (s *Service) PurgeOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return s.repo.DeleteOlderThan(ctx, cutoff)
}
