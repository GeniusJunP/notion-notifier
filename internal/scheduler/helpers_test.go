package scheduler

import (
	"testing"

	"notion-notifier/internal/models"
)

func TestToTemplateEvent_MapsEndDateAndTime(t *testing.T) {
	ev := models.Event{
		Title:     "Deep Work",
		StartDate: "2026-02-13",
		StartTime: "09:00",
		EndDate:   "2026-02-14",
		EndTime:   "10:30",
	}
	got := toTemplateEvent(ev, map[string]string{})
	if got.EndDate != "2026-02-14" {
		t.Fatalf("unexpected end date: got=%s want=%s", got.EndDate, "2026-02-14")
	}
	if got.EndTime != "10:30" {
		t.Fatalf("unexpected end time: got=%s want=%s", got.EndTime, "10:30")
	}
}
