package schedule

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/plugins"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/workspace"
)

var (
	ErrInvalidInput = errors.New("invalid schedule input")
	ErrNotFound     = errors.New("schedule not found")
)

type Repository interface {
	ListByUserID(ctx context.Context, userID string) ([]PrintSchedule, error)
	FindByID(ctx context.Context, userID string, scheduleID string) (*PrintSchedule, error)
	Save(ctx context.Context, schedule PrintSchedule) error
	Delete(ctx context.Context, userID string, scheduleID string) error
	ClaimDue(ctx context.Context, now time.Time, leaseUntil time.Time, limit int) ([]PrintSchedule, error)
}

type Authenticator interface {
	GetCurrentUser(ctx context.Context, accessToken string) (auth.UserDTO, error)
}

type PluginRuntime interface {
	GetInstallation(ctx context.Context, installationID string) (plugins.Installation, plugins.Manifest, error)
	GetBindingForUser(ctx context.Context, installationID string, userID string) (plugins.Binding, map[string]string, error)
	ExecuteFetch(
		ctx context.Context,
		installation plugins.Installation,
		binding plugins.Binding,
		secrets map[string]string,
		scheduleConfig map[string]any,
		trigger plugins.FetchTrigger,
	) (plugins.FetchResult, error)
}

type PrinterRepository interface {
	FindBindingByID(ctx context.Context, userID string, bindingID string) (*printer.Binding, error)
}

type PrinterJobCreator interface {
	CreatePrintJobForUser(ctx context.Context, userID string, input printer.CreateJobInput) (workspace.PrintJob, error)
}

type WorkspaceRepository interface {
	FindByUserID(ctx context.Context, userID string) (*workspace.State, error)
}

type IDGenerator interface {
	New(prefix string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	repo          Repository
	auth          Authenticator
	plugins       PluginRuntime
	printerRepo   PrinterRepository
	printer       PrinterJobCreator
	workspaceRepo WorkspaceRepository
	ids           IDGenerator
	clock         Clock
}

func NewService(
	repo Repository,
	authenticator Authenticator,
	pluginRuntime PluginRuntime,
	printerRepo PrinterRepository,
	printerCreator PrinterJobCreator,
	workspaceRepo WorkspaceRepository,
	ids IDGenerator,
	clock Clock,
) *Service {
	return &Service{
		repo:          repo,
		auth:          authenticator,
		plugins:       pluginRuntime,
		printerRepo:   printerRepo,
		printer:       printerCreator,
		workspaceRepo: workspaceRepo,
		ids:           ids,
		clock:         clock,
	}
}

func (s *Service) List(ctx context.Context, accessToken string) ([]ScheduleView, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	schedules, err := s.repo.ListByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	return s.mapViews(ctx, schedules)
}

func (s *Service) Create(ctx context.Context, accessToken string, input UpsertInput) (ScheduleView, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return ScheduleView{}, err
	}

	scheduleID, err := s.ids.New("schedule")
	if err != nil {
		return ScheduleView{}, err
	}

	now := s.clock.Now()
	schedule, err := s.prepareSchedule(ctx, PrintSchedule{
		ID:                   scheduleID,
		UserID:               currentUser.ID,
		PluginInstallationID: input.PluginInstallationID,
		Title:                input.Title,
		FrequencyType:        input.FrequencyType,
		Timezone:             input.Timezone,
		Hour:                 input.Hour,
		Minute:               input.Minute,
		Weekdays:             input.Weekdays,
		ScheduleConfig:       input.ScheduleConfig,
		DeviceID:             input.DeviceID,
		Enabled:              input.Enabled,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, now)
	if err != nil {
		return ScheduleView{}, err
	}

	if err := s.repo.Save(ctx, schedule); err != nil {
		return ScheduleView{}, err
	}

	views, err := s.mapViews(ctx, []PrintSchedule{schedule})
	if err != nil {
		return ScheduleView{}, err
	}
	return views[0], nil
}

func (s *Service) Update(ctx context.Context, accessToken string, scheduleID string, input UpsertInput) (ScheduleView, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return ScheduleView{}, err
	}

	existing, err := s.repo.FindByID(ctx, currentUser.ID, scheduleID)
	if err != nil {
		return ScheduleView{}, err
	}
	if existing == nil {
		return ScheduleView{}, ErrNotFound
	}

	now := s.clock.Now()
	existing.PluginInstallationID = input.PluginInstallationID
	existing.Title = input.Title
	existing.FrequencyType = input.FrequencyType
	existing.Timezone = input.Timezone
	existing.Hour = input.Hour
	existing.Minute = input.Minute
	existing.Weekdays = input.Weekdays
	existing.ScheduleConfig = input.ScheduleConfig
	existing.DeviceID = input.DeviceID
	existing.Enabled = input.Enabled
	existing.UpdatedAt = now

	updated, err := s.prepareSchedule(ctx, *existing, now)
	if err != nil {
		return ScheduleView{}, err
	}

	if err := s.repo.Save(ctx, updated); err != nil {
		return ScheduleView{}, err
	}

	views, err := s.mapViews(ctx, []PrintSchedule{updated})
	if err != nil {
		return ScheduleView{}, err
	}
	return views[0], nil
}

func (s *Service) Toggle(ctx context.Context, accessToken string, scheduleID string) (ScheduleView, error) {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return ScheduleView{}, err
	}

	existing, err := s.repo.FindByID(ctx, currentUser.ID, scheduleID)
	if err != nil {
		return ScheduleView{}, err
	}
	if existing == nil {
		return ScheduleView{}, ErrNotFound
	}

	existing.Enabled = !existing.Enabled
	existing.UpdatedAt = s.clock.Now()
	if existing.Enabled {
		nextRun, err := NextRunAt(existing.FrequencyType, existing.Timezone, existing.Hour, existing.Minute, existing.Weekdays, s.clock.Now())
		if err != nil {
			return ScheduleView{}, err
		}
		existing.NextRunAt = nextRun
	}

	if err := s.repo.Save(ctx, *existing); err != nil {
		return ScheduleView{}, err
	}

	views, err := s.mapViews(ctx, []PrintSchedule{*existing})
	if err != nil {
		return ScheduleView{}, err
	}
	return views[0], nil
}

func (s *Service) Delete(ctx context.Context, accessToken string, scheduleID string) error {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, currentUser.ID, scheduleID)
}

func (s *Service) ProcessDue(ctx context.Context, limit int) (int, error) {
	now := s.clock.Now()
	claimed, err := s.repo.ClaimDue(ctx, now, now.Add(2*time.Minute), limit)
	if err != nil {
		return 0, err
	}

	for _, current := range claimed {
		s.processSchedule(ctx, current, now)
	}

	return len(claimed), nil
}

func (s *Service) processSchedule(ctx context.Context, current PrintSchedule, now time.Time) {
	scheduledFor := current.NextRunAt
	nextRun, nextErr := NextRunAt(current.FrequencyType, current.Timezone, current.Hour, current.Minute, current.Weekdays, now)
	if nextErr != nil {
		nextRun = now.Add(24 * time.Hour)
	}
	current.LastRunAt = &now
	current.NextRunAt = nextRun
	current.LeaseUntil = nil
	current.UpdatedAt = now

	installation, manifest, err := s.plugins.GetInstallation(ctx, current.PluginInstallationID)
	if err != nil {
		message := err.Error()
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}
	if installation.Status != plugins.InstallationStatusReady {
		message := "插件当前不可用"
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}

	if _, _, fieldErrs := plugins.NormalizeConfigValues(manifest.ScheduleConfigSchema, current.ScheduleConfig, false); len(fieldErrs) > 0 {
		message := fieldErrs[0].Message
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}

	binding, secrets, err := s.plugins.GetBindingForUser(ctx, current.PluginInstallationID, current.UserID)
	if err != nil {
		message := err.Error()
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}
	if !binding.Enabled || binding.Status != plugins.BindingStatusConnected {
		message := "插件连接未启用"
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}

	result, err := s.plugins.ExecuteFetch(ctx, installation, binding, secrets, current.ScheduleConfig, plugins.FetchTrigger{
		ScheduledFor: scheduledFor.In(mustLoadLocation(current.Timezone)).Format(time.RFC3339),
		TriggeredAt:  now.UTC().Format(time.RFC3339),
		Timezone:     current.Timezone,
	})
	if err != nil {
		message := err.Error()
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}

	state, err := s.workspaceRepo.FindByUserID(ctx, current.UserID)
	sendConfirmation := true
	if err == nil {
		if state == nil {
			defaultState := workspace.EmptyState()
			sendConfirmation = defaultState.Preferences.SendConfirmationEnabled
		} else {
			sendConfirmation = workspace.NormalizeState(*state).Preferences.SendConfirmationEnabled
		}
	}

	if _, err := s.printer.CreatePrintJobForUser(ctx, current.UserID, printer.CreateJobInput{
		Title:             result.Title,
		Source:            chooseSourceLabel(result.SourceLabel, installation.DisplayName),
		Content:           result.Content,
		PrinterBindingID:  current.DeviceID,
		SubmitImmediately: !sendConfirmation,
	}); err != nil {
		message := err.Error()
		current.LastError = &message
		_ = s.repo.Save(ctx, current)
		return
	}

	current.LastError = nil
	_ = s.repo.Save(ctx, current)
}

func (s *Service) prepareSchedule(ctx context.Context, current PrintSchedule, now time.Time) (PrintSchedule, error) {
	title := strings.TrimSpace(current.Title)
	if title == "" {
		return PrintSchedule{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	installation, manifest, err := s.plugins.GetInstallation(ctx, strings.TrimSpace(current.PluginInstallationID))
	if err != nil {
		return PrintSchedule{}, err
	}
	if installation.Status != plugins.InstallationStatusReady {
		return PrintSchedule{}, fmt.Errorf("%w: plugin is not ready", ErrInvalidInput)
	}

	binding, _, err := s.plugins.GetBindingForUser(ctx, installation.ID, current.UserID)
	if err != nil {
		return PrintSchedule{}, fmt.Errorf("%w: plugin binding is required", ErrInvalidInput)
	}
	if !binding.Enabled || binding.Status != plugins.BindingStatusConnected {
		return PrintSchedule{}, fmt.Errorf("%w: plugin binding must be enabled", ErrInvalidInput)
	}

	device, err := s.printerRepo.FindBindingByID(ctx, current.UserID, strings.TrimSpace(current.DeviceID))
	if err != nil {
		return PrintSchedule{}, err
	}
	if device == nil || device.Status != workspace.DeviceStatusConnected {
		return PrintSchedule{}, fmt.Errorf("%w: device must be connected", ErrInvalidInput)
	}

	normalizedConfig, _, fieldErrs := plugins.NormalizeConfigValues(manifest.ScheduleConfigSchema, current.ScheduleConfig, false)
	if len(fieldErrs) > 0 {
		return PrintSchedule{}, plugins.ValidationFailure{Errors: fieldErrs}
	}

	weekdays, err := normalizeWeekdays(current.FrequencyType, current.Weekdays)
	if err != nil {
		return PrintSchedule{}, err
	}

	nextRunAt, err := NextRunAt(current.FrequencyType, current.Timezone, current.Hour, current.Minute, weekdays, now)
	if err != nil {
		return PrintSchedule{}, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	current.Title = title
	current.PluginInstallationID = installation.ID
	current.PluginBindingID = binding.ID
	current.ScheduleConfig = normalizedConfig
	current.DeviceID = device.ID
	current.Weekdays = weekdays
	current.NextRunAt = nextRunAt
	current.LeaseUntil = nil

	return current, nil
}

func (s *Service) mapViews(ctx context.Context, schedules []PrintSchedule) ([]ScheduleView, error) {
	installations := map[string]plugins.Installation{}
	views := make([]ScheduleView, 0, len(schedules))

	for _, current := range schedules {
		installation, exists := installations[current.PluginInstallationID]
		if !exists {
			loaded, _, err := s.plugins.GetInstallation(ctx, current.PluginInstallationID)
			if err != nil {
				return nil, err
			}
			installations[current.PluginInstallationID] = loaded
			installation = loaded
		}

		view := ScheduleView{
			ID:                   current.ID,
			Title:                current.Title,
			PluginInstallationID: current.PluginInstallationID,
			PluginBindingID:      current.PluginBindingID,
			PluginDisplayName:    installation.DisplayName,
			FrequencyType:        current.FrequencyType,
			Timezone:             current.Timezone,
			Hour:                 current.Hour,
			Minute:               current.Minute,
			Weekdays:             append([]int{}, current.Weekdays...),
			ScheduleConfig:       cloneMap(current.ScheduleConfig),
			DeviceID:             current.DeviceID,
			Enabled:              current.Enabled,
			NextRunAt:            &current.NextRunAt,
			LastRunAt:            current.LastRunAt,
			TimeLabel:            FormatTimeLabel(current.FrequencyType, current.Hour, current.Minute, current.Weekdays),
			SourceLabel:          installation.DisplayName,
		}
		if current.LastError != nil {
			view.LastError = *current.LastError
		}
		views = append(views, view)
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].Title < views[j].Title
	})

	return views, nil
}

func NextRunAt(frequency FrequencyType, timezone string, hour int, minute int, weekdays []int, now time.Time) (time.Time, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("time must be within 00:00-23:59")
	}

	location, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		return time.Time{}, fmt.Errorf("timezone is invalid")
	}

	current := now.In(location)
	switch frequency {
	case FrequencyTypeDaily:
		candidate := time.Date(current.Year(), current.Month(), current.Day(), hour, minute, 0, 0, location)
		if !candidate.After(current) {
			candidate = candidate.AddDate(0, 0, 1)
		}
		return candidate.UTC(), nil
	case FrequencyTypeWeekly:
		normalized, err := normalizeWeekdays(frequency, weekdays)
		if err != nil {
			return time.Time{}, err
		}
		for offset := 0; offset < 14; offset += 1 {
			day := current.AddDate(0, 0, offset)
			if !containsWeekday(normalized, int(day.Weekday())) {
				continue
			}
			candidate := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, location)
			if candidate.After(current) {
				return candidate.UTC(), nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to resolve next weekly run")
	default:
		return time.Time{}, fmt.Errorf("frequencyType must be daily or weekly")
	}
}

func FormatTimeLabel(frequency FrequencyType, hour int, minute int, weekdays []int) string {
	timeLabel := fmt.Sprintf("%02d:%02d", hour, minute)
	switch frequency {
	case FrequencyTypeDaily:
		return fmt.Sprintf("每天 %s", timeLabel)
	case FrequencyTypeWeekly:
		labels := make([]string, 0, len(weekdays))
		for _, weekday := range weekdays {
			labels = append(labels, weekdayLabel(weekday))
		}
		return fmt.Sprintf("每周%s %s", strings.Join(labels, "、"), timeLabel)
	default:
		return timeLabel
	}
}

func normalizeWeekdays(frequency FrequencyType, weekdays []int) ([]int, error) {
	if frequency == FrequencyTypeDaily {
		return []int{}, nil
	}

	seen := map[int]struct{}{}
	normalized := make([]int, 0, len(weekdays))
	for _, weekday := range weekdays {
		if weekday < 0 || weekday > 6 {
			return nil, fmt.Errorf("%w: weekdays must be within 0-6", ErrInvalidInput)
		}
		if _, exists := seen[weekday]; exists {
			continue
		}
		seen[weekday] = struct{}{}
		normalized = append(normalized, weekday)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("%w: weekly schedules require at least one weekday", ErrInvalidInput)
	}
	sort.Ints(normalized)
	return normalized, nil
}

func chooseSourceLabel(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(fallback)
}

func containsWeekday(weekdays []int, weekday int) bool {
	for _, candidate := range weekdays {
		if candidate == weekday {
			return true
		}
	}
	return false
}

func weekdayLabel(weekday int) string {
	switch weekday {
	case 0:
		return "日"
	case 1:
		return "一"
	case 2:
		return "二"
	case 3:
		return "三"
	case 4:
		return "四"
	case 5:
		return "五"
	case 6:
		return "六"
	default:
		return "?"
	}
}

func mustLoadLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return location
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
