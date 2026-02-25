package scheduler

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
	"notion-notifier/internal/timezone"
)

func (s *Scheduler) RebuildUpcomingSchedules() error {
	return s.withRuntimeOp(rebuildOpTimeout, s.rebuildUpcomingSchedules)
}

func (s *Scheduler) rebuildUpcomingSchedules(ctx context.Context) error {
	cfg := s.cfg.Config()
	loc := timezone.LoadOrLocal(cfg.Timezone)
	now := time.Now().In(loc)
	events, err := s.repo.ListUpcomingEvents(ctx, 30, now)
	if err != nil {
		return err
	}
	schedules := buildUpcomingSchedules(events, cfg, now, loc)
	if err := s.repo.ReplaceUpcomingSchedules(ctx, schedules); err != nil {
		return err
	}
	return s.schedulePendingFromDB(ctx)
}

func (s *Scheduler) SchedulePendingFromDB() error {
	return s.withRuntimeOp(rebuildOpTimeout, s.schedulePendingFromDB)
}

func (s *Scheduler) schedulePendingFromDB(ctx context.Context) error {
	s.clearUpcomingTimers()
	loc := timezone.LoadOrLocal(s.currentTimezone())
	now := time.Now().In(loc)
	schedules, err := s.repo.ListPendingUpcomingSchedules(ctx)
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
			fireCtx, cancel, err := s.newRuntimeOpContext(upcomingFireTimeout)
			if err != nil {
				return
			}
			defer cancel()
			if err := s.fireUpcoming(fireCtx, sched); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					logging.Info("UPCOMING", "upcoming notification skipped: %v", err)
					return
				}
				logging.Error("UPCOMING", "upcoming notification failed: %v", err)
			}
		})
		s.mu.Lock()
		s.upcomingTimers[key] = timer
		s.mu.Unlock()
	}
	return nil
}

func (s *Scheduler) fireUpcoming(ctx context.Context, sched models.UpcomingSchedule) error {
	cfg := s.cfg.Config()
	event, ok, err := s.repo.GetEvent(ctx, sched.NotionPageID)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.repo.MarkUpcomingScheduleFired(ctx, sched.ID); err != nil {
			logging.Error("SCHED", "mark upcoming schedule fired failed (id=%d): %v", sched.ID, err)
		}
	}()
	if !ok {
		return nil
	}
	custom := extractCustomValues(event.RawPropsJSON, cfg.PropertyMap)
	templateEvent := toTemplateEvent(event, custom)
	rule := cfg.Notifications.Upcoming[sched.RuleIndex]
	message, err := s.renderer.RenderSingle(rule.Message, templateEvent, rule.MinutesBefore)
	if err != nil {
		return err
	}
	if err := s.sendWebhook(ctx, notificationTypeUpcoming, message, []models.TemplateEvent{templateEvent}, rule.MinutesBefore, event.NotionPageID); err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) PreviewUpcomingTemplate(ctx context.Context, template string, minutesBefore int) (string, error) {
	cfg := s.cfg.Config()
	loc := timezone.LoadOrLocal(cfg.Timezone)
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

func buildUpcomingSchedules(events []models.Event, cfg config.Config, now time.Time, loc *time.Location) []models.UpcomingSchedule {
	var schedules []models.UpcomingSchedule
	for _, ev := range events {
		startTime := parseEventStart(ev, loc)
		for idx, rule := range cfg.Notifications.Upcoming {
			if !rule.Enabled {
				continue
			}
			if !matchUpcomingConditions(ev, startTime, rule, cfg) {
				continue
			}
			fireAt := startTime.Add(-time.Duration(rule.MinutesBefore) * time.Minute)
			if fireAt.After(startTime) {
				continue
			}
			if fireAt.Before(now.Add(-5 * time.Minute)) {
				continue
			}
			schedules = append(schedules, models.UpcomingSchedule{
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

func matchUpcomingConditions(ev models.Event, start time.Time, rule config.UpcomingNotification, cfg config.Config) bool {
	if !matchesDays(rule.Conditions.DaysOfWeek, weekdayToConfig(start.Weekday())) {
		return false
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

func scheduleKey(notionPageID string, ruleIndex int) string {
	return notionPageID + ":" + strconv.Itoa(ruleIndex)
}

func weekdayToConfig(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}

func matchesDays(days []int, weekday int) bool {
	if len(days) == 0 {
		return true
	}
	for _, day := range days {
		if day == weekday {
			return true
		}
	}
	return false
}
