package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	SnoozeActive  bool
	SnoozeUntil   string
	MuteActive    bool
	MuteUntil     string
	Upcoming      []upcomingView
	History       []historyView
}

type upcomingView struct {
	Title     string
	DateLabel string
	TimeLabel string
	Location  string
	URL       string
	SyncStatus string
}

type historyView struct {
	Title     string
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
	mux.HandleFunc("/api/notifications/preview", s.requireAuth(s.handleAPIPreviewNotification))
	mux.HandleFunc("/api/notifications/manual", s.requireAuth(s.handleAPIManualNotification))
	mux.HandleFunc("/api/calendar/sync", s.requireAuth(s.handleAPICalendarSync))
	mux.HandleFunc("/api/calendar/clear", s.requireAuth(s.handleAPICalendarClear))
	mux.HandleFunc("/api/history/clear", s.requireAuth(s.handleAPIHistoryClear))

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
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"type":"sync_complete","count":%d}}`, count))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current config
	currCfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(currCfg.Timezone)

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
		if m, ok := notifications.(map[string]interface{}); ok {
			if isTruthy(m["advance_clear"]) {
				mergedCfg.Notifications.Advance = nil
			}
			if isTruthy(m["periodic_clear"]) {
				mergedCfg.Notifications.Periodic = nil
			}
			applyAdvanceConditionClears(&mergedCfg, m)
			applyPeriodicDayClears(&mergedCfg, m)
		}
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
		if m, ok := propertyMapping.(map[string]interface{}); ok {
			if isTruthy(m["custom_clear"]) {
				mergedCfg.PropertyMap.Custom = nil
			}
		}
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
			if normalized, err := normalizeDateInput(v, loc); err == nil {
				mergedCfg.SnoozeUntil = normalized
			}
		}
	}
	if mute, ok := updates["mute_until"]; ok {
		if v, ok := mute.(string); ok {
			if normalized, err := normalizeDateInput(v, loc); err == nil {
				mergedCfg.MuteUntil = normalized
			}
		}
	}
	// security.basic_auth.enabled is config-only (credentials are in env.yaml)

	if err := s.cfg.UpdateConfig(mergedCfg); err != nil {
		http.Error(w, "Failed to update config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"showToast":{"type":"config_saved"}}`)
	w.WriteHeader(http.StatusNoContent)
}

type notificationRequest struct {
	Template       string `json:"template"`
	FromDate       string `json:"from_date"`
	ToDate         string `json:"to_date"`
	MinutesBefore  int    `json:"minutes_before"`
	PreviewPayload bool   `json:"preview_payload"`
}

func (s *Server) handleAPIPreviewNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req notificationRequest
	if err := decodeNotificationRequest(r, &req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	from, to, err := parseDateRange(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		http.Error(w, "Invalid date: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.MinutesBefore > 0 {
		message, err := s.scheduler.PreviewAdvanceTemplate(r.Context(), req.Template, req.MinutesBefore)
		if err != nil {
			http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writePreviewHTML(w, message, "")
		return
	}
	if req.PreviewPayload {
		message, payload, err := s.scheduler.PreviewManualPayload(r.Context(), req.Template, from, to)
		if err != nil {
			http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writePreviewHTML(w, message, payload)
		return
	}
	message, err := s.scheduler.PreviewTemplate(r.Context(), req.Template, from, to)
	if err != nil {
		http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writePreviewHTML(w, message, "")
}

func (s *Server) handleAPIManualNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req notificationRequest
	if err := decodeNotificationRequest(r, &req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	from, to, err := parseDateRange(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		http.Error(w, "Invalid date: "+err.Error(), http.StatusBadRequest)
		return
	}
	message, err := s.scheduler.SendManualNotification(r.Context(), req.Template, from, to)
	if err != nil {
		http.Error(w, "Send failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", `{"showToast":{"type":"manual_sent"}}`)
	writePreviewHTML(w, message, "")
}

type calendarSyncRequest struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

func (s *Server) handleAPICalendarSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req calendarSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	from, to, err := parseDateRange(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		http.Error(w, "Invalid date: "+err.Error(), http.StatusBadRequest)
		return
	}
	count, err := s.scheduler.SyncCalendar(r.Context(), from, to)
	if err != nil {
		http.Error(w, "Calendar sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"type":"calendar_sync_complete","count":%d}}`, count))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPICalendarClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repo.ClearSyncRecords(r.Context()); err != nil {
		http.Error(w, "Failed to clear sync records: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", `{"showToast":{"type":"sync_records_cleared"}}`)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIHistoryClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repo.ClearNotificationHistory(r.Context()); err != nil {
		http.Error(w, "Failed to clear history: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", `{"showToast":{"type":"history_cleared"}}`)
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
		cfg, env := s.cfg.Get()
		if !cfg.Security.BasicAuth.Enabled {
			next(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != env.Security.BasicAuth.Username || pass != env.Security.BasicAuth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
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
			Title:     ev.Title,
			DateLabel: formatEventDate(ev, loc),
			TimeLabel: formatEventTime(ev, loc),
			Location:  ev.Location,
			URL:       ev.URL,
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

func isTruthy(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	case float64:
		return v != 0
	default:
		return false
	}
}

func applyAdvanceConditionClears(cfg *config.Config, notifications map[string]interface{}) {
	raw, ok := notifications["advance"].([]interface{})
	if !ok {
		return
	}
	for i, item := range raw {
		if i >= len(cfg.Notifications.Advance) {
			break
		}
		rule, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		cond, ok := rule["conditions"].(map[string]interface{})
		if !ok {
			continue
		}
		if isTruthy(cond["days_of_week_clear"]) {
			cfg.Notifications.Advance[i].Conditions.DaysOfWeek = nil
		}
		if isTruthy(cond["property_filters_clear"]) {
			cfg.Notifications.Advance[i].Conditions.PropertyFilters = nil
		}
	}
}

func applyPeriodicDayClears(cfg *config.Config, notifications map[string]interface{}) {
	raw, ok := notifications["periodic"].([]interface{})
	if !ok {
		return
	}
	for i, item := range raw {
		if i >= len(cfg.Notifications.Periodic) {
			break
		}
		rule, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if isTruthy(rule["days_of_week_clear"]) {
			cfg.Notifications.Periodic[i].DaysOfWeek = nil
		}
	}
}

func decodeNotificationRequest(r *http.Request, req *notificationRequest) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return fmt.Errorf("empty body")
	}
	if err := json.Unmarshal(body, req); err == nil {
		return nil
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	req.Template = values.Get("template")
	req.FromDate = values.Get("from_date")
	req.ToDate = values.Get("to_date")
	if v := values.Get("minutes_before"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			req.MinutesBefore = parsed
		}
	}
	req.PreviewPayload = values.Get("preview_payload") == "true"
	if req.Template == "" && req.FromDate == "" && req.ToDate == "" && req.MinutesBefore == 0 {
		return fmt.Errorf("invalid payload")
	}
	return nil
}

func parseDateRange(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
	current, _ := cfg.Get()
	loc, _ := time.LoadLocation(current.Timezone)
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
		return time.Time{}, time.Time{}, fmt.Errorf("to_date must be after from_date")
	}

	return from, to, nil
}

func parseDateInput(value string, loc *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("date is required")
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
	return time.Time{}, fmt.Errorf("invalid date format")
}

func normalizeDateInput(value string, loc *time.Location) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if len(value) == len("2006-01-02") && strings.Count(value, "-") == 2 {
		parsed, err := time.ParseInLocation("2006-01-02", value, loc)
		if err != nil {
			return "", err
		}
		return parsed.Add(24 * time.Hour).Format(time.RFC3339), nil
	}
	parsed, err := parseDateInput(value, loc)
	if err != nil {
		return "", err
	}
	return parsed.Format(time.RFC3339), nil
}

func writePreviewHTML(w http.ResponseWriter, message string, payload string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	escaped := template.HTMLEscapeString(message)
	parts := []string{
		fmt.Sprintf(`<div class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest mb-2">Message</div><pre class="whitespace-pre-wrap font-mono text-xs leading-relaxed">%s</pre>`, escaped),
	}
	if payload != "" {
		parts = append(parts, fmt.Sprintf(`<div class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest mt-6 mb-2">Payload</div><pre class="whitespace-pre-wrap font-mono text-xs leading-relaxed">%s</pre>`, template.HTMLEscapeString(payload)))
	}
	fmt.Fprintf(w, `<div class="mt-4 rounded-2xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/50 p-4 text-sm text-slate-700 dark:text-slate-200">%s</div>`, strings.Join(parts, ""))
}

func formatDateOnly(value string, loc *time.Location) string {
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	return parsed.In(loc).Format("2006-01-02")
}
