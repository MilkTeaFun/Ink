package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type Processor interface {
	ProcessDue(ctx context.Context, limit int) (int, error)
}

type Runner struct {
	processor Processor
	logger    *slog.Logger
	interval  time.Duration
	limit     int
}

func NewRunner(processor Processor, logger *slog.Logger, interval time.Duration, limit int) *Runner {
	if logger == nil {
		logger = slog.Default()
	}

	return &Runner{
		processor: processor,
		logger:    logger,
		interval:  interval,
		limit:     limit,
	}
}

func (r *Runner) Start(ctx context.Context) {
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

func (r *Runner) runOnce(ctx context.Context) {
	processed, err := r.processor.ProcessDue(ctx, r.limit)
	if err != nil {
		r.logger.Error("schedule processor failed", "error", err)
		return
	}
	if processed > 0 {
		r.logger.Info("processed due schedules", "count", processed)
	}
}
