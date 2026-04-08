package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/platform/clock"
	"github.com/ruhuang/ink/server/internal/platform/config"
	"github.com/ruhuang/ink/server/internal/platform/httpapi"
	"github.com/ruhuang/ink/server/internal/platform/idgen"
	"github.com/ruhuang/ink/server/internal/platform/password"
	"github.com/ruhuang/ink/server/internal/platform/store/postgres"
	"github.com/ruhuang/ink/server/internal/platform/token"
	"github.com/ruhuang/ink/server/internal/workspace"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		panic(err)
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}

	store := postgres.New(db)
	tokenManager, err := token.NewJWTAccessManager(cfg.JWTSecret, cfg.AppName, cfg.AccessTokenTTL)
	if err != nil {
		panic(err)
	}

	service := auth.NewService(
		store,
		store,
		store,
		password.BcryptHasher{},
		tokenManager,
		clock.SystemClock{},
		idgen.Generator{},
		cfg.RefreshTokenTTL,
	)
	workspaceService := workspace.NewService(store, service, clock.SystemClock{})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           httpapi.NewServer(service, workspaceService, logger, cfg.RateLimitWindow, cfg.RateLimitMax).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting auth api", "port", cfg.Port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}
