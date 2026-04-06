package seed

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunFile(ctx context.Context, db *pgxpool.Pool, path string) error {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read seed file %s: %w", path, err)
	}

	if _, err := db.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("apply seed file %s: %w", path, err)
	}

	return nil
}
