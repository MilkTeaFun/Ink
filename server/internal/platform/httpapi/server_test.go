package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/ruhuang/ink/server/internal/ai"
	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/feedback"
	"github.com/ruhuang/ink/server/internal/plugins"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/schedule"
	"github.com/ruhuang/ink/server/internal/workspace"
)

func TestLoginHandlerReturnsTokens(t *testing.T) {
	server := NewServer(fakeAuthService{
		loginResult: auth.AuthResult{
			User: auth.UserDTO{
				ID:    "user-1",
				Email: "name@example.com",
				Name:  "Ink User",
				Role:  "member",
			},
			Token: auth.TokenPair{
				AccessToken:          "access-token",
				RefreshToken:         "refresh-token",
				AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			},
		},
	}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5, 32<<20)

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
	server := NewServer(fakeAuthService{}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5, 32<<20)
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
				Role:  "member",
			},
			Token: auth.TokenPair{
				AccessToken:          "access-token",
				RefreshToken:         "refresh-token",
				AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			},
		},
	}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 1, 32<<20)

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
	server := NewServer(fakeAuthService{}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5, 32<<20)
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

func TestCreateUserRequiresAdminAuthorization(t *testing.T) {
	server := NewServer(
		fakeAuthService{
			createUserResult: auth.UserDTO{
				ID:    "user-2",
				Email: "new-user",
				Name:  "New User",
				Role:  "member",
			},
		},
		fakeWorkspaceService{},
		fakeAIService{},
		fakePrinterService{},
		fakeFeedbackService{},
		fakePluginService{},
		fakeScheduleService{},
		slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		time.Minute,
		5,
		32<<20,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/users",
		bytes.NewBufferString(`{"email":"new-user","name":"New User","password":"demo-password"}`),
	)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestWorkspaceHandlersRequireAuthorization(t *testing.T) {
	server := NewServer(fakeAuthService{}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5, 32<<20)

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/workspace", nil)
	getResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected get workspace to require auth, got %d", getResponse.Code)
	}

	putRequest := httptest.NewRequest(http.MethodPut, "/api/v1/workspace", bytes.NewBufferString(`{}`))
	putResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(putResponse, putRequest)

	if putResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected save workspace to require auth, got %d", putResponse.Code)
	}
}

func TestAIConfigRequiresAuthorization(t *testing.T) {
	server := NewServer(fakeAuthService{}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5, 32<<20)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ai/config", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestListPrintersReturnsDevices(t *testing.T) {
	server := NewServer(
		fakeAuthService{},
		fakeWorkspaceService{},
		fakeAIService{},
		fakePrinterService{
			devices: []workspace.Device{
				{ID: "device-1", Name: "书桌咕咕机", Status: "connected", Note: "默认设备"},
			},
		},
		fakeFeedbackService{},
		fakePluginService{},
		fakeScheduleService{},
		slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		time.Minute,
		5,
		32<<20,
	)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/printers", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

type fakeAuthService struct {
	loginResult       auth.AuthResult
	loginErr          error
	changePasswordErr error
	createUserResult  auth.UserDTO
	createUserErr     error
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
		Role:  "admin",
	}, nil
}

func (f fakeAuthService) ChangePassword(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ auth.ClientMeta,
) error {
	return f.changePasswordErr
}

func (f fakeAuthService) CreateUser(
	_ context.Context,
	_ string,
	_ auth.CreateUserInput,
) (auth.UserDTO, error) {
	return f.createUserResult, f.createUserErr
}

type fakeWorkspaceService struct {
	state workspace.State
	err   error
}

func (f fakeWorkspaceService) GetState(_ context.Context, _ string) (workspace.State, error) {
	return f.state, f.err
}

func (f fakeWorkspaceService) SaveState(
	_ context.Context,
	_ string,
	state workspace.State,
) (workspace.State, error) {
	if f.err != nil {
		return workspace.State{}, f.err
	}
	return state, nil
}

var _ auth.AuthService = fakeAuthService{}
var _ workspace.WorkspaceService = fakeWorkspaceService{}

type fakeAIService struct {
	summary ai.ConfigSummary
	reply   ai.ReplyResult
	err     error
}

func (f fakeAIService) GetConfigSummary(_ context.Context, _ string) (ai.ConfigSummary, error) {
	return f.summary, f.err
}

func (f fakeAIService) UpdateSystemConfig(_ context.Context, _ string, _ ai.UpdateConfigInput) (ai.ConfigSummary, error) {
	return f.summary, f.err
}

func (f fakeAIService) GenerateReply(_ context.Context, _ string, _ ai.ReplyInput) (ai.ReplyResult, error) {
	return f.reply, f.err
}

type fakePrinterService struct {
	devices   []workspace.Device
	printJobs []workspace.PrintJob
	err       error
}

func (f fakePrinterService) ListDevices(_ context.Context, _ string) ([]workspace.Device, error) {
	return f.devices, f.err
}

func (f fakePrinterService) BindDevice(_ context.Context, _ string, _ printer.BindInput) (workspace.Device, error) {
	if f.err != nil {
		return workspace.Device{}, f.err
	}
	if len(f.devices) > 0 {
		return f.devices[0], nil
	}
	return workspace.Device{}, nil
}

func (f fakePrinterService) DeleteDevice(_ context.Context, _ string, _ string) error {
	return f.err
}

func (f fakePrinterService) ListPrintJobs(_ context.Context, _ string) ([]workspace.PrintJob, error) {
	return f.printJobs, f.err
}

func (f fakePrinterService) CreatePrintJob(_ context.Context, _ string, _ printer.CreateJobInput) (workspace.PrintJob, error) {
	if f.err != nil {
		return workspace.PrintJob{}, f.err
	}
	if len(f.printJobs) > 0 {
		return f.printJobs[0], nil
	}
	return workspace.PrintJob{}, nil
}

func (f fakePrinterService) SubmitPrintJob(_ context.Context, _ string, _ string) (workspace.PrintJob, error) {
	if f.err != nil {
		return workspace.PrintJob{}, f.err
	}
	if len(f.printJobs) > 0 {
		return f.printJobs[0], nil
	}
	return workspace.PrintJob{}, nil
}

func (f fakePrinterService) CancelPrintJob(_ context.Context, _ string, _ string) (workspace.PrintJob, error) {
	if f.err != nil {
		return workspace.PrintJob{}, f.err
	}
	if len(f.printJobs) > 0 {
		return f.printJobs[0], nil
	}
	return workspace.PrintJob{}, nil
}

func (f fakePrinterService) UpdatePrintJobDevice(_ context.Context, _ string, _ string, _ printer.UpdateJobDeviceInput) (workspace.PrintJob, error) {
	if f.err != nil {
		return workspace.PrintJob{}, f.err
	}
	if len(f.printJobs) > 0 {
		return f.printJobs[0], nil
	}
	return workspace.PrintJob{}, nil
}

var _ ai.AIService = fakeAIService{}
var _ printer.PrinterService = fakePrinterService{}

type fakeFeedbackService struct {
	err error
}

func (f fakeFeedbackService) Submit(_ context.Context, _ string, _ feedback.SubmitInput) error {
	return f.err
}

var _ FeedbackService = fakeFeedbackService{}

type fakePluginService struct {
	items  []plugins.PluginDetails
	result plugins.PluginDetails
	test   plugins.ValidationResult
	err    error
}

func (f fakePluginService) ListAdminInstallations(_ context.Context, _ string) ([]plugins.PluginDetails, error) {
	return f.items, f.err
}

func (f fakePluginService) UploadPlugin(_ context.Context, _ string, _ string, _ io.Reader) (plugins.PluginDetails, error) {
	return f.result, f.err
}

func (f fakePluginService) DisableInstallation(_ context.Context, _ string, _ string) (plugins.PluginDetails, error) {
	return f.result, f.err
}

func (f fakePluginService) ListUserPlugins(_ context.Context, _ string) ([]plugins.PluginDetails, error) {
	return f.items, f.err
}

func (f fakePluginService) GetUserPlugin(_ context.Context, _ string, _ string) (plugins.PluginDetails, error) {
	return f.result, f.err
}

func (f fakePluginService) SaveBinding(_ context.Context, _ string, _ string, _ plugins.BindingInput) (plugins.PluginDetails, error) {
	return f.result, f.err
}

func (f fakePluginService) TestBinding(_ context.Context, _ string, _ string, _ plugins.BindingInput) (plugins.ValidationResult, error) {
	return f.test, f.err
}

type fakeScheduleService struct {
	items []schedule.ScheduleView
	item  schedule.ScheduleView
	err   error
}

func (f fakeScheduleService) List(_ context.Context, _ string) ([]schedule.ScheduleView, error) {
	return f.items, f.err
}

func (f fakeScheduleService) Create(_ context.Context, _ string, _ schedule.UpsertInput) (schedule.ScheduleView, error) {
	return f.item, f.err
}

func (f fakeScheduleService) Update(_ context.Context, _ string, _ string, _ schedule.UpsertInput) (schedule.ScheduleView, error) {
	return f.item, f.err
}

func (f fakeScheduleService) Toggle(_ context.Context, _ string, _ string) (schedule.ScheduleView, error) {
	return f.item, f.err
}

func (f fakeScheduleService) Delete(_ context.Context, _ string, _ string) error {
	return f.err
}

var _ PluginService = fakePluginService{}
var _ ScheduleService = fakeScheduleService{}

func TestSubmitFeedbackRequiresAuthorization(t *testing.T) {
	server := NewServer(fakeAuthService{}, fakeWorkspaceService{}, fakeAIService{}, fakePrinterService{}, fakeFeedbackService{}, fakePluginService{}, fakeScheduleService{}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), time.Minute, 5, 32<<20)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/print", bytes.NewBufferString(`{"content":"hello"}`))
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestSubmitFeedbackReturnsNoContent(t *testing.T) {
	server := NewServer(
		fakeAuthService{},
		fakeWorkspaceService{},
		fakeAIService{},
		fakePrinterService{},
		fakeFeedbackService{},
		fakePluginService{},
		fakeScheduleService{},
		slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		time.Minute,
		5,
		32<<20,
	)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/print", bytes.NewBufferString(`{"content":"hello"}`))
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
}
