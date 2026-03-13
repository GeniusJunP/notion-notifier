package api

import (
	"net/http"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/timeutil"
)

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
	loc := timeutil.LoadOrLocal(cfg.Timezone)
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
			nextSyncIn = timeutil.FormatDurationShort(next.Sub(now))
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
