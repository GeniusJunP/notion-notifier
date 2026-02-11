package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"notion-notifier/internal/calendar"
	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
	tpl "notion-notifier/internal/template"
	"notion-notifier/internal/webhook"
)

const (
	notificationTypeAdvance  = "advance"
	notificationTypePeriodic = "periodic"
	notificationTypeManual   = "manual"
	syncOpTimeout            = 2 * time.Minute
	calendarOpTimeout        = 3 * time.Minute
	rebuildOpTimeout         = 30 * time.Second
	advanceFireTimeout       = 30 * time.Second
)

var errSchedulerNotRunning = errors.New("scheduler runtime is not running")

type Scheduler struct {
	cfg      *config.Manager
	repo     *db.Repository
	notion   *notion.Client
	webhook  *webhook.Client
	calendar *calendar.Client
	renderer *tpl.Renderer

	mu               sync.Mutex
	advanceTimers    map[string]*time.Timer
	periodicLastSent map[int]string
	notionKey        string
	calendarKey      string
	calendarID       string
	statusMu         sync.RWMutex
	notionStatus     SyncStatus
	periodicMu       sync.Mutex
	opsMu            sync.Mutex
	runtimeMu        sync.RWMutex
	runtimeCtx       context.Context
	runtimeCancel    context.CancelFunc
	wg               sync.WaitGroup
}

type SyncStatus struct {
	LastSyncedAt time.Time
	LastCount    int
	LastError    string
}

func New(cfg *config.Manager, repo *db.Repository, notionClient *notion.Client, webhookClient *webhook.Client, calendarClient *calendar.Client, renderer *tpl.Renderer) *Scheduler {
	return &Scheduler{
		cfg:              cfg,
		repo:             repo,
		notion:           notionClient,
		webhook:          webhookClient,
		calendar:         calendarClient,
		renderer:         renderer,
		advanceTimers:    map[string]*time.Timer{},
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

	if err := s.SchedulePendingFromDB(context.Background()); err != nil {
		log.Printf("schedule pending failed: %v", err)
	}
}

func (s *Scheduler) Stop() {
	s.cancelRuntime()
	s.wg.Wait()
	s.clearAdvanceTimers()
}

func (s *Scheduler) Reload(_ context.Context) error {
	s.periodicMu.Lock()
	s.periodicLastSent = map[int]string{}
	s.periodicMu.Unlock()
	return s.RebuildAdvanceSchedules(context.Background())
}

func (s *Scheduler) syncLoop() {
	defer s.wg.Done()
	runtimeCtx, err := s.runtimeContext()
	if err != nil {
		return
	}
	_, _ = s.SyncNotion(context.Background())
	for {
		cfg, _ := s.cfg.Get()
		interval := time.Duration(cfg.Sync.CheckInterval) * time.Minute
		ticker := time.NewTicker(interval)
		select {
		case <-runtimeCtx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			_, _ = s.SyncNotion(context.Background())
		}
		ticker.Stop()
	}
}

func (s *Scheduler) periodicLoop() {
	defer s.wg.Done()
	runtimeCtx, err := s.runtimeContext()
	if err != nil {
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-runtimeCtx.Done():
			return
		case <-ticker.C:
			cfg, _ := s.cfg.Get()
			loc, _ := time.LoadLocation(cfg.Timezone)
			now := time.Now().In(loc)
			for i, rule := range cfg.Notifications.Periodic {
				if !rule.Enabled {
					continue
				}
				if now.Format("15:04") != rule.Time {
					continue
				}
				if !containsDay(rule.DaysOfWeek, weekdayToConfig(now.Weekday())) {
					continue
				}
				key := now.Format("2006-01-02")
				if s.periodicSent(i, key) {
					continue
				}
				opCtx, cancel, err := s.newRuntimeOpContext(advanceFireTimeout)
				if err != nil {
					return
				}
				err = s.sendPeriodic(opCtx, now, i, rule)
				cancel()
				if err != nil {
					log.Printf("periodic notification failed: %v", err)
				}
				s.markPeriodicSent(i, key)
			}
		}
	}
}

func (s *Scheduler) calendarLoop() {
	defer s.wg.Done()
	runtimeCtx, err := s.runtimeContext()
	if err != nil {
		return
	}

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-runtimeCtx.Done():
			return
		case <-ticker.C:
			cfg, _ := s.cfg.Get()
			if !cfg.CalendarSync.Enabled {
				continue
			}
			interval := time.Duration(cfg.CalendarSync.IntervalHours) * time.Hour
			ticker.Reset(interval)
			lookahead := cfg.CalendarSync.LookaheadDays
			if lookahead <= 0 {
				lookahead = 30
			}
			if _, err := s.SyncCalendar(context.Background(), time.Now(), time.Now().AddDate(0, 0, lookahead)); err != nil {
				log.Printf("calendar sync failed: %v", err)
			}
		}
	}
}

func (s *Scheduler) SyncNotion(_ context.Context) (int, error) {
	return s.withRuntimeIntOp(syncOpTimeout, s.syncNotion)
}

func (s *Scheduler) syncNotion(ctx context.Context) (int, error) {
	cfg, env := s.cfg.Get()
	logging.Info("SYNC", "notion sync started")
	loc, _ := time.LoadLocation(cfg.Timezone)
	if env.Notion.APIKey != "" {
		s.mu.Lock()
		if s.notion == nil || s.notionKey != env.Notion.APIKey {
			s.notion = notion.New(nil, env.Notion.APIKey, retry.Config{})
			s.notionKey = env.Notion.APIKey
		}
		s.mu.Unlock()
	}
	if s.notion == nil {
		err := errors.New("notion client not configured")
		s.setNotionStatus(0, err)
		logging.Error("SYNC", "notion client not configured")
		return 0, err
	}
	pages, err := s.notion.QueryDatabase(ctx, env.Notion.DatabaseID)
	if err != nil {
		s.setNotionStatus(0, err)
		logging.Error("SYNC", "notion query failed: %v", err)
		return 0, err
	}
	events := notion.MapPagesToEvents(pages, cfg.PropertyMap, loc)
	if cfg.ContentRules.StartHeading != "" && s.notion != nil {
		for i := range events {
			content, err := s.notion.FetchContent(ctx, events[i].NotionPageID, cfg.ContentRules)
			if err != nil {
				log.Printf("content extract failed for %s: %v", events[i].NotionPageID, err)
				continue
			}
			events[i].Content = content
		}
	}
	if err := s.repo.UpsertEvents(ctx, events); err != nil {
		s.setNotionStatus(0, err)
		logging.Error("SYNC", "upsert events failed: %v", err)
		return 0, err
	}
	ids := make([]string, 0, len(events))
	for _, ev := range events {
		ids = append(ids, ev.NotionPageID)
	}
	if err := s.repo.DeleteEventsNotIn(ctx, ids); err != nil {
		s.setNotionStatus(len(events), err)
		logging.Error("SYNC", "cleanup stale events failed: %v", err)
		return len(events), err
	}
	if err := s.rebuildAdvanceSchedules(ctx); err != nil {
		s.setNotionStatus(len(events), err)
		logging.Error("SYNC", "rebuild advance schedules failed: %v", err)
		return len(events), err
	}
	s.setNotionStatus(len(events), nil)
	logging.Info("SYNC", "notion sync finished (count=%d)", len(events))
	return len(events), nil
}

func (s *Scheduler) RebuildAdvanceSchedules(_ context.Context) error {
	err := s.withRuntimeErrOp(rebuildOpTimeout, func(ctx context.Context) error {
		return s.rebuildAdvanceSchedules(ctx)
	})
	return err
}

func (s *Scheduler) rebuildAdvanceSchedules(ctx context.Context) error {
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	events, err := s.repo.ListUpcomingEvents(ctx, 30, now)
	if err != nil {
		return err
	}
	schedules := buildAdvanceSchedules(events, cfg, now, loc)
	if err := s.repo.ReplaceAdvanceSchedules(ctx, schedules); err != nil {
		return err
	}
	return s.schedulePendingFromDB(ctx)
}

func (s *Scheduler) SchedulePendingFromDB(_ context.Context) error {
	err := s.withRuntimeErrOp(rebuildOpTimeout, func(ctx context.Context) error {
		return s.schedulePendingFromDB(ctx)
	})
	return err
}

func (s *Scheduler) schedulePendingFromDB(ctx context.Context) error {
	s.clearAdvanceTimers()
	loc, _ := time.LoadLocation(s.currentTimezone())
	now := time.Now().In(loc)
	schedules, err := s.repo.ListPendingAdvanceSchedules(ctx)
	if err != nil {
		return err
	}
	for _, sched := range schedules {
		delay := sched.FireAt.Sub(now)
		if delay < 0 {
			delay = 1 * time.Second
		}
		key := scheduleKey(sched.NotionPageID, sched.RuleIndex)
		sched := sched
		timer := time.AfterFunc(delay, func() {
			fireCtx, cancel, err := s.newRuntimeOpContext(advanceFireTimeout)
			if err != nil {
				return
			}
			defer cancel()
			if err := s.fireAdvance(fireCtx, sched); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					logging.Info("SCHED", "advance notification skipped: %v", err)
					return
				}
				log.Printf("advance notification failed: %v", err)
			}
		})
		s.mu.Lock()
		s.advanceTimers[key] = timer
		s.mu.Unlock()
	}
	return nil
}

func (s *Scheduler) fireAdvance(ctx context.Context, sched models.AdvanceSchedule) error {
	cfg, _ := s.cfg.Get()
	event, ok, err := s.repo.GetEvent(ctx, sched.NotionPageID)
	if err != nil {
		return err
	}
	if !ok {
		return s.repo.MarkAdvanceScheduleFired(ctx, sched.ID)
	}
	custom := extractCustomValues(event.RawPropsJSON, cfg.PropertyMap)
	templateEvent := toTemplateEvent(event, custom)
	rule := cfg.Notifications.Advance[sched.RuleIndex]
	message, err := s.renderer.RenderSingle(rule.Message, templateEvent, rule.MinutesBefore)
	if err != nil {
		_ = s.repo.MarkAdvanceScheduleFired(ctx, sched.ID)
		return err
	}
	if err := s.sendWebhook(ctx, notificationTypeAdvance, message, []models.TemplateEvent{templateEvent}, rule.MinutesBefore, event.NotionPageID, cfg, true); err != nil {
		_ = s.repo.MarkAdvanceScheduleFired(ctx, sched.ID)
		return err
	}
	_ = s.repo.MarkAdvanceScheduleFired(ctx, sched.ID)
	return nil
}

func (s *Scheduler) sendPeriodic(ctx context.Context, now time.Time, idx int, rule config.PeriodicNotification) error {
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, rule.DaysAhead)
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return err
	}
	templateEvents := buildTemplateEvents(events, cfg.PropertyMap)
	message, err := s.renderer.RenderList(rule.Message, templateEvents)
	if err != nil {
		return err
	}
	return s.sendWebhook(ctx, notificationTypePeriodic, message, templateEvents, 0, "", cfg, true)
}

func (s *Scheduler) SendManualNotification(ctx context.Context, template string, from, to time.Time) (string, error) {
	cfg, _ := s.cfg.Get()
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return "", err
	}
	templateEvents := buildTemplateEvents(events, cfg.PropertyMap)
	message, err := s.renderer.RenderList(template, templateEvents)
	if err != nil {
		return "", err
	}
	if err := s.sendWebhook(ctx, notificationTypeManual, message, templateEvents, 0, "", cfg, true); err != nil {
		return message, err
	}
	return message, nil
}

func (s *Scheduler) PreviewTemplate(ctx context.Context, template string, from, to time.Time) (string, error) {
	cfg, _ := s.cfg.Get()
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return "", err
	}
	templateEvents := buildTemplateEvents(events, cfg.PropertyMap)
	return s.renderer.RenderList(template, templateEvents)
}

func (s *Scheduler) PreviewAdvanceTemplate(ctx context.Context, template string, minutesBefore int) (string, error) {
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	events, err := s.repo.ListUpcomingEvents(ctx, 30, now)
	if err != nil {
		return "", err
	}
	if len(events) == 0 {
		return "（プレビュー対象の予定がありません）", nil
	}
	custom := extractCustomValues(events[0].RawPropsJSON, cfg.PropertyMap)
	templateEvent := toTemplateEvent(events[0], custom)
	return s.renderer.RenderSingle(template, templateEvent, minutesBefore)
}

func (s *Scheduler) PreviewManualPayload(ctx context.Context, template string, from, to time.Time) (string, string, error) {
	cfg, _ := s.cfg.Get()
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return "", "", err
	}
	templateEvents := buildTemplateEvents(events, cfg.PropertyMap)
	message, err := s.renderer.RenderList(template, templateEvents)
	if err != nil {
		return "", "", err
	}
	payloadCtx := models.WebhookPayloadContext{
		Type:    notificationTypeManual,
		Message: message,
		Events:  templateEvents,
	}
	if len(templateEvents) > 0 {
		payloadCtx.Event = templateEvents[0]
	}
	payload, err := s.renderer.RenderPayload(cfg.Webhook.Schedule.PayloadTemplate, payloadCtx)
	if err != nil {
		return message, "", err
	}
	return message, payload, nil
}

func (s *Scheduler) SyncCalendar(_ context.Context, from, to time.Time) (int, error) {
	return s.withRuntimeIntOp(calendarOpTimeout, func(ctx context.Context) (int, error) {
		return s.syncCalendar(ctx, from, to)
	})
}

func (s *Scheduler) syncCalendar(ctx context.Context, from, to time.Time) (int, error) {
	cfg, env := s.cfg.Get()
	logging.Info("CALENDAR", "calendar sync started")
	if !cfg.CalendarSync.Enabled {
		logging.Info("CALENDAR", "calendar sync skipped (disabled)")
		return 0, nil
	}
	if env.Google.CalendarID != "" && env.Google.ServiceAccountKey != "" {
		s.mu.Lock()
		if s.calendar == nil || s.calendarKey != env.Google.ServiceAccountKey || s.calendarID != env.Google.CalendarID {
			client, err := calendar.NewClient(ctx, env.Google.CalendarID, env.Google.ServiceAccountKey)
			if err != nil {
				s.mu.Unlock()
				logging.Error("CALENDAR", "calendar client init failed: %v", err)
				return 0, err
			}
			s.calendar = client
			s.calendarKey = env.Google.ServiceAccountKey
			s.calendarID = env.Google.CalendarID
		}
		s.mu.Unlock()
	}
	if s.calendar == nil {
		logging.Error("CALENDAR", "calendar client not configured")
		return 0, errors.New("calendar client not configured")
	}
	loc, _ := time.LoadLocation(cfg.Timezone)

	// 1. Fetch Notion cache from DB (source of truth data).
	dbEvents, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		logging.Error("CALENDAR", "list db events failed: %v", err)
		return 0, err
	}
	dbMap := make(map[string]models.Event, len(dbEvents))
	for _, ev := range dbEvents {
		dbMap[ev.NotionPageID] = ev
	}

	// 2. Load sync_records once and index by Notion page ID.
	syncRecords, err := s.repo.ListSyncRecords(ctx)
	if err != nil {
		logging.Error("CALENDAR", "list sync records failed: %v", err)
		return 0, err
	}
	syncMap := make(map[string]models.SyncRecord, len(syncRecords))
	for _, rec := range syncRecords {
		syncMap[rec.NotionPageID] = rec
	}

	// 3. Fetch tracked Calendar events in range.
	logging.Info("CALENDAR", "fetching calendar events from Google API (range: %s ~ %s)", from.Format("2006-01-02"), to.Format("2006-01-02"))
	calEvents, err := s.calendar.ListEvents(ctx, from, to)
	if err != nil {
		logging.Error("CALENDAR", "list calendar events failed: %v", err)
		return 0, err
	}
	logging.Info("CALENDAR", "fetched %d calendar events from Google API", len(calEvents))
	calGrouped := groupCalendarEvents(calEvents)

	count := 0

	// 4. Calendar-first pass: reverse lookup each tracked Calendar event into DB.
	for notionID, grouped := range calGrouped {
		ev, existsInDB := dbMap[notionID]
		if !existsInDB {
			// No Notion event in cache: this tracked Calendar event must be removed.
			count += s.deleteCalendarEvents(ctx, notionID, grouped)
			if err := s.repo.DeleteSyncRecord(ctx, notionID); err != nil {
				logging.Error("CALENDAR", "sync record delete failed for %s: %v", notionID, err)
			}
			delete(syncMap, notionID)
			continue
		}

		record, hasRecord := syncMap[notionID]
		primary, duplicates := pickPrimaryCalendarEvent(grouped, record, hasRecord)
		if len(duplicates) > 0 {
			count += s.deleteCalendarEvents(ctx, notionID, duplicates)
		}

		needsUpsert := !hasRecord ||
			!record.Synced ||
			record.CalendarEventID != primary.ID ||
			record.NotionUpdatedAt != ev.NotionUpdatedAt ||
			!calendar.EventMatchesNotion(primary, ev, loc)

		if !needsUpsert {
			continue
		}

		newID, _, err := s.calendar.UpsertEvent(ctx, ev, primary.ID, loc)
		if err != nil {
			logging.Error("CALENDAR", "calendar upsert failed for %s: %v", notionID, err)
			_ = s.repo.UpsertSyncRecord(ctx, models.SyncRecord{
				NotionPageID:    notionID,
				CalendarEventID: primary.ID,
				NotionUpdatedAt: ev.NotionUpdatedAt,
				Synced:          false,
			})
			continue
		}

		record = models.SyncRecord{
			NotionPageID:    notionID,
			CalendarEventID: newID,
			NotionUpdatedAt: ev.NotionUpdatedAt,
			Synced:          true,
		}
		_ = s.repo.UpsertSyncRecord(ctx, record)
		syncMap[notionID] = record
		count++
	}

	// 5. DB-first pass: create/fix events missing from fetched Calendar list.
	for _, ev := range dbEvents {
		if _, exists := calGrouped[ev.NotionPageID]; exists {
			continue
		}
		existingCalID := ""
		if rec, ok := syncMap[ev.NotionPageID]; ok {
			existingCalID = rec.CalendarEventID
		}
		newID, _, err := s.calendar.UpsertEvent(ctx, ev, existingCalID, loc)
		if err != nil {
			logging.Error("CALENDAR", "calendar upsert failed for %s: %v", ev.NotionPageID, err)
			_ = s.repo.UpsertSyncRecord(ctx, models.SyncRecord{
				NotionPageID:    ev.NotionPageID,
				CalendarEventID: existingCalID,
				NotionUpdatedAt: ev.NotionUpdatedAt,
				Synced:          false,
			})
			continue
		}
		record := models.SyncRecord{
			NotionPageID:    ev.NotionPageID,
			CalendarEventID: newID,
			NotionUpdatedAt: ev.NotionUpdatedAt,
			Synced:          true,
		}
		_ = s.repo.UpsertSyncRecord(ctx, record)
		syncMap[ev.NotionPageID] = record
		count++
	}

	// 6. Clean up orphaned sync_records (no DB event and no fetched Calendar event).
	orphans, err := s.repo.ListOrphanedSyncRecords(ctx)
	if err == nil {
		for _, rec := range orphans {
			if _, inCal := calGrouped[rec.NotionPageID]; !inCal {
				// Calendar event already gone, just clean up the record
				_ = s.repo.DeleteSyncRecord(ctx, rec.NotionPageID)
			}
		}
	}

	logging.Info("CALENDAR", "calendar sync finished (synced=%d, db_events=%d, cal_events=%d)", count, len(dbEvents), len(calEvents))
	return count, nil
}

func groupCalendarEvents(events []calendar.CalendarEvent) map[string][]calendar.CalendarEvent {
	grouped := make(map[string][]calendar.CalendarEvent, len(events))
	for _, ev := range events {
		grouped[ev.NotionPageID] = append(grouped[ev.NotionPageID], ev)
	}
	return grouped
}

func pickPrimaryCalendarEvent(events []calendar.CalendarEvent, record models.SyncRecord, hasRecord bool) (calendar.CalendarEvent, []calendar.CalendarEvent) {
	if len(events) == 0 {
		return calendar.CalendarEvent{}, nil
	}
	primaryIndex := 0
	if hasRecord && record.CalendarEventID != "" {
		for i, ev := range events {
			if ev.ID == record.CalendarEventID {
				primaryIndex = i
				break
			}
		}
	} else {
		latest := parseCalendarUpdated(events[0].Updated)
		for i := 1; i < len(events); i++ {
			updated := parseCalendarUpdated(events[i].Updated)
			if updated.After(latest) {
				latest = updated
				primaryIndex = i
			}
		}
	}
	primary := events[primaryIndex]
	duplicates := make([]calendar.CalendarEvent, 0, len(events)-1)
	for i, ev := range events {
		if i == primaryIndex {
			continue
		}
		duplicates = append(duplicates, ev)
	}
	return primary, duplicates
}

func parseCalendarUpdated(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (s *Scheduler) deleteCalendarEvents(ctx context.Context, notionID string, events []calendar.CalendarEvent) int {
	deleted := 0
	for _, ev := range events {
		if err := s.calendar.DeleteEvent(ctx, ev.ID); err != nil {
			logging.Error("CALENDAR", "calendar delete failed for %s: %v", ev.ID, err)
			continue
		}
		logging.Info("CALENDAR", "deleted orphaned/duplicate calendar event %s (notion: %s)", ev.ID, notionID)
		deleted++
	}
	return deleted
}

func (s *Scheduler) sendWebhook(ctx context.Context, typ, message string, events []models.TemplateEvent, minutesBefore int, notionPageID string, cfg config.Config, scheduled bool) error {
	if config.IsMuted(cfg, time.Now()) {
		return nil
	}
	if scheduled && config.IsSnoozed(cfg, time.Now()) {
		return nil
	}
	envCfg, env := s.cfg.Get()
	target := envCfg.Webhook.Notification
	url := env.Webhook.NotificationURL
	if scheduled {
		target = envCfg.Webhook.Schedule
		url = env.Webhook.ScheduleURL
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
	payload, err := s.renderer.RenderPayload(target.PayloadTemplate, payloadCtx)
	if err != nil {
		return err
	}
	status := "success"
	errStr := ""
	if err := s.webhook.Send(ctx, url, target.ContentType, []byte(payload)); err != nil {
		status = "failed"
		errStr = err.Error()
	}
	history := models.NotificationHistory{
		Type:         typ,
		Status:       status,
		Message:      message,
		NotionPageID: notionPageID,
		Error:        errStr,
		SentAt:       time.Now(),
	}
	_ = s.repo.InsertNotificationHistory(ctx, history)
	if status == "failed" {
		logging.Error("WEBHOOK", "send failed (%s): %s", typ, errStr)
		return errors.New(errStr)
	}
	logging.Info("WEBHOOK", "send ok (%s)", typ)
	return nil
}

func buildAdvanceSchedules(events []models.Event, cfg config.Config, now time.Time, loc *time.Location) []models.AdvanceSchedule {
	var schedules []models.AdvanceSchedule
	for _, ev := range events {
		startTime := parseEventStart(ev, loc)
		for idx, rule := range cfg.Notifications.Advance {
			if !rule.Enabled {
				continue
			}
			if !matchAdvanceConditions(ev, rule, cfg) {
				continue
			}
			fireAt := startTime.Add(-time.Duration(rule.MinutesBefore) * time.Minute)
			if fireAt.After(startTime) {
				continue
			}
			if fireAt.Before(now.Add(-5 * time.Minute)) {
				continue
			}
			schedules = append(schedules, models.AdvanceSchedule{
				NotionPageID: ev.NotionPageID,
				RuleIndex:    idx,
				FireAt:       fireAt,
			})
		}
	}
	return schedules
}

func parseEventStart(ev models.Event, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	if ev.StartDate == "" {
		return time.Now().In(loc)
	}
	if ev.StartTime == "" {
		t, err := time.ParseInLocation("2006-01-02", ev.StartDate, loc)
		if err != nil {
			return time.Now().In(loc)
		}
		return t
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", ev.StartDate+" "+ev.StartTime, loc)
	if err != nil {
		return time.Now().In(loc)
	}
	return t
}

func matchAdvanceConditions(ev models.Event, rule config.AdvanceNotification, cfg config.Config) bool {
	if !rule.Conditions.Enabled {
		return true
	}
	if len(rule.Conditions.DaysOfWeek) > 0 {
		start := parseEventStart(ev, nil)
		if !containsDay(rule.Conditions.DaysOfWeek, weekdayToConfig(start.Weekday())) {
			return false
		}
	}
	if len(rule.Conditions.PropertyFilters) == 0 {
		return true
	}
	values := buildFilterValues(ev, cfg)
	for _, filter := range rule.Conditions.PropertyFilters {
		val := values[filter.Property]
		if !matchFilter(val, filter.Operator, filter.Value) {
			return false
		}
	}
	return true
}

func buildFilterValues(ev models.Event, cfg config.Config) map[string]string {
	values := map[string]string{
		"title":    ev.Title,
		"location": ev.Location,
	}
	custom := extractCustomValues(ev.RawPropsJSON, cfg.PropertyMap)
	for k, v := range custom {
		values[k] = v
	}
	return values
}

func matchFilter(value, operator, expected string) bool {
	switch strings.ToLower(operator) {
	case "eq", "equals", "=":
		return value == expected
	case "neq", "not_equals", "!=":
		return value != expected
	case "contains":
		return strings.Contains(value, expected)
	case "not_contains":
		return !strings.Contains(value, expected)
	default:
		return false
	}
}

func buildTemplateEvents(events []models.Event, mapping config.PropertyMapping) []models.TemplateEvent {
	var out []models.TemplateEvent
	for _, ev := range events {
		custom := extractCustomValues(ev.RawPropsJSON, mapping)
		out = append(out, toTemplateEvent(ev, custom))
	}
	return out
}

func extractCustomValues(raw string, mapping config.PropertyMapping) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var props map[string]any
	if err := json.Unmarshal([]byte(raw), &props); err != nil {
		return map[string]string{}
	}
	custom := map[string]string{}
	for _, cm := range mapping.Custom {
		custom[cm.Variable] = notion.ExtractString(props[cm.Property])
	}
	return custom
}

func toTemplateEvent(ev models.Event, custom map[string]string) models.TemplateEvent {
	return models.TemplateEvent{
		Name:     ev.Title,
		Date:     ev.StartDate,
		Time:     ev.StartTime,
		EndTime:  ev.EndTime,
		IsAllDay: ev.IsAllDay,
		Location: ev.Location,
		URL:      ev.URL,
		Content:  ev.Content,
		Custom:   custom,
	}
}

func scheduleKey(notionPageID string, ruleIndex int) string {
	return notionPageID + ":" + strconv.Itoa(ruleIndex)
}

func containsDay(days []int, day int) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func weekdayToConfig(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}

func (s *Scheduler) setRuntimeContext(parent context.Context) {
	s.runtimeMu.Lock()
	defer s.runtimeMu.Unlock()
	if s.runtimeCancel != nil {
		s.runtimeCancel()
	}
	s.runtimeCtx, s.runtimeCancel = context.WithCancel(parent)
}

func (s *Scheduler) cancelRuntime() {
	s.runtimeMu.Lock()
	defer s.runtimeMu.Unlock()
	if s.runtimeCancel != nil {
		s.runtimeCancel()
	}
	s.runtimeCtx = nil
	s.runtimeCancel = nil
}

func (s *Scheduler) runtimeContext() (context.Context, error) {
	s.runtimeMu.RLock()
	defer s.runtimeMu.RUnlock()
	if s.runtimeCtx == nil {
		return nil, errSchedulerNotRunning
	}
	return s.runtimeCtx, nil
}

func (s *Scheduler) newRuntimeOpContext(timeout time.Duration) (context.Context, context.CancelFunc, error) {
	runtimeCtx, err := s.runtimeContext()
	if err != nil {
		return nil, nil, err
	}
	if timeout <= 0 {
		ctx, cancel := context.WithCancel(runtimeCtx)
		return ctx, cancel, nil
	}
	ctx, cancel := context.WithTimeout(runtimeCtx, timeout)
	return ctx, cancel, nil
}

func (s *Scheduler) withRuntimeErrOp(timeout time.Duration, fn func(context.Context) error) error {
	s.opsMu.Lock()
	defer s.opsMu.Unlock()
	opCtx, cancel, err := s.newRuntimeOpContext(timeout)
	if err != nil {
		return err
	}
	defer cancel()
	return fn(opCtx)
}

func (s *Scheduler) withRuntimeIntOp(timeout time.Duration, fn func(context.Context) (int, error)) (int, error) {
	s.opsMu.Lock()
	defer s.opsMu.Unlock()
	opCtx, cancel, err := s.newRuntimeOpContext(timeout)
	if err != nil {
		return 0, err
	}
	defer cancel()
	return fn(opCtx)
}

func (s *Scheduler) periodicSent(idx int, key string) bool {
	s.periodicMu.Lock()
	defer s.periodicMu.Unlock()
	return s.periodicLastSent[idx] == key
}

func (s *Scheduler) markPeriodicSent(idx int, key string) {
	s.periodicMu.Lock()
	defer s.periodicMu.Unlock()
	s.periodicLastSent[idx] = key
}

func (s *Scheduler) clearAdvanceTimers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, timer := range s.advanceTimers {
		timer.Stop()
	}
	s.advanceTimers = map[string]*time.Timer{}
}

func (s *Scheduler) currentTimezone() string {
	cfg, _ := s.cfg.Get()
	if cfg.Timezone == "" {
		return time.Local.String()
	}
	return cfg.Timezone
}

func (s *Scheduler) NotionSyncStatus() SyncStatus {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	return s.notionStatus
}

func (s *Scheduler) setNotionStatus(count int, err error) {
	status := SyncStatus{
		LastSyncedAt: time.Now(),
		LastCount:    count,
	}
	if err != nil {
		status.LastError = err.Error()
	}
	s.statusMu.Lock()
	s.notionStatus = status
	s.statusMu.Unlock()
}
