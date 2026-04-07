package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/session"
	"github.com/ruhuang/ink/server/internal/user"
)

func TestLoginSuccess(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	users := &fakeUserRepo{
		byEmail: map[string]*user.User{
			"name@example.com": {
				ID:           "user-1",
				Email:        "name@example.com",
				DisplayName:  "Ink User",
				Status:       user.StatusActive,
				PasswordHash: "stored-hash",
			},
		},
	}
	sessions := newFakeSessionRepo()
	service := NewService(
		users,
		sessions,
		&fakeAuditLogger{},
		fakeHasher{expectedHash: "stored-hash", expectedPassword: "demo-password"},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		30*24*time.Hour,
	)

	result, err := service.Login(context.Background(), LoginInput{
		Email:    "Name@example.com",
		Password: "demo-password",
		Meta: ClientMeta{
			ClientType: session.ClientTypeWeb,
		},
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if result.User.Email != "name@example.com" {
		t.Fatalf("unexpected user email: %s", result.User.Email)
	}
	if result.Token.AccessToken == "" || result.Token.RefreshToken == "" {
		t.Fatalf("expected token pair to be populated")
	}
	if users.updatedLastLoginAt == nil {
		t.Fatalf("expected last login to be updated")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	service := NewService(
		&fakeUserRepo{},
		newFakeSessionRepo(),
		&fakeAuditLogger{},
		fakeHasher{compareErr: errors.New("invalid password")},
		fakeAccessTokenManager{},
		fakeClock{now: time.Now().UTC()},
		fakeIDGenerator{},
		time.Hour,
	)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "missing@example.com",
		Password: "wrong",
		Meta:     ClientMeta{ClientType: session.ClientTypeWeb},
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestRefreshRotatesTokens(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	users := &fakeUserRepo{
		byID: map[string]*user.User{
			"user-1": {
				ID:          "user-1",
				Email:       "name@example.com",
				DisplayName: "Ink User",
				Status:      user.StatusActive,
			},
		},
	}
	sessions := newFakeSessionRepo()
	sessions.create(session.Session{
		ID:               "ss_old",
		FamilyID:         "sf_1",
		UserID:           "user-1",
		RefreshTokenHash: HashRefreshToken("refresh-1"),
		ClientType:       session.ClientTypeWeb,
		ExpiresAt:        now.Add(time.Hour),
		CreatedAt:        now,
		LastUsedAt:       now,
	})

	service := NewService(
		users,
		sessions,
		&fakeAuditLogger{},
		fakeHasher{},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		24*time.Hour,
	)

	result, err := service.Refresh(context.Background(), "refresh-1", ClientMeta{
		ClientType: session.ClientTypeWeb,
	})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if result.Token.RefreshToken == "refresh-1" {
		t.Fatalf("expected refresh token rotation")
	}
	current := sessions.byID["ss_old"]
	if current.RotatedAt == nil {
		t.Fatalf("expected old session to be marked rotated")
	}
	if len(sessions.byID) != 2 {
		t.Fatalf("expected new session row to be created")
	}
}

func TestRefreshReuseRevokesFamily(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	users := &fakeUserRepo{
		byID: map[string]*user.User{
			"user-1": {
				ID:          "user-1",
				Email:       "name@example.com",
				DisplayName: "Ink User",
				Status:      user.StatusActive,
			},
		},
	}
	sessions := newFakeSessionRepo()
	rotatedAt := now.Add(-time.Minute)
	sessions.create(session.Session{
		ID:               "ss_old",
		FamilyID:         "sf_1",
		UserID:           "user-1",
		RefreshTokenHash: HashRefreshToken("refresh-1"),
		ClientType:       session.ClientTypeWeb,
		ExpiresAt:        now.Add(time.Hour),
		RotatedAt:        &rotatedAt,
		CreatedAt:        now.Add(-2 * time.Minute),
		LastUsedAt:       now.Add(-time.Minute),
	})

	service := NewService(
		users,
		sessions,
		&fakeAuditLogger{},
		fakeHasher{},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		24*time.Hour,
	)

	_, err := service.Refresh(context.Background(), "refresh-1", ClientMeta{
		ClientType: session.ClientTypeWeb,
	})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("expected invalid refresh token, got %v", err)
	}
	if sessions.byID["ss_old"].RevokedAt == nil {
		t.Fatalf("expected session family to be revoked on refresh reuse")
	}
}

func TestLogoutRevokesSessionFamily(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	sessions := newFakeSessionRepo()
	sessions.create(session.Session{
		ID:               "ss_old",
		FamilyID:         "sf_1",
		UserID:           "user-1",
		RefreshTokenHash: HashRefreshToken("refresh-1"),
		ClientType:       session.ClientTypeWeb,
		ExpiresAt:        now.Add(time.Hour),
		CreatedAt:        now,
		LastUsedAt:       now,
	})

	service := NewService(
		&fakeUserRepo{},
		sessions,
		&fakeAuditLogger{},
		fakeHasher{},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		24*time.Hour,
	)

	if err := service.Logout(context.Background(), "", "refresh-1"); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if sessions.byID["ss_old"].RevokedAt == nil {
		t.Fatalf("expected session to be revoked")
	}
}

func TestChangePasswordUpdatesHashAndRevokesSessions(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	users := &fakeUserRepo{
		byID: map[string]*user.User{
			"user-1": {
				ID:           "user-1",
				Email:        "name@example.com",
				DisplayName:  "Ink User",
				Status:       user.StatusActive,
				PasswordHash: "stored-hash",
			},
		},
	}
	sessions := newFakeSessionRepo()
	sessions.create(session.Session{
		ID:               "ss_old",
		FamilyID:         "sf_1",
		UserID:           "user-1",
		RefreshTokenHash: HashRefreshToken("refresh-1"),
		ClientType:       session.ClientTypeWeb,
		ExpiresAt:        now.Add(time.Hour),
		CreatedAt:        now,
		LastUsedAt:       now,
	})

	service := NewService(
		users,
		sessions,
		&fakeAuditLogger{},
		fakeHasher{
			expectedHash:     "stored-hash",
			expectedPassword: "demo-password",
			nextHash:         "new-hash",
		},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		24*time.Hour,
	)

	err := service.ChangePassword(
		context.Background(),
		"access-token",
		"demo-password",
		"next-password",
		ClientMeta{ClientType: session.ClientTypeWeb},
	)
	if err != nil {
		t.Fatalf("change password failed: %v", err)
	}
	if users.updatedPasswordHash != "new-hash" {
		t.Fatalf("expected password hash to be updated, got %q", users.updatedPasswordHash)
	}
	if sessions.byID["ss_old"].RevokedAt == nil {
		t.Fatalf("expected all user sessions to be revoked")
	}
}

func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	users := &fakeUserRepo{
		byID: map[string]*user.User{
			"user-1": {
				ID:           "user-1",
				Email:        "name@example.com",
				DisplayName:  "Ink User",
				Status:       user.StatusActive,
				PasswordHash: "stored-hash",
			},
		},
	}
	sessions := newFakeSessionRepo()
	sessions.create(session.Session{
		ID:               "ss_old",
		FamilyID:         "sf_1",
		UserID:           "user-1",
		RefreshTokenHash: HashRefreshToken("refresh-1"),
		ClientType:       session.ClientTypeWeb,
		ExpiresAt:        now.Add(time.Hour),
		CreatedAt:        now,
		LastUsedAt:       now,
	})

	service := NewService(
		users,
		sessions,
		&fakeAuditLogger{},
		fakeHasher{compareErr: errors.New("mismatch")},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		24*time.Hour,
	)

	err := service.ChangePassword(
		context.Background(),
		"access-token",
		"wrong-password",
		"next-password",
		ClientMeta{ClientType: session.ClientTypeWeb},
	)
	if !errors.Is(err, ErrCurrentPassword) {
		t.Fatalf("expected current password mismatch error, got %v", err)
	}
}

func TestChangePasswordRejectsWhitespaceOnlyDifference(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	users := &fakeUserRepo{
		byID: map[string]*user.User{
			"user-1": {
				ID:           "user-1",
				Email:        "name@example.com",
				DisplayName:  "Ink User",
				Status:       user.StatusActive,
				PasswordHash: "stored-hash",
			},
		},
	}
	sessions := newFakeSessionRepo()
	sessions.create(session.Session{
		ID:               "ss_old",
		FamilyID:         "sf_1",
		UserID:           "user-1",
		RefreshTokenHash: HashRefreshToken("refresh-1"),
		ClientType:       session.ClientTypeWeb,
		ExpiresAt:        now.Add(time.Hour),
		CreatedAt:        now,
		LastUsedAt:       now,
	})

	service := NewService(
		users,
		sessions,
		&fakeAuditLogger{},
		fakeHasher{
			expectedHash:     "stored-hash",
			expectedPassword: "demo-password",
		},
		fakeAccessTokenManager{},
		fakeClock{now: now},
		fakeIDGenerator{},
		24*time.Hour,
	)

	err := service.ChangePassword(
		context.Background(),
		"access-token",
		"demo-password",
		" demo-password ",
		ClientMeta{ClientType: session.ClientTypeWeb},
	)
	if !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected weak password error, got %v", err)
	}
}

type fakeUserRepo struct {
	byEmail             map[string]*user.User
	byID                map[string]*user.User
	updatedLastLoginAt  *time.Time
	updatedPasswordHash string
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if account, ok := f.byEmail[email]; ok {
		return account, nil
	}
	return nil, nil
}

func (f *fakeUserRepo) FindUserByID(_ context.Context, id string) (*user.User, error) {
	if account, ok := f.byID[id]; ok {
		return account, nil
	}
	for _, account := range f.byEmail {
		if account.ID == id {
			return account, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) UpdateLastLoginAt(_ context.Context, _ string, at time.Time) error {
	f.updatedLastLoginAt = &at
	return nil
}

func (f *fakeUserRepo) UpdatePasswordHash(
	_ context.Context,
	userID string,
	passwordHash string,
	_ time.Time,
) error {
	f.updatedPasswordHash = passwordHash
	account, ok := f.FindUserByID(context.Background(), userID)
	if ok != nil {
		return ok
	}
	if account != nil {
		account.PasswordHash = passwordHash
	}
	return nil
}

type fakeSessionRepo struct {
	byID   map[string]*session.Session
	byHash map[string]*session.Session
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{
		byID:   make(map[string]*session.Session),
		byHash: make(map[string]*session.Session),
	}
}

func (f *fakeSessionRepo) create(current session.Session) {
	copy := current
	f.byID[current.ID] = &copy
	f.byHash[current.RefreshTokenHash] = &copy
}

func (f *fakeSessionRepo) Create(_ context.Context, current session.Session) error {
	f.create(current)
	return nil
}

func (f *fakeSessionRepo) FindByRefreshTokenHash(_ context.Context, hash string) (*session.Session, error) {
	if current, ok := f.byHash[hash]; ok {
		copy := *current
		return &copy, nil
	}
	return nil, nil
}

func (f *fakeSessionRepo) FindSessionByID(_ context.Context, id string) (*session.Session, error) {
	if current, ok := f.byID[id]; ok {
		copy := *current
		return &copy, nil
	}
	return nil, nil
}

func (f *fakeSessionRepo) Rotate(
	_ context.Context,
	current session.Session,
	next session.Session,
	rotatedAt time.Time,
) error {
	existing, ok := f.byID[current.ID]
	if !ok {
		return fmt.Errorf("missing session %s", current.ID)
	}
	existing.RotatedAt = &rotatedAt
	existing.LastUsedAt = rotatedAt
	f.create(next)
	return nil
}

func (f *fakeSessionRepo) RevokeFamily(_ context.Context, familyID string, revokedAt time.Time) error {
	for _, current := range f.byID {
		if current.FamilyID == familyID {
			current.RevokedAt = &revokedAt
		}
	}
	return nil
}

func (f *fakeSessionRepo) RevokeByID(_ context.Context, sessionID string, revokedAt time.Time) error {
	if current, ok := f.byID[sessionID]; ok {
		current.RevokedAt = &revokedAt
	}
	return nil
}

func (f *fakeSessionRepo) RevokeByUserID(_ context.Context, userID string, revokedAt time.Time) error {
	for _, current := range f.byID {
		if current.UserID == userID {
			current.RevokedAt = &revokedAt
		}
	}
	return nil
}

type fakeAuditLogger struct{}

func (fakeAuditLogger) Log(_ context.Context, _ AuditEvent) error {
	return nil
}

type fakeHasher struct {
	expectedHash     string
	expectedPassword string
	compareErr       error
	nextHash         string
}

func (f fakeHasher) Hash(password string) (string, error) {
	if f.nextHash != "" {
		return f.nextHash, nil
	}
	return "hashed-" + password, nil
}

func (f fakeHasher) Compare(hash string, password string) error {
	if f.compareErr != nil {
		return f.compareErr
	}
	if f.expectedHash != "" && hash != f.expectedHash {
		return errors.New("unexpected hash")
	}
	if f.expectedPassword != "" && password != f.expectedPassword {
		return errors.New("unexpected password")
	}
	return nil
}

type fakeAccessTokenManager struct{}

func (fakeAccessTokenManager) Issue(account user.User, sessionID string, _ time.Time) (string, time.Time, error) {
	return "access-" + account.ID + "-" + sessionID, time.Now().UTC().Add(15 * time.Minute), nil
}

func (fakeAccessTokenManager) Parse(token string) (*AccessClaims, error) {
	return &AccessClaims{
		UserID:    "user-1",
		SessionID: "ss_old",
	}, nil
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}

type fakeIDGenerator struct{}

func (fakeIDGenerator) New(prefix string) (string, error) {
	return prefix + "_next", nil
}
