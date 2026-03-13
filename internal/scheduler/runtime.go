package scheduler

import (
	"context"
	"sync"
	"time"
)

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

func (s *Scheduler) runWithTimeout(mu *sync.Mutex, timeout time.Duration, fn func(context.Context) error) error {
	mu.Lock()
	defer mu.Unlock()
	opCtx, cancel, err := s.newRuntimeOpContext(timeout)
	if err != nil {
		return err
	}
	defer cancel()
	return fn(opCtx)
}

func (s *Scheduler) withSyncOp(timeout time.Duration, fn func(context.Context) error) error {
	return s.runWithTimeout(&s.syncMu, timeout, fn)
}

func (s *Scheduler) withCalendarOp(timeout time.Duration, fn func(context.Context) error) error {
	return s.runWithTimeout(&s.calendarMu, timeout, fn)
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

func (s *Scheduler) clearUpcomingTimers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, timer := range s.upcomingTimers {
		timer.Stop()
	}
	s.upcomingTimers = map[string]*time.Timer{}
}

func (s *Scheduler) currentTimezone() string {
	cfg := s.cfg.Config()
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
