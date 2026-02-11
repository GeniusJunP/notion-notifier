package scheduler

import (
	"reflect"
	"testing"

	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
)

func TestExtractAttendees(t *testing.T) {
	raw := `{
		"Members": {
			"type": "people",
			"people": [
				{"person": {"email": "first@example.com"}},
				{"person": {"email": "second@example.com"}}
			]
		}
	}`
	mapping := config.PropertyMapping{
		Attendees:        "Members",
		AttendeesEnabled: true,
	}

	got := extractAttendees(raw, mapping)
	want := []string{"first@example.com", "second@example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected attendees: got=%v want=%v", got, want)
	}
}

func TestHydrateEventAttendees(t *testing.T) {
	events := []models.Event{
		{
			RawPropsJSON: `{"Members":{"type":"people","people":[{"person":{"email":"first@example.com"}}]}}`,
		},
	}
	mapping := config.PropertyMapping{
		Attendees:        "Members",
		AttendeesEnabled: true,
	}

	hydrateEventAttendees(events, mapping)
	want := []string{"first@example.com"}
	if !reflect.DeepEqual(events[0].Attendees, want) {
		t.Fatalf("unexpected hydrated attendees: got=%v want=%v", events[0].Attendees, want)
	}
}
