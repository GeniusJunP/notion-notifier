package config

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultTemplatesContainExpectedTokens(t *testing.T) {
	templates := DefaultTemplates()

	upcoming, ok := templates["upcoming"]
	if !ok {
		t.Fatalf("upcoming template is missing")
	}
	periodic, ok := templates["periodic"]
	if !ok {
		t.Fatalf("periodic template is missing")
	}
	manual, ok := templates["manual"]
	if !ok {
		t.Fatalf("manual template is missing")
	}

	upcomingTokens := []string{
		"## 予定リマインド！⏰",
		"{{.MinutesBefore}}",
		"{{.Name}}",
	}
	for _, token := range upcomingTokens {
		if !strings.Contains(upcoming, token) {
			t.Fatalf("upcoming template must contain %q", token)
		}
	}

	periodicTokens := []string{
		"## 今週の予定！📣",
		"{{len .Events}}",
		"{{range .Events}}",
		"### {{.Name}}",
	}
	for _, token := range periodicTokens {
		if !strings.Contains(periodic, token) {
			t.Fatalf("periodic template must contain %q", token)
		}
	}
	if manual != periodic {
		t.Fatalf("manual template must match periodic template by default")
	}
}

func TestApplyEnvOverridesBasicAuthEnabled(t *testing.T) {
	t.Setenv("BASIC_AUTH_ENABLED", "true")
	env := ApplyEnvOverrides(Env{})
	if !env.Security.BasicAuth.Enabled {
		t.Fatalf("BASIC_AUTH_ENABLED=true must enable basic auth")
	}

	t.Setenv("BASIC_AUTH_ENABLED", "false")
	env = ApplyEnvOverrides(Env{
		Security: SecurityEnv{
			BasicAuth: BasicAuthEnv{Enabled: true},
		},
	})
	if env.Security.BasicAuth.Enabled {
		t.Fatalf("BASIC_AUTH_ENABLED=false must disable basic auth")
	}
}

func TestApplyEnvOverridesAppPort(t *testing.T) {
	t.Setenv("APP_PORT", "19090")
	env := ApplyEnvOverrides(Env{})
	if env.Server.Port != 19090 {
		t.Fatalf("APP_PORT must override server port: got=%d want=%d", env.Server.Port, 19090)
	}
}

func TestApplyEnvOverridesTLSFiles(t *testing.T) {
	t.Setenv("APP_TLS_CERT_FILE", "/tmp/cert.pem")
	t.Setenv("APP_TLS_KEY_FILE", "/tmp/key.pem")
	env := ApplyEnvOverrides(Env{})
	if env.Server.TLS.CertFile != "/tmp/cert.pem" {
		t.Fatalf("APP_TLS_CERT_FILE must override cert path")
	}
	if env.Server.TLS.KeyFile != "/tmp/key.pem" {
		t.Fatalf("APP_TLS_KEY_FILE must override key path")
	}
}

func TestApplyEnvOverridesWebhookURLs(t *testing.T) {
	t.Setenv("NOTIFICATION_WEBHOOK_URL", "https://example.com/notification")
	t.Setenv("INTERNAL_NOTIFICATION_WEBHOOK_URL", "https://example.com/internal")

	env := ApplyEnvOverrides(Env{})
	if env.Webhook.NotificationURL != "https://example.com/notification" {
		t.Fatalf("NOTIFICATION_WEBHOOK_URL must override webhook notification url")
	}
	if env.Webhook.InternalNotificationURL != "https://example.com/internal" {
		t.Fatalf("INTERNAL_NOTIFICATION_WEBHOOK_URL must override webhook internal notification url")
	}
}

func TestApplyEnvOverridesGoogleServiceAccount(t *testing.T) {
	t.Setenv("GOOGLE_CALENDAR_ID", "calendar@example.com")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_KEY_FILE", "/run/secrets/google-sa.json")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_KEY_JSON", `{"client_email":"sa@example.com"}`)

	env := ApplyEnvOverrides(Env{})
	if env.Google.CalendarID != "calendar@example.com" {
		t.Fatalf("GOOGLE_CALENDAR_ID must override calendar id")
	}
	if env.Google.ServiceAccountKeyFile != "/run/secrets/google-sa.json" {
		t.Fatalf("GOOGLE_SERVICE_ACCOUNT_KEY_FILE must override service account key file")
	}
	if env.Google.ServiceAccountKeyJSON != `{"client_email":"sa@example.com"}` {
		t.Fatalf("GOOGLE_SERVICE_ACCOUNT_KEY_JSON must override service account key json")
	}
}

func TestNormalizeConfigSnoozeDefaultsAndNormalizesUntil(t *testing.T) {
	cfg := NormalizeConfig(Config{
		Timezone:     "Asia/Tokyo",
		Sync:         SyncConfig{CheckInterval: 15},
		CalendarSync: CalendarSyncConfig{IntervalHours: 6, LookaheadDays: 30},
		Snooze: SnoozeConfig{
			Until: "2026-05-04T10:30",
		},
	})

	if !cfg.Snooze.MuteUpcoming || !cfg.Snooze.MutePeriodic {
		t.Fatalf("snooze defaults must mute upcoming and periodic")
	}
	if cfg.Snooze.Until != "2026-05-04T10:30:00+09:00" {
		t.Fatalf("unexpected normalized snooze until: %q", cfg.Snooze.Until)
	}
}

func TestIsSnoozedByNotificationType(t *testing.T) {
	now := time.Date(2026, 5, 4, 9, 0, 0, 0, time.UTC)
	cfg := Config{
		Timezone: "UTC",
		Snooze: SnoozeConfig{
			Until:        "2026-05-04T10:00:00Z",
			MuteUpcoming: true,
			MutePeriodic: false,
		},
	}

	if !IsSnoozed(cfg, "upcoming", now) {
		t.Fatalf("upcoming must be snoozed")
	}
	if IsSnoozed(cfg, "periodic", now) {
		t.Fatalf("periodic must not be snoozed when mute_periodic is false")
	}
	if IsSnoozed(cfg, "manual", now) {
		t.Fatalf("manual must never be snoozed")
	}
}
