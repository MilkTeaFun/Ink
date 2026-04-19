package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/ruhuang/ink/server/internal/dispatch"
)

// DispatchProcessor retries inbox items that previously failed to dispatch.
type DispatchProcessor interface {
	RetryFailed(ctx context.Context, limit int) (dispatch.FlushResult, error)
}

// DispatchRunner periodically invokes RetryFailed on the dispatch service so
// transient delivery failures self-heal without operator intervention.
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
	result, err := r.processor.RetryFailed(ctx, r.limit)
	if err != nil {
		r.logger.Error("dispatch retry failed", "error", err)
		return
	}
	if result.Printed > 0 || result.Failed > 0 {
		r.logger.Info("retried failed inbox items",
			"printed", result.Printed,
			"failed", result.Failed,
			"skipped", result.Skipped,
		)
	}
}
