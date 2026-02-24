package scheduler

import (
	"context"
	"errors"
	"time"

	"notion-notifier/internal/calendar"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/models"
)

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
			cfg := s.cfg.Config()
			if !cfg.CalendarSync.Enabled {
				continue
			}
			interval := time.Duration(cfg.CalendarSync.IntervalHours) * time.Hour
			ticker.Reset(interval)
			lookahead := cfg.CalendarSync.LookaheadDays
			if lookahead <= 0 {
				lookahead = 30
			}
			if _, err := s.SyncCalendar(time.Now(), time.Now().AddDate(0, 0, lookahead)); err != nil {
				logging.Error("CALENDAR", "calendar sync failed: %v", err)
			}
		}
	}
}

func (s *Scheduler) SyncCalendar(from, to time.Time) (int, error) {
	count := 0
	err := s.withRuntimeOp(calendarOpTimeout, func(ctx context.Context) error {
		var err error
		count, err = s.syncCalendar(ctx, from, to)
		return err
	})
	return count, err
}

func (s *Scheduler) syncCalendar(ctx context.Context, from, to time.Time) (int, error) {
	cfg, env := s.cfg.Snapshot()
	logging.Info("CALENDAR", "calendar sync started")
	if !cfg.CalendarSync.Enabled {
		logging.Info("CALENDAR", "calendar sync skipped (disabled)")
		return 0, nil
	}
	calendarOpts := calendar.ClientOptions{
		CalendarID:        env.Google.CalendarID,
		OAuthClientID:     env.Google.OAuthClientID,
		OAuthClientSecret: env.Google.OAuthClientSecret,
		OAuthRefreshToken: env.Google.OAuthRefreshToken,
	}
	if err := calendarOpts.Validate(); err != nil {
		s.mu.Lock()
		s.calendar = nil
		s.calendarFingerprint = ""
		s.mu.Unlock()
		logging.Error("CALENDAR", "calendar oauth config invalid: %v", err)
		return 0, err
	}
	fingerprint := calendarOpts.Fingerprint()

	s.mu.Lock()
	if s.calendar == nil || s.calendarFingerprint != fingerprint {
		client, err := calendar.NewClient(ctx, calendarOpts)
		if err != nil {
			s.mu.Unlock()
			logging.Error("CALENDAR", "calendar client init failed: %v", err)
			return 0, err
		}
		s.calendar = client
		s.calendarFingerprint = fingerprint
	}
	s.mu.Unlock()
	if s.calendar == nil {
		logging.Error("CALENDAR", "calendar client not configured")
		return 0, errors.New("calendar client not configured")
	}
	loc := loadLocationOrLocal(cfg.Timezone)

	dbEvents, err := s.repo.ListEventsBetween(ctx, from, to)
	if err != nil {
		logging.Error("CALENDAR", "list db events failed: %v", err)
		return 0, err
	}
	dbMap := make(map[string]models.Event, len(dbEvents))
	for _, ev := range dbEvents {
		dbMap[ev.NotionPageID] = ev
	}

	syncRecords, err := s.repo.ListSyncRecords(ctx)
	if err != nil {
		logging.Error("CALENDAR", "list sync records failed: %v", err)
		return 0, err
	}
	syncMap := make(map[string]models.SyncRecord, len(syncRecords))
	for _, rec := range syncRecords {
		syncMap[rec.NotionPageID] = rec
	}

	logging.Info("CALENDAR", "fetching calendar events from Google API (range: %s ~ %s)", from.Format("2006-01-02"), to.Format("2006-01-02"))
	calEvents, err := s.calendar.ListEvents(ctx, from, to)
	if err != nil {
		logging.Error("CALENDAR", "list calendar events failed: %v", err)
		return 0, err
	}
	logging.Info("CALENDAR", "fetched %d calendar events from Google API", len(calEvents))
	calGrouped := groupCalendarEvents(calEvents)

	count := 0

	for notionID, grouped := range calGrouped {
		ev, existsInDB := dbMap[notionID]
		if !existsInDB {
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
			!record.Attempted ||
			!record.Synced ||
			record.CalendarEventID != primary.ID ||
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
				Attempted:       true,
				Synced:          false,
			})
			continue
		}

		record = models.SyncRecord{
			NotionPageID:    notionID,
			CalendarEventID: newID,
			Attempted:       true,
			Synced:          true,
		}
		_ = s.repo.UpsertSyncRecord(ctx, record)
		syncMap[notionID] = record
		count++
	}

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
				Attempted:       true,
				Synced:          false,
			})
			continue
		}
		record := models.SyncRecord{
			NotionPageID:    ev.NotionPageID,
			CalendarEventID: newID,
			Attempted:       true,
			Synced:          true,
		}
		_ = s.repo.UpsertSyncRecord(ctx, record)
		syncMap[ev.NotionPageID] = record
		count++
	}

	orphans, err := s.repo.ListOrphanedSyncRecords(ctx)
	if err == nil {
		for _, rec := range orphans {
			if _, inCal := calGrouped[rec.NotionPageID]; !inCal {
				_ = s.repo.DeleteSyncRecord(ctx, rec.NotionPageID)
			}
		}
	}

	logging.Info("CALENDAR", "calendar sync finished (synced=%d, db_events=%d, cal_events=%d)", count, len(dbEvents), len(calEvents))
	return count, nil
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
		latest, _ := time.Parse(time.RFC3339, events[0].Updated)
		for i := 1; i < len(events); i++ {
			updated, _ := time.Parse(time.RFC3339, events[i].Updated)
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
