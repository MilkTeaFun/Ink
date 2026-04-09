package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/plugins"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/workspace"
)

type scheduleRepo struct {
	schedules map[string]PrintSchedule
	claimed   []PrintSchedule
}

func newScheduleRepo() *scheduleRepo {
	return &scheduleRepo{
		schedules: map[string]PrintSchedule{},
	}
}

func (r *scheduleRepo) ListByUserID(_ context.Context, userID string) ([]PrintSchedule, error) {
	result := make([]PrintSchedule, 0, len(r.schedules))
	for _, schedule := range r.schedules {
		if schedule.UserID == userID {
			result = append(result, schedule)
		}
	}
	return result, nil
}

func (r *scheduleRepo) FindByID(_ context.Context, userID string, scheduleID string) (*PrintSchedule, error) {
	schedule, exists := r.schedules[scheduleID]
	if !exists || schedule.UserID != userID {
		return nil, nil
	}
	copy := schedule
	return &copy, nil
}

func (r *scheduleRepo) Save(_ context.Context, schedule PrintSchedule) error {
	r.schedules[schedule.ID] = schedule
	return nil
}

func (r *scheduleRepo) Delete(_ context.Context, userID string, scheduleID string) error {
	schedule, exists := r.schedules[scheduleID]
	if exists && schedule.UserID == userID {
		delete(r.schedules, scheduleID)
	}
	return nil
}

func (r *scheduleRepo) ClaimDue(_ context.Context, _ time.Time, _ time.Time, limit int) ([]PrintSchedule, error) {
	if len(r.claimed) > limit {
		return append([]PrintSchedule{}, r.claimed[:limit]...), nil
	}
	return append([]PrintSchedule{}, r.claimed...), nil
}

type fakeScheduleAuth struct{}

func (fakeScheduleAuth) GetCurrentUser(_ context.Context, accessToken string) (auth.UserDTO, error) {
	return auth.UserDTO{
		ID:    "member-user",
		Email: accessToken,
		Name:  "Member",
		Role:  "member",
	}, nil
}

type fakePluginRuntime struct {
	fetchResult plugins.FetchResult
}

func (r fakePluginRuntime) GetInstallation(_ context.Context, installationID string) (plugins.Installation, plugins.Manifest, error) {
	manifest := plugins.Manifest{
		SchemaVersion: 1,
		Kind:          "source",
		PluginKey:     "fixture-source",
		Name:          "Fixture Source",
		Version:       "1.0.0",
		Description:   "fixture",
		Runtime: plugins.RuntimeSpec{
			Type: "node",
		},
		Entrypoints: plugins.Entrypoints{
			Validate: plugins.CommandSpec{Command: []string{"node", "validate.mjs"}},
			Fetch:    plugins.CommandSpec{Command: []string{"node", "fetch.mjs"}},
		},
		ScheduleConfigSchema: []plugins.FieldSpec{
			{
				Key:      "message",
				Label:    "Message",
				Type:     plugins.FieldTypeText,
				Required: true,
			},
		},
	}

	return plugins.Installation{
		ID:          installationID,
		DisplayName: "Fixture Source",
		Status:      plugins.InstallationStatusReady,
		ManifestJSON: []byte(`{
			"schemaVersion": 1,
			"kind": "source",
			"pluginKey": "fixture-source",
			"name": "Fixture Source",
			"version": "1.0.0",
			"description": "fixture",
			"runtime": { "type": "node" },
			"entrypoints": {
				"validate": { "command": ["node", "validate.mjs"] },
				"fetch": { "command": ["node", "fetch.mjs"] }
			},
			"workspaceConfigSchema": [],
			"scheduleConfigSchema": [
				{ "key": "message", "label": "Message", "type": "text", "required": true }
			]
		}`),
	}, manifest, nil
}

func (fakePluginRuntime) GetBindingForUser(_ context.Context, installationID string, userID string) (plugins.Binding, map[string]string, error) {
	return plugins.Binding{
		ID:                   "binding-1",
		PluginInstallationID: installationID,
		UserID:               userID,
		Enabled:              true,
		Status:               plugins.BindingStatusConnected,
		Config: map[string]any{
			"sourceName": "Fixture Source",
		},
	}, map[string]string{}, nil
}

func (r fakePluginRuntime) ExecuteFetch(_ context.Context, _ plugins.Installation, _ plugins.Binding, _ map[string]string, scheduleConfig map[string]any, _ plugins.FetchTrigger) (plugins.FetchResult, error) {
	result := r.fetchResult
	if message, ok := scheduleConfig["message"].(string); ok {
		result.Content = message
	}
	return result, nil
}

type fakePrinterRepo struct{}

func (fakePrinterRepo) FindBindingByID(_ context.Context, userID string, bindingID string) (*printer.Binding, error) {
	return &printer.Binding{
		ID:        bindingID,
		UserID:    userID,
		Name:      "Desk Printer",
		Status:    workspace.DeviceStatusConnected,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type capturedPrinter struct {
	inputs []printer.CreateJobInput
}

func (p *capturedPrinter) CreatePrintJobForUser(_ context.Context, userID string, input printer.CreateJobInput) (workspace.PrintJob, error) {
	p.inputs = append(p.inputs, input)

	status := workspace.PrintStatusPending
	if input.SubmitImmediately {
		status = workspace.PrintStatusQueued
	}

	return workspace.PrintJob{
		ID:        "print-job-1",
		Title:     input.Title,
		Source:    input.Source,
		DeviceID:  input.PrinterBindingID,
		Status:    status,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Content:   input.Content,
	}, nil
}

type fakeWorkspaceRepo struct {
	state *workspace.State
}

func (r fakeWorkspaceRepo) FindByUserID(_ context.Context, _ string) (*workspace.State, error) {
	return r.state, nil
}

type fakeScheduleIDs struct{}

func (fakeScheduleIDs) New(prefix string) (string, error) {
	return prefix + "-1", nil
}

type fakeScheduleClock struct {
	now time.Time
}

func (c fakeScheduleClock) Now() time.Time {
	return c.now
}

func TestCreateAndProcessDueSchedule(t *testing.T) {
	t.Parallel()

	repo := newScheduleRepo()
	printerCreator := &capturedPrinter{}
	now := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	service := NewService(
		repo,
		fakeScheduleAuth{},
		fakePluginRuntime{
			fetchResult: plugins.FetchResult{
				Title:       "Fixture Source Digest",
				Content:     "fallback",
				SourceLabel: "Fixture Source",
			},
		},
		fakePrinterRepo{},
		printerCreator,
		fakeWorkspaceRepo{
			state: &workspace.State{
				Preferences: workspace.Preferences{
					SendConfirmationEnabled: true,
				},
			},
		},
		fakeScheduleIDs{},
		fakeScheduleClock{now: now},
	)

	created, err := service.Create(context.Background(), "member-token", UpsertInput{
		Title:                "Morning Digest",
		PluginInstallationID: "plugin-1",
		FrequencyType:        FrequencyTypeDaily,
		Timezone:             "Asia/Shanghai",
		Hour:                 9,
		Minute:               30,
		Weekdays:             nil,
		ScheduleConfig: map[string]any{
			"message": "hello schedule",
		},
		DeviceID: "device-1",
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if created.TimeLabel != "每天 09:30" {
		t.Fatalf("unexpected time label: %s", created.TimeLabel)
	}
	if created.PluginDisplayName != "Fixture Source" {
		t.Fatalf("unexpected plugin display name: %s", created.PluginDisplayName)
	}

	saved := repo.schedules["schedule-1"]
	saved.NextRunAt = now
	repo.schedules[saved.ID] = saved
	repo.claimed = []PrintSchedule{saved}

	processed, err := service.ProcessDue(context.Background(), 10)
	if err != nil {
		t.Fatalf("process due: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected 1 processed schedule, got %d", processed)
	}
	if len(printerCreator.inputs) != 1 {
		t.Fatalf("expected 1 created print job, got %d", len(printerCreator.inputs))
	}

	input := printerCreator.inputs[0]
	if input.Title != "Fixture Source Digest" {
		t.Fatalf("unexpected print title: %s", input.Title)
	}
	if input.Source != "Fixture Source" {
		t.Fatalf("unexpected print source: %s", input.Source)
	}
	if input.Content != "hello schedule" {
		t.Fatalf("unexpected print content: %s", input.Content)
	}
	if input.SubmitImmediately {
		t.Fatalf("expected pending print job when confirmation is enabled")
	}

	updated := repo.schedules["schedule-1"]
	if updated.LastRunAt == nil {
		t.Fatalf("expected last run at to be populated")
	}
	if !updated.NextRunAt.After(now) {
		t.Fatalf("expected next run to move forward, got %s", updated.NextRunAt)
	}
	if updated.LastError != nil {
		t.Fatalf("expected no last error, got %s", *updated.LastError)
	}
}

func TestNextRunAtWeekly(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC) // Friday morning UTC.
	next, err := NextRunAt(FrequencyTypeWeekly, "Asia/Shanghai", 9, 30, []int{1, 3, 5}, now)
	if err != nil {
		t.Fatalf("next run at: %v", err)
	}

	location, _ := time.LoadLocation("Asia/Shanghai")
	local := next.In(location)
	if local.Weekday() != time.Friday {
		t.Fatalf("expected Friday run, got %s", local.Weekday())
	}
	if local.Hour() != 9 || local.Minute() != 30 {
		t.Fatalf("expected 09:30 local time, got %02d:%02d", local.Hour(), local.Minute())
	}
}
