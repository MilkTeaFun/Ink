package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ruhuang/ink/server/internal/ai"
	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/plugins"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/schedule"
	"github.com/ruhuang/ink/server/internal/session"
	"github.com/ruhuang/ink/server/internal/workspace"
)

type PluginService interface {
	ListAdminInstallations(ctx context.Context, accessToken string) ([]plugins.PluginDetails, error)
	UploadPlugin(ctx context.Context, accessToken string, filename string, source io.Reader) (plugins.PluginDetails, error)
	DisableInstallation(ctx context.Context, accessToken string, installationID string) (plugins.PluginDetails, error)
	ListUserPlugins(ctx context.Context, accessToken string) ([]plugins.PluginDetails, error)
	GetUserPlugin(ctx context.Context, accessToken string, installationID string) (plugins.PluginDetails, error)
	SaveBinding(ctx context.Context, accessToken string, installationID string, input plugins.BindingInput) (plugins.PluginDetails, error)
	TestBinding(ctx context.Context, accessToken string, installationID string, input plugins.BindingInput) (plugins.ValidationResult, error)
}

type ScheduleService interface {
	List(ctx context.Context, accessToken string) ([]schedule.ScheduleView, error)
	Create(ctx context.Context, accessToken string, input schedule.UpsertInput) (schedule.ScheduleView, error)
	Update(ctx context.Context, accessToken string, scheduleID string, input schedule.UpsertInput) (schedule.ScheduleView, error)
	Toggle(ctx context.Context, accessToken string, scheduleID string) (schedule.ScheduleView, error)
	Delete(ctx context.Context, accessToken string, scheduleID string) error
}

const pluginUploadMultipartMemory int64 = 32 << 10

// Server exposes the HTTP handlers for authentication endpoints.
type Server struct {
	auth                 auth.AuthService
	workspace            workspace.WorkspaceService
	ai                   ai.AIService
	printer              printer.PrinterService
	plugins              PluginService
	schedules            ScheduleService
	logger               *slog.Logger
	rateLimiter          *LoginRateLimiter
	pluginUploadMaxBytes int64
}

// NewServer wires the auth service, logger, and login rate limiter into an HTTP server.
func NewServer(
	authService auth.AuthService,
	workspaceService workspace.WorkspaceService,
	aiService ai.AIService,
	printerService printer.PrinterService,
	pluginService PluginService,
	scheduleService ScheduleService,
	logger *slog.Logger,
	rateWindow time.Duration,
	rateMax int,
	pluginUploadMaxBytes int64,
) *Server {
	return &Server{
		auth:                 authService,
		workspace:            workspaceService,
		ai:                   aiService,
		printer:              printerService,
		plugins:              pluginService,
		schedules:            scheduleService,
		logger:               logger,
		rateLimiter:          NewLoginRateLimiter(rateWindow, rateMax),
		pluginUploadMaxBytes: pluginUploadMaxBytes,
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
	mux.HandleFunc("POST /api/v1/admin/users", s.wrap(s.handleCreateUser))
	mux.HandleFunc("GET /api/v1/workspace", s.wrap(s.handleGetWorkspace))
	mux.HandleFunc("PUT /api/v1/workspace", s.wrap(s.handleSaveWorkspace))
	mux.HandleFunc("GET /api/v1/ai/config", s.wrap(s.handleGetAIConfig))
	mux.HandleFunc("PUT /api/v1/admin/ai/config", s.wrap(s.handleSaveAIConfig))
	mux.HandleFunc("POST /api/v1/ai/reply", s.wrap(s.handleGenerateAIReply))
	mux.HandleFunc("GET /api/v1/admin/plugins", s.wrap(s.handleListAdminPlugins))
	mux.HandleFunc("POST /api/v1/admin/plugins/upload", s.wrap(s.handleUploadPlugin))
	mux.HandleFunc("POST /api/v1/admin/plugins/{installationID}/disable", s.wrap(s.handleDisablePlugin))
	mux.HandleFunc("GET /api/v1/plugins", s.wrap(s.handleListPlugins))
	mux.HandleFunc("GET /api/v1/plugins/{installationID}", s.wrap(s.handleGetPlugin))
	mux.HandleFunc("PUT /api/v1/plugins/{installationID}/binding", s.wrap(s.handleSavePluginBinding))
	mux.HandleFunc("POST /api/v1/plugins/{installationID}/test", s.wrap(s.handleTestPluginBinding))
	mux.HandleFunc("GET /api/v1/printers", s.wrap(s.handleListPrinters))
	mux.HandleFunc("POST /api/v1/printers/bind", s.wrap(s.handleBindPrinter))
	mux.HandleFunc("DELETE /api/v1/printers/{printerID}", s.wrap(s.handleDeletePrinter))
	mux.HandleFunc("GET /api/v1/print-jobs", s.wrap(s.handleListPrintJobs))
	mux.HandleFunc("POST /api/v1/print-jobs", s.wrap(s.handleCreatePrintJob))
	mux.HandleFunc("POST /api/v1/print-jobs/{jobID}/submit", s.wrap(s.handleSubmitPrintJob))
	mux.HandleFunc("POST /api/v1/print-jobs/{jobID}/cancel", s.wrap(s.handleCancelPrintJob))
	mux.HandleFunc("PUT /api/v1/print-jobs/{jobID}/device", s.wrap(s.handleUpdatePrintJobDevice))
	mux.HandleFunc("GET /api/v1/print-schedules", s.wrap(s.handleListPrintSchedules))
	mux.HandleFunc("POST /api/v1/print-schedules", s.wrap(s.handleCreatePrintSchedule))
	mux.HandleFunc("PUT /api/v1/print-schedules/{scheduleID}", s.wrap(s.handleUpdatePrintSchedule))
	mux.HandleFunc("POST /api/v1/print-schedules/{scheduleID}/toggle", s.wrap(s.handleTogglePrintSchedule))
	mux.HandleFunc("DELETE /api/v1/print-schedules/{scheduleID}", s.wrap(s.handleDeletePrintSchedule))
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

type createUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type aiConfigRequest struct {
	ProviderName string `json:"providerName"`
	ProviderType string `json:"providerType"`
	BaseURL      string `json:"baseUrl"`
	Model        string `json:"model"`
	APIKey       string `json:"apiKey"`
}

type aiReplyRequest struct {
	Messages []ai.ChatMessage `json:"messages"`
}

type bindPrinterRequest struct {
	Name     string `json:"name"`
	Note     string `json:"note"`
	DeviceID string `json:"deviceId"`
}

type createPrintJobRequest struct {
	Title             string `json:"title"`
	Source            string `json:"source"`
	Content           string `json:"content"`
	PrinterBindingID  string `json:"printerBindingId"`
	SubmitImmediately bool   `json:"submitImmediately"`
}

type updatePrintJobDeviceRequest struct {
	PrinterBindingID string `json:"printerBindingId"`
}

type pluginBindingRequest struct {
	Enabled bool              `json:"enabled"`
	Config  map[string]any    `json:"config"`
	Secrets map[string]string `json:"secrets"`
}

type printScheduleRequest struct {
	Title                string                 `json:"title"`
	PluginInstallationID string                 `json:"pluginInstallationId"`
	FrequencyType        schedule.FrequencyType `json:"frequencyType"`
	Timezone             string                 `json:"timezone"`
	Hour                 int                    `json:"hour"`
	Minute               int                    `json:"minute"`
	Weekdays             []int                  `json:"weekdays"`
	ScheduleConfig       map[string]any         `json:"scheduleConfig"`
	DeviceID             string                 `json:"deviceId"`
	Enabled              bool                   `json:"enabled"`
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

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	created, err := s.auth.CreateUser(r.Context(), accessToken, auth.CreateUserInput{
		Email:    payload.Email,
		Name:     payload.Name,
		Password: payload.Password,
	})
	if err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]auth.UserDTO{"user": created})
}

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	state, err := s.workspace.GetState(r.Context(), accessToken)
	if err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, state)
}

func (s *Server) handleSaveWorkspace(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var state workspace.State
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	saved, err := s.workspace.SaveState(r.Context(), accessToken, state)
	if err != nil {
		s.writeAuthError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) handleGetAIConfig(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	summary, err := s.ai.GetConfigSummary(r.Context(), accessToken)
	if err != nil {
		s.writeAIError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleSaveAIConfig(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload aiConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	summary, err := s.ai.UpdateSystemConfig(r.Context(), accessToken, ai.UpdateConfigInput{
		ProviderName: payload.ProviderName,
		ProviderType: payload.ProviderType,
		BaseURL:      payload.BaseURL,
		Model:        payload.Model,
		APIKey:       payload.APIKey,
	})
	if err != nil {
		s.writeAIError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleGenerateAIReply(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload aiReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	reply, err := s.ai.GenerateReply(r.Context(), accessToken, ai.ReplyInput{Messages: payload.Messages})
	if err != nil {
		s.writeAIError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, reply)
}

func (s *Server) handleListAdminPlugins(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	items, err := s.plugins.ListAdminInstallations(r.Context(), accessToken)
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]plugins.PluginDetails{"plugins": items})
}

func (s *Server) handleUploadPlugin(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.pluginUploadMaxBytes)
	if err := r.ParseMultipartForm(pluginUploadMultipartMemory); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_upload", "插件上传包无效或体积过大。")
		return
	}
	if r.MultipartForm != nil {
		defer func() {
			_ = r.MultipartForm.RemoveAll()
		}()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, requestID, http.StatusBadRequest, "missing_file", "请上传 ZIP 插件包。")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	details, err := s.plugins.UploadPlugin(r.Context(), accessToken, header.Filename, file)
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]plugins.PluginDetails{"plugin": details})
}

func (s *Server) handleDisablePlugin(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	details, err := s.plugins.DisableInstallation(r.Context(), accessToken, r.PathValue("installationID"))
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]plugins.PluginDetails{"plugin": details})
}

func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	items, err := s.plugins.ListUserPlugins(r.Context(), accessToken)
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]plugins.PluginDetails{"plugins": items})
}

func (s *Server) handleGetPlugin(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	details, err := s.plugins.GetUserPlugin(r.Context(), accessToken, r.PathValue("installationID"))
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]plugins.PluginDetails{"plugin": details})
}

func (s *Server) handleSavePluginBinding(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload pluginBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	details, err := s.plugins.SaveBinding(r.Context(), accessToken, r.PathValue("installationID"), plugins.BindingInput{
		Enabled: payload.Enabled,
		Config:  payload.Config,
		Secrets: payload.Secrets,
	})
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]plugins.PluginDetails{"plugin": details})
}

func (s *Server) handleTestPluginBinding(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload pluginBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	result, err := s.plugins.TestBinding(r.Context(), accessToken, r.PathValue("installationID"), plugins.BindingInput{
		Enabled: payload.Enabled,
		Config:  payload.Config,
		Secrets: payload.Secrets,
	})
	if err != nil {
		s.writePluginError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]plugins.ValidationResult{"result": result})
}

func (s *Server) handleListPrinters(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	devices, err := s.printer.ListDevices(r.Context(), accessToken)
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]workspace.Device{"devices": devices})
}

func (s *Server) handleBindPrinter(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload bindPrinterRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	device, err := s.printer.BindDevice(r.Context(), accessToken, printer.BindInput{
		Name:     payload.Name,
		Note:     payload.Note,
		DeviceID: payload.DeviceID,
	})
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]workspace.Device{"device": device})
}

func (s *Server) handleDeletePrinter(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	if err := s.printer.DeleteDevice(r.Context(), accessToken, r.PathValue("printerID")); err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListPrintJobs(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	jobs, err := s.printer.ListPrintJobs(r.Context(), accessToken)
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]workspace.PrintJob{"printJobs": jobs})
}

func (s *Server) handleCreatePrintJob(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload createPrintJobRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	job, err := s.printer.CreatePrintJob(r.Context(), accessToken, printer.CreateJobInput{
		Title:             payload.Title,
		Source:            payload.Source,
		Content:           payload.Content,
		PrinterBindingID:  payload.PrinterBindingID,
		SubmitImmediately: payload.SubmitImmediately,
	})
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]workspace.PrintJob{"printJob": job})
}

func (s *Server) handleSubmitPrintJob(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	job, err := s.printer.SubmitPrintJob(r.Context(), accessToken, r.PathValue("jobID"))
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]workspace.PrintJob{"printJob": job})
}

func (s *Server) handleCancelPrintJob(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	job, err := s.printer.CancelPrintJob(r.Context(), accessToken, r.PathValue("jobID"))
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]workspace.PrintJob{"printJob": job})
}

func (s *Server) handleUpdatePrintJobDevice(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload updatePrintJobDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	job, err := s.printer.UpdatePrintJobDevice(r.Context(), accessToken, r.PathValue("jobID"), printer.UpdateJobDeviceInput{
		PrinterBindingID: payload.PrinterBindingID,
	})
	if err != nil {
		s.writePrinterError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]workspace.PrintJob{"printJob": job})
}

func (s *Server) handleListPrintSchedules(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	items, err := s.schedules.List(r.Context(), accessToken)
	if err != nil {
		s.writeScheduleError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]schedule.ScheduleView{"schedules": items})
}

func (s *Server) handleCreatePrintSchedule(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload printScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	item, err := s.schedules.Create(r.Context(), accessToken, schedule.UpsertInput{
		Title:                payload.Title,
		PluginInstallationID: payload.PluginInstallationID,
		FrequencyType:        payload.FrequencyType,
		Timezone:             payload.Timezone,
		Hour:                 payload.Hour,
		Minute:               payload.Minute,
		Weekdays:             payload.Weekdays,
		ScheduleConfig:       payload.ScheduleConfig,
		DeviceID:             payload.DeviceID,
		Enabled:              payload.Enabled,
	})
	if err != nil {
		s.writeScheduleError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]schedule.ScheduleView{"schedule": item})
}

func (s *Server) handleUpdatePrintSchedule(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	var payload printScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, requestID, http.StatusBadRequest, "invalid_request", "请求格式不正确。")
		return
	}

	item, err := s.schedules.Update(r.Context(), accessToken, r.PathValue("scheduleID"), schedule.UpsertInput{
		Title:                payload.Title,
		PluginInstallationID: payload.PluginInstallationID,
		FrequencyType:        payload.FrequencyType,
		Timezone:             payload.Timezone,
		Hour:                 payload.Hour,
		Minute:               payload.Minute,
		Weekdays:             payload.Weekdays,
		ScheduleConfig:       payload.ScheduleConfig,
		DeviceID:             payload.DeviceID,
		Enabled:              payload.Enabled,
	})
	if err != nil {
		s.writeScheduleError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]schedule.ScheduleView{"schedule": item})
}

func (s *Server) handleTogglePrintSchedule(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	item, err := s.schedules.Toggle(r.Context(), accessToken, r.PathValue("scheduleID"))
	if err != nil {
		s.writeScheduleError(w, requestID, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]schedule.ScheduleView{"schedule": item})
}

func (s *Server) handleDeletePrintSchedule(w http.ResponseWriter, r *http.Request, requestID string) {
	accessToken := bearerToken(r.Header.Get("Authorization"))
	if accessToken == "" {
		writeError(w, requestID, http.StatusUnauthorized, "unauthorized", "请先登录。")
		return
	}

	if err := s.schedules.Delete(r.Context(), accessToken, r.PathValue("scheduleID")); err != nil {
		s.writeScheduleError(w, requestID, err)
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
			"密码至少 8 位。",
		)
	case errors.Is(err, auth.ErrInvalidProfile):
		writeError(w, requestID, http.StatusBadRequest, "invalid_profile", "请输入有效的账号信息。")
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, requestID, http.StatusConflict, "email_taken", "该账号已存在。")
	case errors.Is(err, auth.ErrForbidden):
		writeError(w, requestID, http.StatusForbidden, "forbidden", "当前账号没有该操作权限。")
	case errors.Is(err, auth.ErrUserDisabled):
		writeError(w, requestID, http.StatusLocked, "user_disabled", "账号已被禁用。")
	default:
		s.logger.Error("auth handler failed", "request_id", requestID, "error", err)
		writeError(w, requestID, http.StatusInternalServerError, "internal_error", "服务暂时不可用，请稍后重试。")
	}
}

func (s *Server) writeAIError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, ai.ErrForbidden):
		writeError(w, requestID, http.StatusForbidden, "forbidden", "当前账号没有该操作权限。")
	case errors.Is(err, ai.ErrNotConfigured):
		writeError(w, requestID, http.StatusPreconditionFailed, "ai_not_configured", "当前还没有配置 AI 服务。")
	case errors.Is(err, ai.ErrMissingSecret):
		writeError(w, requestID, http.StatusServiceUnavailable, "ai_secret_missing", "服务端尚未配置 AI 加密密钥。")
	case errors.Is(err, ai.ErrInvalidConfig):
		writeError(w, requestID, http.StatusBadRequest, "invalid_ai_config", "请输入有效的 AI 服务配置。")
	case errors.Is(err, ai.ErrInvalidInput):
		writeError(w, requestID, http.StatusBadRequest, "invalid_ai_input", "请输入有效的对话内容。")
	case errors.Is(err, ai.ErrProviderUnavailable):
		writeError(w, requestID, http.StatusBadGateway, "ai_provider_unavailable", "AI 服务暂时不可用，请稍后重试。")
	default:
		s.writeAuthError(w, requestID, err)
	}
}

func (s *Server) writePrinterError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, printer.ErrForbidden):
		writeError(w, requestID, http.StatusForbidden, "forbidden", "当前账号没有该操作权限。")
	case errors.Is(err, printer.ErrNotConfigured):
		writeError(w, requestID, http.StatusPreconditionFailed, "printer_not_configured", "当前还没有配置 Memobird 服务。")
	case errors.Is(err, printer.ErrNotFound):
		writeError(w, requestID, http.StatusNotFound, "printer_resource_not_found", "指定的设备或打印任务不存在。")
	case errors.Is(err, printer.ErrInvalidInput):
		writeError(w, requestID, http.StatusBadRequest, "invalid_printer_input", "请输入有效的设备或打印信息。")
	case errors.Is(err, printer.ErrUnavailable):
		writeError(w, requestID, http.StatusBadGateway, "printer_unavailable", "咕咕机服务暂时不可用，请稍后重试。")
	default:
		s.writeAuthError(w, requestID, err)
	}
}

func (s *Server) writePluginError(w http.ResponseWriter, requestID string, err error) {
	var validationFailure plugins.ValidationFailure
	switch {
	case errors.Is(err, plugins.ErrForbidden):
		writeError(w, requestID, http.StatusForbidden, "forbidden", "当前账号没有该操作权限。")
	case errors.Is(err, plugins.ErrNotFound):
		writeError(w, requestID, http.StatusNotFound, "plugin_not_found", "指定插件不存在。")
	case errors.Is(err, plugins.ErrMissingSecret):
		writeError(w, requestID, http.StatusServiceUnavailable, "plugin_secret_missing", "服务端尚未配置插件加密密钥。")
	case errors.As(err, &validationFailure):
		writeError(w, requestID, http.StatusBadRequest, "invalid_plugin_config", validationFailure.Error())
	case errors.Is(err, plugins.ErrInvalidInput):
		writeError(w, requestID, http.StatusBadRequest, "invalid_plugin_input", "请输入有效的插件配置。")
	case errors.Is(err, plugins.ErrInvalidPlugin):
		writeError(w, requestID, http.StatusBadRequest, "invalid_plugin_package", err.Error())
	case errors.Is(err, plugins.ErrExecutionFailed):
		writeError(w, requestID, http.StatusBadGateway, "plugin_execution_failed", err.Error())
	default:
		s.writeAuthError(w, requestID, err)
	}
}

func (s *Server) writeScheduleError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, schedule.ErrNotFound):
		writeError(w, requestID, http.StatusNotFound, "schedule_not_found", "指定定时任务不存在。")
	case errors.Is(err, schedule.ErrInvalidInput):
		writeError(w, requestID, http.StatusBadRequest, "invalid_schedule_input", err.Error())
	default:
		var validationFailure plugins.ValidationFailure
		if errors.As(err, &validationFailure) {
			writeError(w, requestID, http.StatusBadRequest, "invalid_schedule_config", validationFailure.Error())
			return
		}
		s.writePluginError(w, requestID, err)
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
