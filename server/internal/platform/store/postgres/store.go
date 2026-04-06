package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/session"
	"github.com/ruhuang/ink/server/internal/user"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row := s.db.QueryRow(ctx, `
		select id, email, password_hash, display_name, status, created_at, updated_at, last_login_at
		from users
		where email = $1
	`, email)

	return scanUser(row)
}

func (s *Store) FindUserByID(ctx context.Context, id string) (*user.User, error) {
	row := s.db.QueryRow(ctx, `
		select id, email, password_hash, display_name, status, created_at, updated_at, last_login_at
		from users
		where id = $1
	`, id)

	return scanUser(row)
}

func (s *Store) UpdateLastLoginAt(ctx context.Context, userID string, at time.Time) error {
	_, err := s.db.Exec(ctx, `update users set last_login_at = $2, updated_at = $2 where id = $1`, userID, at)
	return err
}

func (s *Store) UpdatePasswordHash(
	ctx context.Context,
	userID string,
	passwordHash string,
	at time.Time,
) error {
	_, err := s.db.Exec(ctx, `
		update users
		set password_hash = $2, updated_at = $3
		where id = $1
	`, userID, passwordHash, at)
	return err
}

func (s *Store) Create(ctx context.Context, current session.Session) error {
	_, err := s.db.Exec(ctx, `
		insert into auth_sessions (
			id, family_id, user_id, refresh_token_hash, client_type, user_agent, ip_address,
			expires_at, rotated_at, revoked_at, created_at, last_used_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, null, null, $9, $10)
	`,
		current.ID,
		current.FamilyID,
		current.UserID,
		current.RefreshTokenHash,
		current.ClientType,
		nullableString(current.UserAgent),
		nullableString(current.IPAddress),
		current.ExpiresAt,
		current.CreatedAt,
		current.LastUsedAt,
	)
	return err
}

func (s *Store) FindByRefreshTokenHash(ctx context.Context, hash string) (*session.Session, error) {
	row := s.db.QueryRow(ctx, `
		select id, family_id, user_id, refresh_token_hash, client_type, user_agent, ip_address,
			expires_at, rotated_at, revoked_at, created_at, last_used_at
		from auth_sessions
		where refresh_token_hash = $1
	`, hash)

	return scanSession(row)
}

func (s *Store) FindSessionByID(ctx context.Context, id string) (*session.Session, error) {
	row := s.db.QueryRow(ctx, `
		select id, family_id, user_id, refresh_token_hash, client_type, user_agent, ip_address,
			expires_at, rotated_at, revoked_at, created_at, last_used_at
		from auth_sessions
		where id = $1
	`, id)

	return scanSession(row)
}

func (s *Store) Rotate(
	ctx context.Context,
	current session.Session,
	next session.Session,
	rotatedAt time.Time,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := tx.Exec(ctx, `
		update auth_sessions
		set rotated_at = $2, last_used_at = $2
		where id = $1 and rotated_at is null and revoked_at is null
	`, current.ID, rotatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("refresh session %s was already rotated or revoked", current.ID)
	}

	if _, err := tx.Exec(ctx, `
		insert into auth_sessions (
			id, family_id, user_id, refresh_token_hash, client_type, user_agent, ip_address,
			expires_at, rotated_at, revoked_at, created_at, last_used_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, null, null, $9, $10)
	`,
		next.ID,
		next.FamilyID,
		next.UserID,
		next.RefreshTokenHash,
		next.ClientType,
		nullableString(next.UserAgent),
		nullableString(next.IPAddress),
		next.ExpiresAt,
		next.CreatedAt,
		next.LastUsedAt,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		update auth_sessions
		set revoked_at = coalesce(revoked_at, $2)
		where family_id = $1
	`, familyID, revokedAt)
	return err
}

func (s *Store) RevokeByID(ctx context.Context, sessionID string, revokedAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		update auth_sessions
		set revoked_at = coalesce(revoked_at, $2)
		where id = $1
	`, sessionID, revokedAt)
	return err
}

func (s *Store) RevokeByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		update auth_sessions
		set revoked_at = coalesce(revoked_at, $2)
		where user_id = $1
	`, userID, revokedAt)
	return err
}

func (s *Store) Log(ctx context.Context, event auth.AuditEvent) error {
	var detail any
	if event.Detail != nil {
		payload, err := json.Marshal(event.Detail)
		if err != nil {
			return err
		}
		detail = payload
	}

	_, err := s.db.Exec(ctx, `
		insert into auth_audit_logs (
			id, user_id, event_type, client_type, ip_address, user_agent, detail, created_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		fmt.Sprintf("al_%d", time.Now().UnixNano()),
		event.UserID,
		event.EventType,
		event.ClientType,
		nullableString(event.IPAddress),
		nullableString(event.UserAgent),
		detail,
		event.CreatedAt,
	)
	return err
}

func scanUser(row pgx.Row) (*user.User, error) {
	var account user.User
	var lastLoginAt *time.Time
	if err := row.Scan(
		&account.ID,
		&account.Email,
		&account.PasswordHash,
		&account.DisplayName,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
		&lastLoginAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	account.LastLoginAt = lastLoginAt
	return &account, nil
}

func scanSession(row pgx.Row) (*session.Session, error) {
	var current session.Session
	if err := row.Scan(
		&current.ID,
		&current.FamilyID,
		&current.UserID,
		&current.RefreshTokenHash,
		&current.ClientType,
		&current.UserAgent,
		&current.IPAddress,
		&current.ExpiresAt,
		&current.RotatedAt,
		&current.RevokedAt,
		&current.CreatedAt,
		&current.LastUsedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &current, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}
