package schedule

import "time"

type FrequencyType string

const (
	FrequencyTypeDaily  FrequencyType = "daily"
	FrequencyTypeWeekly FrequencyType = "weekly"
)

type PrintSchedule struct {
	ID                   string
	UserID               string
	PluginInstallationID string
	PluginBindingID      string
	Title                string
	FrequencyType        FrequencyType
	Timezone             string
	Hour                 int
	Minute               int
	Weekdays             []int
	ScheduleConfig       map[string]any
	DeviceID             string
	Enabled              bool
	NextRunAt            time.Time
	LastRunAt            *time.Time
	LeaseUntil           *time.Time
	LastError            *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UpsertInput struct {
	Title                string         `json:"title"`
	PluginInstallationID string         `json:"pluginInstallationId"`
	FrequencyType        FrequencyType  `json:"frequencyType"`
	Timezone             string         `json:"timezone"`
	Hour                 int            `json:"hour"`
	Minute               int            `json:"minute"`
	Weekdays             []int          `json:"weekdays"`
	ScheduleConfig       map[string]any `json:"scheduleConfig"`
	DeviceID             string         `json:"deviceId"`
	Enabled              bool           `json:"enabled"`
}

type ScheduleView struct {
	ID                   string         `json:"id"`
	Title                string         `json:"title"`
	PluginInstallationID string         `json:"pluginInstallationId"`
	PluginBindingID      string         `json:"pluginBindingId"`
	PluginDisplayName    string         `json:"pluginDisplayName"`
	FrequencyType        FrequencyType  `json:"frequencyType"`
	Timezone             string         `json:"timezone"`
	Hour                 int            `json:"hour"`
	Minute               int            `json:"minute"`
	Weekdays             []int          `json:"weekdays"`
	ScheduleConfig       map[string]any `json:"scheduleConfig"`
	DeviceID             string         `json:"deviceId"`
	Enabled              bool           `json:"enabled"`
	NextRunAt            *time.Time     `json:"nextRunAt,omitempty"`
	LastRunAt            *time.Time     `json:"lastRunAt,omitempty"`
	LastError            string         `json:"lastError,omitempty"`
	TimeLabel            string         `json:"timeLabel"`
	SourceLabel          string         `json:"sourceLabel"`
}
