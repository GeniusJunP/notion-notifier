package api

import (
	"encoding/json"
	"net/http"

	"notion-notifier/internal/config"
	"notion-notifier/internal/timeutil"
)

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

	// Upcoming template preview
	if req.MinutesBefore > 0 {
		message, err := h.sched.PreviewUpcomingTemplate(r.Context(), req.Template, req.MinutesBefore)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "preview failed: "+err.Error())
			return
		}
		respondJSON(w, http.StatusOK, previewResponse{Message: message})
		return
	}

	from, to, err := timeutil.ParseDateRange(req.FromDate, req.ToDate, h.cfg)
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

	from, to, err := timeutil.ParseDateRange(req.FromDate, req.ToDate, h.cfg)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date: "+err.Error())
		return
	}

	template := config.SanitizeTemplate(req.Template)

	message, err := h.sched.SendManualNotification(r.Context(), template, from, to)
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
