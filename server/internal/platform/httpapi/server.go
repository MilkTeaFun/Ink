package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/session"
)

// Server exposes the HTTP handlers for authentication endpoints.
type Server struct {
	auth        auth.AuthService
	logger      *slog.Logger
	rateLimiter *LoginRateLimiter
}

// NewServer wires the auth service, logger, and login rate limiter into an HTTP server.
func NewServer(authService auth.AuthService, logger *slog.Logger, rateWindow time.Duration, rateMax int) *Server {
	return &Server{
		auth:        authService,
		logger:      logger,
		rateLimiter: NewLoginRateLimiter(rateWindow, rateMax),
	}
}

// Handler builds the HTTP handler tree for the auth API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /api/v1/auth/login", s.wrap(s.handleLogin))
	mux.HandleFunc("POST /api/v1/auth/refresh", s.wrap(s.handleRefresh))
	mux.HandleFunc("POST /api/v1/auth/logout", s.wrap(s.handleLogout))
	mux.HandleFunc("POST /api/v1/auth/change-password", s.wrap(s.handleChangePassword))
	mux.HandleFunc("GET /api/v1/auth/me", s.wrap(s.handleMe))
	return mux
}

type responseEnvelope struct {
	User         auth.UserDTO `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int64        `json:"expiresIn"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type errorEnvelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func (s *Server) wrap(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("Content-Type", "application/json")
		next(w, r, requestID)
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request, requestID string) {
	var payload loginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	meta := clientMetaFromRequest(r)
	if !s.rateLimiter.Allow(meta.IPAddress + ":" + auth.NormalizeEmail(payload.Email)) {
		writeError(w, requestID, http.StatusTooManyRequests, "rate_limited", "登录尝试过于频繁，请稍后再试。")
		return
	}

	result, err := s.auth.Login(r.Context(), auth.LoginInput{
		Email:    payload.Email,
		Password: payload.Password,
		Meta:     meta,
	})
	if err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{
		User:         result.User,
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		ExpiresIn:    int64(time.Until(result.Token.AccessTokenExpiresAt).Seconds()),
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request, requestID string) {
	var payload refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	result, err := s.auth.Refresh(r.Context(), payload.RefreshToken, clientMetaFromRequest(r))
	if err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{
		User:         result.User,
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		ExpiresIn:    int64(time.Until(result.Token.AccessTokenExpiresAt).Seconds()),
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	account, err := s.auth.GetCurrentUser(r.Context(), accessToken)
	if err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]auth.UserDTO{"user": account})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request, requestID string) {
	var payload logoutRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	}

	if err := s.auth.Logout(r.Context(), bearerToken(r.Header.Get("Authorization")), payload.RefreshToken); err != nil {
		s.logger.Warn("logout failed", "request_id", requestID, "error", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	if err := s.auth.ChangePassword(
		r.Context(),
		accessToken,
		payload.CurrentPassword,
		payload.NewPassword,
		clientMetaFromRequest(r),
	); err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeAuthError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, requestID, http.StatusUnauthorized, "invalid_credentials", "账号或密码不正确。")
	case errors.Is(err, auth.ErrCurrentPassword):
		writeError(w, requestID, http.StatusUnauthorized, "current_password_incorrect", "当前密码不正确。")
	case errors.Is(err, auth.ErrInvalidRefreshToken), errors.Is(err, auth.ErrInvalidAccessToken):
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "登录状态已失效，请重新登录。")
	case errors.Is(err, auth.ErrWeakPassword):
		writeError(
			w,
			requestID,
			http.StatusBadRequest,
			"invalid_password",
			"新密码至少 8 位，且不能与当前密码相同。",
		)
	case errors.Is(err, auth.ErrUserDisabled):
		writeError(w, requestID, http.StatusLocked, "user_disabled", "账号已被禁用。")
	default:
		s.logger.Error("auth handler failed", "request_id", requestID, "error", err)
		writeError(w, requestID, http.StatusInternalServerError, "internal_error", "服务暂时不可用，请稍后重试。")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, requestID string, status int, code string, message string) {
	writeJSON(w, status, errorEnvelope{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func clientMetaFromRequest(r *http.Request) auth.ClientMeta {
	return auth.ClientMeta{
		ClientType: session.ClientTypeWeb,
		UserAgent:  r.UserAgent(),
		IPAddress:  requestIP(r),
	}
}

func requestIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

// LoginRateLimiter limits repeated login attempts within a fixed time window.
type LoginRateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	max    int
	hits   map[string][]time.Time
}

// NewLoginRateLimiter creates a rate limiter for login attempts.
func NewLoginRateLimiter(window time.Duration, max int) *LoginRateLimiter {
	return &LoginRateLimiter{
		window: window,
		max:    max,
		hits:   make(map[string][]time.Time),
	}
}

// Allow records a login attempt and reports whether it is still within the limit.
func (l *LoginRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	windowHits := l.hits[key][:0]

	for _, hit := range l.hits[key] {
		if hit.After(cutoff) {
			windowHits = append(windowHits, hit)
		}
	}

	if len(windowHits) >= l.max {
		l.hits[key] = windowHits
		return false
	}

	windowHits = append(windowHits, now)
	l.hits[key] = windowHits
	return true
}

type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID stores the request identifier on a context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}
