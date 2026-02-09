package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/scheduler"
)

type Server struct {
	addr      string
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
}

func New(addr string, cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) *Server {
	return &Server{
		addr:      addr,
		cfg:       cfg,
		repo:      repo,
		scheduler: sched,
	}
}

func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.Handler(),
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	return srv.ListenAndServe()
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.basicAuthMiddleware)

	r.Get("/api/health", s.handleHealth)
	r.Get("/api/config", s.handleGetConfig)
	r.Put("/api/config", s.handleUpdateConfig)
	r.Post("/api/reload", s.handleReload)

	r.Get("/api/events", s.handleListEvents)
	r.Post("/api/sync/notion", s.handleSyncNotion)

	r.Post("/api/notifications/preview", s.handlePreviewNotification)
	r.Post("/api/notifications/manual", s.handleManualNotification)
	r.Get("/api/notifications/history", s.handleHistory)
	r.Post("/api/notifications/history/clear", s.handleClearHistory)
	r.Post("/api/notifications/snooze", s.handleSnooze)
	r.Post("/api/notifications/mute", s.handleMute)

	r.Post("/api/calendar/sync", s.handleCalendarSync)
	r.Get("/api/calendar/status", s.handleCalendarStatus)
	r.Post("/api/calendar/clear", s.handleCalendarClear)
	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, _ := s.cfg.Get()
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	if err := s.cfg.UpdateConfig(cfg); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	if err := s.scheduler.Reload(r.Context()); err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if err := s.cfg.Reload(); err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.scheduler.Reload(r.Context()); err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	days := 14
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	events, err := s.repo.ListUpcomingEvents(r.Context(), days, time.Now().In(loc))
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleSyncNotion(w http.ResponseWriter, r *http.Request) {
	count, err := s.scheduler.SyncNotion(r.Context())
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "count": count})
}

type notificationRequest struct {
	Template string `json:"template"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

func (s *Server) handlePreviewNotification(w http.ResponseWriter, r *http.Request) {
	var req notificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	from, to, err := parseDateRangeRequest(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	message, err := s.scheduler.PreviewTemplate(r.Context(), req.Template, from, to)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": message})
}

func (s *Server) handleManualNotification(w http.ResponseWriter, r *http.Request) {
	var req notificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	from, to, err := parseDateRangeRequest(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	message, err := s.scheduler.SendManualNotification(r.Context(), req.Template, from, to)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": message})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	history, err := s.repo.ListNotificationHistory(r.Context(), 50)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (s *Server) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if err := s.repo.ClearNotificationHistory(r.Context()); err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type snoozeRequest struct {
	Until string `json:"until"`
	Days  int    `json:"days"`
}

func (s *Server) handleSnooze(w http.ResponseWriter, r *http.Request) {
	var req snoozeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	until, err := parseUntil(req.Until, req.Days, loc)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	cfg.SnoozeUntil = until
	if err := s.cfg.UpdateConfig(cfg); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "until": cfg.SnoozeUntil})
}

func (s *Server) handleMute(w http.ResponseWriter, r *http.Request) {
	var req snoozeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	until, err := parseUntil(req.Until, req.Days, loc)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	cfg.MuteUntil = until
	if err := s.cfg.UpdateConfig(cfg); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "until": cfg.MuteUntil})
}

type calendarSyncRequest struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

func (s *Server) handleCalendarSync(w http.ResponseWriter, r *http.Request) {
	var req calendarSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	from, to, err := parseDateRangeRequest(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}
	count, err := s.scheduler.SyncCalendar(r.Context(), from, to)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "count": count})
}

func (s *Server) handleCalendarStatus(w http.ResponseWriter, r *http.Request) {
	last, count, err := s.repo.GetSyncSummary(r.Context())
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"last_synced_at": last, "count": count})
}

func (s *Server) handleCalendarClear(w http.ResponseWriter, r *http.Request) {
	if err := s.repo.ClearSyncRecords(r.Context()); err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseDateRangeRequest(fromStr, toStr string, cfg *config.Manager) (time.Time, time.Time, error) {
	c, _ := cfg.Get()
	loc, _ := time.LoadLocation(c.Timezone)
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, errors.New("from_date and to_date are required")
	}
	from, err := time.ParseInLocation("2006-01-02", fromStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}

func parseUntil(until string, days int, loc *time.Location) (string, error) {
	if days > 0 {
		return time.Now().In(loc).AddDate(0, 0, days).Format(time.RFC3339), nil
	}
	if until == "" {
		return "", nil
	}
	t, err := time.Parse(time.RFC3339, until)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC3339), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func errorJSON(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
