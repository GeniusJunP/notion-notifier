package api

import (
	"encoding/json"
	"errors"
	"net/http"

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
		respondJSON(w, http.StatusOK, h.cfg.Config())
	case http.MethodPut:
		var incoming config.Config
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		saved, err := h.cfg.UpdateConfig(incoming)
		if err == nil && h.sched != nil {
			err = h.sched.Reload()
		}

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
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
