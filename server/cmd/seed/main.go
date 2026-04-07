package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruhuang/ink/server/internal/platform/config"
	"github.com/ruhuang/ink/server/internal/platform/seed"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "dev" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/seed dev")
		os.Exit(1)
	}

	if err := config.LoadDotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "load .env: %v\n", err)
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelConnect()

	db, err := pgxpool.New(connectCtx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(connectCtx); err != nil {
		fmt.Fprintf(os.Stderr, "ping database: %v\n", err)
		os.Exit(1)
	}

	serverDir := filepath.Dir(config.ResolveProjectPath(".env.example"))
	seedCtx, cancelSeed := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelSeed()

	result, err := seed.EnsureDevAdmin(seedCtx, db, seed.DevAdminOptions{
		CredentialsPath: filepath.Join(serverDir, ".dev-admin-password"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run seed: %v\n", err)
		os.Exit(1)
	}

	if result.Created {
		fmt.Println("seeded development admin account")
		fmt.Printf("admin login: %s\n", result.Login)
		fmt.Printf("initial password: %s\n", result.Password)
		fmt.Printf("saved to: %s\n", result.CredentialsPath)
		return
	}

	fmt.Printf("development admin account already exists: %s\n", result.Login)
	if result.CredentialsPath != "" && credentialsFileExists(result.CredentialsPath) {
		fmt.Printf("saved credentials: %s\n", result.CredentialsPath)
	}
}

func credentialsFileExists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
}
