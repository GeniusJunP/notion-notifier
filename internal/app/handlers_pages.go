package app

import (
	"net/http"
	"time"

	"notion-notifier/internal/config"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.AddDate(0, 0, 1)

	todayEvents, _ := s.repo.ListEventsBetween(r.Context(), todayStart, todayEnd)
	upcomingEvents, _ := s.repo.ListUpcomingEvents(r.Context(), 14, now)
	historyItems, _ := s.repo.ListNotificationHistory(r.Context(), 50)

	status := s.scheduler.NotionSyncStatus()
	nextSyncLabel := "—"
	nextSyncSub := ""
	if !status.LastSyncedAt.IsZero() {
		next := status.LastSyncedAt.In(loc).Add(time.Duration(cfg.Sync.CheckInterval) * time.Minute)
		nextSyncLabel = next.Format("15:04")
		if next.After(now) {
			nextSyncSub = "in " + formatDurationShort(next.Sub(now))
		} else {
			nextSyncSub = "due"
		}
	}

	lastSyncLabel := "未実行"
	if !status.LastSyncedAt.IsZero() {
		lastSyncLabel = status.LastSyncedAt.In(loc).Format("2006-01-02 15:04")
	}

	snoozeActive := config.IsSnoozed(cfg, now)
	muteActive := config.IsMuted(cfg, now)

	syncMap := map[string]string{}
	if len(upcomingEvents) > 0 {
		ids := make([]string, 0, len(upcomingEvents))
		for _, ev := range upcomingEvents {
			if ev.NotionPageID != "" {
				ids = append(ids, ev.NotionPageID)
			}
		}
		if len(ids) > 0 {
			if statusMap, err := s.repo.GetSyncStatusMap(r.Context(), ids); err == nil {
				syncMap = statusMap
			}
		}
	}

	view := dashboardView{
		TodayCount:    len(todayEvents),
		NextSyncLabel: nextSyncLabel,
		NextSyncSub:   nextSyncSub,
		LastSyncLabel: lastSyncLabel,
		LastSyncCount: status.LastCount,
		LastSyncError: status.LastError,
		SnoozeActive:  snoozeActive,
		SnoozeUntil:   formatDateOnly(cfg.SnoozeUntil, loc),
		MuteActive:    muteActive,
		MuteUntil:     formatDateOnly(cfg.MuteUntil, loc),
		Upcoming:      buildUpcomingViews(upcomingEvents, 10, loc, syncMap),
		History:       buildHistoryViews(historyItems, loc),
	}

	manualTemplates := map[string]string{
		"periodic": "",
	}
	if len(cfg.Notifications.Periodic) > 0 {
		manualTemplates["periodic"] = cfg.Notifications.Periodic[0].Message
	}
	manualPreset := "periodic"
	if manualTemplates["periodic"] == "" {
		manualPreset = "custom"
	}
	defaultTemplate := manualTemplates[manualPreset]

	s.render(w, "dashboard.html", map[string]interface{}{
		"Page":            "dashboard",
		"PageTitle":       "ダッシュボード - Notion Notifier",
		"Config":          cfg,
		"Dashboard":       view,
		"ManualTemplates": manualTemplates,
		"ManualPreset":    manualPreset,
		"ManualTemplate":  defaultTemplate,
		"ManualFrom":      now.Format("2006-01-02"),
		"ManualTo":        now.Format("2006-01-02"),
	})
}

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	s.render(w, "notifications.html", map[string]interface{}{
		"Page":         "notifications",
		"PageTitle":    "通知設定 - Notion Notifier",
		"Config":       cfg,
		"SnoozeDate":   formatDateOnly(cfg.SnoozeUntil, loc),
		"MuteDate":     formatDateOnly(cfg.MuteUntil, loc),
		"SnoozeActive": config.IsSnoozed(cfg, now),
		"MuteActive":   config.IsMuted(cfg, now),
	})
}

func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	lookahead := cfg.CalendarSync.LookaheadDays
	if lookahead <= 0 {
		lookahead = 30
	}
	lastSyncedAt, syncCount, _ := s.repo.GetSyncSummary(r.Context())
	lastSyncLabel := "未実行"
	if lastSyncedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, lastSyncedAt); err == nil {
			lastSyncLabel = parsed.In(loc).Format("2006-01-02 15:04")
		}
	}
	s.render(w, "calendar.html", map[string]interface{}{
		"Page":          "calendar",
		"PageTitle":     "カレンダー連携 - Notion Notifier",
		"Config":        cfg,
		"CalendarLast":  lastSyncLabel,
		"CalendarCount": syncCount,
		"CalendarFrom":  now.Format("2006-01-02"),
		"CalendarTo":    now.AddDate(0, 0, lookahead).Format("2006-01-02"),
	})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	s.render(w, "settings.html", map[string]interface{}{
		"Page":      "settings",
		"PageTitle": "システム設定 - Notion Notifier",
		"Config":    cfg,
	})
}
