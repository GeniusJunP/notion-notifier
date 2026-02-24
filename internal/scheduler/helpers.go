package scheduler

import (
	"encoding/json"
	"strings"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
	"notion-notifier/internal/notion"
)

func buildTemplateEvents(events []models.Event, mapping config.PropertyMapping) []models.TemplateEvent {
	var out []models.TemplateEvent
	for _, ev := range events {
		custom := extractCustomValues(ev.RawPropsJSON, mapping)
		out = append(out, toTemplateEvent(ev, custom))
	}
	return out
}

func extractCustomValues(raw string, mapping config.PropertyMapping) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var props map[string]any
	if err := json.Unmarshal([]byte(raw), &props); err != nil {
		return map[string]string{}
	}
	custom := map[string]string{}
	for _, cm := range mapping.Custom {
		custom[cm.Variable] = notion.ExtractString(props[cm.Property])
	}
	return custom
}

func toTemplateEvent(ev models.Event, custom map[string]string) models.TemplateEvent {
	return models.TemplateEvent{
		Name:     ev.Title,
		Date:     ev.StartDate,
		Time:     ev.StartTime,
		EndDate:  ev.EndDate,
		EndTime:  ev.EndTime,
		IsAllDay: ev.IsAllDay,
		Location: ev.Location,
		URL:      ev.URL,
		Content:  ev.Content,
		Custom:   custom,
	}
}

func loadLocationOrLocal(name string) *time.Location {
	if strings.TrimSpace(name) == "" {
		return time.Local
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}
