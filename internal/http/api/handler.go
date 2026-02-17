package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
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
	respondJSON(w, http.StatusOK, h.cfg.Config())
}

func (h *Handler) putConfig(w http.ResponseWriter, r *http.Request) {
	var incoming config.Config
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	saved, err := h.cfg.UpdateConfig(incoming)
	if err != nil {
		var vErr config.ValidationError
		if errors.As(err, &vErr) {
			respondValidationError(w, "validation failed", map[string]string{
				"config": vErr.Error(),
			})
			return
		}
		logging.Error("CONF", "update failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to save config")
		return
	}

	logging.Info("CONF", "config updated from %s", r.RemoteAddr)
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	cfg := h.cfg.Config()
	loc := loadLocationOrLocal(cfg.Timezone)
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.AddDate(0, 0, 1)

	todayEvents, err := h.repo.ListEventsBetween(r.Context(), todayStart, todayEnd)
	if err != nil {
		logging.Error("DASH", "list events failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}
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
	NotionPageID  string `json:"notion_page_id"`
	Title         string `json:"title"`
	StartDate     string `json:"start_date"`
	StartTime     string `json:"start_time"`
	EndDate       string `json:"end_date,omitempty"`
	EndTime       string `json:"end_time,omitempty"`
	IsAllDay      bool   `json:"is_all_day"`
	Location      string `json:"location,omitempty"`
	URL           string `json:"url,omitempty"`
	CalendarState string `json:"calendar_state"`
}

func (h *Handler) handleUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	cfg := h.cfg.Config()
	loc := loadLocationOrLocal(cfg.Timezone)
	now := time.Now().In(loc)

	events, err := h.repo.ListUpcomingEvents(r.Context(), 14, now)
	if err != nil {
		logging.Error("EVENTS", "list upcoming events failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to load events")
		return
	}

	// Build sync record map
	ids := make([]string, 0, len(events))
	for _, ev := range events {
		if ev.NotionPageID != "" {
			ids = append(ids, ev.NotionPageID)
		}
	}
	syncMap := map[string]models.SyncRecord{}
	if len(ids) > 0 {
		syncMap, err = h.repo.GetSyncRecordMap(r.Context(), ids)
		if err != nil {
			logging.Error("EVENTS", "load sync records failed: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to load events")
			return
		}
	}

	out := make([]eventResponse, 0, len(events))
	for _, ev := range events {
		calendarState := "disabled"
		if cfg.CalendarSync.Enabled {
			record, ok := syncMap[ev.NotionPageID]
			switch {
			case !ok || !record.Attempted:
				calendarState = "needs_sync"
			case record.Synced:
				calendarState = "synced"
			default:
				calendarState = "error"
			}
		}
		out = append(out, eventResponse{
			NotionPageID:  ev.NotionPageID,
			Title:         ev.Title,
			StartDate:     ev.StartDate,
			StartTime:     ev.StartTime,
			EndDate:       ev.EndDate,
			EndTime:       ev.EndTime,
			IsAllDay:      ev.IsAllDay,
			Location:      ev.Location,
			URL:           ev.URL,
			CalendarState: calendarState,
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	items, err := h.repo.ListNotificationHistory(r.Context(), 50)
	if err != nil {
		logging.Error("HISTORY", "load history failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to load history")
		return
	}
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
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	count, err := h.sched.SyncNotion()
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
	if !requireMethod(w, r, http.MethodPost) {
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

	count, err := h.sched.SyncCalendar(from, to)
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
	if !requireMethod(w, r, http.MethodPost) {
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
	if !requireMethod(w, r, http.MethodPost) {
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
}

func (h *Handler) handlePreviewNotification(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
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

	message, err := h.sched.PreviewManualTemplate(r.Context(), req.Template, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "preview failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, previewResponse{Message: message})
}

// --- POST /api/notifications/manual ---

func (h *Handler) handleManualNotification(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
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

	cfg := h.cfg.Config()
	cfg.Notifications.Manual = req.Template
	saved, err := h.cfg.UpdateConfig(cfg)
	if err != nil {
		logging.Error("CONF", "manual template save failed: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to save manual template")
		return
	}

	message, err := h.sched.SendManualNotification(r.Context(), saved.Notifications.Manual, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "send failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, previewResponse{Message: message})
}

// --- GET /api/templates/defaults ---

func (h *Handler) handleDefaultTemplates(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	respondJSON(w, http.StatusOK, config.DefaultTemplates())
}
