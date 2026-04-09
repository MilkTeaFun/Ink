package printer

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/integrations/memobird"
	"github.com/ruhuang/ink/server/internal/workspace"
)

var (
	ErrForbidden     = errors.New("forbidden")
	ErrNotConfigured = errors.New("printer service not configured")
	ErrNotFound      = errors.New("printer resource not found")
	ErrInvalidInput  = errors.New("invalid printer input")
	ErrUnavailable   = errors.New("printer provider unavailable")
)

type Binding struct {
	ID               string
	UserID           string
	Name             string
	Note             string
	DeviceIdentifier string
	ProviderUserID   int
	Status           workspace.DeviceStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Job struct {
	ID                     string
	UserID                 string
	PrinterBindingID       string
	Title                  string
	Source                 string
	Content                string
	Status                 workspace.PrintStatus
	ProviderPrintContentID *int
	ProviderSmartGUID      *string
	ErrorMessage           *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Repository interface {
	ListBindingsByUserID(ctx context.Context, userID string) ([]Binding, error)
	FindBindingByID(ctx context.Context, userID string, bindingID string) (*Binding, error)
	SaveBinding(ctx context.Context, binding Binding) error
	DeleteBinding(ctx context.Context, userID string, bindingID string) error
	ListJobsByUserID(ctx context.Context, userID string) ([]Job, error)
	FindJobByID(ctx context.Context, userID string, jobID string) (*Job, error)
	SaveJob(ctx context.Context, job Job) error
}

type Authenticator interface {
	GetCurrentUser(ctx context.Context, accessToken string) (auth.UserDTO, error)
}

type IDGenerator interface {
	New(prefix string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	repo      Repository
	auth      Authenticator
	ids       IDGenerator
	clock     Clock
	accessKey string
	baseURL   string
	timeout   time.Duration
}

type PrinterService interface {
	ListDevices(ctx context.Context, accessToken string) ([]workspace.Device, error)
	BindDevice(ctx context.Context, accessToken string, input BindInput) (workspace.Device, error)
	DeleteDevice(ctx context.Context, accessToken string, bindingID string) error
	ListPrintJobs(ctx context.Context, accessToken string) ([]workspace.PrintJob, error)
	CreatePrintJob(ctx context.Context, accessToken string, input CreateJobInput) (workspace.PrintJob, error)
	SubmitPrintJob(ctx context.Context, accessToken string, jobID string) (workspace.PrintJob, error)
	CancelPrintJob(ctx context.Context, accessToken string, jobID string) (workspace.PrintJob, error)
	UpdatePrintJobDevice(ctx context.Context, accessToken string, jobID string, input UpdateJobDeviceInput) (workspace.PrintJob, error)
}

type BindInput struct {
	Name     string `json:"name"`
	Note     string `json:"note"`
	DeviceID string `json:"deviceId"`
}

type CreateJobInput struct {
	Title             string `json:"title"`
	Source            string `json:"source"`
	Content           string `json:"content"`
	PrinterBindingID  string `json:"printerBindingId"`
	SubmitImmediately bool   `json:"submitImmediately"`
}

type UpdateJobDeviceInput struct {
	PrinterBindingID string `json:"printerBindingId"`
}

func NewService(
	repo Repository,
	authenticator Authenticator,
	ids IDGenerator,
	clock Clock,
	accessKey string,
	baseURL string,
	timeout time.Duration,
) *Service {
	return &Service{
		repo:      repo,
		auth:      authenticator,
		ids:       ids,
		clock:     clock,
		accessKey: strings.TrimSpace(accessKey),
		baseURL:   strings.TrimSpace(baseURL),
		timeout:   timeout,
	}
}

func (s *Service) ListDevices(ctx context.Context, accessToken string) ([]workspace.Device, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	bindings, err := s.repo.ListBindingsByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	devices := make([]workspace.Device, 0, len(bindings))
	for _, binding := range bindings {
		devices = append(devices, workspace.Device{
			ID:     binding.ID,
			Name:   binding.Name,
			Status: binding.Status,
			Note:   binding.Note,
		})
	}

	return devices, nil
}

func (s *Service) BindDevice(ctx context.Context, accessToken string, input BindInput) (workspace.Device, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return workspace.Device{}, err
	}
	if s.accessKey == "" {
		return workspace.Device{}, ErrNotConfigured
	}

	name := strings.TrimSpace(input.Name)
	deviceID := strings.TrimSpace(input.DeviceID)
	if name == "" || deviceID == "" {
		return workspace.Device{}, ErrInvalidInput
	}

	client := memobird.NewClient(memobird.Config{
		AccessKey: s.accessKey,
		DeviceID:  deviceID,
		BaseURL:   s.baseURL,
		Timeout:   s.timeout,
	})

	userIdentifying := fmt.Sprintf("ink:%s:%s", currentUser.ID, deviceID)
	resp, err := client.BindAndRemember(ctx, userIdentifying)
	if err != nil {
		return workspace.Device{}, fmt.Errorf("%w: %s", ErrUnavailable, err.Error())
	}

	now := s.clock.Now()
	existingBindings, err := s.repo.ListBindingsByUserID(ctx, currentUser.ID)
	if err != nil {
		return workspace.Device{}, err
	}

	bindingID := ""
	createdAt := now
	for _, candidate := range existingBindings {
		if candidate.DeviceIdentifier == deviceID {
			bindingID = candidate.ID
			createdAt = candidate.CreatedAt
			break
		}
	}
	if bindingID == "" {
		bindingID, err = s.ids.New("device")
		if err != nil {
			return workspace.Device{}, err
		}
	}

	binding := Binding{
		ID:               bindingID,
		UserID:           currentUser.ID,
		Name:             name,
		Note:             strings.TrimSpace(input.Note),
		DeviceIdentifier: deviceID,
		ProviderUserID:   resp.UserID,
		Status:           workspace.DeviceStatusConnected,
		CreatedAt:        createdAt,
		UpdatedAt:        now,
	}
	if err := s.repo.SaveBinding(ctx, binding); err != nil {
		return workspace.Device{}, err
	}

	return workspace.Device{
		ID:     binding.ID,
		Name:   binding.Name,
		Status: binding.Status,
		Note:   binding.Note,
	}, nil
}

func (s *Service) DeleteDevice(ctx context.Context, accessToken string, bindingID string) error {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return err
	}
	if strings.TrimSpace(bindingID) == "" {
		return ErrInvalidInput
	}

	binding, err := s.repo.FindBindingByID(ctx, currentUser.ID, bindingID)
	if err != nil {
		return err
	}
	if binding == nil {
		return ErrNotFound
	}
	if binding.Status == workspace.DeviceStatusOffline {
		return nil
	}

	binding.Status = workspace.DeviceStatusOffline
	binding.Note = archivedBindingNote(binding.Note)
	binding.UpdatedAt = s.clock.Now()
	return s.repo.SaveBinding(ctx, *binding)
}

func (s *Service) ListPrintJobs(ctx context.Context, accessToken string) ([]workspace.PrintJob, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	jobs, err := s.repo.ListJobsByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	if s.accessKey != "" {
		for index := range jobs {
			if jobs[index].Status != workspace.PrintStatusQueued || jobs[index].ProviderPrintContentID == nil {
				continue
			}

			binding, findErr := s.repo.FindBindingByID(ctx, currentUser.ID, jobs[index].PrinterBindingID)
			if findErr != nil || binding == nil {
				continue
			}
			client := s.newClient(*binding)
			statusResp, statusErr := client.GetPrintStatus(ctx, *jobs[index].ProviderPrintContentID)
			if statusErr == nil && statusResp.IsPrinted() {
				jobs[index].Status = workspace.PrintStatusCompleted
				jobs[index].UpdatedAt = s.clock.Now()
				_ = s.repo.SaveJob(ctx, jobs[index])
			}
		}
	}

	printJobs := make([]workspace.PrintJob, 0, len(jobs))
	for _, job := range jobs {
		printJobs = append(printJobs, mapJob(job))
	}

	return printJobs, nil
}

func (s *Service) CreatePrintJob(ctx context.Context, accessToken string, input CreateJobInput) (workspace.PrintJob, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return workspace.PrintJob{}, err
	}

	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	source := strings.TrimSpace(input.Source)
	bindingID := strings.TrimSpace(input.PrinterBindingID)
	if title == "" || content == "" || bindingID == "" {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	binding, err := s.repo.FindBindingByID(ctx, currentUser.ID, bindingID)
	if err != nil {
		return workspace.PrintJob{}, err
	}
	if binding == nil {
		return workspace.PrintJob{}, ErrNotFound
	}
	if binding.Status != workspace.DeviceStatusConnected {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	jobID, err := s.ids.New("print")
	if err != nil {
		return workspace.PrintJob{}, err
	}

	now := s.clock.Now()
	job := Job{
		ID:               jobID,
		UserID:           currentUser.ID,
		PrinterBindingID: bindingID,
		Title:            title,
		Source:           chooseString(source, "手动打印"),
		Content:          content,
		Status:           workspace.PrintStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if input.SubmitImmediately && s.accessKey == "" {
		return workspace.PrintJob{}, ErrNotConfigured
	}
	if input.SubmitImmediately {
		job.Status = workspace.PrintStatusQueued
	}

	if err := s.repo.SaveJob(ctx, job); err != nil {
		return workspace.PrintJob{}, err
	}

	if input.SubmitImmediately {
		submitted, submitErr := s.submitJob(ctx, *binding, job)
		if submitErr != nil {
			return workspace.PrintJob{}, submitErr
		}
		return mapJob(submitted), nil
	}

	return mapJob(job), nil
}

func (s *Service) SubmitPrintJob(ctx context.Context, accessToken string, jobID string) (workspace.PrintJob, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return workspace.PrintJob{}, err
	}

	job, err := s.repo.FindJobByID(ctx, currentUser.ID, strings.TrimSpace(jobID))
	if err != nil {
		return workspace.PrintJob{}, err
	}
	if job == nil {
		return workspace.PrintJob{}, ErrNotFound
	}
	if job.Status != workspace.PrintStatusPending {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	binding, err := s.repo.FindBindingByID(ctx, currentUser.ID, job.PrinterBindingID)
	if err != nil {
		return workspace.PrintJob{}, err
	}
	if binding == nil {
		return workspace.PrintJob{}, ErrNotFound
	}
	if binding.Status != workspace.DeviceStatusConnected {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	submitted, err := s.submitJob(ctx, *binding, *job)
	if err != nil {
		return workspace.PrintJob{}, err
	}
	return mapJob(submitted), nil
}

func (s *Service) CancelPrintJob(ctx context.Context, accessToken string, jobID string) (workspace.PrintJob, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return workspace.PrintJob{}, err
	}

	job, err := s.repo.FindJobByID(ctx, currentUser.ID, strings.TrimSpace(jobID))
	if err != nil {
		return workspace.PrintJob{}, err
	}
	if job == nil {
		return workspace.PrintJob{}, ErrNotFound
	}

	if job.Status != workspace.PrintStatusPending {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	job.Status = workspace.PrintStatusCancelled
	job.UpdatedAt = s.clock.Now()
	job.ErrorMessage = nil
	if err := s.repo.SaveJob(ctx, *job); err != nil {
		return workspace.PrintJob{}, err
	}

	return mapJob(*job), nil
}

func (s *Service) UpdatePrintJobDevice(ctx context.Context, accessToken string, jobID string, input UpdateJobDeviceInput) (workspace.PrintJob, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return workspace.PrintJob{}, err
	}

	job, err := s.repo.FindJobByID(ctx, currentUser.ID, strings.TrimSpace(jobID))
	if err != nil {
		return workspace.PrintJob{}, err
	}
	if job == nil {
		return workspace.PrintJob{}, ErrNotFound
	}
	if job.Status != workspace.PrintStatusPending {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	bindingID := strings.TrimSpace(input.PrinterBindingID)
	binding, err := s.repo.FindBindingByID(ctx, currentUser.ID, bindingID)
	if err != nil {
		return workspace.PrintJob{}, err
	}
	if binding == nil {
		return workspace.PrintJob{}, ErrNotFound
	}
	if binding.Status != workspace.DeviceStatusConnected {
		return workspace.PrintJob{}, ErrInvalidInput
	}

	job.PrinterBindingID = bindingID
	job.UpdatedAt = s.clock.Now()
	if err := s.repo.SaveJob(ctx, *job); err != nil {
		return workspace.PrintJob{}, err
	}

	return mapJob(*job), nil
}

func (s *Service) submitJob(ctx context.Context, binding Binding, job Job) (Job, error) {
	if s.accessKey == "" {
		return Job{}, ErrNotConfigured
	}

	client := s.newClient(binding)
	resp, err := client.PrintHTML(ctx, renderPrintHTML(job.Title, job.Content))
	if err != nil {
		message := err.Error()
		job.Status = workspace.PrintStatusFailed
		job.ErrorMessage = &message
		job.UpdatedAt = s.clock.Now()
		_ = s.repo.SaveJob(ctx, job)
		return Job{}, fmt.Errorf("%w: %s", ErrUnavailable, err.Error())
	}

	job.Status = workspace.PrintStatusQueued
	job.UpdatedAt = s.clock.Now()
	job.ErrorMessage = nil
	if resp.PrintContentID != 0 {
		printID := resp.PrintContentID
		job.ProviderPrintContentID = &printID
	}
	if strings.TrimSpace(resp.SmartGuid) != "" {
		guid := strings.TrimSpace(resp.SmartGuid)
		job.ProviderSmartGUID = &guid
	}

	if job.ProviderPrintContentID != nil {
		if statusResp, statusErr := client.GetPrintStatus(ctx, *job.ProviderPrintContentID); statusErr == nil && statusResp.IsPrinted() {
			job.Status = workspace.PrintStatusCompleted
		}
	}

	if err := s.repo.SaveJob(ctx, job); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *Service) newClient(binding Binding) *memobird.Client {
	return memobird.NewClient(memobird.Config{
		AccessKey: s.accessKey,
		DeviceID:  binding.DeviceIdentifier,
		UserID:    binding.ProviderUserID,
		BaseURL:   s.baseURL,
		Timeout:   s.timeout,
	})
}

func renderPrintHTML(title string, content string) string {
	escapedTitle := html.EscapeString(strings.TrimSpace(title))
	escapedContent := strings.ReplaceAll(html.EscapeString(strings.TrimSpace(content)), "\n", "<br>")

	return fmt.Sprintf(
		`<article style="font-family: sans-serif; width: 320px; padding: 12px 8px; color: #111827;"><h1 style="font-size: 22px; margin: 0 0 12px 0;">%s</h1><div style="font-size: 16px; line-height: 1.7; white-space: normal;">%s</div></article>`,
		escapedTitle,
		escapedContent,
	)
}

func mapJob(job Job) workspace.PrintJob {
	return workspace.PrintJob{
		ID:        job.ID,
		Title:     job.Title,
		Source:    job.Source,
		DeviceID:  job.PrinterBindingID,
		Status:    job.Status,
		CreatedAt: job.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: job.UpdatedAt.UTC().Format(time.RFC3339),
		Content:   job.Content,
	}
}

func chooseString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func archivedBindingNote(note string) string {
	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return "已解绑，仅保留历史记录"
	}
	if strings.Contains(trimmed, "已解绑") {
		return trimmed
	}
	return trimmed + " · 已解绑，仅保留历史记录"
}

var _ PrinterService = (*Service)(nil)
