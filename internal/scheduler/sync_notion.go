package scheduler

import (
	"context"
	"errors"
	"log"
	"time"

	"notion-notifier/internal/logging"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
)

func (s *Scheduler) syncLoop() {
	defer s.wg.Done()
	runtimeCtx, err := s.runtimeContext()
	if err != nil {
		return
	}
	_, _ = s.SyncNotion()
	for {
		cfg := s.cfg.Config()
		interval := time.Duration(cfg.Sync.CheckInterval) * time.Minute
		ticker := time.NewTicker(interval)
		select {
		case <-runtimeCtx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			_, _ = s.SyncNotion()
		}
		ticker.Stop()
	}
}

func (s *Scheduler) SyncNotion() (int, error) {
	count := 0
	err := s.withRuntimeOp(syncOpTimeout, func(ctx context.Context) error {
		var err error
		count, err = s.syncNotion(ctx)
		return err
	})
	return count, err
}

func (s *Scheduler) syncNotion(ctx context.Context) (int, error) {
	cfg, env := s.cfg.Snapshot()
	logging.Info("SYNC", "notion sync started")
	loc := loadLocationOrLocal(cfg.Timezone)
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
	fromDate := notionOnOrAfterDate(time.Now(), loc)
	pages, err := s.notion.QueryDatabaseOnOrAfter(ctx, env.Notion.DatabaseID, cfg.PropertyMap.Date, fromDate)
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
	if err := s.rebuildUpcomingSchedules(ctx); err != nil {
		s.setNotionStatus(len(events), err)
		logging.Error("SYNC", "rebuild advance schedules failed: %v", err)
		return len(events), err
	}
	s.setNotionStatus(len(events), nil)
	logging.Info("SYNC", "notion sync finished (count=%d)", len(events))
	return len(events), nil
}
