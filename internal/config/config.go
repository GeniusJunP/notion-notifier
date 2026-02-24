package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
}

type SyncConfig struct {
	CheckInterval int `yaml:"check_interval" json:"check_interval"`
}

type Notifications struct {
	Upcoming  []UpcomingNotification  `yaml:"upcoming" json:"upcoming"`
	Periodic []PeriodicNotification `yaml:"periodic" json:"periodic"`
	Manual   string                 `yaml:"manual" json:"manual"`
}

type WebhookConfig struct {
	IsTest               bool          `yaml:"is_test" json:"is_test"`
	Notification         WebhookTarget `yaml:"notification" json:"notification"`
	InternalNotification WebhookTarget `yaml:"internal_notification" json:"internal_notification"`
}

type WebhookTarget struct {
	ContentType     string `yaml:"content_type" json:"content_type"`
	PayloadTemplate string `yaml:"payload_template" json:"payload_template"`
}

type UpcomingNotification struct {
	Enabled       bool              `yaml:"enabled" json:"enabled"`
	MinutesBefore int               `yaml:"minutes_before" json:"minutes_before"`
	Message       string            `yaml:"message" json:"message"`
	Conditions    UpcomingConditions `yaml:"conditions" json:"conditions"`
}

type UpcomingConditions struct {
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
	Enabled       bool `yaml:"enabled" json:"enabled"`
	IntervalHours int  `yaml:"interval_hours" json:"interval_hours"`
	LookaheadDays int  `yaml:"lookahead_days" json:"lookahead_days"`
}

type PropertyMapping struct {
	Title            string          `yaml:"title" json:"title"`
	Date             string          `yaml:"date" json:"date"`
	Location         string          `yaml:"location" json:"location"`
	Attendees        string          `yaml:"attendees" json:"attendees"`                 // Notion people property for Calendar attendee emails
	AttendeesEnabled bool            `yaml:"attendees_enabled" json:"attendees_enabled"` // Enable attendee email sync
	Custom           []CustomMapping `yaml:"custom" json:"custom"`
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
	DelimiterText     string `yaml:"delimiter_text" json:"delimiter_text"`
}

type Env struct {
	Notion   NotionEnv   `yaml:"notion" json:"notion"`
	Webhook  WebhookEnv  `yaml:"webhook" json:"webhook"`
	Google   GoogleEnv   `yaml:"google" json:"google"`
	Server   ServerEnv   `yaml:"server" json:"server"`
	Security SecurityEnv `yaml:"security" json:"security"`
}

type NotionEnv struct {
	APIKey     string `yaml:"api_key" json:"api_key"`
	DatabaseID string `yaml:"database_id" json:"database_id"`
}

type WebhookEnv struct {
	NotificationURL         string `yaml:"notification_url" json:"notification_url"`
	InternalNotificationURL string `yaml:"internal_notification_url" json:"internal_notification_url"`
}

type GoogleEnv struct {
	CalendarID        string `yaml:"calendar_id" json:"calendar_id"`
	OAuthClientID     string `yaml:"oauth_client_id" json:"oauth_client_id"`
	OAuthClientSecret string `yaml:"oauth_client_secret" json:"oauth_client_secret"`
	OAuthRefreshToken string `yaml:"oauth_refresh_token" json:"oauth_refresh_token"`
}

type ServerEnv struct {
	Port int       `yaml:"port" json:"port"`
	TLS  TLSServer `yaml:"tls" json:"tls"`
}

type TLSServer struct {
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

type SecurityEnv struct {
	BasicAuth BasicAuthEnv `yaml:"basic_auth" json:"basic_auth"`
}

type BasicAuthEnv struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
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
	env.Webhook.NotificationURL = pickEnv("NOTIFICATION_WEBHOOK_URL", env.Webhook.NotificationURL)
	env.Webhook.InternalNotificationURL = pickEnv("INTERNAL_NOTIFICATION_WEBHOOK_URL", env.Webhook.InternalNotificationURL)
	env.Google.CalendarID = pickEnv("GOOGLE_CALENDAR_ID", env.Google.CalendarID)
	env.Google.OAuthClientID = pickEnv("GOOGLE_OAUTH_CLIENT_ID", env.Google.OAuthClientID)
	env.Google.OAuthClientSecret = pickEnv("GOOGLE_OAUTH_CLIENT_SECRET", env.Google.OAuthClientSecret)
	env.Google.OAuthRefreshToken = pickEnv("GOOGLE_OAUTH_REFRESH_TOKEN", env.Google.OAuthRefreshToken)
	env.Server.Port = pickEnvInt("APP_PORT", env.Server.Port)
	env.Server.TLS.CertFile = pickEnv("APP_TLS_CERT_FILE", env.Server.TLS.CertFile)
	env.Server.TLS.KeyFile = pickEnv("APP_TLS_KEY_FILE", env.Server.TLS.KeyFile)
	env.Security.BasicAuth.Enabled = pickEnvBool("BASIC_AUTH_ENABLED", env.Security.BasicAuth.Enabled)
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

func pickEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func pickEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
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
	for i, adv := range cfg.Notifications.Upcoming {
		if adv.MinutesBefore <= 0 {
			return fmt.Errorf("notifications.upcoming[%d].minutes_before must be > 0", i)
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
	for i, adv := range cfg.Notifications.Upcoming {
		for _, d := range adv.Conditions.DaysOfWeek {
			if d < 1 || d > 7 {
				return fmt.Errorf("notifications.upcoming[%d].conditions.days_of_week must be 1-7", i)
			}
		}
	}
	if cfg.SnoozeUntil != "" {
		if _, err := time.Parse(time.RFC3339, cfg.SnoozeUntil); err != nil {
			return fmt.Errorf("snooze_until must be RFC3339: %w", err)
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
	if cfg.Notifications.Manual == "" {
		cfg.Notifications.Manual = DefaultManualMessage
	}
	cfg.Notifications.Manual = SanitizeTemplate(cfg.Notifications.Manual)

	// Timezone default
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Tokyo"
	}

	// Sync defaults
	if cfg.Sync.CheckInterval <= 0 {
		cfg.Sync.CheckInterval = 15
	}

	// Calendar defaults
	if cfg.CalendarSync.IntervalHours <= 0 {
		cfg.CalendarSync.IntervalHours = 6
	}

	// Webhook defaults
	if cfg.Webhook.Notification.ContentType == "" {
		cfg.Webhook.Notification.ContentType = "application/json"
	}
	if cfg.Webhook.InternalNotification.ContentType == "" {
		cfg.Webhook.InternalNotification.ContentType = "application/json"
	}
	defaultPayload := `{"content":{{json .Message}}}`
	if cfg.Webhook.Notification.PayloadTemplate == "" {
		cfg.Webhook.Notification.PayloadTemplate = defaultPayload
	}
	if cfg.Webhook.InternalNotification.PayloadTemplate == "" {
		cfg.Webhook.InternalNotification.PayloadTemplate = defaultPayload
	}
	if cfg.CalendarSync.LookaheadDays <= 0 {
		cfg.CalendarSync.LookaheadDays = 30
	}
	cfg.Webhook.Notification.PayloadTemplate = SanitizeTemplate(cfg.Webhook.Notification.PayloadTemplate)
	cfg.Webhook.InternalNotification.PayloadTemplate = SanitizeTemplate(cfg.Webhook.InternalNotification.PayloadTemplate)
	return cfg
}

func SanitizeTemplate(input string) string {
	return strings.ReplaceAll(input, "\r\n", "\n")
}

// DefaultUpcomingMessage is the default message template for upcoming notifications.
const DefaultUpcomingMessage = "## 予定リマインド！⏰\n" +
	"@everyone **{{.Name}}** が **{{.MinutesBefore}}分後** に始まります！\n\n" +
	"### 詳細\n" +
	"- **日時:** {{.Date}} {{if .IsAllDay}}(終日){{else}}`{{.Time}}`{{end}}{{if .EndDate}} 〜 {{.EndDate}} {{if .EndTime}}`{{.EndTime}}`{{end}}{{end}}\n" +
	"{{if .Location}}- **場所:** {{.Location}}{{end}}\n" +
	"{{if .URL}}- **詳細:** {{.URL}}{{end}}\n" +
	"{{with .Content}}- **メモ:** {{.}}{{end}}"

// DefaultPeriodicMessage is the default message template for periodic notifications.
const DefaultPeriodicMessage = "{{if .Events}}\n" +
	"## 今週の予定！📣\n" +
	"@everyone **今週は {{len .Events}} 件** あります！\n\n" +
	"{{range .Events}}\n" +
	"### {{.Name}}\n" +
	"- **日時:** {{.Date}} {{if .IsAllDay}}(終日){{else}}`{{.Time}}`{{end}}{{if .EndDate}} 〜 {{.EndDate}} {{if .EndTime}}`{{.EndTime}}`{{end}}{{end}}\n" +
	"{{if .Location}}- **場所:** {{.Location}}{{end}}\n" +
	"{{if .URL}}- **詳細:** {{.URL}}{{end}}\n" +
	"{{with .Content}}- **メモ:** {{.}}{{end}}\n\n" +
	"{{end}}\n" +
	"{{else}}\n" +
	"## 今週の予定！📣\n" +
	"@everyone 今週の予定はありません！\n" +
	"{{end}}"

// DefaultManualMessage is the default message template for manual notifications.
const DefaultManualMessage = DefaultPeriodicMessage

// DefaultTemplates returns the default message templates keyed by type.
func DefaultTemplates() map[string]string {
	return map[string]string{
		"upcoming": DefaultUpcomingMessage,
		"periodic": DefaultPeriodicMessage,
		"manual":   DefaultManualMessage,
	}
}
