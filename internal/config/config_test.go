package config

import (
	"strings"
	"testing"
)

func TestDefaultTemplatesContainExpectedTokens(t *testing.T) {
	templates := DefaultTemplates()

	advance, ok := templates["advance"]
	if !ok {
		t.Fatalf("advance template is missing")
	}
	periodic, ok := templates["periodic"]
	if !ok {
		t.Fatalf("periodic template is missing")
	}

	advanceTokens := []string{
		"## 予定リマインド！⏰",
		"{{.MinutesBefore}}",
		"{{.Name}}",
	}
	for _, token := range advanceTokens {
		if !strings.Contains(advance, token) {
			t.Fatalf("advance template must contain %q", token)
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
