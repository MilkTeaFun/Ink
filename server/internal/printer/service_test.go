package printer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/workspace"
)

func TestCancelPrintJobRejectsQueuedJobs(t *testing.T) {
	repo := newFakePrinterRepo()
	repo.jobs["job-1"] = Job{
		ID:               "job-1",
		UserID:           "user-1",
		PrinterBindingID: "device-1",
		Status:           workspace.PrintStatusQueued,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeIDGenerator{},
		fakeClock{now: time.Now().UTC()},
		"",
		"",
		time.Second,
	)

	_, err := service.CancelPrintJob(context.Background(), "access-token", "job-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected queued job cancellation to be rejected, got %v", err)
	}
}

func TestUpdatePrintJobDeviceRejectsQueuedJobs(t *testing.T) {
	repo := newFakePrinterRepo()
	repo.jobs["job-1"] = Job{
		ID:               "job-1",
		UserID:           "user-1",
		PrinterBindingID: "device-1",
		Status:           workspace.PrintStatusQueued,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.bindings["device-2"] = Binding{
		ID:               "device-2",
		UserID:           "user-1",
		DeviceIdentifier: "m1-2",
		Status:           workspace.DeviceStatusConnected,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeIDGenerator{},
		fakeClock{now: time.Now().UTC()},
		"",
		"",
		time.Second,
	)

	_, err := service.UpdatePrintJobDevice(context.Background(), "access-token", "job-1", UpdateJobDeviceInput{
		PrinterBindingID: "device-2",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected queued job rebind to be rejected, got %v", err)
	}
}

func TestDeleteDeviceMarksBindingOfflineInsteadOfRemovingHistory(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	repo := newFakePrinterRepo()
	repo.bindings["device-1"] = Binding{
		ID:               "device-1",
		UserID:           "user-1",
		Name:             "书桌咕咕机",
		DeviceIdentifier: "m1-1",
		Status:           workspace.DeviceStatusConnected,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Hour),
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeIDGenerator{},
		fakeClock{now: now},
		"",
		"",
		time.Second,
	)

	if err := service.DeleteDevice(context.Background(), "access-token", "device-1"); err != nil {
		t.Fatalf("delete device failed: %v", err)
	}

	binding := repo.bindings["device-1"]
	if binding.Status != workspace.DeviceStatusOffline {
		t.Fatalf("expected device to be marked offline, got %s", binding.Status)
	}
	if !binding.UpdatedAt.Equal(now) {
		t.Fatalf("expected updated_at to be refreshed")
	}
}

func TestCreatePrintJobRejectsOfflineBindings(t *testing.T) {
	repo := newFakePrinterRepo()
	repo.bindings["device-1"] = Binding{
		ID:               "device-1",
		UserID:           "user-1",
		DeviceIdentifier: "m1-1",
		Status:           workspace.DeviceStatusOffline,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeIDGenerator{},
		fakeClock{now: time.Now().UTC()},
		"",
		"",
		time.Second,
	)

	_, err := service.CreatePrintJob(context.Background(), "access-token", CreateJobInput{
		Title:            "测试",
		Source:           "手动打印",
		Content:          "内容",
		PrinterBindingID: "device-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected offline binding to be rejected, got %v", err)
	}
}

func TestCreatePrintJobSubmitImmediatelyRejectsUnconfiguredService(t *testing.T) {
	repo := newFakePrinterRepo()
	repo.bindings["device-1"] = Binding{
		ID:               "device-1",
		UserID:           "user-1",
		DeviceIdentifier: "m1-1",
		Status:           workspace.DeviceStatusConnected,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeIDGenerator{},
		fakeClock{now: time.Now().UTC()},
		"",
		"",
		time.Second,
	)

	_, err := service.CreatePrintJob(context.Background(), "access-token", CreateJobInput{
		Title:             "测试",
		Source:            "手动打印",
		Content:           "内容",
		PrinterBindingID:  "device-1",
		SubmitImmediately: true,
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected unconfigured immediate submit to be rejected, got %v", err)
	}
	if len(repo.jobs) != 0 {
		t.Fatalf("expected no print job to be persisted when service is unconfigured")
	}
}

func TestSubmitPrintJobRejectsNonPendingJobs(t *testing.T) {
	now := time.Now().UTC()
	repo := newFakePrinterRepo()
	repo.jobs["job-1"] = Job{
		ID:               "job-1",
		UserID:           "user-1",
		PrinterBindingID: "device-1",
		Status:           workspace.PrintStatusQueued,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	repo.bindings["device-1"] = Binding{
		ID:               "device-1",
		UserID:           "user-1",
		DeviceIdentifier: "m1-1",
		Status:           workspace.DeviceStatusConnected,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeIDGenerator{},
		fakeClock{now: now},
		"access-key",
		"",
		time.Second,
	)

	_, err := service.SubmitPrintJob(context.Background(), "access-token", "job-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected non-pending job submission to be rejected, got %v", err)
	}
	if repo.jobs["job-1"].Status != workspace.PrintStatusQueued {
		t.Fatalf("expected rejected submission to leave job status unchanged, got %s", repo.jobs["job-1"].Status)
	}
}

type fakePrinterRepo struct {
	bindings map[string]Binding
	jobs     map[string]Job
}

func newFakePrinterRepo() *fakePrinterRepo {
	return &fakePrinterRepo{
		bindings: map[string]Binding{},
		jobs:     map[string]Job{},
	}
}

func (f *fakePrinterRepo) ListBindingsByUserID(_ context.Context, userID string) ([]Binding, error) {
	bindings := make([]Binding, 0, len(f.bindings))
	for _, binding := range f.bindings {
		if binding.UserID == userID {
			bindings = append(bindings, binding)
		}
	}
	return bindings, nil
}

func (f *fakePrinterRepo) FindBindingByID(_ context.Context, userID string, bindingID string) (*Binding, error) {
	binding, ok := f.bindings[bindingID]
	if !ok || binding.UserID != userID {
		return nil, nil
	}

	copy := binding
	return &copy, nil
}

func (f *fakePrinterRepo) SaveBinding(_ context.Context, binding Binding) error {
	f.bindings[binding.ID] = binding
	return nil
}

func (f *fakePrinterRepo) DeleteBinding(_ context.Context, userID string, bindingID string) error {
	binding, ok := f.bindings[bindingID]
	if ok && binding.UserID == userID {
		delete(f.bindings, bindingID)
	}
	return nil
}

func (f *fakePrinterRepo) ListJobsByUserID(_ context.Context, userID string) ([]Job, error) {
	jobs := make([]Job, 0, len(f.jobs))
	for _, job := range f.jobs {
		if job.UserID == userID {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (f *fakePrinterRepo) FindJobByID(_ context.Context, userID string, jobID string) (*Job, error) {
	job, ok := f.jobs[jobID]
	if !ok || job.UserID != userID {
		return nil, nil
	}

	copy := job
	return &copy, nil
}

func (f *fakePrinterRepo) SaveJob(_ context.Context, job Job) error {
	f.jobs[job.ID] = job
	return nil
}

type fakeAuthenticator struct{}

func (fakeAuthenticator) GetCurrentUser(_ context.Context, _ string) (auth.UserDTO, error) {
	return auth.UserDTO{
		ID:    "user-1",
		Email: "name@example.com",
		Name:  "Ink User",
		Role:  "member",
	}, nil
}

type fakeIDGenerator struct{}

func (fakeIDGenerator) New(prefix string) (string, error) {
	return prefix + "-generated", nil
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}
