package template

import (
	"bytes"
	"encoding/json"
	"text/template"

	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
)

type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

func newTemplate(name string) *template.Template {
	return template.New(name).Funcs(template.FuncMap{
		"json": func(v any) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
	})
}

func (r *Renderer) RenderSingle(tmpl string, event models.TemplateEvent, minutesBefore int) (string, error) {
	tmpl = config.SanitizeTemplate(tmpl)
	t, err := newTemplate("message").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	ctx := models.TemplateContext{
		Events:        []models.TemplateEvent{event},
		MinutesBefore: minutesBefore,
	}
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *Renderer) RenderList(tmpl string, events []models.TemplateEvent) (string, error) {
	tmpl = config.SanitizeTemplate(tmpl)
	t, err := newTemplate("message").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	ctx := models.TemplateContext{Events: events}
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *Renderer) RenderPayload(tmpl string, ctx any) (string, error) {
	tmpl = config.SanitizeTemplate(tmpl)
	t, err := newTemplate("payload").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}
