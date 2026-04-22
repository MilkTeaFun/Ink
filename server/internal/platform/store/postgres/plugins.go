package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ruhuang/ink/server/internal/plugins"
)

var _ plugins.Repository = (*Store)(nil)

const pluginBindingColumns = `id, plugin_installation_id, user_id, enabled, config_json, secret_ciphertext,
			secret_nonce, cursor_json, max_prints_per_run, max_prints_per_day,
			status, last_validated_at, last_error, next_fetch_at, last_fetch_at,
			fetch_lease_until, last_fetch_error, created_at, updated_at`

const pluginBindingColumnsQualified = `bindings.id, bindings.plugin_installation_id, bindings.user_id, bindings.enabled, bindings.config_json, bindings.secret_ciphertext,
			bindings.secret_nonce, bindings.cursor_json, bindings.max_prints_per_run, bindings.max_prints_per_day,
			bindings.status, bindings.last_validated_at, bindings.last_error, bindings.next_fetch_at, bindings.last_fetch_at,
			bindings.fetch_lease_until, bindings.last_fetch_error, bindings.created_at, bindings.updated_at`

const pluginInstallationColumns = `id, plugin_key, source_type, display_name, version, runtime_type, manifest_json,
			current_path, status, last_error, installed_by, repo_url, repo_ref, repo_commit_sha,
			repo_subdir, created_at, updated_at`

func (s *Store) ListInstallations(ctx context.Context) ([]plugins.Installation, error) {
	rows, err := s.db.Query(ctx, `
		select `+pluginInstallationColumns+`
		from plugin_installations
		order by updated_at desc, created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []plugins.Installation{}
	for rows.Next() {
		current, err := scanPluginInstallation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *current)
	}

	return result, rows.Err()
}

func (s *Store) FindInstallationByID(ctx context.Context, installationID string) (*plugins.Installation, error) {
	row := s.db.QueryRow(ctx, `
		select `+pluginInstallationColumns+`
		from plugin_installations
		where id = $1
	`, installationID)
	return scanPluginInstallation(row)
}

func (s *Store) FindInstallationByPluginKey(ctx context.Context, pluginKey string) (*plugins.Installation, error) {
	row := s.db.QueryRow(ctx, `
		select `+pluginInstallationColumns+`
		from plugin_installations
		where plugin_key = $1
	`, pluginKey)
	return scanPluginInstallation(row)
}

func (s *Store) SaveInstallation(ctx context.Context, installation plugins.Installation) error {
	_, err := s.db.Exec(ctx, `
		insert into plugin_installations (
			id, plugin_key, source_type, display_name, version, runtime_type, manifest_json,
			current_path, status, last_error, installed_by, repo_url, repo_ref, repo_commit_sha,
			repo_subdir, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		on conflict (id)
		do update set
			plugin_key = excluded.plugin_key,
			source_type = excluded.source_type,
			display_name = excluded.display_name,
			version = excluded.version,
			runtime_type = excluded.runtime_type,
			manifest_json = excluded.manifest_json,
			current_path = excluded.current_path,
			status = excluded.status,
			last_error = excluded.last_error,
			installed_by = excluded.installed_by,
			repo_url = excluded.repo_url,
			repo_ref = excluded.repo_ref,
			repo_commit_sha = excluded.repo_commit_sha,
			repo_subdir = excluded.repo_subdir,
			updated_at = excluded.updated_at
	`,
		installation.ID,
		installation.PluginKey,
		installation.SourceType,
		installation.DisplayName,
		installation.Version,
		installation.RuntimeType,
		installation.ManifestJSON,
		installation.CurrentPath,
		installation.Status,
		installation.LastError,
		installation.InstalledBy,
		installation.RepoURL,
		installation.RepoRef,
		installation.RepoCommitSHA,
		installation.RepoSubdir,
		installation.CreatedAt,
		installation.UpdatedAt,
	)
	return err
}

func (s *Store) ListPluginBindingsByUserID(ctx context.Context, userID string) ([]plugins.Binding, error) {
	rows, err := s.db.Query(ctx, `
		select `+pluginBindingColumns+`
		from plugin_bindings
		where user_id = $1
		order by updated_at desc, created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []plugins.Binding{}
	for rows.Next() {
		current, err := scanPluginBinding(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *current)
	}

	return result, rows.Err()
}

func (s *Store) FindPluginBindingByInstallationAndUserID(ctx context.Context, installationID string, userID string) (*plugins.Binding, error) {
	row := s.db.QueryRow(ctx, `
		select `+pluginBindingColumns+`
		from plugin_bindings
		where plugin_installation_id = $1 and user_id = $2
	`, installationID, userID)
	return scanPluginBinding(row)
}

func (s *Store) FindPluginBindingByID(ctx context.Context, bindingID string) (*plugins.Binding, error) {
	row := s.db.QueryRow(ctx, `
		select `+pluginBindingColumns+`
		from plugin_bindings
		where id = $1
	`, bindingID)
	return scanPluginBinding(row)
}

func (s *Store) SavePluginBinding(ctx context.Context, binding plugins.Binding) error {
	configJSON, err := json.Marshal(binding.Config)
	if err != nil {
		return err
	}
	cursorJSON, err := json.Marshal(binding.Cursor)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		insert into plugin_bindings (
			id, plugin_installation_id, user_id, enabled, config_json, secret_ciphertext,
			secret_nonce, cursor_json, max_prints_per_run, max_prints_per_day,
			status, last_validated_at, last_error, next_fetch_at, last_fetch_at,
			fetch_lease_until, last_fetch_error, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		on conflict (id)
		do update set
			plugin_installation_id = excluded.plugin_installation_id,
			user_id = excluded.user_id,
			enabled = excluded.enabled,
			config_json = excluded.config_json,
			secret_ciphertext = excluded.secret_ciphertext,
			secret_nonce = excluded.secret_nonce,
			cursor_json = excluded.cursor_json,
			max_prints_per_run = excluded.max_prints_per_run,
			max_prints_per_day = excluded.max_prints_per_day,
			status = excluded.status,
			last_validated_at = excluded.last_validated_at,
			last_error = excluded.last_error,
			next_fetch_at = excluded.next_fetch_at,
			last_fetch_at = excluded.last_fetch_at,
			fetch_lease_until = excluded.fetch_lease_until,
			last_fetch_error = excluded.last_fetch_error,
			updated_at = excluded.updated_at
	`,
		binding.ID,
		binding.PluginInstallationID,
		binding.UserID,
		binding.Enabled,
		configJSON,
		nullableBytes(binding.Ciphertext),
		nullableBytes(binding.Nonce),
		cursorJSON,
		binding.MaxPrintsPerRun,
		binding.MaxPrintsPerDay,
		binding.Status,
		binding.LastValidatedAt,
		binding.LastError,
		binding.NextFetchAt,
		binding.LastFetchAt,
		binding.FetchLeaseUntil,
		binding.LastFetchError,
		binding.CreatedAt,
		binding.UpdatedAt,
	)
	return err
}

func (s *Store) UpdatePluginBindingCursor(ctx context.Context, bindingID string, cursor *string, updatedAt time.Time) error {
	cursorJSON, err := json.Marshal(cursor)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		update plugin_bindings
		set cursor_json = $2, updated_at = $3
		where id = $1
	`, bindingID, cursorJSON, updatedAt)
	return err
}

func (s *Store) ClaimBindingsDueForFetch(ctx context.Context, now time.Time, leaseUntil time.Time, limit int) ([]plugins.Binding, error) {
	rows, err := s.db.Query(ctx, `
		with due as (
			select id
			from plugin_bindings
			where enabled = true
			  and status = $1
			  and next_fetch_at is not null
			  and next_fetch_at <= $2
			  and (fetch_lease_until is null or fetch_lease_until < $2)
			order by next_fetch_at asc
			limit $3
			for update skip locked
		)
		update plugin_bindings as bindings
		set fetch_lease_until = $4, updated_at = $2
		from due
		where bindings.id = due.id
		returning `+pluginBindingColumnsQualified+`
	`, string(plugins.BindingStatusConnected), now, limit, leaseUntil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []plugins.Binding{}
	for rows.Next() {
		current, err := scanPluginBinding(rows)
		if err != nil {
			return nil, err
		}
		if current != nil {
			result = append(result, *current)
		}
	}

	return result, rows.Err()
}

func scanPluginInstallation(row pgx.Row) (*plugins.Installation, error) {
	var current plugins.Installation
	var lastError *string
	var installedBy *string
	if err := row.Scan(
		&current.ID,
		&current.PluginKey,
		&current.SourceType,
		&current.DisplayName,
		&current.Version,
		&current.RuntimeType,
		&current.ManifestJSON,
		&current.CurrentPath,
		&current.Status,
		&lastError,
		&installedBy,
		&current.RepoURL,
		&current.RepoRef,
		&current.RepoCommitSHA,
		&current.RepoSubdir,
		&current.CreatedAt,
		&current.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	current.LastError = lastError
	current.InstalledBy = installedBy
	return &current, nil
}

func scanPluginBinding(row pgx.Row) (*plugins.Binding, error) {
	var current plugins.Binding
	var configJSON []byte
	var cursorJSON []byte
	var lastValidatedAt *time.Time
	var lastError *string
	var nextFetchAt *time.Time
	var lastFetchAt *time.Time
	var fetchLeaseUntil *time.Time
	var lastFetchError *string
	var ciphertext []byte
	var nonce []byte
	if err := row.Scan(
		&current.ID,
		&current.PluginInstallationID,
		&current.UserID,
		&current.Enabled,
		&configJSON,
		&ciphertext,
		&nonce,
		&cursorJSON,
		&current.MaxPrintsPerRun,
		&current.MaxPrintsPerDay,
		&current.Status,
		&lastValidatedAt,
		&lastError,
		&nextFetchAt,
		&lastFetchAt,
		&fetchLeaseUntil,
		&lastFetchError,
		&current.CreatedAt,
		&current.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &current.Config); err != nil {
			return nil, err
		}
	}
	if current.Config == nil {
		current.Config = map[string]any{}
	}
	if len(cursorJSON) > 0 {
		if err := json.Unmarshal(cursorJSON, &current.Cursor); err != nil {
			return nil, err
		}
	}
	current.Ciphertext = ciphertext
	current.Nonce = nonce
	current.LastValidatedAt = lastValidatedAt
	current.LastError = lastError
	current.NextFetchAt = nextFetchAt
	current.LastFetchAt = lastFetchAt
	current.FetchLeaseUntil = fetchLeaseUntil
	current.LastFetchError = lastFetchError
	return &current, nil
}
