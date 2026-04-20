package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ruhuang/ink/server/internal/inbox"
	"github.com/ruhuang/ink/server/internal/plugins"
)

var _ inbox.Repository = (*Store)(nil)

const pluginItemColumns = `id, user_id, plugin_installation_id, plugin_binding_id, device_id, external_id,
	title, source_label, published_at, blocks_json, status, attempt_count, last_error,
	print_job_id, fetched_at, created_at, updated_at`

// InsertItem stores a new plugin_items row. When a conflict on
// (plugin_binding_id, external_id) occurs the call returns false without
// mutating the existing row.
func (s *Store) InsertItem(ctx context.Context, item inbox.Item) (bool, error) {
	blocksJSON, err := json.Marshal(item.Blocks)
	if err != nil {
		return false, err
	}

	tag, err := s.db.Exec(ctx, `
		insert into plugin_items (
			id, user_id, plugin_installation_id, plugin_binding_id, device_id, external_id,
			title, source_label, published_at, blocks_json, status, attempt_count, last_error,
			print_job_id, fetched_at, created_at, updated_at
		) values (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
		on conflict (plugin_binding_id, external_id) do nothing
	`,
		item.ID,
		item.UserID,
		item.PluginInstallationID,
		item.PluginBindingID,
		item.DeviceID,
		item.ExternalID,
		item.Title,
		item.SourceLabel,
		item.PublishedAt,
		blocksJSON,
		string(item.Status),
		item.AttemptCount,
		item.LastError,
		item.PrintJobID,
		item.FetchedAt,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() > 0, nil
}

// FindInboxItemByID loads a single inbox item by id.
func (s *Store) FindInboxItemByID(ctx context.Context, itemID string) (*inbox.Item, error) {
	row := s.db.QueryRow(ctx, `select `+pluginItemColumns+` from plugin_items where id = $1`, itemID)
	return scanInboxItem(row)
}

// ListPendingByBinding returns pending items for a binding in insertion order.
func (s *Store) ListPendingByBinding(ctx context.Context, bindingID string, limit int) ([]inbox.Item, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `
		select `+pluginItemColumns+`
		from plugin_items
		where plugin_binding_id = $1
		  and status = $2
		order by created_at asc
		limit $3
	`, bindingID, string(inbox.StatusPending), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []inbox.Item{}
	for rows.Next() {
		current, err := scanInboxItem(rows)
		if err != nil {
			return nil, err
		}
		if current == nil {
			continue
		}
		result = append(result, *current)
	}
	return result, rows.Err()
}

// ListPendingBindingIDs returns bindings with pending items ordered by their
// oldest pending row so backlog draining stays fair across bindings.
func (s *Store) ListPendingBindingIDs(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `
		select plugin_binding_id
		from plugin_items
		where status = $1
		group by plugin_binding_id
		order by min(created_at) asc
		limit $2
	`, string(inbox.StatusPending), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []string{}
	for rows.Next() {
		var bindingID string
		if err := rows.Scan(&bindingID); err != nil {
			return nil, err
		}
		result = append(result, bindingID)
	}
	return result, rows.Err()
}

// ListRetryable returns failed items whose attempt count is below the retry
// ceiling and whose last update is older than olderThan.
func (s *Store) ListRetryable(ctx context.Context, olderThan time.Time, limit int) ([]inbox.Item, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `
		select `+pluginItemColumns+`
		from plugin_items
		where status = $1
		  and attempt_count < $2
		  and updated_at < $3
		order by updated_at asc
		limit $4
	`, string(inbox.StatusFailed), inbox.MaxDispatchAttempts, olderThan, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []inbox.Item{}
	for rows.Next() {
		current, err := scanInboxItem(rows)
		if err != nil {
			return nil, err
		}
		if current == nil {
			continue
		}
		result = append(result, *current)
	}
	return result, rows.Err()
}

// UpdateStatus persists a status transition for an existing item.
func (s *Store) UpdateStatus(ctx context.Context, item inbox.Item) error {
	_, err := s.db.Exec(ctx, `
		update plugin_items set
			status = $2,
			attempt_count = $3,
			last_error = $4,
			print_job_id = $5,
			updated_at = $6
		where id = $1
	`,
		item.ID,
		string(item.Status),
		item.AttemptCount,
		item.LastError,
		item.PrintJobID,
		item.UpdatedAt,
	)
	return err
}

// DeletePrintedOlderThan prunes printed items whose updated_at falls before
// the given cutoff.
func (s *Store) DeletePrintedOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := s.db.Exec(ctx, `
		delete from plugin_items
		where status = $1 and updated_at < $2
	`, string(inbox.StatusPrinted), cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// CountPrintedInLast24h returns how many items for a binding were printed
// after the given instant.
func (s *Store) CountPrintedInLast24h(ctx context.Context, bindingID string, since time.Time) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		select count(*)
		from plugin_items
		where plugin_binding_id = $1
		  and status = $2
		  and updated_at >= $3
	`, bindingID, string(inbox.StatusPrinted), since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func scanInboxItem(row pgx.Row) (*inbox.Item, error) {
	var current inbox.Item
	var deviceID *string
	var publishedAt *time.Time
	var blocksJSON []byte
	var status string
	var lastError *string
	var printJobID *string
	if err := row.Scan(
		&current.ID,
		&current.UserID,
		&current.PluginInstallationID,
		&current.PluginBindingID,
		&deviceID,
		&current.ExternalID,
		&current.Title,
		&current.SourceLabel,
		&publishedAt,
		&blocksJSON,
		&status,
		&current.AttemptCount,
		&lastError,
		&printJobID,
		&current.FetchedAt,
		&current.CreatedAt,
		&current.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	current.DeviceID = deviceID
	current.PublishedAt = publishedAt
	current.Status = inbox.ItemStatus(status)
	current.LastError = lastError
	current.PrintJobID = printJobID
	if len(blocksJSON) > 0 {
		if err := json.Unmarshal(blocksJSON, &current.Blocks); err != nil {
			return nil, err
		}
	}
	if current.Blocks == nil {
		current.Blocks = []plugins.ContentBlock{}
	}
	return &current, nil
}
