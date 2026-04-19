package inbox

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/plugins"
)

type memoryRepo struct {
	mu       sync.Mutex
	items    map[string]Item
	inserted []Item
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{items: map[string]Item{}}
}

func (r *memoryRepo) InsertItem(_ context.Context, item Item) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.items {
		if existing.PluginBindingID == item.PluginBindingID && existing.ExternalID == item.ExternalID {
			return false, nil
		}
	}
	r.items[item.ID] = item
	r.inserted = append(r.inserted, item)
	return true, nil
}

func (r *memoryRepo) FindInboxItemByID(_ context.Context, itemID string) (*Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.items[itemID]
	if !exists {
		return nil, nil
	}
	copy := item
	return &copy, nil
}

func (r *memoryRepo) ListPendingByBinding(_ context.Context, bindingID string, limit int) ([]Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Item, 0)
	for _, item := range r.items {
		if item.PluginBindingID != bindingID {
			continue
		}
		if item.Status != StatusPending && item.Status != StatusFailed {
			continue
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *memoryRepo) ListRetryable(_ context.Context, olderThan time.Time, limit int) ([]Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Item, 0)
	for _, item := range r.items {
		if item.Status != StatusFailed {
			continue
		}
		if item.AttemptCount >= MaxDispatchAttempts {
			continue
		}
		if !item.UpdatedAt.Before(olderThan) {
			continue
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UpdatedAt.Before(result[j].UpdatedAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *memoryRepo) UpdateStatus(_ context.Context, item Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.items[item.ID]
	if !exists {
		return errors.New("not found")
	}
	existing.Status = item.Status
	existing.AttemptCount = item.AttemptCount
	existing.LastError = item.LastError
	existing.PrintJobID = item.PrintJobID
	existing.UpdatedAt = item.UpdatedAt
	r.items[item.ID] = existing
	return nil
}

func (r *memoryRepo) DeletePrintedOlderThan(_ context.Context, cutoff time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	removed := int64(0)
	for id, item := range r.items {
		if item.Status == StatusPrinted && item.UpdatedAt.Before(cutoff) {
			delete(r.items, id)
			removed++
		}
	}
	return removed, nil
}

type incrementingIDs struct {
	mu      sync.Mutex
	counter int
}

func (g *incrementingIDs) New(prefix string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counter++
	return prefix + "-" + string(rune('a'+g.counter-1)), nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

func validItem(external string, title string) plugins.Item {
	return plugins.Item{
		ExternalID: external,
		Title:      title,
		Blocks: []plugins.ContentBlock{
			{Type: plugins.BlockHeading, Level: 1, Text: title},
			{Type: plugins.BlockParagraph, Text: "body"},
		},
	}
}

func buildIngestInput() IngestInput {
	return IngestInput{
		UserID:               "user-1",
		PluginInstallationID: "install-1",
		PluginBindingID:      "binding-1",
		DeviceID:             "device-1",
		SourceLabelFallback:  "Fallback Source",
		Items: []plugins.Item{
			validItem("ext-1", "First"),
			validItem("ext-1", "First Duplicate"),          // dup of ext-1
			validItem("ext-2", "Second"),                   // new
			{ExternalID: "", Title: "Missing external id"}, // invalid
			{ExternalID: "ext-3", Title: ""},               // invalid title
			{ExternalID: "ext-4", Title: "Bad block", Blocks: []plugins.ContentBlock{{Type: plugins.BlockHeading}}},
		},
	}
}

func assertIngestCounts(t *testing.T, result IngestResult) {
	t.Helper()
	if result.Inserted != 2 {
		t.Fatalf("expected 2 inserted, got %d", result.Inserted)
	}
	if result.Duplicates != 1 {
		t.Fatalf("expected 1 duplicate, got %d", result.Duplicates)
	}
	if result.Invalid != 3 {
		t.Fatalf("expected 3 invalid, got %d", result.Invalid)
	}
	if len(result.ItemIDs) != 5 {
		t.Fatalf("expected 5 item ids in result (inserted + invalid), got %d", len(result.ItemIDs))
	}
}

func assertFallbackApplied(t *testing.T, items []Item) {
	t.Helper()
	for _, item := range items {
		if item.SourceLabel != "Fallback Source" {
			t.Fatalf("expected fallback source label to be applied, got %q", item.SourceLabel)
		}
		if item.DeviceID == nil || *item.DeviceID != "device-1" {
			t.Fatalf("expected device id to be stored, got %+v", item.DeviceID)
		}
	}
}

func TestIngestInsertsDedupesAndValidates(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	service := NewService(repo, &incrementingIDs{}, fixedClock{now: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)})
	input := buildIngestInput()

	result, err := service.Ingest(context.Background(), input)
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}

	assertIngestCounts(t, result)
	assertFallbackApplied(t, repo.inserted)

	// A second ingest with the same items should be a no-op.
	second, err := service.Ingest(context.Background(), input)
	if err != nil {
		t.Fatalf("second ingest: %v", err)
	}
	if second.Inserted != 0 {
		t.Fatalf("expected idempotent re-ingest to insert 0, got %d", second.Inserted)
	}
}

func TestIngestRequiresBinding(t *testing.T) {
	t.Parallel()

	service := NewService(newMemoryRepo(), &incrementingIDs{}, fixedClock{})
	_, err := service.Ingest(context.Background(), IngestInput{
		Items: []plugins.Item{validItem("ext-1", "First")},
	})
	if err == nil {
		t.Fatalf("expected error when binding id is missing")
	}
}

func ingestOne(t *testing.T, service *Service, bindingID string, externalID string, title string) Item {
	t.Helper()
	if _, err := service.Ingest(context.Background(), IngestInput{
		PluginBindingID: bindingID,
		Items:           []plugins.Item{validItem(externalID, title)},
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	pending, err := service.ListPendingByBinding(context.Background(), bindingID, 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending item, got %d", len(pending))
	}
	return pending[0]
}

func TestMarkPrintedRemovesItemFromPending(t *testing.T) {
	t.Parallel()

	service := NewService(newMemoryRepo(), &incrementingIDs{}, fixedClock{now: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)})
	item := ingestOne(t, service, "binding-1", "ext-1", "Good")

	if err := service.MarkPrinted(context.Background(), item, "print-job-1"); err != nil {
		t.Fatalf("mark printed: %v", err)
	}
	afterPrinted, err := service.ListPendingByBinding(context.Background(), "binding-1", 10)
	if err != nil {
		t.Fatalf("list pending after print: %v", err)
	}
	if len(afterPrinted) != 0 {
		t.Fatalf("expected 0 pending items after mark-printed, got %d", len(afterPrinted))
	}
}

func TestMarkFailedMovesItemToRetryable(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	service := NewService(newMemoryRepo(), &incrementingIDs{}, fixedClock{now: now})
	item := ingestOne(t, service, "binding-1", "ext-2", "Retry me")

	if err := service.MarkFailed(context.Background(), item, "network down"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	retryable, err := service.ListRetryable(context.Background(), now.Add(1*time.Hour), 10)
	if err != nil {
		t.Fatalf("list retryable: %v", err)
	}
	if len(retryable) != 1 {
		t.Fatalf("expected 1 retryable item, got %d", len(retryable))
	}
	r := retryable[0]
	if r.AttemptCount != 1 || r.LastError == nil || *r.LastError != "network down" {
		t.Fatalf("unexpected retryable state: %+v", r)
	}
}

func TestPurgePrintedOlderThan(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	now := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	service := NewService(repo, &incrementingIDs{}, fixedClock{now: now})

	if _, err := service.Ingest(context.Background(), IngestInput{
		PluginBindingID: "binding-1",
		Items:           []plugins.Item{validItem("ext-1", "Old")},
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	pending, _ := service.ListPendingByBinding(context.Background(), "binding-1", 10)
	if err := service.MarkPrinted(context.Background(), pending[0], "job-1"); err != nil {
		t.Fatalf("mark printed: %v", err)
	}

	// Now advance clock and purge with cutoff in the future.
	service.clock = fixedClock{now: now.Add(31 * 24 * time.Hour)}
	removed, err := service.PurgePrinted(context.Background(), now.Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 purged, got %d", removed)
	}
}
