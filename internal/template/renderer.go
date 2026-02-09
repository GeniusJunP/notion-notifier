package template

import (
	"bytes"
	"text/template"

	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
)

type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

func (r *Renderer) RenderSingle(tmpl string, event models.TemplateEvent, minutesBefore int) (string, error) {
	tmpl = config.SanitizeTemplate(tmpl)
	t, err := template.New("message").Parse(tmpl)
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
	t, err := template.New("message").Parse(tmpl)
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
