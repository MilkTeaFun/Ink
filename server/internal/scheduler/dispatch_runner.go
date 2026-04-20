package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/ruhuang/ink/server/internal/dispatch"
)

// DispatchProcessor drains pending backlog and retries inbox items that
// previously failed to dispatch.
type DispatchProcessor interface {
	DrainPending(ctx context.Context, limit int) (dispatch.FlushResult, error)
	RetryFailed(ctx context.Context, limit int) (dispatch.FlushResult, error)
}

// DispatchRunner periodically drains pending backlog and retries failed items
// so transient delivery failures self-heal and over-budget fetches keep moving.
type DispatchRunner struct {
	processor DispatchProcessor
	logger    *slog.Logger
	interval  time.Duration
	limit     int
}

func NewDispatchRunner(processor DispatchProcessor, logger *slog.Logger, interval time.Duration, limit int) *DispatchRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &DispatchRunner{
		processor: processor,
		logger:    logger,
		interval:  interval,
		limit:     limit,
	}
}

func (r *DispatchRunner) Start(ctx context.Context) {
	if r.processor == nil || r.interval <= 0 {
		return
	}

	ticker := time.NewTicker(r.interval)
	go func() {
		defer ticker.Stop()

		r.runOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.runOnce(ctx)
			}
		}
	}()
}

func (r *DispatchRunner) runOnce(ctx context.Context) {
	pendingResult, err := r.processor.DrainPending(ctx, r.limit)
	if err != nil {
		r.logger.Error("pending backlog drain failed", "error", err)
	} else if pendingResult.Printed > 0 || pendingResult.Failed > 0 || pendingResult.Skipped > 0 {
		r.logger.Info("drained pending inbox items",
			"printed", pendingResult.Printed,
			"failed", pendingResult.Failed,
			"skipped", pendingResult.Skipped,
		)
	}

	retryResult, err := r.processor.RetryFailed(ctx, r.limit)
	if err != nil {
		r.logger.Error("dispatch retry failed", "error", err)
		return
	}
	if retryResult.Printed > 0 || retryResult.Failed > 0 || retryResult.Skipped > 0 {
		r.logger.Info("retried failed inbox items",
			"printed", retryResult.Printed,
			"failed", retryResult.Failed,
			"skipped", retryResult.Skipped,
		)
	}
}
