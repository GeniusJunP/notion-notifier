package scheduler

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"notion-notifier/internal/calendar"
	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
	"notion-notifier/internal/notion"
	tpl "notion-notifier/internal/template"
	"notion-notifier/internal/webhook"
)

const (
	notificationTypeUpcoming  = "upcoming"
	notificationTypePeriodic  = "periodic"
	notificationTypeManual    = "manual"
	notificationStatusSuccess = "success"
	notificationStatusFailed  = "failed"
	notificationStatusSkipped = "skipped"
	syncOpTimeout             = 2 * time.Minute
	calendarOpTimeout         = 3 * time.Minute
	rebuildOpTimeout          = 30 * time.Second
	upcomingFireTimeout       = 30 * time.Second
)

var errSchedulerNotRunning = errors.New("scheduler runtime is not running")

type Scheduler struct {
	cfg      *config.Manager
	repo     *db.Repository
	notion   *notion.Client
	webhook  *webhook.Client
	calendar *calendar.Client

	mu                  sync.Mutex
	upcomingTimers      map[string]*time.Timer
	periodicLastSent    map[int]string
	notionKey           string
	calendarFingerprint string
	statusMu            sync.RWMutex
	notionStatus        SyncStatus
	periodicMu          sync.Mutex
	syncMu              sync.Mutex // guards Notion sync operations
	calendarMu          sync.Mutex // guards Calendar sync operations
	runtimeMu           sync.RWMutex
	runtimeCtx          context.Context
	runtimeCancel       context.CancelFunc
	wg                  sync.WaitGroup
}

type SyncStatus struct {
	LastSyncedAt time.Time
	LastCount    int
	LastError    string
}

func New(cfg *config.Manager, repo *db.Repository, notionClient *notion.Client, webhookClient *webhook.Client, calendarClient *calendar.Client) *Scheduler {
	return &Scheduler{
		cfg:              cfg,
		repo:             repo,
		notion:           notionClient,
		webhook:          webhookClient,
		calendar:         calendarClient,
		upcomingTimers:   map[string]*time.Timer{},
		periodicLastSent: map[int]string{},
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.setRuntimeContext(ctx)

	s.wg.Add(1)
	go s.syncLoop()

	s.wg.Add(1)
	go s.periodicLoop()

	s.wg.Add(1)
	go s.calendarLoop()

	if err := s.SchedulePendingFromDB(); err != nil {
		logging.Error("SCHED", "schedule pending failed: %v", err)
	}
}

func (s *Scheduler) Stop() {
	s.cancelRuntime()
	s.wg.Wait()
	s.clearUpcomingTimers()
}

func (s *Scheduler) Reload() error {
	s.periodicMu.Lock()
	s.periodicLastSent = map[int]string{}
	s.periodicMu.Unlock()
	if _, err := s.runtimeContext(); err != nil {
		if errors.Is(err, errSchedulerNotRunning) {
			return nil
		}
		return err
	}
	return s.RebuildUpcomingSchedules()
}

func (s *Scheduler) SendManualNotification(ctx context.Context, template string, from, to time.Time) (string, error) {
	message, templateEvents, err := s.renderListFromRange(ctx, template, from, to)
	if err != nil {
		return "", err
	}
	if err := s.sendWebhook(ctx, notificationTypeManual, message, templateEvents, 0, ""); err != nil {
		return message, err
	}
	return message, nil
}

func (s *Scheduler) PreviewManualTemplate(ctx context.Context, template string, from, to time.Time) (string, error) {
	message, _, err := s.renderListFromRange(ctx, template, from, to)
	return message, err
}

func (s *Scheduler) renderListFromRange(ctx context.Context, template string, from, to time.Time) (string, []models.TemplateEvent, error) {
	cfg := s.cfg.Config()
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return "", nil, err
	}
	templateEvents := buildTemplateEvents(events, cfg.PropertyMap)
	message, err := tpl.RenderList(template, templateEvents)
	if err != nil {
		return "", nil, err
	}
	return message, templateEvents, nil
}

func (s *Scheduler) sendWebhook(ctx context.Context, typ, message string, events []models.TemplateEvent, minutesBefore int, notionPageID string) error {
	logging.Info("WBHK", "sending (%s)", typ)
	envCfg, env := s.cfg.Snapshot()
	now := time.Now()
	if config.IsSnoozed(envCfg, typ, now) {
		history := models.NotificationHistory{
			Type:         typ,
			Status:       notificationStatusSkipped,
			Message:      message,
			NotionPageID: notionPageID,
			Error:        "snoozed",
			SentAt:       now,
		}
		if err := s.repo.InsertNotificationHistory(ctx, history); err != nil {
			logging.Error("WBHK", "save notification history failed: %v", err)
		}
		logging.Info("WBHK", "send skipped (%s): snoozed", typ)
		return nil
	}
	payloadTarget, url := envCfg.Webhook.Notification, strings.TrimSpace(env.Webhook.NotificationURL)
	if envCfg.Webhook.IsTest {
		payloadTarget, url = envCfg.Webhook.InternalNotification, strings.TrimSpace(env.Webhook.InternalNotificationURL)
	}
	payloadCtx := models.WebhookPayloadContext{
		Type:          typ,
		Message:       message,
		Events:        events,
		MinutesBefore: minutesBefore,
	}
	if len(events) > 0 {
		payloadCtx.Event = events[0]
	}
	status, errStr := notificationStatusSuccess, ""
	if url == "" {
		status, errStr = notificationStatusFailed, "webhook url is empty"
	} else if s.webhook == nil {
		status, errStr = notificationStatusFailed, "webhook client not configured"
	} else if payload, err := tpl.RenderPayload(payloadTarget.PayloadTemplate, payloadCtx); err != nil {
		status, errStr = notificationStatusFailed, err.Error()
	} else if err := s.webhook.Send(ctx, url, payloadTarget.ContentType, []byte(payload)); err != nil {
		status, errStr = notificationStatusFailed, err.Error()
	}
	history := models.NotificationHistory{
		Type:         typ,
		Status:       status,
		Message:      message,
		NotionPageID: notionPageID,
		Error:        errStr,
		SentAt:       now,
	}
	if err := s.repo.InsertNotificationHistory(ctx, history); err != nil {
		logging.Error("WBHK", "save notification history failed: %v", err)
	}
	if status == notificationStatusFailed {
		logging.Error("WBHK", "send failed (%s): %s", typ, errStr)
		return errors.New(errStr)
	}
	logging.Info("WBHK", "send ok (%s)", typ)
	return nil
}
