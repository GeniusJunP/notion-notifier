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
