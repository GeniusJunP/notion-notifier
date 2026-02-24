package config

import (
	"strings"
	"testing"
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
