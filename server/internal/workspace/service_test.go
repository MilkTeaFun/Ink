package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
)

func TestGetStateSeedsMissingWorkspace(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeClock{now: time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)},
	)

	state, err := service.GetState(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("get state failed: %v", err)
	}

	if len(state.Devices) != 0 || len(state.Conversations) != 0 || len(state.PrintJobs) != 0 {
		t.Fatalf("expected empty account workspace, got %+v", state)
	}
	if repo.savedUserID != "user-1" {
		t.Fatalf("expected seeded workspace to be saved for current user")
	}
}

func TestSaveStatePersistsNormalizedWorkspace(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(
		repo,
		fakeAuthenticator{},
		fakeClock{now: time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)},
	)

	saved, err := service.SaveState(context.Background(), "access-token", State{})
	if err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	if saved.Devices == nil || repo.savedState == nil || repo.savedState.ServiceBinding.ModelName != "Ink AI" {
		t.Fatalf("expected normalized workspace to be persisted")
	}
}

type fakeRepository struct {
	current     *State
	savedUserID string
	savedState  *State
}

func (f *fakeRepository) FindByUserID(_ context.Context, _ string) (*State, error) {
	return f.current, nil
}

func (f *fakeRepository) SaveByUserID(
	_ context.Context,
	userID string,
	state State,
	_ time.Time,
) error {
	copy := state
	f.savedUserID = userID
	f.savedState = &copy
	f.current = &copy
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

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}
