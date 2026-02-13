package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/scheduler"
)

// Handler holds dependencies for all API endpoints.
type Handler struct {
	cfg   *config.Manager
	repo  *db.Repository
	sched *scheduler.Scheduler
}

// NewHandler creates a new API handler.
func NewHandler(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) *Handler {
	return &Handler{cfg: cfg, repo: repo, sched: sched}
}

// Register mounts all API routes on the given mux under /api/.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/config", h.handleConfig)
	mux.HandleFunc("/api/dashboard", h.handleDashboard)
	mux.HandleFunc("/api/history", h.handleHistory)
	mux.HandleFunc("/api/events/upcoming", h.handleUpcomingEvents)
	mux.HandleFunc("/api/sync", h.handleSync)
	mux.HandleFunc("/api/calendar/sync", h.handleCalendarSync)
	mux.HandleFunc("/api/calendar/clear", h.handleCalendarClear)
	mux.HandleFunc("/api/history/clear", h.handleHistoryClear)
	mux.HandleFunc("/api/notifications/preview", h.handlePreviewNotification)
	mux.HandleFunc("/api/notifications/manual", h.handleManualNotification)
	mux.HandleFunc("/api/templates/defaults", h.handleDefaultTemplates)
}

// --- GET /api/config, PUT /api/config ---

func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getConfig(w, r)
	case http.MethodPut:
		h.putConfig(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) getConfig(w http.ResponseWriter, _ *http.Request) {
	cfg, _ := h.cfg.Get()
	respondJSON(w, http.StatusOK, cfg)
}

func (h *Handler) putConfig(w http.ResponseWriter, r *http.Request) {
	var incoming config.Config
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	incoming = config.NormalizeConfig(incoming)
	if err := config.ValidateConfig(incoming); err != nil {
		respondValidationError(w, "validation failed", map[string]string{
			"config": err.Error(),
		})
		return
	}

	if err := h.cfg.UpdateConfig(incoming); err != nil {
		logging.Error("CONF", "update failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to save config")
		return
	}

	logging.Info("CONF", "config updated from %s", r.RemoteAddr)
	// Return the normalized config
	saved, _ := h.cfg.Get()
	respondJSON(w, http.StatusOK, saved)
}

// --- GET /api/dashboard ---

type dashboardResponse struct {
	TodayCount    int    `json:"today_count"`
	NextSync      string `json:"next_sync"`
	NextSyncIn    string `json:"next_sync_in"`
	LastSync      string `json:"last_sync"`
	LastSyncCount int    `json:"last_sync_count"`
	LastSyncError string `json:"last_sync_error,omitempty"`
	SnoozeActive  bool   `json:"snooze_active"`
	SnoozeUntil   string `json:"snooze_until,omitempty"`
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfg, _ := h.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.AddDate(0, 0, 1)

	todayEvents, _ := h.repo.ListEventsBetween(r.Context(), todayStart, todayEnd)
	status := h.sched.NotionSyncStatus()

	nextSync := ""
	nextSyncIn := ""
	if !status.LastSyncedAt.IsZero() {
		next := status.LastSyncedAt.In(loc).Add(time.Duration(cfg.Sync.CheckInterval) * time.Minute)
		nextSync = next.Format(time.RFC3339)
		if next.After(now) {
			nextSyncIn = formatDurationShort(next.Sub(now))
		}
	}

	lastSync := ""
	if !status.LastSyncedAt.IsZero() {
		lastSync = status.LastSyncedAt.In(loc).Format(time.RFC3339)
	}

	resp := dashboardResponse{
		TodayCount:    len(todayEvents),
		NextSync:      nextSync,
		NextSyncIn:    nextSyncIn,
		LastSync:      lastSync,
		LastSyncCount: status.LastCount,
		LastSyncError: status.LastError,
		SnoozeActive:  config.IsSnoozed(cfg, now),
		SnoozeUntil:   cfg.SnoozeUntil,
	}
	respondJSON(w, http.StatusOK, resp)
}

// --- GET /api/events/upcoming ---

type eventResponse struct {
	NotionPageID string `json:"notion_page_id"`
	Title        string `json:"title"`
	StartDate    string `json:"start_date"`
	StartTime    string `json:"start_time"`
	EndDate      string `json:"end_date,omitempty"`
	EndTime      string `json:"end_time,omitempty"`
	IsAllDay     bool   `json:"is_all_day"`
	Location     string `json:"location,omitempty"`
	URL          string `json:"url,omitempty"`
	CacheStatus  string `json:"cache_status"`
}

func (h *Handler) handleUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfg, _ := h.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)

	events, _ := h.repo.ListUpcomingEvents(r.Context(), 14, now)

	// Build sync status map
	ids := make([]string, 0, len(events))
	for _, ev := range events {
		if ev.NotionPageID != "" {
			ids = append(ids, ev.NotionPageID)
		}
	}
	syncMap := map[string]string{}
	if len(ids) > 0 {
		if m, err := h.repo.GetSyncStatusMap(r.Context(), ids); err == nil {
			syncMap = m
		}
	}

	out := make([]eventResponse, 0, len(events))
	for _, ev := range events {
		cacheStatus := "unsynced"
		if v, ok := syncMap[ev.NotionPageID]; ok && v != "" {
			cacheStatus = v
		}
		out = append(out, eventResponse{
			NotionPageID: ev.NotionPageID,
			Title:        ev.Title,
			StartDate:    ev.StartDate,
			StartTime:    ev.StartTime,
			EndDate:      ev.EndDate,
			EndTime:      ev.EndTime,
			IsAllDay:     ev.IsAllDay,
			Location:     ev.Location,
			URL:          ev.URL,
			CacheStatus:  cacheStatus,
		})
	}
	respondJSON(w, http.StatusOK, out)
}

// --- GET /api/history ---

type historyResponse struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	SentAt  string `json:"sent_at"`
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	items, _ := h.repo.ListNotificationHistory(r.Context(), 50)
	out := make([]historyResponse, 0, len(items))
	for _, item := range items {
		out = append(out, historyResponse{
			ID:      item.ID,
			Type:    item.Type,
			Status:  item.Status,
			Message: item.Message,
			Error:   item.Error,
			SentAt:  item.SentAt.Format(time.RFC3339),
		})
	}
	respondJSON(w, http.StatusOK, out)
}

// --- POST /api/sync ---

func (h *Handler) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	count, err := h.sched.SyncNotion(r.Context())
	if err != nil {
		logging.Error("SYNC", "notion sync failed: %v", err)
		respondError(w, http.StatusInternalServerError, "sync failed: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]int{"count": count})
}

// --- POST /api/calendar/sync ---

type calendarSyncRequest struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

func (h *Handler) handleCalendarSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req calendarSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	from, to, err := parseDateRange(req.FromDate, req.ToDate, h.cfg)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date: "+err.Error())
		return
	}

	count, err := h.sched.SyncCalendar(r.Context(), from, to)
	if err != nil {
		logging.Error("CALENDAR", "calendar sync failed: %v", err)
		respondError(w, http.StatusInternalServerError, "calendar sync failed: "+err.Error())
		return
	}

	logging.Info("CALENDAR", "calendar sync complete (count=%d)", count)
	respondJSON(w, http.StatusOK, map[string]int{"count": count})
}

// --- POST /api/calendar/clear ---

func (h *Handler) handleCalendarClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := h.repo.ClearSyncRecords(r.Context()); err != nil {
		logging.Error("CALENDAR", "clear sync records failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to clear sync records")
		return
	}

	logging.Info("CALENDAR", "sync records cleared")
	w.WriteHeader(http.StatusNoContent)
}

// --- POST /api/history/clear ---

func (h *Handler) handleHistoryClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := h.repo.ClearNotificationHistory(r.Context()); err != nil {
		logging.Error("HISTORY", "clear history failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to clear history")
		return
	}

	logging.Info("HISTORY", "notification history cleared")
	w.WriteHeader(http.StatusNoContent)
}

// --- POST /api/notifications/preview ---

type notificationRequest struct {
	Template      string `json:"template"`
	FromDate      string `json:"from_date"`
	ToDate        string `json:"to_date"`
	MinutesBefore int    `json:"minutes_before"`
}

type previewResponse struct {
	Message string `json:"message"`
	Payload string `json:"payload,omitempty"`
}

func (h *Handler) handlePreviewNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req notificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Advance template preview
	if req.MinutesBefore > 0 {
		message, err := h.sched.PreviewAdvanceTemplate(r.Context(), req.Template, req.MinutesBefore)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "preview failed: "+err.Error())
			return
		}
		respondJSON(w, http.StatusOK, previewResponse{Message: message})
		return
	}

	from, to, err := parseDateRange(req.FromDate, req.ToDate, h.cfg)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date: "+err.Error())
		return
	}

	message, payload, err := h.sched.PreviewManualPayload(r.Context(), req.Template, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "preview failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, previewResponse{Message: message, Payload: payload})
}

// --- POST /api/notifications/manual ---

func (h *Handler) handleManualNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req notificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	from, to, err := parseDateRange(req.FromDate, req.ToDate, h.cfg)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date: "+err.Error())
		return
	}

	message, err := h.sched.SendManualNotification(r.Context(), req.Template, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "send failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, previewResponse{Message: message})
}

// --- GET /api/templates/defaults ---

func (h *Handler) handleDefaultTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	respondJSON(w, http.StatusOK, config.DefaultTemplates())
}

// --- Utility ---

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

var (
	errToBeforeFrom      = fmt.Errorf("to_date must be after from_date")
	errDateRequired      = fmt.Errorf("date is required")
	errInvalidDateFormat = fmt.Errorf("invalid date format")
)
