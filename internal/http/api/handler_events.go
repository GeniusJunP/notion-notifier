package api

import (
	"net/http"
	"time"

	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
	"notion-notifier/internal/timeutil"
)

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
	loc := timeutil.LoadOrLocal(cfg.Timezone)
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
