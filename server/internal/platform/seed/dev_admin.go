package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruhuang/ink/server/internal/platform/password"
)

const (
	devAdminID          = "user_admin"
	devAdminLogin       = "admin"
	devAdminDisplayName = "Administrator"
)

// DevAdminOptions configures how the development admin account is bootstrapped.
type DevAdminOptions struct {
	CredentialsPath string
}

// DevAdminResult describes whether the development admin user was created.
type DevAdminResult struct {
	Created         bool
	Login           string
	Password        string
	CredentialsPath string
}

// EnsureDevAdmin creates the local development admin account when it is missing.
func EnsureDevAdmin(ctx context.Context, db *pgxpool.Pool, options DevAdminOptions) (DevAdminResult, error) {
	result := DevAdminResult{
		Login:           devAdminLogin,
		CredentialsPath: options.CredentialsPath,
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DevAdminResult{}, fmt.Errorf("begin dev admin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	exists, err := devAdminExists(ctx, tx)
	if err != nil {
		return DevAdminResult{}, err
	}

	if exists {
		return result, nil
	}

	initialPassword, err := generatePassword(18)
	if err != nil {
		return DevAdminResult{}, fmt.Errorf("generate initial password: %w", err)
	}

	passwordHash, err := password.BcryptHasher{}.Hash(initialPassword)
	if err != nil {
		return DevAdminResult{}, fmt.Errorf("hash initial password: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`insert into users (id, email, password_hash, display_name, role, status, created_at, updated_at)
		 values ($1, $2, $3, $4, 'admin', 'active', $5, $5)`,
		devAdminID,
		devAdminLogin,
		passwordHash,
		devAdminDisplayName,
		time.Now().UTC(),
	); err != nil {
		return DevAdminResult{}, fmt.Errorf("insert dev admin: %w", err)
	}

	if err := writeCredentialsFile(options.CredentialsPath, devAdminLogin, initialPassword); err != nil {
		return DevAdminResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return DevAdminResult{}, fmt.Errorf("commit dev admin transaction: %w", err)
	}

	result.Created = true
	result.Password = initialPassword
	return result, nil
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func devAdminExists(ctx context.Context, db queryRower) (bool, error) {
	var exists bool
	if err := db.QueryRow(
		ctx,
		`select exists(select 1 from users where id = $1 or email = $2)`,
		devAdminID,
		devAdminLogin,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check dev admin: %w", err)
	}

	return exists, nil
}

func generatePassword(byteLength int) (string, error) {
	payload := make([]byte, byteLength)
	if _, err := rand.Read(payload); err != nil {
		return "", err
	}

	password := base64.RawURLEncoding.EncodeToString(payload)
	return strings.TrimRight(password, "="), nil
}

func writeCredentialsFile(path string, login string, password string) error {
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("prepare credentials directory: %w", err)
	}

	content := fmt.Sprintf("login=%s\npassword=%s\n", login, password)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write credentials file %s: %w", path, err)
	}

	return nil
}
