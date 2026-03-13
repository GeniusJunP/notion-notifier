package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/models"
	"notion-notifier/internal/retry"
	"notion-notifier/internal/webhook"

	"gopkg.in/yaml.v3"
)

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
			Upcoming: []config.UpcomingNotification{},
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

	sched := New(cfgMgr, repo, nil, webhook.New(nil, retry.Config{}), nil)
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

func TestSendManualNotificationRoutesByIsTest(t *testing.T) {
	var prodHits atomic.Int64
	prodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prodHits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer prodServer.Close()

	var internalHits atomic.Int64
	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalHits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer internalServer.Close()

	tests := []struct {
		name             string
		isTest           bool
		notificationURL  string
		internalNotifURL string
		wantProdHits     int64
		wantInternalHits int64
	}{
		{
			name:             "prod route uses notification_url",
			isTest:           false,
			notificationURL:  prodServer.URL,
			internalNotifURL: internalServer.URL,
			wantProdHits:     1,
			wantInternalHits: 0,
		},
		{
			name:             "test route uses internal_notification_url",
			isTest:           true,
			notificationURL:  prodServer.URL,
			internalNotifURL: internalServer.URL,
			wantProdHits:     0,
			wantInternalHits: 1,
		},
		{
			name:             "prod route fails when notification_url is empty",
			isTest:           false,
			notificationURL:  "",
			internalNotifURL: internalServer.URL,
			wantProdHits:     0,
			wantInternalHits: 0,
		},
		{
			name:             "test route fails when internal_notification_url is empty",
			isTest:           true,
			notificationURL:  prodServer.URL,
			internalNotifURL: "",
			wantProdHits:     0,
			wantInternalHits: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prodHits.Store(0)
			internalHits.Store(0)

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
					Upcoming: []config.UpcomingNotification{},
					Periodic: []config.PeriodicNotification{},
				},
				Webhook: config.WebhookConfig{
					IsTest: tc.isTest,
					Notification: config.WebhookTarget{
						ContentType:     "application/json",
						PayloadTemplate: `{"route":"prod","content":{{json .Message}}}`,
					},
					InternalNotification: config.WebhookTarget{
						ContentType:     "application/json",
						PayloadTemplate: `{"route":"internal","content":{{json .Message}}}`,
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
					NotificationURL:         tc.notificationURL,
					InternalNotificationURL: tc.internalNotifURL,
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

			sched := New(cfgMgr, repo, nil, webhook.New(nil, retry.Config{}), nil)
			loc, _ := time.LoadLocation("Asia/Tokyo")
			day := time.Now().In(loc).AddDate(0, 0, 1)
			from := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
			to := from.AddDate(0, 0, 1)
			if err := repo.UpsertEvents(context.Background(), []models.Event{
				{
					NotionPageID: "event-route-test",
					Title:        "Routing Test",
					StartDate:    from.Format("2006-01-02"),
					RawPropsJSON: "{}",
					FetchedAt:    time.Now(),
				},
			}); err != nil {
				t.Fatalf("upsert event: %v", err)
			}

			_, err = sched.SendManualNotification(context.Background(), "manual", from, to)
			if tc.wantProdHits == 0 && tc.wantInternalHits == 0 {
				if err == nil {
					t.Fatalf("SendManualNotification must fail when selected webhook url is empty")
				}
				if !strings.Contains(err.Error(), "webhook url is empty") {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err != nil {
				t.Fatalf("SendManualNotification failed: %v", err)
			}

			if got := prodHits.Load(); got != tc.wantProdHits {
				t.Fatalf("prod hits = %d, want %d", got, tc.wantProdHits)
			}
			if got := internalHits.Load(); got != tc.wantInternalHits {
				t.Fatalf("internal hits = %d, want %d", got, tc.wantInternalHits)
			}
		})
	}
}
