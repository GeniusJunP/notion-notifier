package api

import (
	"fmt"
	"strings"
	"time"

	"notion-notifier/internal/config"
)

func parseDateRange(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
	current := cfg.Config()
	loc := loadLocationOrLocal(current.Timezone)
	now := time.Now().In(loc)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := from

	fromStr = strings.TrimSpace(fromStr)
	toStr = strings.TrimSpace(toStr)

	if fromStr != "" {
		parsed, err := parseDateInput(fromStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed
	}

	if toStr != "" {
		parsed, err := parseDateInput(toStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed
	} else if fromStr != "" {
		to = from
	}

	if to.Before(from) {
		return time.Time{}, time.Time{}, errToBeforeFrom
	}

	return from, to, nil
}

func parseDateInput(value string, loc *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errDateRequired
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.In(loc), nil
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, loc); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errInvalidDateFormat
}

func formatDurationShort(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Minute {
		return "< 1m"
	}
	m := int(d.Minutes())
	if d < time.Hour {
		return fmt.Sprintf("%dm", m)
	}
	h := int(d.Hours())
	rm := m % 60
	if rm == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, rm)
}

func loadLocationOrLocal(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

var (
	errToBeforeFrom      = fmt.Errorf("to_date must be after from_date")
	errDateRequired      = fmt.Errorf("date is required")
	errInvalidDateFormat = fmt.Errorf("invalid date format")
)
