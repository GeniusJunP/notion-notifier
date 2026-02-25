package scheduler

import (
	"context"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/timezone"
)

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
			cfg := s.cfg.Config()
			loc := timezone.LoadOrLocal(cfg.Timezone)
			now := time.Now().In(loc)
			for i, rule := range cfg.Notifications.Periodic {
				if !rule.Enabled {
					continue
				}
				if now.Format("15:04") != rule.Time {
					continue
				}
				if !matchesDays(rule.DaysOfWeek, weekdayToConfig(now.Weekday())) {
					continue
				}
				key := now.Format("2006-01-02")
				if s.periodicSent(i, key) {
					continue
				}
				opCtx, cancel, err := s.newRuntimeOpContext(upcomingFireTimeout)
				if err != nil {
					return
				}
				err = s.sendPeriodic(opCtx, now, rule)
				cancel()
				if err != nil {
					logging.Error("PERIODIC", "periodic notification failed: %v", err)
				}
				s.markPeriodicSent(i, key)
			}
		}
	}
}

func (s *Scheduler) sendPeriodic(ctx context.Context, now time.Time, rule config.PeriodicNotification) error {
	cfg := s.cfg.Config()
	loc := timezone.LoadOrLocal(cfg.Timezone)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, rule.DaysAhead)
	message, templateEvents, err := s.renderListFromRange(ctx, rule.Message, from, to)
	if err != nil {
		return err
	}
	return s.sendWebhook(ctx, notificationTypePeriodic, message, templateEvents, 0, "")
}
