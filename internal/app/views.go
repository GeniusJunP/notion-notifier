package app

import (
	"fmt"
	"strings"
	"time"

	"notion-notifier/internal/models"
)

type dashboardView struct {
	TodayCount    int
	NextSyncLabel string
	NextSyncSub   string
	LastSyncLabel string
	LastSyncCount int
	LastSyncError string
	SnoozeActive  bool
	SnoozeUntil   string
	MuteActive    bool
	MuteUntil     string
	Upcoming      []upcomingView
	History       []historyView
}

type upcomingView struct {
	Title      string
	DateLabel  string
	TimeLabel  string
	Location   string
	URL        string
	SyncStatus string
}

type historyView struct {
	Title     string
	Status    string
	TimeLabel string
}

func buildUpcomingViews(events []models.Event, limit int, loc *time.Location, syncMap map[string]string) []upcomingView {
	out := make([]upcomingView, 0, limit)
	for _, ev := range events {
		if limit > 0 && len(out) >= limit {
			break
		}
		status := "unsynced"
		if syncMap != nil {
			if v, ok := syncMap[ev.NotionPageID]; ok && v != "" {
				status = v
			}
		}
		out = append(out, upcomingView{
			Title:      ev.Title,
			DateLabel:  formatEventDate(ev, loc),
			TimeLabel:  formatEventTime(ev, loc),
			Location:   ev.Location,
			URL:        ev.URL,
			SyncStatus: status,
		})
	}
	return out
}

func buildHistoryViews(items []models.NotificationHistory, loc *time.Location) []historyView {
	out := make([]historyView, 0, len(items))
	for _, item := range items {
		title := firstLine(item.Message)
		if title == "" {
			title = item.Type
		}
		timeLabel := item.SentAt.In(loc).Format("01/02 15:04")
		out = append(out, historyView{
			Title:     title,
			Status:    item.Status,
			TimeLabel: timeLabel,
		})
	}
	return out
}

func formatEventTime(ev models.Event, loc *time.Location) string {
	if ev.IsAllDay || ev.StartTime == "" {
		return "終日"
	}
	return ev.StartTime
}

func formatEventDate(ev models.Event, loc *time.Location) string {
	if ev.StartDate == "" {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	if parsed, err := time.ParseInLocation("2006-01-02", ev.StartDate, loc); err == nil {
		return parsed.Format("01/02")
	}
	return ev.StartDate
}

func formatDurationShort(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := hours / 24
	remHours := hours % 24
	if remHours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, remHours)
}

func firstLine(input string) string {
	line := strings.TrimSpace(input)
	if line == "" {
		return ""
	}
	parts := strings.SplitN(line, "\n", 2)
	return strings.TrimSpace(parts[0])
}
