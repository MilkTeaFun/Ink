package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config contains the runtime settings required by the auth service.
type Config struct {
	AppName         string
	Port            int
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	RateLimitWindow time.Duration
	RateLimitMax    int
}

// Load reads application configuration from the current environment.
func Load() (Config, error) {
	port, err := envInt("PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	accessTokenTTL, err := envDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	refreshTokenTTL, err := envDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	rateLimitWindow, err := envDuration("LOGIN_RATE_LIMIT_WINDOW", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	rateLimitMax, err := envInt("LOGIN_RATE_LIMIT_MAX", 10)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppName:         envString("APP_NAME", "ink-auth"),
		Port:            port,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
		RateLimitWindow: rateLimitWindow,
		RateLimitMax:    rateLimitMax,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	if cfg.Port <= 0 {
		return Config{}, fmt.Errorf("PORT must be positive")
	}

	if cfg.AccessTokenTTL <= 0 {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_TTL must be positive")
	}

	if cfg.RefreshTokenTTL <= 0 {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be positive")
	}

	if cfg.RateLimitWindow <= 0 {
		return Config{}, fmt.Errorf("LOGIN_RATE_LIMIT_WINDOW must be positive")
	}

	if cfg.RateLimitMax <= 0 {
		return Config{}, fmt.Errorf("LOGIN_RATE_LIMIT_MAX must be positive")
	}

	return cfg, nil
}

// LoadDotEnv loads the first local dotenv file that exists.
func LoadDotEnv() error {
	candidates := []string{
		".env",
		filepath.Join("server", ".env"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return godotenv.Load(candidate)
		}
	}

	return nil
}

// ResolveProjectPath resolves a path from either the repo root or server directory.
func ResolveProjectPath(path string) string {
	candidates := []string{
		path,
		filepath.Join("server", path),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return path
}

func envString(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return parsed, nil
}
