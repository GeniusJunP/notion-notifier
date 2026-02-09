package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"notion-notifier/internal/config"
)

func (s *Server) handleAPISync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	count, err := s.scheduler.SyncNotion(r.Context())
	if err != nil {
		http.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"type":"sync_complete","count":%d}}`, count))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current config
	currCfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(currCfg.Timezone)

	// Decode into a map first to see what fields are present
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Re-encode and decode into the struct to apply updates
	// This is a simple way to merge JSON into existing struct
	b, _ := json.Marshal(currCfg)
	var mergedCfg config.Config
	json.Unmarshal(b, &mergedCfg)

	// Since we want to support nested updates, we'll use a more targeted approach
	// For now, let's just handle the sections we have
	if notifications, ok := updates["notifications"]; ok {
		if m, ok := notifications.(map[string]interface{}); ok {
			if advance, ok := m["advance"]; ok {
				normalizeBoolSlice(advance, "enabled")
				if items, ok := advance.([]interface{}); ok {
					for _, item := range items {
						rule, ok := item.(map[string]interface{})
						if !ok {
							continue
						}
						if cond, ok := rule["conditions"].(map[string]interface{}); ok {
							normalizeBoolField(cond, "enabled")
						}
					}
				}
			}
			if periodic, ok := m["periodic"]; ok {
				normalizeBoolSlice(periodic, "enabled")
			}
		}
		nb, _ := json.Marshal(notifications)
		json.Unmarshal(nb, &mergedCfg.Notifications)
		if m, ok := notifications.(map[string]interface{}); ok {
			if isTruthy(m["advance_clear"]) {
				mergedCfg.Notifications.Advance = nil
			}
			if isTruthy(m["periodic_clear"]) {
				mergedCfg.Notifications.Periodic = nil
			}
			applyAdvanceConditionClears(&mergedCfg, m)
			applyPeriodicDayClears(&mergedCfg, m)
		}
	}
	if webhook, ok := updates["webhook"]; ok {
		wb, _ := json.Marshal(webhook)
		json.Unmarshal(wb, &mergedCfg.Webhook)
	}
	if calendarSync, ok := updates["calendar_sync"]; ok {
		if m, ok := calendarSync.(map[string]interface{}); ok {
			normalizeBoolField(m, "enabled")
		}
		cb, _ := json.Marshal(calendarSync)
		json.Unmarshal(cb, &mergedCfg.CalendarSync)
	}
	if propertyMapping, ok := updates["property_mapping"]; ok {
		pb, _ := json.Marshal(propertyMapping)
		json.Unmarshal(pb, &mergedCfg.PropertyMap)
		if m, ok := propertyMapping.(map[string]interface{}); ok {
			if isTruthy(m["custom_clear"]) {
				mergedCfg.PropertyMap.Custom = nil
			}
		}
	}
	if contentRules, ok := updates["content_rules"]; ok {
		if m, ok := contentRules.(map[string]interface{}); ok {
			normalizeBoolField(m, "include_start_heading")
			normalizeBoolField(m, "stop_at_next_heading")
			normalizeBoolField(m, "stop_at_delimiter")
		}
		cr, _ := json.Marshal(contentRules)
		json.Unmarshal(cr, &mergedCfg.ContentRules)
	}
	if sync, ok := updates["sync"]; ok {
		sb, _ := json.Marshal(sync)
		json.Unmarshal(sb, &mergedCfg.Sync)
	}
	if timezone, ok := updates["timezone"]; ok {
		if v, ok := timezone.(string); ok {
			mergedCfg.Timezone = v
		}
	}
	if snooze, ok := updates["snooze_until"]; ok {
		if v, ok := snooze.(string); ok {
			if normalized, err := normalizeDateInput(v, loc); err == nil {
				mergedCfg.SnoozeUntil = normalized
			}
		}
	}
	if mute, ok := updates["mute_until"]; ok {
		if v, ok := mute.(string); ok {
			if normalized, err := normalizeDateInput(v, loc); err == nil {
				mergedCfg.MuteUntil = normalized
			}
		}
	}
	// security.basic_auth.enabled is config-only (credentials are in env.yaml)

	if err := s.cfg.UpdateConfig(mergedCfg); err != nil {
		http.Error(w, "Failed to update config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"showToast":{"type":"config_saved"}}`)
	w.WriteHeader(http.StatusNoContent)
}

type notificationRequest struct {
	Template       string `json:"template"`
	FromDate       string `json:"from_date"`
	ToDate         string `json:"to_date"`
	MinutesBefore  int    `json:"minutes_before"`
	PreviewPayload bool   `json:"preview_payload"`
}

func (s *Server) handleAPIPreviewNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req notificationRequest
	if err := decodeNotificationRequest(r, &req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	from, to, err := parseDateRange(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		http.Error(w, "Invalid date: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.MinutesBefore > 0 {
		message, err := s.scheduler.PreviewAdvanceTemplate(r.Context(), req.Template, req.MinutesBefore)
		if err != nil {
			http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writePreviewHTML(w, message, "")
		return
	}
	if req.PreviewPayload {
		message, payload, err := s.scheduler.PreviewManualPayload(r.Context(), req.Template, from, to)
		if err != nil {
			http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writePreviewHTML(w, message, payload)
		return
	}
	message, err := s.scheduler.PreviewTemplate(r.Context(), req.Template, from, to)
	if err != nil {
		http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writePreviewHTML(w, message, "")
}

func (s *Server) handleAPIManualNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req notificationRequest
	if err := decodeNotificationRequest(r, &req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	from, to, err := parseDateRange(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		http.Error(w, "Invalid date: "+err.Error(), http.StatusBadRequest)
		return
	}
	message, err := s.scheduler.SendManualNotification(r.Context(), req.Template, from, to)
	if err != nil {
		http.Error(w, "Send failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", `{"showToast":{"type":"manual_sent"}}`)
	writePreviewHTML(w, message, "")
}

type calendarSyncRequest struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

func (s *Server) handleAPICalendarSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req calendarSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	from, to, err := parseDateRange(req.FromDate, req.ToDate, s.cfg)
	if err != nil {
		http.Error(w, "Invalid date: "+err.Error(), http.StatusBadRequest)
		return
	}
	count, err := s.scheduler.SyncCalendar(r.Context(), from, to)
	if err != nil {
		http.Error(w, "Calendar sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"type":"calendar_sync_complete","count":%d}}`, count))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPICalendarClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repo.ClearSyncRecords(r.Context()); err != nil {
		http.Error(w, "Failed to clear sync records: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", `{"showToast":{"type":"sync_records_cleared"}}`)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIHistoryClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repo.ClearNotificationHistory(r.Context()); err != nil {
		http.Error(w, "Failed to clear history: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", `{"showToast":{"type":"history_cleared"}}`)
	w.WriteHeader(http.StatusNoContent)
}
