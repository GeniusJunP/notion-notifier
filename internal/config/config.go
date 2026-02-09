package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Timezone      string             `yaml:"timezone" json:"timezone"`
	Sync          SyncConfig         `yaml:"sync" json:"sync"`
	Notifications Notifications      `yaml:"notifications" json:"notifications"`
	Webhook       WebhookConfig      `yaml:"webhook" json:"webhook"`
	CalendarSync  CalendarSyncConfig `yaml:"calendar_sync" json:"calendar_sync"`
	PropertyMap   PropertyMapping    `yaml:"property_mapping" json:"property_mapping"`
	ContentRules  ContentRules       `yaml:"content_rules" json:"content_rules"`
	SnoozeUntil   string             `yaml:"snooze_until" json:"snooze_until"`
	MuteUntil     string             `yaml:"mute_until" json:"mute_until"`
	Security      SecurityConfig     `yaml:"security" json:"security"`
}

type SyncConfig struct {
	CheckInterval int `yaml:"check_interval" json:"check_interval"`
}

type Notifications struct {
	Advance  []AdvanceNotification   `yaml:"advance" json:"advance"`
	Periodic []PeriodicNotification  `yaml:"periodic" json:"periodic"`
	Weekly   []PeriodicNotification  `yaml:"weekly" json:"-"`
}

type WebhookConfig struct {
	Schedule     WebhookTarget `yaml:"schedule" json:"schedule"`
	Notification WebhookTarget `yaml:"notification" json:"notification"`
}

type WebhookTarget struct {
	ContentType     string `yaml:"content_type" json:"content_type"`
	PayloadTemplate string `yaml:"payload_template" json:"payload_template"`
}

type AdvanceNotification struct {
	Enabled       bool              `yaml:"enabled" json:"enabled"`
	MinutesBefore int               `yaml:"minutes_before" json:"minutes_before"`
	Message       string            `yaml:"message" json:"message"`
	Location      string            `yaml:"location" json:"location"`
	URL           string            `yaml:"url" json:"url"`
	Conditions    AdvanceConditions `yaml:"conditions" json:"conditions"`
}

type AdvanceConditions struct {
	Enabled         bool             `yaml:"enabled" json:"enabled"`
	DaysOfWeek      []int            `yaml:"days_of_week" json:"days_of_week"`
	PropertyFilters []PropertyFilter `yaml:"property_filters" json:"property_filters"`
}

type PropertyFilter struct {
	Property string `yaml:"property" json:"property"`
	Operator string `yaml:"operator" json:"operator"`
	Value    string `yaml:"value" json:"value"`
}

type PeriodicNotification struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	DaysOfWeek []int  `yaml:"days_of_week" json:"days_of_week"`
	Time       string `yaml:"time" json:"time"`
	DaysAhead  int    `yaml:"days_ahead" json:"days_ahead"`
	Message    string `yaml:"message" json:"message"`
}

type CalendarSyncConfig struct {
	Enabled        bool `yaml:"enabled" json:"enabled"`
	IntervalHours  int  `yaml:"interval_hours" json:"interval_hours"`
	LookaheadDays  int  `yaml:"lookahead_days" json:"lookahead_days"`
}

type PropertyMapping struct {
	Title    string          `yaml:"title" json:"title"`
	Date     string          `yaml:"date" json:"date"`
	Location string          `yaml:"location" json:"location"`
	Custom   []CustomMapping `yaml:"custom" json:"custom"`
}

type CustomMapping struct {
	Variable string `yaml:"variable" json:"variable"`
	Property string `yaml:"property" json:"property"`
}

type ContentRules struct {
	StartHeading      string `yaml:"start_heading" json:"start_heading"`
	IncludeStart      bool   `yaml:"include_start_heading" json:"include_start_heading"`
	StopAtNextHeading bool   `yaml:"stop_at_next_heading" json:"stop_at_next_heading"`
	StopAtDelimiter   bool   `yaml:"stop_at_delimiter" json:"stop_at_delimiter"`
}

type SecurityConfig struct {
	BasicAuth BasicAuthConfig `yaml:"basic_auth" json:"basic_auth"`
}

type BasicAuthConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Env struct {
	Notion   NotionEnv   `yaml:"notion" json:"notion"`
	Webhook  WebhookEnv  `yaml:"webhook" json:"webhook"`
	Google   GoogleEnv   `yaml:"google" json:"google"`
	Security SecurityEnv `yaml:"security" json:"security"`
}

type NotionEnv struct {
	APIKey     string `yaml:"api_key" json:"api_key"`
	DatabaseID string `yaml:"database_id" json:"database_id"`
}

type WebhookEnv struct {
	ScheduleURL     string `yaml:"schedule_url" json:"schedule_url"`
	NotificationURL string `yaml:"notification_url" json:"notification_url"`
}

type GoogleEnv struct {
	CalendarID        string `yaml:"calendar_id" json:"calendar_id"`
	ServiceAccountKey string `yaml:"service_account_key" json:"service_account_key"`
}

type SecurityEnv struct {
	BasicAuth BasicAuthEnv `yaml:"basic_auth" json:"basic_auth"`
}

type BasicAuthEnv struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	cfg = NormalizeConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func LoadEnv(path string) (Env, error) {
	var env Env
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return env, nil
		}
		return env, err
	}
	if err := yaml.Unmarshal(data, &env); err != nil {
		return env, err
	}
	return env, nil
}

func ApplyEnvOverrides(env Env) Env {
	env.Notion.APIKey = pickEnv("NOTION_API_KEY", env.Notion.APIKey)
	env.Notion.DatabaseID = pickEnv("NOTION_DATABASE_ID", env.Notion.DatabaseID)
	env.Webhook.ScheduleURL = pickEnv("SCHEDULE_WEBHOOK_URL", env.Webhook.ScheduleURL)
	env.Webhook.NotificationURL = pickEnv("NOTIFICATION_WEBHOOK_URL", env.Webhook.NotificationURL)
	env.Google.CalendarID = pickEnv("GOOGLE_CALENDAR_ID", env.Google.CalendarID)
	env.Google.ServiceAccountKey = pickEnv("GOOGLE_SERVICE_ACCOUNT_KEY", env.Google.ServiceAccountKey)
	env.Security.BasicAuth.Username = pickEnv("BASIC_AUTH_USERNAME", env.Security.BasicAuth.Username)
	env.Security.BasicAuth.Password = pickEnv("BASIC_AUTH_PASSWORD", env.Security.BasicAuth.Password)
	return env
}

func pickEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func ValidateConfig(cfg Config) error {
	if cfg.Timezone == "" {
		return errors.New("timezone is required")
	}
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}
	if cfg.Sync.CheckInterval <= 0 {
		return errors.New("sync.check_interval must be > 0")
	}
	if cfg.CalendarSync.IntervalHours <= 0 {
		return errors.New("calendar_sync.interval_hours must be > 0")
	}
	if cfg.CalendarSync.LookaheadDays <= 0 {
		return errors.New("calendar_sync.lookahead_days must be > 0")
	}
	for i, adv := range cfg.Notifications.Advance {
		if adv.MinutesBefore <= 0 {
			return fmt.Errorf("notifications.advance[%d].minutes_before must be > 0", i)
		}
	}
	for i, periodic := range cfg.Notifications.Periodic {
		if periodic.DaysAhead <= 0 {
			return fmt.Errorf("notifications.periodic[%d].days_ahead must be > 0", i)
		}
		if err := validateHHMM(periodic.Time); err != nil {
			return fmt.Errorf("notifications.periodic[%d].time: %w", i, err)
		}
		for _, d := range periodic.DaysOfWeek {
			if d < 1 || d > 7 {
				return fmt.Errorf("notifications.periodic[%d].days_of_week must be 1-7", i)
			}
		}
	}
	for i, adv := range cfg.Notifications.Advance {
		for _, d := range adv.Conditions.DaysOfWeek {
			if d < 1 || d > 7 {
				return fmt.Errorf("notifications.advance[%d].conditions.days_of_week must be 1-7", i)
			}
		}
	}
	if cfg.SnoozeUntil != "" {
		if _, err := time.Parse(time.RFC3339, cfg.SnoozeUntil); err != nil {
			return fmt.Errorf("snooze_until must be RFC3339: %w", err)
		}
	}
	if cfg.MuteUntil != "" {
		if _, err := time.Parse(time.RFC3339, cfg.MuteUntil); err != nil {
			return fmt.Errorf("mute_until must be RFC3339: %w", err)
		}
	}
	return nil
}

func validateHHMM(value string) error {
	if value == "" {
		return errors.New("time is required")
	}
	_, err := time.Parse("15:04", value)
	return err
}

func IsMuted(cfg Config, now time.Time) bool {
	if cfg.MuteUntil == "" {
		return false
	}
	until, err := time.Parse(time.RFC3339, cfg.MuteUntil)
	if err != nil {
		return false
	}
	return now.Before(until)
}

func IsSnoozed(cfg Config, now time.Time) bool {
	if cfg.SnoozeUntil == "" {
		return false
	}
	until, err := time.Parse(time.RFC3339, cfg.SnoozeUntil)
	if err != nil {
		return false
	}
	return now.Before(until)
}

func WriteConfig(path string, cfg Config) error {
	cfg = NormalizeConfig(cfg)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func NormalizeConfig(cfg Config) Config {
	if len(cfg.Notifications.Periodic) == 0 && len(cfg.Notifications.Weekly) > 0 {
		cfg.Notifications.Periodic = cfg.Notifications.Weekly
	}
	cfg.Notifications.Weekly = nil
	if cfg.Webhook.Schedule.ContentType == "" {
		cfg.Webhook.Schedule.ContentType = "application/json"
	}
	if cfg.Webhook.Notification.ContentType == "" {
		cfg.Webhook.Notification.ContentType = "application/json"
	}
	if cfg.Webhook.Schedule.PayloadTemplate == "" {
		cfg.Webhook.Schedule.PayloadTemplate = `{"content":"{{.Message}}"}`
	}
	if cfg.Webhook.Notification.PayloadTemplate == "" {
		cfg.Webhook.Notification.PayloadTemplate = `{"content":"{{.Message}}"}`
	}
	if cfg.CalendarSync.LookaheadDays <= 0 {
		cfg.CalendarSync.LookaheadDays = 30
	}
	cfg.Webhook.Schedule.PayloadTemplate = SanitizeTemplate(cfg.Webhook.Schedule.PayloadTemplate)
	cfg.Webhook.Notification.PayloadTemplate = SanitizeTemplate(cfg.Webhook.Notification.PayloadTemplate)
	return cfg
}

func SanitizeTemplate(input string) string {
	return strings.ReplaceAll(input, "\r\n", "\n")
}
