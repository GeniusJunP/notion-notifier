package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"notion-notifier/internal/models"
)

func TestUpsertEventsPersistsAttendees(t *testing.T) {
	repo, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer repo.Close()

	ev := models.Event{
		NotionPageID:    "page-1",
		Title:           "Team Meeting",
		StartDate:       "2026-02-20",
		EndDate:         "2026-02-20",
		IsAllDay:        true,
		RawPropsJSON:    "{}",
		NotionUpdatedAt: "2026-02-10T00:00:00Z",
		FetchedAt:       time.Now(),
		Attendees:       []string{"first@example.com", "second@example.com"},
	}
	if err := repo.UpsertEvents(context.Background(), []models.Event{ev}); err != nil {
		t.Fatalf("upsert events: %v", err)
	}

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)
	events, err := repo.ListEventsBetween(context.Background(), from, to)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("unexpected events len: got=%d want=1", len(events))
	}
	if !reflect.DeepEqual(events[0].Attendees, ev.Attendees) {
		t.Fatalf("unexpected attendees: got=%v want=%v", events[0].Attendees, ev.Attendees)
	}
}

func TestReplaceAdvanceSchedulesPreservesFiredForSameFireAt(t *testing.T) {
	repo, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	fireAt := time.Date(2026, 2, 12, 13, 59, 0, 0, time.UTC)
	sched := models.AdvanceSchedule{
		NotionPageID: "page-1",
		RuleIndex:    0,
		FireAt:       fireAt,
	}

	if err := repo.ReplaceAdvanceSchedules(ctx, []models.AdvanceSchedule{sched}); err != nil {
		t.Fatalf("replace schedules (initial): %v", err)
	}
	pending, err := repo.ListPendingAdvanceSchedules(ctx)
	if err != nil {
		t.Fatalf("list pending (initial): %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("unexpected pending len (initial): got=%d want=1", len(pending))
	}
	if err := repo.MarkAdvanceScheduleFired(ctx, pending[0].ID); err != nil {
		t.Fatalf("mark fired: %v", err)
	}

	if err := repo.ReplaceAdvanceSchedules(ctx, []models.AdvanceSchedule{sched}); err != nil {
		t.Fatalf("replace schedules (same fire_at): %v", err)
	}
	pending, err = repo.ListPendingAdvanceSchedules(ctx)
	if err != nil {
		t.Fatalf("list pending (same fire_at): %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("unexpected pending len after preserving fired: got=%d want=0", len(pending))
	}
}

func TestReplaceAdvanceSchedulesResetsFiredWhenFireAtChangesAndDeletesStale(t *testing.T) {
	repo, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	fireAtA := time.Date(2026, 2, 12, 13, 59, 0, 0, time.UTC)
	fireAtB := time.Date(2026, 2, 12, 14, 30, 0, 0, time.UTC)
	schedules := []models.AdvanceSchedule{
		{NotionPageID: "page-a", RuleIndex: 0, FireAt: fireAtA},
		{NotionPageID: "page-b", RuleIndex: 1, FireAt: fireAtB},
	}
	if err := repo.ReplaceAdvanceSchedules(ctx, schedules); err != nil {
		t.Fatalf("replace schedules (initial): %v", err)
	}
	pending, err := repo.ListPendingAdvanceSchedules(ctx)
	if err != nil {
		t.Fatalf("list pending (initial): %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("unexpected pending len (initial): got=%d want=2", len(pending))
	}

	var pageAID int64
	for _, p := range pending {
		if p.NotionPageID == "page-a" {
			pageAID = p.ID
			break
		}
	}
	if pageAID == 0 {
		t.Fatalf("page-a schedule id not found")
	}
	if err := repo.MarkAdvanceScheduleFired(ctx, pageAID); err != nil {
		t.Fatalf("mark fired (page-a): %v", err)
	}

	updatedA := models.AdvanceSchedule{
		NotionPageID: "page-a",
		RuleIndex:    0,
		FireAt:       fireAtA.Add(1 * time.Minute),
	}
	if err := repo.ReplaceAdvanceSchedules(ctx, []models.AdvanceSchedule{updatedA}); err != nil {
		t.Fatalf("replace schedules (updated): %v", err)
	}

	pending, err = repo.ListPendingAdvanceSchedules(ctx)
	if err != nil {
		t.Fatalf("list pending (updated): %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("unexpected pending len (updated): got=%d want=1", len(pending))
	}
	if pending[0].NotionPageID != "page-a" {
		t.Fatalf("unexpected pending schedule notion_page_id: got=%s want=page-a", pending[0].NotionPageID)
	}
	if !pending[0].FireAt.Equal(updatedA.FireAt) {
		t.Fatalf("unexpected fire_at: got=%s want=%s", pending[0].FireAt.Format(time.RFC3339), updatedA.FireAt.Format(time.RFC3339))
	}

	var count int
	row := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM advance_schedules WHERE notion_page_id = ?;`, "page-b")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("count stale schedules: %v", err)
	}
	if count != 0 {
		t.Fatalf("stale schedule was not deleted: count=%d", count)
	}
}

func TestReplaceAdvanceSchedulesClearsAllWhenEmpty(t *testing.T) {
	repo, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	fireAt := time.Date(2026, 2, 12, 10, 0, 0, 0, time.UTC)
	if err := repo.ReplaceAdvanceSchedules(ctx, []models.AdvanceSchedule{
		{NotionPageID: "page-1", RuleIndex: 0, FireAt: fireAt},
	}); err != nil {
		t.Fatalf("replace schedules (initial): %v", err)
	}
	if err := repo.ReplaceAdvanceSchedules(ctx, nil); err != nil {
		t.Fatalf("replace schedules (empty): %v", err)
	}

	var count int
	row := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM advance_schedules;`)
	if err := row.Scan(&count); err != nil && err != sql.ErrNoRows {
		t.Fatalf("count schedules: %v", err)
	}
	if count != 0 {
		t.Fatalf("schedules were not cleared: count=%d", count)
	}
}

func TestUpsertSyncRecordPersistsAttempted(t *testing.T) {
	repo, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.UpsertSyncRecord(ctx, models.SyncRecord{
		NotionPageID:    "page-1",
		CalendarEventID: "cal-1",
		Attempted:       true,
		Synced:          false,
	}); err != nil {
		t.Fatalf("upsert sync record: %v", err)
	}

	records, err := repo.ListSyncRecords(ctx)
	if err != nil {
		t.Fatalf("list sync records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("unexpected records len: got=%d want=1", len(records))
	}
	if !records[0].Attempted {
		t.Fatalf("attempted was not persisted: got=%v want=true", records[0].Attempted)
	}
	if records[0].Synced {
		t.Fatalf("synced was not persisted: got=%v want=false", records[0].Synced)
	}
}

func TestMigrateSyncRecordsAddsAttemptedFromLegacySchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)

	legacyDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	if _, err := legacyDB.Exec(`CREATE TABLE sync_records (
		notion_page_id TEXT PRIMARY KEY,
		calendar_event_id TEXT NOT NULL,
		synced INTEGER DEFAULT 0
	);`); err != nil {
		legacyDB.Close()
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := legacyDB.Exec(`INSERT INTO sync_records (notion_page_id, calendar_event_id, synced) VALUES
		('page-1', 'cal-1', 1),
		('page-2', '', 0);`); err != nil {
		legacyDB.Close()
		t.Fatalf("insert legacy records: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	repo, err := Open(path)
	if err != nil {
		t.Fatalf("open migrated db: %v", err)
	}
	defer repo.Close()

	records, err := repo.ListSyncRecords(context.Background())
	if err != nil {
		t.Fatalf("list sync records: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("unexpected records len: got=%d want=2", len(records))
	}
	for _, rec := range records {
		if !rec.Attempted {
			t.Fatalf("legacy record attempted should be true: notion_page_id=%s", rec.NotionPageID)
		}
	}
}
