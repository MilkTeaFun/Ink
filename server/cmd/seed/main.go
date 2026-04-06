package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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

	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	serverDir := filepath.Dir(config.ResolveProjectPath(".env.example"))
	result, err := seed.EnsureDevAdmin(ctx, db, seed.DevAdminOptions{
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
