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
)

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
	stopCh           chan struct{}
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
		stopCh:           make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.syncLoop(ctx)

	s.wg.Add(1)
	go s.periodicLoop(ctx)

	s.wg.Add(1)
	go s.calendarLoop(ctx)

	if err := s.SchedulePendingFromDB(ctx); err != nil {
		log.Printf("schedule pending failed: %v", err)
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.clearAdvanceTimers()
}

func (s *Scheduler) Reload(ctx context.Context) error {
	s.periodicLastSent = map[int]string{}
	return s.RebuildAdvanceSchedules(ctx)
}

func (s *Scheduler) syncLoop(ctx context.Context) {
	defer s.wg.Done()
	_, _ = s.SyncNotion(ctx)
	for {
		cfg, _ := s.cfg.Get()
		interval := time.Duration(cfg.Sync.CheckInterval) * time.Minute
		ticker := time.NewTicker(interval)
		select {
		case <-s.stopCh:
			ticker.Stop()
			return
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			_, _ = s.SyncNotion(ctx)
		}
		ticker.Stop()
	}
}

func (s *Scheduler) periodicLoop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
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
				if s.periodicLastSent[i] == key {
					continue
				}
				err := s.sendPeriodic(ctx, now, i, rule)
				if err != nil {
					log.Printf("periodic notification failed: %v", err)
				}
				s.periodicLastSent[i] = key
			}
		}
	}
}

func (s *Scheduler) calendarLoop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
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
			if _, err := s.SyncCalendar(ctx, time.Now(), time.Now().AddDate(0, 0, lookahead)); err != nil {
				log.Printf("calendar sync failed: %v", err)
			}
		}
	}
}

func (s *Scheduler) SyncNotion(ctx context.Context) (int, error) {
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
	if err := s.RebuildAdvanceSchedules(ctx); err != nil {
		log.Printf("rebuild advance schedules failed: %v", err)
	}
	s.setNotionStatus(len(events), nil)
	logging.Info("SYNC", "notion sync finished (count=%d)", len(events))
	return len(events), nil
}

func (s *Scheduler) RebuildAdvanceSchedules(ctx context.Context) error {
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
	return s.SchedulePendingFromDB(ctx)
}

func (s *Scheduler) SchedulePendingFromDB(ctx context.Context) error {
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
		timer := time.AfterFunc(delay, func() {
			if err := s.fireAdvance(ctx, sched); err != nil {
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
	if err := s.sendWebhook(ctx, notificationTypeManual, message, templateEvents, 0, "", cfg, false); err != nil {
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
	payload, err := s.renderer.RenderPayload(cfg.Webhook.Notification.PayloadTemplate, payloadCtx)
	if err != nil {
		return message, "", err
	}
	return message, payload, nil
}

func (s *Scheduler) SyncCalendar(ctx context.Context, from, to time.Time) (int, error) {
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

	// 1. Get Notion events from DB cache
	dbEvents, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		logging.Error("CALENDAR", "list db events failed: %v", err)
		return 0, err
	}
	dbMap := make(map[string]models.Event, len(dbEvents))
	for _, ev := range dbEvents {
		dbMap[ev.NotionPageID] = ev
	}

	// 2. Get Calendar events from Google Calendar API
	logging.Info("CALENDAR", "fetching calendar events from Google API (range: %s ~ %s)", from.Format("2006-01-02"), to.Format("2006-01-02"))
	calEvents, err := s.calendar.ListEvents(ctx, from, to)
	if err != nil {
		logging.Error("CALENDAR", "list calendar events failed: %v", err)
		return 0, err
	}
	logging.Info("CALENDAR", "fetched %d calendar events from Google API", len(calEvents))
	calMap := make(map[string]calendar.CalendarEvent, len(calEvents))
	for _, ce := range calEvents {
		calMap[ce.NotionPageID] = ce
	}

	count := 0

	// 3. Create or update Calendar events for DB events
	for _, ev := range dbEvents {
		record, hasRecord, err := s.repo.GetSyncRecord(ctx, ev.NotionPageID)
		if err != nil {
			return count, err
		}

		existingCalID := ""
		if hasRecord {
			existingCalID = record.CalendarEventID
		} else if ce, inCal := calMap[ev.NotionPageID]; inCal {
			// Calendar has the event but no sync_record — adopt it
			existingCalID = ce.ID
		}

		// Skip if already synced, Notion hasn't changed, AND the calendar event still exists
		if hasRecord && record.Synced && record.NotionUpdatedAt == ev.NotionUpdatedAt {
			if _, stillExists := calMap[ev.NotionPageID]; stillExists {
				continue
			}
			// Calendar event was deleted externally — re-create it
			logging.Info("CALENDAR", "calendar event missing for %s, re-creating", ev.NotionPageID)
			existingCalID = ""
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
		_ = s.repo.UpsertSyncRecord(ctx, models.SyncRecord{
			NotionPageID:    ev.NotionPageID,
			CalendarEventID: newID,
			NotionUpdatedAt: ev.NotionUpdatedAt,
			Synced:          true,
		})
		count++
	}

	// 4. Delete Calendar events that no longer exist in Notion DB
	for notionID, ce := range calMap {
		if _, exists := dbMap[notionID]; !exists {
			if err := s.calendar.DeleteEvent(ctx, ce.ID); err != nil {
				logging.Error("CALENDAR", "calendar delete failed for %s: %v", ce.ID, err)
				continue
			}
			_ = s.repo.DeleteSyncRecord(ctx, notionID)
			logging.Info("CALENDAR", "deleted orphaned calendar event %s (notion: %s)", ce.ID, notionID)
			count++
		}
	}

	// 5. Clean up orphaned sync_records (no DB event and no Calendar event)
	orphans, err := s.repo.ListOrphanedSyncRecords(ctx)
	if err == nil {
		for _, rec := range orphans {
			if _, inCal := calMap[rec.NotionPageID]; !inCal {
				// Calendar event already gone, just clean up the record
				_ = s.repo.DeleteSyncRecord(ctx, rec.NotionPageID)
			}
		}
	}

	logging.Info("CALENDAR", "calendar sync finished (synced=%d, db_events=%d, cal_events=%d)", count, len(dbEvents), len(calEvents))
	return count, nil
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

// CalendarStatusMap returns calendar presence status for each Notion page ID.
// Status values: present | missing | disabled | unavailable
func (s *Scheduler) CalendarStatusMap(ctx context.Context, from, to time.Time, notionIDs []string) (map[string]string, error) {
	statuses := make(map[string]string, len(notionIDs))
	if len(notionIDs) == 0 {
		return statuses, nil
	}

	cfg, env := s.cfg.Get()
	if !cfg.CalendarSync.Enabled {
		for _, id := range notionIDs {
			statuses[id] = "disabled"
		}
		return statuses, nil
	}
	if env.Google.CalendarID == "" || env.Google.ServiceAccountKey == "" {
		for _, id := range notionIDs {
			statuses[id] = "unavailable"
		}
		return statuses, nil
	}

	s.mu.Lock()
	if s.calendar == nil || s.calendarKey != env.Google.ServiceAccountKey || s.calendarID != env.Google.CalendarID {
		client, err := calendar.NewClient(ctx, env.Google.CalendarID, env.Google.ServiceAccountKey)
		if err != nil {
			s.mu.Unlock()
			return statuses, err
		}
		s.calendar = client
		s.calendarKey = env.Google.ServiceAccountKey
		s.calendarID = env.Google.CalendarID
	}
	s.mu.Unlock()

	if s.calendar == nil {
		return statuses, errors.New("calendar client not configured")
	}

	calEvents, err := s.calendar.ListEvents(ctx, from, to)
	if err != nil {
		return statuses, err
	}
	present := make(map[string]struct{}, len(calEvents))
	for _, ce := range calEvents {
		if ce.NotionPageID == "" {
			continue
		}
		present[ce.NotionPageID] = struct{}{}
	}
	for _, id := range notionIDs {
		if _, ok := present[id]; ok {
			statuses[id] = "present"
		} else {
			statuses[id] = "missing"
		}
	}
	return statuses, nil
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
