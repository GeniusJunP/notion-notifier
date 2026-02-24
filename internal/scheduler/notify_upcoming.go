package scheduler

import (
	"context"
	"errors"
	"log"
	"time"

	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
)

func (s *Scheduler) RebuildUpcomingSchedules() error {
	return s.withRuntimeOp(rebuildOpTimeout, s.rebuildUpcomingSchedules)
}

func (s *Scheduler) rebuildUpcomingSchedules(ctx context.Context) error {
	cfg := s.cfg.Config()
	loc := loadLocationOrLocal(cfg.Timezone)
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
	loc := loadLocationOrLocal(s.currentTimezone())
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
					logging.Info("SCHED", "upcoming notification skipped: %v", err)
					return
				}
				log.Printf("upcoming notification failed: %v", err)
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
		_ = s.repo.MarkUpcomingScheduleFired(ctx, sched.ID)
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
	loc := loadLocationOrLocal(cfg.Timezone)
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
