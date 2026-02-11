package db

import (
	"context"
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
