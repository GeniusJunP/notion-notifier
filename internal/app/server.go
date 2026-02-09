package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/models"
	"notion-notifier/internal/scheduler"
)

//go:embed web/templates/*.html
var templatesFS embed.FS

type Server struct {
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
	tmpl      *template.Template
}

type dashboardView struct {
	TodayCount    int
	NextSyncLabel string
	NextSyncSub   string
	LastSyncLabel string
	LastSyncCount int
	LastSyncError string
	Upcoming      []upcomingView
	History       []historyView
}

type upcomingView struct {
	Title     string
	TimeLabel string
	URL       string
}

type historyView struct {
	Title      string
	Status    string
	TimeLabel string
}

func NewServer(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) (*Server, error) {
	tmpl := template.New("base").Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	})
	tmpl, err := tmpl.ParseFS(templatesFS, "web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		cfg:       cfg,
		repo:      repo,
		scheduler: sched,
		tmpl:      tmpl,
	}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Page routes
	mux.HandleFunc("/", s.requireAuth(s.handleDashboard))
	mux.HandleFunc("/notifications", s.requireAuth(s.handleNotifications))
	mux.HandleFunc("/calendar", s.requireAuth(s.handleCalendar))
	mux.HandleFunc("/settings", s.requireAuth(s.handleSettings))

	// API routes
	mux.HandleFunc("/api/sync", s.requireAuth(s.handleAPISync))
	mux.HandleFunc("/api/config", s.requireAuth(s.handleAPIConfig))

	return mux
}

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
	historyItems, _ := s.repo.ListNotificationHistory(r.Context(), 10)

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

	view := dashboardView{
		TodayCount:    len(todayEvents),
		NextSyncLabel: nextSyncLabel,
		NextSyncSub:   nextSyncSub,
		LastSyncLabel: lastSyncLabel,
		LastSyncCount: status.LastCount,
		LastSyncError: status.LastError,
		Upcoming:      buildUpcomingViews(upcomingEvents, 5, loc),
		History:       buildHistoryViews(historyItems, loc),
	}

	s.render(w, "dashboard.html", map[string]interface{}{
		"Page":            "dashboard",
		"PageTitle":       "ダッシュボード - Notion Notifier",
		"Config":          cfg,
		"Dashboard":       view,
	})
}

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	s.render(w, "notifications.html", map[string]interface{}{
		"Page":            "notifications",
		"PageTitle":       "通知設定 - Notion Notifier",
		"Config":          cfg,
	})
}

func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	s.render(w, "calendar.html", map[string]interface{}{
		"Page":            "calendar",
		"PageTitle":       "カレンダー連携 - Notion Notifier",
		"Config":          cfg,
	})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	s.render(w, "settings.html", map[string]interface{}{
		"Page":            "settings",
		"PageTitle":       "システム設定 - Notion Notifier",
		"Config":          cfg,
	})
}

func (s *Server) handleAPISync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	count, err := s.scheduler.SyncNotion(r.Context())
	if err != nil {
		http.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": "同期を完了しました (%d件)"}`, count))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current config
	currCfg, _ := s.cfg.Get()

	// Decode into a map first to see what fields are present
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Re-encode and decode into the struct to apply updates
	// This is a simple way to merge JSON into existing struct
	b, _ := json.Marshal(currCfg)
	var mergedCfg config.Config
	json.Unmarshal(b, &mergedCfg)

	// Since we want to support nested updates, we'll use a more targeted approach
	// For now, let's just handle the sections we have
	if notifications, ok := updates["notifications"]; ok {
		nb, _ := json.Marshal(notifications)
		json.Unmarshal(nb, &mergedCfg.Notifications)
	}
	if webhook, ok := updates["webhook"]; ok {
		wb, _ := json.Marshal(webhook)
		json.Unmarshal(wb, &mergedCfg.Webhook)
	}
	if calendarSync, ok := updates["calendar_sync"]; ok {
		cb, _ := json.Marshal(calendarSync)
		json.Unmarshal(cb, &mergedCfg.CalendarSync)
	}
	if propertyMapping, ok := updates["property_mapping"]; ok {
		pb, _ := json.Marshal(propertyMapping)
		json.Unmarshal(pb, &mergedCfg.PropertyMap)
	}
	if contentRules, ok := updates["content_rules"]; ok {
		cr, _ := json.Marshal(contentRules)
		json.Unmarshal(cr, &mergedCfg.ContentRules)
	}
	if sync, ok := updates["sync"]; ok {
		sb, _ := json.Marshal(sync)
		json.Unmarshal(sb, &mergedCfg.Sync)
	}
	if timezone, ok := updates["timezone"]; ok {
		if v, ok := timezone.(string); ok {
			mergedCfg.Timezone = v
		}
	}
	if snooze, ok := updates["snooze_until"]; ok {
		if v, ok := snooze.(string); ok {
			mergedCfg.SnoozeUntil = v
		}
	}
	if mute, ok := updates["mute_until"]; ok {
		if v, ok := mute.(string); ok {
			mergedCfg.MuteUntil = v
		}
	}
	// security.basic_auth is config-only (not updated via UI)

	if err := s.cfg.UpdateConfig(mergedCfg); err != nil {
		http.Error(w, "Failed to update config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"showToast": "設定を保存しました"}`)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) render(w http.ResponseWriter, name string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, _ := s.cfg.Get()
		if !cfg.Security.BasicAuth.Enabled {
			next(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != cfg.Security.BasicAuth.Username || pass != cfg.Security.BasicAuth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func buildUpcomingViews(events []models.Event, limit int, loc *time.Location) []upcomingView {
	out := make([]upcomingView, 0, limit)
	for _, ev := range events {
		if limit > 0 && len(out) >= limit {
			break
		}
		out = append(out, upcomingView{
			Title:     ev.Title,
			TimeLabel: formatEventTime(ev, loc),
			URL:       ev.URL,
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
			Title:      title,
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
