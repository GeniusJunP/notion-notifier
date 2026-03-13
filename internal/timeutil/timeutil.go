package timeutil

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"notion-notifier/internal/config"
)

// LoadOrLocal loads the named timezone, falling back to time.Local
// if the name is empty or invalid.
func LoadOrLocal(name string) *time.Location {
	if strings.TrimSpace(name) == "" {
		return time.Local
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

// ParseDateRange parses fromStr and toStr using the timezone configured in cfg.
func ParseDateRange(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
	current := cfg.Config()
	loc := LoadOrLocal(current.Timezone)
	now := time.Now().In(loc)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := from

	fromStr = strings.TrimSpace(fromStr)
	toStr = strings.TrimSpace(toStr)

	if fromStr != "" {
		parsed, err := ParseDateInput(fromStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed
	}

	if toStr != "" {
		parsed, err := ParseDateInput(toStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed
	} else if fromStr != "" {
		to = from
	}

	if to.Before(from) {
		return time.Time{}, time.Time{}, ErrToBeforeFrom
	}

	return from, to, nil
}

// ParseDateInput parses a date input string in various formats in the given location.
func ParseDateInput(value string, loc *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, ErrDateRequired
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
	return time.Time{}, ErrInvalidDateFormat
}

// FormatDurationShort formats a duration into a short string (e.g., "1h30m", "< 1m").
func FormatDurationShort(d time.Duration) string {
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

// Exported errors
var (
	ErrToBeforeFrom      = errors.New("to_date must be after from_date")
	ErrDateRequired      = errors.New("date is required")
	ErrInvalidDateFormat = errors.New("invalid date format")
)
