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
	"notion-notifier/internal/discord"
	"notion-notifier/internal/models"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
	tpl "notion-notifier/internal/template"
)

const (
	notificationTypeAdvance = "advance"
	notificationTypeDaily   = "daily"
	notificationTypeWeekly  = "weekly"
	notificationTypeManual  = "manual"
)

type Scheduler struct {
	cfg       *config.Manager
	repo      *db.Repository
	notion    *notion.Client
	discord   *discord.Client
	calendar  *calendar.Client
	renderer  *tpl.Renderer

	mu             sync.Mutex
	advanceTimers  map[string]*time.Timer
	dailyLastSent  string
	weeklyLastSent map[int]string
	notionKey      string
	calendarKey    string
	calendarID     string
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

func New(cfg *config.Manager, repo *db.Repository, notionClient *notion.Client, discordClient *discord.Client, calendarClient *calendar.Client, renderer *tpl.Renderer) *Scheduler {
	return &Scheduler{
		cfg:            cfg,
		repo:           repo,
		notion:         notionClient,
		discord:        discordClient,
		calendar:       calendarClient,
		renderer:       renderer,
		advanceTimers:  map[string]*time.Timer{},
		weeklyLastSent: map[int]string{},
		stopCh:         make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.syncLoop(ctx)

	s.wg.Add(1)
	go s.dailyLoop(ctx)

	s.wg.Add(1)
	go s.weeklyLoop(ctx)

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
	s.dailyLastSent = ""
	s.weeklyLastSent = map[int]string{}
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

func (s *Scheduler) dailyLoop(ctx context.Context) {
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
			if !cfg.Notifications.Daily.Enabled {
				continue
			}
			loc, _ := time.LoadLocation(cfg.Timezone)
			now := time.Now().In(loc)
			if now.Format("15:04") != cfg.Notifications.Daily.Time {
				continue
			}
			today := now.Format("2006-01-02")
			if s.dailyLastSent == today {
				continue
			}
			err := s.sendDaily(ctx, now)
			if err != nil {
				log.Printf("daily notification failed: %v", err)
			}
			s.dailyLastSent = today
		}
	}
}

func (s *Scheduler) weeklyLoop(ctx context.Context) {
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
			for i, rule := range cfg.Notifications.Weekly {
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
				if s.weeklyLastSent[i] == key {
					continue
				}
				err := s.sendWeekly(ctx, now, i, rule)
				if err != nil {
					log.Printf("weekly notification failed: %v", err)
				}
				s.weeklyLastSent[i] = key
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
			if _, err := s.SyncCalendar(ctx, time.Now(), time.Now().AddDate(0, 0, 30)); err != nil {
				log.Printf("calendar sync failed: %v", err)
			}
		}
	}
}

func (s *Scheduler) SyncNotion(ctx context.Context) (int, error) {
	cfg, env := s.cfg.Get()
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
		return 0, errors.New("notion client not configured")
	}
	pages, err := s.notion.QueryDatabase(ctx, env.Notion.DatabaseID)
	if err != nil {
		return 0, err
	}
	events := notion.MapPagesToEvents(pages, cfg.PropertyMap, loc)
	if err := s.repo.UpsertEvents(ctx, events); err != nil {
		return 0, err
	}
	ids := make([]string, 0, len(events))
	for _, ev := range events {
		ids = append(ids, ev.NotionPageID)
	}
	if err := s.repo.DeleteEventsNotIn(ctx, ids); err != nil {
		return len(events), err
	}
	if err := s.RebuildAdvanceSchedules(ctx); err != nil {
		log.Printf("rebuild advance schedules failed: %v", err)
	}
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
	if err := s.sendDiscord(ctx, notificationTypeAdvance, message, event.NotionPageID, cfg, true); err != nil {
		_ = s.repo.MarkAdvanceScheduleFired(ctx, sched.ID)
		return err
	}
	_ = s.repo.MarkAdvanceScheduleFired(ctx, sched.ID)
	return nil
}

func (s *Scheduler) sendDaily(ctx context.Context, now time.Time) error {
	cfg, _ := s.cfg.Get()
	loc, _ := time.LoadLocation(cfg.Timezone)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := from
	if !cfg.Notifications.Daily.TodayOnly {
		to = from.AddDate(0, 0, 1)
	}
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return err
	}
	templateEvents := buildTemplateEvents(events, cfg.PropertyMap)
	message, err := s.renderer.RenderList(cfg.Notifications.Daily.Message, templateEvents)
	if err != nil {
		return err
	}
	return s.sendDiscord(ctx, notificationTypeDaily, message, "", cfg, true)
}

func (s *Scheduler) sendWeekly(ctx context.Context, now time.Time, idx int, rule config.WeeklyNotification) error {
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
	return s.sendDiscord(ctx, notificationTypeWeekly, message, "", cfg, true)
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
	if err := s.sendDiscord(ctx, notificationTypeManual, message, "", cfg, false); err != nil {
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

func (s *Scheduler) SyncCalendar(ctx context.Context, from, to time.Time) (int, error) {
	cfg, env := s.cfg.Get()
	if !cfg.CalendarSync.Enabled {
		return 0, nil
	}
	if env.Google.CalendarID != "" && env.Google.ServiceAccountKey != "" {
		s.mu.Lock()
		if s.calendar == nil || s.calendarKey != env.Google.ServiceAccountKey || s.calendarID != env.Google.CalendarID {
			client, err := calendar.NewClient(ctx, env.Google.CalendarID, env.Google.ServiceAccountKey)
			if err != nil {
				s.mu.Unlock()
				return 0, err
			}
			s.calendar = client
			s.calendarKey = env.Google.ServiceAccountKey
			s.calendarID = env.Google.CalendarID
		}
		s.mu.Unlock()
	}
	if s.calendar == nil {
		return 0, errors.New("calendar client not configured")
	}
	loc, _ := time.LoadLocation(cfg.Timezone)
	events, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, ev := range events {
		record, exists, err := s.repo.GetSyncRecord(ctx, ev.NotionPageID)
		if err != nil {
			return count, err
		}
		if exists && record.NotionUpdatedAt == ev.NotionUpdatedAt && record.SyncStatus == "synced" {
			continue
		}
		eventID := ""
		if exists {
			eventID = record.CalendarEventID
		}
		newID, updatedAt, err := s.calendar.UpsertEvent(ctx, ev, eventID, loc)
		if err != nil {
			_ = s.repo.UpsertSyncRecord(ctx, models.SyncRecord{
				NotionPageID:    ev.NotionPageID,
				CalendarEventID: eventID,
				NotionUpdatedAt: ev.NotionUpdatedAt,
				LastSyncedAt:    time.Now().In(loc).Format(time.RFC3339),
				SyncStatus:      "error",
			})
			return count, err
		}
		record = models.SyncRecord{
			NotionPageID:       ev.NotionPageID,
			CalendarEventID:    newID,
			NotionUpdatedAt:    ev.NotionUpdatedAt,
			CalendarUpdatedAt:  updatedAt,
			LastSyncedAt:       time.Now().In(loc).Format(time.RFC3339),
			SyncStatus:         "synced",
		}
		if err := s.repo.UpsertSyncRecord(ctx, record); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Scheduler) sendDiscord(ctx context.Context, typ, message, notionPageID string, cfg config.Config, scheduled bool) error {
	if config.IsMuted(cfg, time.Now()) {
		return nil
	}
	if scheduled && config.IsSnoozed(cfg, time.Now()) {
		return nil
	}
	webhook := ""
	_, env := s.cfg.Get()
	if scheduled {
		webhook = env.Discord.ScheduleWebhook
	} else {
		webhook = env.Discord.NotificationWebhook
	}
	status := "success"
	errStr := ""
	if err := s.discord.Send(ctx, webhook, message); err != nil {
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
		return errors.New(errStr)
	}
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
