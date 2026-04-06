package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/session"
	"github.com/ruhuang/ink/server/internal/user"
)

var (
	// Authentication domain errors returned by the service.
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrUserDisabled        = errors.New("user disabled")
	ErrCurrentPassword     = errors.New("current password mismatch")
	ErrWeakPassword        = errors.New("weak password")
)

// LoginInput contains the credentials and client metadata for a login attempt.
type LoginInput struct {
	Email    string
	Password string
	Meta     ClientMeta
}

// ClientMeta describes the client that initiated an auth action.
type ClientMeta struct {
	ClientType session.ClientType
	UserAgent  string
	IPAddress  string
}

// TokenPair bundles the issued access and refresh tokens.
type TokenPair struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

// AuthResult contains the signed-in user and their active session tokens.
type AuthResult struct {
	User  UserDTO
	Token TokenPair
}

// UserDTO is the public representation returned to clients.
type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AccessClaims represents the validated claims extracted from an access token.
type AccessClaims struct {
	UserID    string
	SessionID string
}

// AuthService defines the authentication use cases exposed to transports.
type AuthService interface {
	Login(ctx context.Context, input LoginInput) (AuthResult, error)
	Refresh(ctx context.Context, refreshToken string, meta ClientMeta) (AuthResult, error)
	Logout(ctx context.Context, accessToken string, refreshToken string) error
	GetCurrentUser(ctx context.Context, accessToken string) (UserDTO, error)
	ChangePassword(
		ctx context.Context,
		accessToken string,
		currentPassword string,
		newPassword string,
		meta ClientMeta,
	) error
}

// UserRepository provides user persistence required by the auth service.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindUserByID(ctx context.Context, id string) (*user.User, error)
	UpdateLastLoginAt(ctx context.Context, userID string, at time.Time) error
	UpdatePasswordHash(ctx context.Context, userID string, passwordHash string, at time.Time) error
}

// SessionRepository provides session persistence required by the auth service.
type SessionRepository interface {
	Create(ctx context.Context, current session.Session) error
	FindByRefreshTokenHash(ctx context.Context, hash string) (*session.Session, error)
	FindSessionByID(ctx context.Context, id string) (*session.Session, error)
	Rotate(ctx context.Context, current session.Session, next session.Session, rotatedAt time.Time) error
	RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error
	RevokeByID(ctx context.Context, sessionID string, revokedAt time.Time) error
	RevokeByUserID(ctx context.Context, userID string, revokedAt time.Time) error
}

// AuditLogger records authentication audit events.
type AuditLogger interface {
	Log(ctx context.Context, event AuditEvent) error
}

// PasswordHasher hashes and verifies user passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

// AccessTokenManager issues and parses access tokens.
type AccessTokenManager interface {
	Issue(user user.User, sessionID string, now time.Time) (token string, expiresAt time.Time, err error)
	Parse(token string) (*AccessClaims, error)
}

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

// IDGenerator creates stable prefixed identifiers.
type IDGenerator interface {
	New(prefix string) string
}

// AuditEvent captures a security-relevant authentication event.
type AuditEvent struct {
	UserID     *string
	EventType  string
	ClientType session.ClientType
	IPAddress  string
	UserAgent  string
	Detail     map[string]any
	CreatedAt  time.Time
}

// Service implements the auth use cases over repository interfaces.
type Service struct {
	users      UserRepository
	sessions   SessionRepository
	audit      AuditLogger
	hasher     PasswordHasher
	tokens     AccessTokenManager
	clock      Clock
	ids        IDGenerator
	refreshTTL time.Duration
}

// NewService wires the auth service dependencies.
func NewService(
	users UserRepository,
	sessions SessionRepository,
	audit AuditLogger,
	hasher PasswordHasher,
	tokens AccessTokenManager,
	clock Clock,
	ids IDGenerator,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		users:      users,
		sessions:   sessions,
		audit:      audit,
		hasher:     hasher,
		tokens:     tokens,
		clock:      clock,
		ids:        ids,
		refreshTTL: refreshTTL,
	}
}

// NormalizeEmail trims and lowercases user input before lookup.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// MapUser converts a stored user record into the public DTO.
func MapUser(entity user.User) UserDTO {
	return UserDTO{
		ID:    entity.ID,
		Email: entity.Email,
		Name:  entity.DisplayName,
	}
}

// Login verifies credentials and creates a fresh session token pair.
func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	now := s.clock.Now()
	email := NormalizeEmail(input.Email)

	account, err := s.users.FindByEmail(ctx, email)
	if err != nil || account == nil {
		s.logEvent(ctx, AuditEvent{
			EventType:  "login_failed",
			ClientType: input.Meta.ClientType,
			IPAddress:  input.Meta.IPAddress,
			UserAgent:  input.Meta.UserAgent,
			Detail:     map[string]any{"email": email},
			CreatedAt:  now,
		})
		return AuthResult{}, ErrInvalidCredentials
	}

	if account.Status == user.StatusDisabled {
		s.logEvent(ctx, AuditEvent{
			UserID:     &account.ID,
			EventType:  "login_failed",
			ClientType: input.Meta.ClientType,
			IPAddress:  input.Meta.IPAddress,
			UserAgent:  input.Meta.UserAgent,
			Detail:     map[string]any{"reason": "disabled"},
			CreatedAt:  now,
		})
		return AuthResult{}, ErrUserDisabled
	}

	if err := s.hasher.Compare(account.PasswordHash, input.Password); err != nil {
		s.logEvent(ctx, AuditEvent{
			UserID:     &account.ID,
			EventType:  "login_failed",
			ClientType: input.Meta.ClientType,
			IPAddress:  input.Meta.IPAddress,
			UserAgent:  input.Meta.UserAgent,
			Detail:     map[string]any{"email": email},
			CreatedAt:  now,
		})
		return AuthResult{}, ErrInvalidCredentials
	}

	tokenPair, familyID, err := s.newSessionTokens(ctx, *account, input.Meta, now)
	if err != nil {
		return AuthResult{}, err
	}

	if err := s.users.UpdateLastLoginAt(ctx, account.ID, now); err != nil {
		return AuthResult{}, err
	}

	s.logEvent(ctx, AuditEvent{
		UserID:     &account.ID,
		EventType:  "login_success",
		ClientType: input.Meta.ClientType,
		IPAddress:  input.Meta.IPAddress,
		UserAgent:  input.Meta.UserAgent,
		Detail:     map[string]any{"familyId": familyID},
		CreatedAt:  now,
	})

	return AuthResult{
		User:  MapUser(*account),
		Token: tokenPair,
	}, nil
}

// Refresh rotates an existing refresh token into a new session.
func (s *Service) Refresh(ctx context.Context, refreshToken string, meta ClientMeta) (AuthResult, error) {
	now := s.clock.Now()
	current, err := s.sessions.FindByRefreshTokenHash(ctx, HashRefreshToken(refreshToken))
	if err != nil || current == nil {
		return AuthResult{}, ErrInvalidRefreshToken
	}

	if current.RevokedAt != nil || current.ExpiresAt.Before(now) {
		return AuthResult{}, ErrInvalidRefreshToken
	}

	if current.RotatedAt != nil {
		_ = s.sessions.RevokeFamily(ctx, current.FamilyID, now)
		s.logEvent(ctx, AuditEvent{
			UserID:     &current.UserID,
			EventType:  "refresh_reuse_detected",
			ClientType: meta.ClientType,
			IPAddress:  meta.IPAddress,
			UserAgent:  meta.UserAgent,
			Detail:     map[string]any{"familyId": current.FamilyID},
			CreatedAt:  now,
		})
		return AuthResult{}, ErrInvalidRefreshToken
	}

	account, err := s.users.FindUserByID(ctx, current.UserID)
	if err != nil || account == nil {
		return AuthResult{}, ErrInvalidRefreshToken
	}

	if account.Status == user.StatusDisabled {
		return AuthResult{}, ErrUserDisabled
	}

	next, pair, err := s.buildRotatedSession(*account, *current, meta, now)
	if err != nil {
		return AuthResult{}, err
	}

	if err := s.sessions.Rotate(ctx, *current, next, now); err != nil {
		return AuthResult{}, err
	}

	s.logEvent(ctx, AuditEvent{
		UserID:     &account.ID,
		EventType:  "token_refreshed",
		ClientType: meta.ClientType,
		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
		Detail:     map[string]any{"familyId": current.FamilyID},
		CreatedAt:  now,
	})

	return AuthResult{
		User:  MapUser(*account),
		Token: pair,
	}, nil
}

// Logout revokes the current session family or access-token session when possible.
func (s *Service) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	now := s.clock.Now()

	if refreshToken != "" {
		current, err := s.sessions.FindByRefreshTokenHash(ctx, HashRefreshToken(refreshToken))
		if err == nil && current != nil {
			s.logEvent(ctx, AuditEvent{
				UserID:     &current.UserID,
				EventType:  "logout",
				ClientType: current.ClientType,
				IPAddress:  "",
				UserAgent:  "",
				Detail:     map[string]any{"familyId": current.FamilyID},
				CreatedAt:  now,
			})
			return s.sessions.RevokeFamily(ctx, current.FamilyID, now)
		}
	}

	if accessToken == "" {
		return nil
	}

	claims, err := s.tokens.Parse(accessToken)
	if err != nil {
		return nil
	}

	return s.sessions.RevokeByID(ctx, claims.SessionID, now)
}

// GetCurrentUser resolves the user behind a valid access token.
func (s *Service) GetCurrentUser(ctx context.Context, accessToken string) (UserDTO, error) {
	claims, err := s.tokens.Parse(accessToken)
	if err != nil {
		return UserDTO{}, ErrInvalidAccessToken
	}

	current, err := s.sessions.FindSessionByID(ctx, claims.SessionID)
	if err != nil || current == nil || current.RevokedAt != nil {
		return UserDTO{}, ErrInvalidAccessToken
	}

	account, err := s.users.FindUserByID(ctx, claims.UserID)
	if err != nil || account == nil || account.Status != user.StatusActive {
		return UserDTO{}, ErrInvalidAccessToken
	}

	return MapUser(*account), nil
}

// ChangePassword verifies the current password and revokes existing sessions.
func (s *Service) ChangePassword(
	ctx context.Context,
	accessToken string,
	currentPassword string,
	newPassword string,
	meta ClientMeta,
) error {
	now := s.clock.Now()
	claims, err := s.tokens.Parse(accessToken)
	if err != nil {
		return ErrInvalidAccessToken
	}

	currentSession, err := s.sessions.FindSessionByID(ctx, claims.SessionID)
	if err != nil || currentSession == nil || currentSession.RevokedAt != nil {
		return ErrInvalidAccessToken
	}

	account, err := s.users.FindUserByID(ctx, claims.UserID)
	if err != nil || account == nil || account.Status != user.StatusActive {
		return ErrInvalidAccessToken
	}

	if err := s.hasher.Compare(account.PasswordHash, currentPassword); err != nil {
		return ErrCurrentPassword
	}

	if !isStrongEnoughPassword(currentPassword, newPassword) {
		return ErrWeakPassword
	}

	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePasswordHash(ctx, account.ID, passwordHash, now); err != nil {
		return err
	}

	if err := s.sessions.RevokeByUserID(ctx, account.ID, now); err != nil {
		return err
	}

	s.logEvent(ctx, AuditEvent{
		UserID:     &account.ID,
		EventType:  "password_changed",
		ClientType: meta.ClientType,
		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
		Detail:     map[string]any{"sessionId": claims.SessionID},
		CreatedAt:  now,
	})

	return nil
}

func (s *Service) newSessionTokens(
	ctx context.Context,
	account user.User,
	meta ClientMeta,
	now time.Time,
) (TokenPair, string, error) {
	familyID := s.ids.New("sf")
	sessionID := s.ids.New("ss")
	refreshToken, err := NewRefreshToken()
	if err != nil {
		return TokenPair{}, "", err
	}

	refreshExpiresAt := now.Add(s.refreshTTL)
	current := session.Session{
		ID:               sessionID,
		FamilyID:         familyID,
		UserID:           account.ID,
		RefreshTokenHash: HashRefreshToken(refreshToken),
		ClientType:       meta.ClientType,
		UserAgent:        meta.UserAgent,
		IPAddress:        meta.IPAddress,
		ExpiresAt:        refreshExpiresAt,
		CreatedAt:        now,
		LastUsedAt:       now,
	}

	if err := s.sessions.Create(ctx, current); err != nil {
		return TokenPair{}, "", err
	}

	accessToken, accessExpiresAt, err := s.tokens.Issue(account, sessionID, now)
	if err != nil {
		return TokenPair{}, "", err
	}

	return TokenPair{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, familyID, nil
}

func (s *Service) buildRotatedSession(
	account user.User,
	current session.Session,
	meta ClientMeta,
	now time.Time,
) (session.Session, TokenPair, error) {
	refreshToken, err := NewRefreshToken()
	if err != nil {
		return session.Session{}, TokenPair{}, err
	}

	nextSession := session.Session{
		ID:               s.ids.New("ss"),
		FamilyID:         current.FamilyID,
		UserID:           current.UserID,
		RefreshTokenHash: HashRefreshToken(refreshToken),
		ClientType:       current.ClientType,
		UserAgent:        chooseString(meta.UserAgent, current.UserAgent),
		IPAddress:        chooseString(meta.IPAddress, current.IPAddress),
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedAt:        now,
		LastUsedAt:       now,
	}

	accessToken, accessExpiresAt, err := s.tokens.Issue(account, nextSession.ID, now)
	if err != nil {
		return session.Session{}, TokenPair{}, err
	}

	return nextSession, TokenPair{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: nextSession.ExpiresAt,
	}, nil
}

// NewRefreshToken generates a new opaque refresh token.
func NewRefreshToken() (string, error) {
	payload := make([]byte, 32)
	if _, err := rand.Read(payload); err != nil {
		return "", fmt.Errorf("read refresh token entropy: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

// HashRefreshToken returns the SHA-256 digest used for refresh-token lookup.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func chooseString(preferred string, fallback string) string {
	if strings.TrimSpace(preferred) != "" {
		return preferred
	}

	return fallback
}

func isStrongEnoughPassword(currentPassword string, nextPassword string) bool {
	trimmedCurrent := strings.TrimSpace(currentPassword)
	trimmedNext := strings.TrimSpace(nextPassword)

	return len(trimmedNext) >= 8 && trimmedNext != trimmedCurrent
}

func (s *Service) logEvent(ctx context.Context, event AuditEvent) {
	if s.audit == nil {
		return
	}

	_ = s.audit.Log(ctx, event)
}
