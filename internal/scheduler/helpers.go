package scheduler

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"notion-notifier/internal/calendar"
	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
	"notion-notifier/internal/notion"
)

func groupCalendarEvents(events []calendar.CalendarEvent) map[string][]calendar.CalendarEvent {
	grouped := make(map[string][]calendar.CalendarEvent, len(events))
	for _, ev := range events {
		grouped[ev.NotionPageID] = append(grouped[ev.NotionPageID], ev)
	}
	return grouped
}

func pickPrimaryCalendarEvent(events []calendar.CalendarEvent, record models.SyncRecord, hasRecord bool) (calendar.CalendarEvent, []calendar.CalendarEvent) {
	if len(events) == 0 {
		return calendar.CalendarEvent{}, nil
	}
	primaryIndex := 0
	if hasRecord && record.CalendarEventID != "" {
		for i, ev := range events {
			if ev.ID == record.CalendarEventID {
				primaryIndex = i
				break
			}
		}
	} else {
		latest, _ := time.Parse(time.RFC3339, events[0].Updated)
		for i := 1; i < len(events); i++ {
			updated, _ := time.Parse(time.RFC3339, events[i].Updated)
			if updated.After(latest) {
				latest = updated
				primaryIndex = i
			}
		}
	}
	primary := events[primaryIndex]
	duplicates := make([]calendar.CalendarEvent, 0, len(events)-1)
	for i, ev := range events {
		if i == primaryIndex {
			continue
		}
		duplicates = append(duplicates, ev)
	}
	return primary, duplicates
}

func buildUpcomingSchedules(events []models.Event, cfg config.Config, now time.Time, loc *time.Location) []models.UpcomingSchedule {
	var schedules []models.UpcomingSchedule
	for _, ev := range events {
		startTime := parseEventStart(ev, loc)
		for idx, rule := range cfg.Notifications.Upcoming {
			if !rule.Enabled {
				continue
			}
			if !matchUpcomingConditions(ev, startTime, rule, cfg) {
				continue
			}
			fireAt := startTime.Add(-time.Duration(rule.MinutesBefore) * time.Minute)
			if fireAt.After(startTime) {
				continue
			}
			if fireAt.Before(now.Add(-5 * time.Minute)) {
				continue
			}
			schedules = append(schedules, models.UpcomingSchedule{
				NotionPageID: ev.NotionPageID,
				RuleIndex:    idx,
				FireAt:       fireAt,
			})
		}
	}
	return schedules
}

func parseEventStart(ev models.Event, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	if ev.StartDate == "" {
		return time.Now().In(loc)
	}
	if ev.StartTime == "" {
		t, err := time.ParseInLocation("2006-01-02", ev.StartDate, loc)
		if err != nil {
			return time.Now().In(loc)
		}
		return t
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", ev.StartDate+" "+ev.StartTime, loc)
	if err != nil {
		return time.Now().In(loc)
	}
	return t
}

func notionOnOrAfterDate(now time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	localNow := now.In(loc)
	localMidnight := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
	// Notion date-only filter is effectively compared at UTC day granularity.
	// Convert local day start to UTC day so early local-time events are included.
	return localMidnight.UTC().Format("2006-01-02")
}

func matchUpcomingConditions(ev models.Event, start time.Time, rule config.UpcomingNotification, cfg config.Config) bool {
	if !matchesDays(rule.Conditions.DaysOfWeek, weekdayToConfig(start.Weekday())) {
		return false
	}
	if len(rule.Conditions.PropertyFilters) == 0 {
		return true
	}
	values := buildFilterValues(ev, cfg)
	for _, filter := range rule.Conditions.PropertyFilters {
		val := values[filter.Property]
		if !matchFilter(val, filter.Operator, filter.Value) {
			return false
		}
	}
	return true
}

func buildFilterValues(ev models.Event, cfg config.Config) map[string]string {
	values := map[string]string{
		"title":    ev.Title,
		"location": ev.Location,
	}
	custom := extractCustomValues(ev.RawPropsJSON, cfg.PropertyMap)
	for k, v := range custom {
		values[k] = v
	}
	return values
}

func matchFilter(value, operator, expected string) bool {
	switch strings.ToLower(operator) {
	case "eq", "equals", "=":
		return value == expected
	case "neq", "not_equals", "!=":
		return value != expected
	case "contains":
		return strings.Contains(value, expected)
	case "not_contains":
		return !strings.Contains(value, expected)
	default:
		return false
	}
}

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

func scheduleKey(notionPageID string, ruleIndex int) string {
	return notionPageID + ":" + strconv.Itoa(ruleIndex)
}

func weekdayToConfig(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}

func matchesDays(days []int, weekday int) bool {
	if len(days) == 0 {
		return true
	}
	for _, day := range days {
		if day == weekday {
			return true
		}
	}
	return false
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
