package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/models"
	"notion-notifier/internal/retry"
	tpl "notion-notifier/internal/template"
	"notion-notifier/internal/webhook"

	"gopkg.in/yaml.v3"
)

func TestMatchesDays(t *testing.T) {
	weekday := weekdayToConfig(time.Monday)

	if !matchesDays(nil, weekday) {
		t.Fatalf("expected empty days to match")
	}
	if !matchesDays([]int{}, weekday) {
		t.Fatalf("expected empty days to match")
	}
	if !matchesDays([]int{weekday}, weekday) {
		t.Fatalf("expected configured weekday to match")
	}
	if matchesDays([]int{weekdayToConfig(time.Tuesday)}, weekday) {
		t.Fatalf("expected non-matching weekday to fail")
	}
}

func TestMatchAdvanceConditions(t *testing.T) {
	start := time.Date(2026, 2, 16, 9, 0, 0, 0, time.UTC)
	event := models.Event{
		Title:        "Weekly Review",
		Location:     "Room A",
		StartDate:    start.Format("2006-01-02"),
		StartTime:    start.Format("15:04"),
		RawPropsJSON: "{}",
	}
	cfg := config.Config{}
	weekday := weekdayToConfig(start.Weekday())
	otherDay := weekdayToConfig(time.Sunday)
	if otherDay == weekday {
		otherDay = weekdayToConfig(time.Saturday)
	}

	tests := []struct {
		name string
		rule config.AdvanceNotification
		want bool
	}{
		{
			name: "days empty and filters empty",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek:      nil,
					PropertyFilters: nil,
				},
			},
			want: true,
		},
		{
			name: "days mismatch",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek: []int{otherDay},
				},
			},
			want: false,
		},
		{
			name: "one filter mismatched",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek: []int{weekday},
					PropertyFilters: []config.PropertyFilter{
						{Property: "location", Operator: "eq", Value: "Room A"},
						{Property: "title", Operator: "eq", Value: "Another Title"},
					},
				},
			},
			want: false,
		},
		{
			name: "all filters matched",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek: []int{weekday},
					PropertyFilters: []config.PropertyFilter{
						{Property: "location", Operator: "eq", Value: "Room A"},
						{Property: "title", Operator: "eq", Value: "Weekly Review"},
					},
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchAdvanceConditions(event, start, tc.rule, cfg)
			if got != tc.want {
				t.Fatalf("matchAdvanceConditions() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNotionOnOrAfterDate_JSTEarlyMorningUsesPreviousUTCDate(t *testing.T) {
	loc := time.FixedZone("JST", 9*60*60)
	now := time.Date(2026, 2, 13, 3, 0, 0, 0, loc)

	got := notionOnOrAfterDate(now, loc)
	want := "2026-02-12"
	if got != want {
		t.Fatalf("notionOnOrAfterDate() = %s, want %s", got, want)
	}
}

func TestNotionOnOrAfterDate_PSTUsesSameUTCDate(t *testing.T) {
	loc := time.FixedZone("PST", -8*60*60)
	now := time.Date(2026, 2, 13, 3, 0, 0, 0, loc)

	got := notionOnOrAfterDate(now, loc)
	want := "2026-02-13"
	if got != want {
		t.Fatalf("notionOnOrAfterDate() = %s, want %s", got, want)
	}
}

func TestToTemplateEvent_MapsEndDateAndTime(t *testing.T) {
	ev := models.Event{
		Title:     "Deep Work",
		StartDate: "2026-02-13",
		StartTime: "09:00",
		EndDate:   "2026-02-14",
		EndTime:   "10:30",
	}
	got := toTemplateEvent(ev, map[string]string{})
	if got.EndDate != "2026-02-14" {
		t.Fatalf("unexpected end date: got=%s want=%s", got.EndDate, "2026-02-14")
	}
	if got.EndTime != "10:30" {
		t.Fatalf("unexpected end time: got=%s want=%s", got.EndTime, "10:30")
	}
}

func TestSendWebhookRecordsHistoryOnPayloadRenderError(t *testing.T) {
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
			Notification: config.WebhookTarget{
				ContentType:     "application/json",
				PayloadTemplate: "{{if",
			},
			InternalNotification: config.WebhookTarget{
				ContentType:     "application/json",
				PayloadTemplate: `{"content":{{json .Message}}}`,
			},
		},
		CalendarSync: config.CalendarSyncConfig{
			Enabled:       false,
			IntervalHours: 6,
			LookaheadDays: 30,
		},
	})
	if err := config.WriteConfig(cfgPath, cfg); err != nil {
		t.Fatalf("write config: %v", err)
	}
	envData, err := yaml.Marshal(config.Env{
		Webhook: config.WebhookEnv{
			NotificationURL: "http://127.0.0.1:1",
		},
	})
	if err != nil {
		t.Fatalf("marshal env: %v", err)
	}
	if err := os.WriteFile(envPath, envData, 0o644); err != nil {
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
	defer repo.Close()

	sched := New(cfgMgr, repo, nil, webhook.New(nil, retry.Config{}), nil, tpl.New())
	loc, _ := time.LoadLocation("Asia/Tokyo")
	day := time.Now().In(loc).AddDate(0, 0, 1)
	from := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, 1)
	if err := repo.UpsertEvents(context.Background(), []models.Event{
		{
			NotionPageID: "event-1",
			Title:        "Planning",
			StartDate:    from.Format("2006-01-02"),
			RawPropsJSON: "{}",
			FetchedAt:    time.Now(),
		},
	}); err != nil {
		t.Fatalf("upsert event: %v", err)
	}

	if _, err := sched.SendManualNotification(context.Background(), "manual", from, to); err == nil {
		t.Fatalf("SendManualNotification must fail on payload render error")
	}

	history, err := repo.ListNotificationHistory(context.Background(), 10)
	if err != nil {
		t.Fatalf("list history: %v", err)
	}
	if len(history) == 0 {
		t.Fatalf("notification history must include failed send")
	}
	if history[0].Type != "manual" {
		t.Fatalf("unexpected history type: got=%q want=%q", history[0].Type, "manual")
	}
	if history[0].Status != "failed" {
		t.Fatalf("unexpected history status: got=%q want=%q", history[0].Status, "failed")
	}
	if strings.TrimSpace(history[0].Error) == "" {
		t.Fatalf("failed history must include error text")
	}
}
