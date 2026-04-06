package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/ruhuang/ink/server/internal/auth"
)

func TestLoginHandlerReturnsTokens(t *testing.T) {
	server := NewServer(fakeAuthService{
		loginResult: auth.AuthResult{
			User: auth.UserDTO{
				ID:    "user-1",
				Email: "name@example.com",
				Name:  "Ink User",
			},
			Token: auth.TokenPair{
				AccessToken:          "access-token",
				RefreshToken:         "refresh-token",
				AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			},
		},
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"name@example.com","password":"demo-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["accessToken"] != "access-token" {
		t.Fatalf("expected access token in response")
	}
}

func TestMeRequiresBearerToken(t *testing.T) {
	server := NewServer(fakeAuthService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestLoginRateLimit(t *testing.T) {
	server := NewServer(fakeAuthService{
		loginResult: auth.AuthResult{
			User: auth.UserDTO{
				ID:    "user-1",
				Email: "name@example.com",
				Name:  "Ink User",
			},
			Token: auth.TokenPair{
				AccessToken:          "access-token",
				RefreshToken:         "refresh-token",
				AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			},
		},
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 1)

	first := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"name@example.com","password":"demo-password"}`))
	first.RemoteAddr = "127.0.0.1:1234"
	second := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"name@example.com","password":"demo-password"}`))
	second.RemoteAddr = "127.0.0.1:1234"
	firstResponse := httptest.NewRecorder()
	secondResponse := httptest.NewRecorder()

	server.Handler().ServeHTTP(firstResponse, first)
	server.Handler().ServeHTTP(secondResponse, second)

	if firstResponse.Code != http.StatusOK {
		t.Fatalf("expected first login to succeed, got %d", firstResponse.Code)
	}
	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second login to be rate limited, got %d", secondResponse.Code)
	}
}

func TestChangePasswordReturnsNoContent(t *testing.T) {
	server := NewServer(fakeAuthService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/change-password",
		bytes.NewBufferString(`{"currentPassword":"demo-password","newPassword":"next-password"}`),
	)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
}

type fakeAuthService struct {
	loginResult           auth.AuthResult
	loginErr              error
	changePasswordErr     error
	lastChangeAccessToken string
	lastCurrentPassword   string
	lastNewPassword       string
}

func (f fakeAuthService) Login(_ context.Context, _ auth.LoginInput) (auth.AuthResult, error) {
	return f.loginResult, f.loginErr
}

func (f fakeAuthService) Refresh(_ context.Context, _ string, _ auth.ClientMeta) (auth.AuthResult, error) {
	return auth.AuthResult{}, nil
}

func (f fakeAuthService) Logout(_ context.Context, _ string, _ string) error {
	return nil
}

func (f fakeAuthService) GetCurrentUser(_ context.Context, _ string) (auth.UserDTO, error) {
	return auth.UserDTO{
		ID:    "user-1",
		Email: "name@example.com",
		Name:  "Ink User",
	}, nil
}

func (f fakeAuthService) ChangePassword(
	_ context.Context,
	accessToken string,
	currentPassword string,
	newPassword string,
	_ auth.ClientMeta,
) error {
	f.lastChangeAccessToken = accessToken
	f.lastCurrentPassword = currentPassword
	f.lastNewPassword = newPassword
	return f.changePasswordErr
}

var _ auth.AuthService = fakeAuthService{}
