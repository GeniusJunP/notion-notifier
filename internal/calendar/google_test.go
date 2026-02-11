package calendar

import (
	"reflect"
	"testing"
	"time"

	calendarapi "google.golang.org/api/calendar/v3"

	"notion-notifier/internal/models"
)

func TestEventMatchesNotion(t *testing.T) {
	loc := time.FixedZone("JST", 9*60*60)
	ev := models.Event{
		NotionPageID: "page-1",
		Title:        "Weekly Sync",
		StartDate:    "2026-02-11",
		StartTime:    "10:00",
		EndDate:      "2026-02-11",
		EndTime:      "11:00",
		Location:     "Room A",
		URL:          "https://notion.so/page-1",
		Content:      "Agenda",
		Attendees:    []string{"Bob@example.com", "alice@example.com"},
	}

	mapped := mapEvent(ev, loc)
	calEvent := CalendarEvent{
		NotionPageID:  ev.NotionPageID,
		Summary:       mapped.Summary,
		Location:      mapped.Location,
		Description:   mapped.Description,
		StartDate:     mapped.Start.Date,
		StartDateTime: mapped.Start.DateTime,
		EndDate:       mapped.End.Date,
		EndDateTime:   mapped.End.DateTime,
		Attendees:     []string{"alice@example.com", "bob@example.com"},
	}
	if !EventMatchesNotion(calEvent, ev, loc) {
		t.Fatalf("expected matching event to be treated as synced")
	}

	calEvent.Attendees = []string{"alice@example.com"}
	if EventMatchesNotion(calEvent, ev, loc) {
		t.Fatalf("expected attendee drift to be detected")
	}
}

func TestExtractEmails(t *testing.T) {
	got := extractEmails([]*calendarapi.EventAttendee{
		{Email: " Bob@example.com "},
		{Email: "alice@example.com"},
		{Email: "bob@example.com"},
		{Email: ""},
		nil,
	})
	want := []string{"alice@example.com", "bob@example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected emails: got=%v want=%v", got, want)
	}
}
