package api

import (
	"net/http"

	"notion-notifier/internal/logging"
)

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
