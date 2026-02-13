package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/models"
	"notion-notifier/internal/scheduler"
	tpl "notion-notifier/internal/template"
)

func TestHandleUpcomingEventsCalendarState(t *testing.T) {
	mux, repo, cfgMgr := setupAPIHandler(t, true)
	defer repo.Close()

	loc, _ := time.LoadLocation("Asia/Tokyo")
	start := time.Now().In(loc).AddDate(0, 0, 1)
	startDate := start.Format("2006-01-02")

	events := []models.Event{
		{NotionPageID: "event-needs-sync-no-record", Title: "A", StartDate: startDate, RawPropsJSON: "{}", FetchedAt: time.Now()},
		{NotionPageID: "event-needs-sync-not-attempted", Title: "B", StartDate: startDate, RawPropsJSON: "{}", FetchedAt: time.Now()},
		{NotionPageID: "event-synced", Title: "C", StartDate: startDate, RawPropsJSON: "{}", FetchedAt: time.Now()},
		{NotionPageID: "event-error", Title: "D", StartDate: startDate, RawPropsJSON: "{}", FetchedAt: time.Now()},
	}
	if err := repo.UpsertEvents(context.Background(), events); err != nil {
		t.Fatalf("upsert events: %v", err)
	}
	if err := repo.UpsertSyncRecord(context.Background(), models.SyncRecord{
		NotionPageID:    "event-needs-sync-not-attempted",
		CalendarEventID: "",
		Attempted:       false,
		Synced:          false,
	}); err != nil {
		t.Fatalf("upsert sync record (not attempted): %v", err)
	}
	if err := repo.UpsertSyncRecord(context.Background(), models.SyncRecord{
		NotionPageID:    "event-synced",
		CalendarEventID: "cal-1",
		Attempted:       true,
		Synced:          true,
	}); err != nil {
		t.Fatalf("upsert sync record (synced): %v", err)
	}
	if err := repo.UpsertSyncRecord(context.Background(), models.SyncRecord{
		NotionPageID:    "event-error",
		CalendarEventID: "",
		Attempted:       true,
		Synced:          false,
	}); err != nil {
		t.Fatalf("upsert sync record (error): %v", err)
	}

	states := fetchCalendarStates(t, mux)
	if got := states["event-needs-sync-no-record"]; got != "needs_sync" {
		t.Fatalf("unexpected state for no record: got=%q want=%q", got, "needs_sync")
	}
	if got := states["event-needs-sync-not-attempted"]; got != "needs_sync" {
		t.Fatalf("unexpected state for not attempted: got=%q want=%q", got, "needs_sync")
	}
	if got := states["event-synced"]; got != "synced" {
		t.Fatalf("unexpected state for synced: got=%q want=%q", got, "synced")
	}
	if got := states["event-error"]; got != "error" {
		t.Fatalf("unexpected state for error: got=%q want=%q", got, "error")
	}

	cfg, _ := cfgMgr.Get()
	cfg.CalendarSync.Enabled = false
	if err := cfgMgr.UpdateConfig(cfg); err != nil {
		t.Fatalf("disable calendar sync: %v", err)
	}

	states = fetchCalendarStates(t, mux)
	for notionID, state := range states {
		if state != "disabled" {
			t.Fatalf("unexpected state when disabled: notion_page_id=%s got=%q want=%q", notionID, state, "disabled")
		}
	}
}

func TestHandlePreviewNotificationReturnsMessageOnly(t *testing.T) {
	mux, repo, _ := setupAPIHandler(t, true)
	defer repo.Close()

	loc, _ := time.LoadLocation("Asia/Tokyo")
	start := time.Now().In(loc).AddDate(0, 0, 1)
	startDate := start.Format("2006-01-02")

	if err := repo.UpsertEvents(context.Background(), []models.Event{
		{
			NotionPageID: "event-1",
			Title:        "Planning",
			StartDate:    startDate,
			StartTime:    "10:00",
			RawPropsJSON: "{}",
			FetchedAt:    time.Now(),
		},
	}); err != nil {
		t.Fatalf("upsert event: %v", err)
	}

	manualReq := map[string]any{
		"template":  "count={{len .Events}}",
		"from_date": startDate,
		"to_date":   startDate,
	}
	manualResp := postJSON(t, mux, "/api/notifications/preview", manualReq)
	if _, ok := manualResp["payload"]; ok {
		t.Fatalf("manual preview response must not include payload field")
	}
	if _, ok := manualResp["message"]; !ok {
		t.Fatalf("manual preview response must include message field")
	}

	advanceReq := map[string]any{
		"template":       "{{.Name}}/{{.MinutesBefore}}",
		"minutes_before": 30,
	}
	advanceResp := postJSON(t, mux, "/api/notifications/preview", advanceReq)
	if _, ok := advanceResp["payload"]; ok {
		t.Fatalf("advance preview response must not include payload field")
	}
	if _, ok := advanceResp["message"]; !ok {
		t.Fatalf("advance preview response must include message field")
	}
}

func TestHandleDefaultTemplates(t *testing.T) {
	mux, repo, _ := setupAPIHandler(t, true)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/templates/defaults", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	advance, ok := payload["advance"]
	if !ok {
		t.Fatalf("response must include advance template")
	}
	periodic, ok := payload["periodic"]
	if !ok {
		t.Fatalf("response must include periodic template")
	}

	if !strings.Contains(advance, "## 予定リマインド！⏰") {
		t.Fatalf("advance template must include new heading")
	}
	if !strings.Contains(periodic, "## 今週の予定！📣") {
		t.Fatalf("periodic template must include new heading")
	}
	if strings.Contains(periodic, "次の予定に備えましょう！") {
		t.Fatalf("periodic template must not include removed phrase")
	}
}

func setupAPIHandler(t *testing.T, calendarEnabled bool) (*http.ServeMux, *db.Repository, *config.Manager) {
	t.Helper()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	envPath := filepath.Join(dir, "env.yaml")
	dbPath := filepath.Join(dir, "test.db")

	cfg := config.NormalizeConfig(config.Config{
		Timezone: "Asia/Tokyo",
		Sync: config.SyncConfig{
			CheckInterval: 15,
		},
		Notifications: config.Notifications{
			Advance:  []config.AdvanceNotification{},
			Periodic: []config.PeriodicNotification{},
		},
		Webhook: config.WebhookConfig{
			Schedule: config.WebhookTarget{
				ContentType:     "application/json",
				PayloadTemplate: `{"content":{{json .Message}}}`,
			},
			Notification: config.WebhookTarget{
				ContentType:     "application/json",
				PayloadTemplate: `{"content":{{json .Message}}}`,
			},
		},
		CalendarSync: config.CalendarSyncConfig{
			Enabled:       calendarEnabled,
			IntervalHours: 6,
			LookaheadDays: 30,
		},
		PropertyMap: config.PropertyMapping{
			Title:            "名前",
			Date:             "日付",
			Location:         "場所",
			Attendees:        "",
			AttendeesEnabled: false,
			Custom:           nil,
		},
		ContentRules: config.ContentRules{
			StartHeading:      "",
			IncludeStart:      false,
			StopAtNextHeading: true,
			StopAtDelimiter:   true,
			DelimiterText:     "",
		},
	})
	if err := config.WriteConfig(cfgPath, cfg); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(envPath, []byte{}, 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfgMgr, err := config.NewManager(cfgPath, envPath)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	repo, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sched := scheduler.New(cfgMgr, repo, nil, nil, nil, tpl.New())
	handler := NewHandler(cfgMgr, repo, sched)

	mux := http.NewServeMux()
	handler.Register(mux)
	return mux, repo, cfgMgr
}

func fetchCalendarStates(t *testing.T, mux *http.ServeMux) map[string]string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/events/upcoming", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result := make(map[string]string, len(payload))
	for _, row := range payload {
		notionID, _ := row["notion_page_id"].(string)
		state, _ := row["calendar_state"].(string)
		result[notionID] = state
	}
	return result
}

func postJSON(t *testing.T, mux *http.ServeMux, path string, body any) map[string]any {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}
