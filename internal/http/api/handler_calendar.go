package api

import (
	"encoding/json"
	"net/http"

	"notion-notifier/internal/logging"
)

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
