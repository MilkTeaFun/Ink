package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runner applies SQL migrations in filename order.
type Runner struct {
	db *pgxpool.Pool
}

// NewRunner constructs a migration runner for the provided database pool.
func NewRunner(db *pgxpool.Pool) *Runner {
	return &Runner{db: db}
}

// Up applies every migration file that has not been recorded yet.
func (r *Runner) Up(ctx context.Context, dir string) ([]string, error) {
	if err := r.ensureSchemaMigrationsTable(ctx); err != nil {
		return nil, err
	}

	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	files, err := migrationFiles(dir)
	if err != nil {
		return nil, err
	}

	appliedNow := make([]string, 0)

	for _, file := range files {
		version := filepath.Base(file)
		if applied[version] {
			continue
		}

		if err := r.applyFile(ctx, file, version); err != nil {
			return appliedNow, err
		}

		appliedNow = append(appliedNow, version)
	}

	return appliedNow, nil
}

func (r *Runner) ensureSchemaMigrationsTable(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		create table if not exists schema_migrations (
			version text primary key,
			applied_at timestamptz not null
		)
	`)
	return err
}

func (r *Runner) appliedVersions(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.Query(ctx, `select version from schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = true
	}

	return versions, rows.Err()
}

func (r *Runner) applyFile(ctx context.Context, path string, version string) error {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", version, err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("apply migration %s: %w", version, err)
	}

	if _, err := tx.Exec(ctx, `
		insert into schema_migrations (version, applied_at)
		values ($1, $2)
	`, version, time.Now().UTC()); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}

	return tx.Commit(ctx)
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %s: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		files = append(files, filepath.Join(dir, entry.Name()))
	}

	sort.Strings(files)
	return files, nil
}
