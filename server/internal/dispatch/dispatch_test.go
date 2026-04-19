package dispatch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/inbox"
	"github.com/ruhuang/ink/server/internal/plugins"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/workspace"
)

// --- test doubles --------------------------------------------------------

type memInboxRepo struct {
	mu    sync.Mutex
	items map[string]inbox.Item
}

func newMemInboxRepo() *memInboxRepo {
	return &memInboxRepo{items: map[string]inbox.Item{}}
}

func (r *memInboxRepo) InsertItem(_ context.Context, item inbox.Item) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.items {
		if existing.PluginBindingID == item.PluginBindingID && existing.ExternalID == item.ExternalID {
			return false, nil
		}
	}
	r.items[item.ID] = item
	return true, nil
}

func (r *memInboxRepo) FindInboxItemByID(_ context.Context, itemID string) (*inbox.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[itemID]
	if !ok {
		return nil, nil
	}
	copy := item
	return &copy, nil
}

func (r *memInboxRepo) ListPendingByBinding(_ context.Context, bindingID string, limit int) ([]inbox.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]inbox.Item, 0)
	for _, item := range r.items {
		if item.PluginBindingID != bindingID {
			continue
		}
		if item.Status != inbox.StatusPending && item.Status != inbox.StatusFailed {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *memInboxRepo) ListRetryable(_ context.Context, olderThan time.Time, limit int) ([]inbox.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]inbox.Item, 0)
	for _, item := range r.items {
		if item.Status != inbox.StatusFailed {
			continue
		}
		if item.AttemptCount >= inbox.MaxDispatchAttempts {
			continue
		}
		if !item.UpdatedAt.Before(olderThan) {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.Before(out[j].UpdatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *memInboxRepo) UpdateStatus(_ context.Context, item inbox.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[item.ID]
	if !ok {
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

func (r *memInboxRepo) DeletePrintedOlderThan(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

type incrementingIDs struct {
	mu    sync.Mutex
	count int
}

func (g *incrementingIDs) New(prefix string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count++
	return fmt.Sprintf("%s-%d", prefix, g.count), nil
}

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

type stubPluginRuntime struct {
	binding      plugins.Binding
	installation plugins.Installation
	manifest     plugins.Manifest
}

func (s *stubPluginRuntime) GetBindingByID(_ context.Context, _ string) (plugins.Binding, map[string]string, error) {
	return s.binding, map[string]string{}, nil
}

func (s *stubPluginRuntime) GetInstallation(_ context.Context, _ string) (plugins.Installation, plugins.Manifest, error) {
	return s.installation, s.manifest, nil
}

type stubWorkspace struct{}

func (stubWorkspace) FindByUserID(_ context.Context, _ string) (*workspace.State, error) {
	state := workspace.EmptyState()
	return &state, nil
}

type stubCounter struct{ printedToday int }

func (s stubCounter) CountPrintedInLast24h(_ context.Context, _ string, _ time.Time) (int, error) {
	return s.printedToday, nil
}

type stubPrinter struct {
	mu       sync.Mutex
	created  []printer.CreateJobInput
	nextID   int
	failNext error
	failOnce bool
}

func (p *stubPrinter) CreatePrintJobForUser(_ context.Context, _ string, input printer.CreateJobInput) (workspace.PrintJob, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.failNext != nil {
		err := p.failNext
		if p.failOnce {
			p.failNext = nil
		}
		return workspace.PrintJob{}, err
	}
	p.nextID++
	p.created = append(p.created, input)
	return workspace.PrintJob{ID: fmt.Sprintf("job-%d", p.nextID)}, nil
}

// --- helpers -------------------------------------------------------------

func newTestService(t *testing.T, binding plugins.Binding, counter DailyCounter, p PrinterJobCreator, clk Clock) (*Service, *inbox.Service, *memInboxRepo) {
	t.Helper()
	repo := newMemInboxRepo()
	inboxService := inbox.NewService(repo, &incrementingIDs{}, clk)
	runtime := &stubPluginRuntime{
		binding:      binding,
		installation: plugins.Installation{ID: binding.PluginInstallationID, DisplayName: "Hello Plugin"},
	}
	return NewService(inboxService, runtime, p, stubWorkspace{}, counter, clk), inboxService, repo
}

func ingestItem(t *testing.T, svc *inbox.Service, bindingID, deviceID, externalID, title string) inbox.Item {
	t.Helper()
	items := []plugins.Item{{
		ExternalID:  externalID,
		Title:       title,
		SourceLabel: "Hello Plugin",
		Blocks:      []plugins.ContentBlock{{Type: plugins.BlockParagraph, Text: "body"}},
	}}
	result, err := svc.Ingest(context.Background(), inbox.IngestInput{
		UserID:               "user-1",
		PluginInstallationID: "install-1",
		PluginBindingID:      bindingID,
		DeviceID:             deviceID,
		SourceLabelFallback:  "Hello Plugin",
		Items:                items,
	})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if result.Inserted != 1 {
		t.Fatalf("expected 1 inserted, got %d", result.Inserted)
	}
	// Fetch back to get the full item with its generated ID.
	pending, err := svc.ListPendingByBinding(context.Background(), bindingID, 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	for _, p := range pending {
		if p.ExternalID == externalID {
			return p
		}
	}
	t.Fatalf("ingested item %s not found", externalID)
	return inbox.Item{}
}

// --- tests ---------------------------------------------------------------

func TestFlushBindingPrintsItemsAndRespectsPerRunLimit(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      2,
		MaxPrintsPerDay:      10,
	}
	printerStub := &stubPrinter{}
	svc, ibx, _ := newTestService(t, binding, stubCounter{}, printerStub, clk)

	for i := 0; i < 5; i++ {
		ingestItem(t, ibx, binding.ID, "dev-1", fmt.Sprintf("ext-%d", i), fmt.Sprintf("Item %d", i))
	}

	result, err := svc.FlushBinding(context.Background(), binding.ID, "dev-1")
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if result.Printed != 2 {
		t.Fatalf("expected 2 printed (per-run cap), got %d", result.Printed)
	}
	if len(printerStub.created) != 2 {
		t.Fatalf("expected 2 print jobs, got %d", len(printerStub.created))
	}
	if len(result.PrintJobIDs) != 2 {
		t.Fatalf("expected 2 job ids, got %d", len(result.PrintJobIDs))
	}
}

func TestFlushBindingRespectsPerDayLimit(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      10,
		MaxPrintsPerDay:      3,
	}
	printerStub := &stubPrinter{}
	svc, ibx, _ := newTestService(t, binding, stubCounter{printedToday: 2}, printerStub, clk)

	for i := 0; i < 5; i++ {
		ingestItem(t, ibx, binding.ID, "dev-1", fmt.Sprintf("ext-%d", i), fmt.Sprintf("Item %d", i))
	}

	result, err := svc.FlushBinding(context.Background(), binding.ID, "dev-1")
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if result.Printed != 1 {
		t.Fatalf("expected 1 printed (daily cap remaining), got %d", result.Printed)
	}
}

func TestFlushBindingSkipsWhenDailyCapReached(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      10,
		MaxPrintsPerDay:      3,
	}
	printerStub := &stubPrinter{}
	svc, ibx, _ := newTestService(t, binding, stubCounter{printedToday: 3}, printerStub, clk)
	ingestItem(t, ibx, binding.ID, "dev-1", "ext-0", "Item 0")

	result, err := svc.FlushBinding(context.Background(), binding.ID, "dev-1")
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if result.Printed != 0 {
		t.Fatalf("expected 0 printed when daily cap reached, got %d", result.Printed)
	}
	if len(printerStub.created) != 0 {
		t.Fatalf("printer should not have been called")
	}
}

func TestFlushBindingMarksItemFailedOnPrinterError(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      5,
		MaxPrintsPerDay:      5,
	}
	printerStub := &stubPrinter{failNext: errors.New("printer down")}
	svc, ibx, repo := newTestService(t, binding, stubCounter{}, printerStub, clk)
	item := ingestItem(t, ibx, binding.ID, "dev-1", "ext-0", "Item 0")

	result, err := svc.FlushBinding(context.Background(), binding.ID, "dev-1")
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if result.Failed != 1 || result.Printed != 0 {
		t.Fatalf("expected 1 failed, got %+v", result)
	}
	stored, _ := repo.FindInboxItemByID(context.Background(), item.ID)
	if stored == nil || stored.Status != inbox.StatusFailed || stored.AttemptCount != 1 {
		t.Fatalf("expected item in failed with attempt=1, got %+v", stored)
	}
}

func TestFlushBindingFailsItemWithoutDevice(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      5,
		MaxPrintsPerDay:      5,
	}
	printerStub := &stubPrinter{}
	svc, ibx, repo := newTestService(t, binding, stubCounter{}, printerStub, clk)
	item := ingestItem(t, ibx, binding.ID, "", "ext-0", "Item 0")

	result, err := svc.FlushBinding(context.Background(), binding.ID, "")
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if result.Failed != 1 || result.Printed != 0 {
		t.Fatalf("expected 1 failed due to missing device, got %+v", result)
	}
	stored, _ := repo.FindInboxItemByID(context.Background(), item.ID)
	if stored == nil || stored.Status != inbox.StatusFailed {
		t.Fatalf("expected item failed, got %+v", stored)
	}
	if stored.LastError == nil || *stored.LastError != "no device bound for item" {
		t.Fatalf("expected missing-device error, got %v", stored.LastError)
	}
}

func TestFlushBindingSkipsItemsPastAttemptBudget(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      5,
		MaxPrintsPerDay:      5,
	}
	printerStub := &stubPrinter{}
	svc, ibx, repo := newTestService(t, binding, stubCounter{}, printerStub, clk)
	item := ingestItem(t, ibx, binding.ID, "dev-1", "ext-0", "Item 0")

	// Bump attempts to the ceiling.
	for i := 0; i < inbox.MaxDispatchAttempts; i++ {
		if err := ibx.MarkFailed(context.Background(), item, "boom"); err != nil {
			t.Fatalf("mark failed: %v", err)
		}
		refreshed, _ := repo.FindInboxItemByID(context.Background(), item.ID)
		item = *refreshed
	}

	result, err := svc.FlushBinding(context.Background(), binding.ID, "dev-1")
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	if result.Skipped != 1 {
		t.Fatalf("expected 1 skipped due to attempt budget, got %+v", result)
	}
	if len(printerStub.created) != 0 {
		t.Fatalf("printer should not be called for exhausted items")
	}
}

func TestRetryFailedDispatchesAcrossBindings(t *testing.T) {
	past := time.Now().UTC().Add(-1 * time.Hour)
	clk := &fixedClock{now: past.Add(30 * time.Minute)}
	binding := plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: "install-1",
		UserID:               "user-1",
		MaxPrintsPerRun:      5,
		MaxPrintsPerDay:      5,
	}
	printerStub := &stubPrinter{}
	svc, ibx, repo := newTestService(t, binding, stubCounter{}, printerStub, clk)
	item := ingestItem(t, ibx, binding.ID, "dev-1", "ext-0", "Item 0")

	// Simulate a past failure older than the 15m retry cutoff.
	clk.now = past
	if err := ibx.MarkFailed(context.Background(), item, "transient"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	clk.now = past.Add(30 * time.Minute)

	result, err := svc.RetryFailed(context.Background(), 10)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if result.Printed != 1 {
		t.Fatalf("expected 1 printed on retry, got %+v", result)
	}
	stored, _ := repo.FindInboxItemByID(context.Background(), item.ID)
	if stored == nil || stored.Status != inbox.StatusPrinted {
		t.Fatalf("expected item to move to printed, got %+v", stored)
	}
}

func TestFlushBindingRejectsEmptyID(t *testing.T) {
	clk := &fixedClock{now: time.Now().UTC()}
	svc, _, _ := newTestService(t, plugins.Binding{ID: "binding-1"}, stubCounter{}, &stubPrinter{}, clk)
	if _, err := svc.FlushBinding(context.Background(), "   ", "dev-1"); err == nil {
		t.Fatalf("expected error for empty binding id")
	}
}
