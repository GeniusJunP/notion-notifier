package api

import (
	"net/http"
	"time"

	"notion-notifier/internal/logging"
)

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
