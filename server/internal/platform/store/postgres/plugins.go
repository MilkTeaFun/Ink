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

func (s *Store) ListInstallations(ctx context.Context) ([]plugins.Installation, error) {
	rows, err := s.db.Query(ctx, `
		select id, plugin_key, source_type, display_name, version, runtime_type, manifest_json,
			current_path, status, last_error, installed_by, created_at, updated_at
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
		select id, plugin_key, source_type, display_name, version, runtime_type, manifest_json,
			current_path, status, last_error, installed_by, created_at, updated_at
		from plugin_installations
		where id = $1
	`, installationID)
	return scanPluginInstallation(row)
}

func (s *Store) FindInstallationByPluginKey(ctx context.Context, pluginKey string) (*plugins.Installation, error) {
	row := s.db.QueryRow(ctx, `
		select id, plugin_key, source_type, display_name, version, runtime_type, manifest_json,
			current_path, status, last_error, installed_by, created_at, updated_at
		from plugin_installations
		where plugin_key = $1
	`, pluginKey)
	return scanPluginInstallation(row)
}

func (s *Store) SaveInstallation(ctx context.Context, installation plugins.Installation) error {
	_, err := s.db.Exec(ctx, `
		insert into plugin_installations (
			id, plugin_key, source_type, display_name, version, runtime_type, manifest_json,
			current_path, status, last_error, installed_by, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
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
		installation.CreatedAt,
		installation.UpdatedAt,
	)
	return err
}

func (s *Store) ListPluginBindingsByUserID(ctx context.Context, userID string) ([]plugins.Binding, error) {
	rows, err := s.db.Query(ctx, `
		select id, plugin_installation_id, user_id, enabled, config_json, secret_ciphertext,
			secret_nonce, status, last_validated_at, last_error, created_at, updated_at
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
		select id, plugin_installation_id, user_id, enabled, config_json, secret_ciphertext,
			secret_nonce, status, last_validated_at, last_error, created_at, updated_at
		from plugin_bindings
		where plugin_installation_id = $1 and user_id = $2
	`, installationID, userID)
	return scanPluginBinding(row)
}

func (s *Store) SavePluginBinding(ctx context.Context, binding plugins.Binding) error {
	configJSON, err := json.Marshal(binding.Config)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		insert into plugin_bindings (
			id, plugin_installation_id, user_id, enabled, config_json, secret_ciphertext,
			secret_nonce, status, last_validated_at, last_error, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		on conflict (id)
		do update set
			plugin_installation_id = excluded.plugin_installation_id,
			user_id = excluded.user_id,
			enabled = excluded.enabled,
			config_json = excluded.config_json,
			secret_ciphertext = excluded.secret_ciphertext,
			secret_nonce = excluded.secret_nonce,
			status = excluded.status,
			last_validated_at = excluded.last_validated_at,
			last_error = excluded.last_error,
			updated_at = excluded.updated_at
	`,
		binding.ID,
		binding.PluginInstallationID,
		binding.UserID,
		binding.Enabled,
		configJSON,
		nullableBytes(binding.Ciphertext),
		nullableBytes(binding.Nonce),
		binding.Status,
		binding.LastValidatedAt,
		binding.LastError,
		binding.CreatedAt,
		binding.UpdatedAt,
	)
	return err
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
	var lastValidatedAt *time.Time
	var lastError *string
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
		&current.Status,
		&lastValidatedAt,
		&lastError,
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
	current.Ciphertext = ciphertext
	current.Nonce = nonce
	current.LastValidatedAt = lastValidatedAt
	current.LastError = lastError
	return &current, nil
}
