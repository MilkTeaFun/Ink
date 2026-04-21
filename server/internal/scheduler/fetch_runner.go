package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type FetchProcessor interface {
	ProcessDue(ctx context.Context, limit int) (int, error)
}

type FetchRunner struct {
	processor FetchProcessor
	logger    *slog.Logger
	interval  time.Duration
	limit     int
}

func NewFetchRunner(processor FetchProcessor, logger *slog.Logger, interval time.Duration, limit int) *FetchRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &FetchRunner{
		processor: processor,
		logger:    logger,
		interval:  interval,
		limit:     limit,
	}
}

func (r *FetchRunner) Start(ctx context.Context) {
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

func (r *FetchRunner) runOnce(ctx context.Context) {
	processed, err := r.processor.ProcessDue(ctx, r.limit)
	if err != nil {
		r.logger.Error("plugin fetch processor failed", "error", err)
		return
	}
	if processed > 0 {
		r.logger.Info("processed due plugin fetches", "count", processed)
	}
}
