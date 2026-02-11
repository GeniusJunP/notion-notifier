package notion

import (
	"reflect"
	"testing"
	"time"

	"notion-notifier/internal/config"
)

func TestMapPagesToEvents_MapsAttendees(t *testing.T) {
	loc := time.FixedZone("JST", 9*60*60)
	pages := []page{
		{
			ID:             "page-1",
			URL:            "https://notion.so/page-1",
			LastEditedTime: "2026-02-10T00:00:00Z",
			Properties: map[string]any{
				"Name": map[string]any{
					"type":  "title",
					"title": []any{map[string]any{"plain_text": "Team Meeting"}},
				},
				"Date": map[string]any{
					"type": "date",
					"date": map[string]any{
						"start": "2026-02-20T09:00:00+09:00",
						"end":   "2026-02-20T10:00:00+09:00",
					},
				},
				"Members": map[string]any{
					"type": "people",
					"people": []any{
						map[string]any{"person": map[string]any{"email": "first@example.com"}},
						map[string]any{"person": map[string]any{"email": "second@example.com"}},
					},
				},
			},
		},
	}
	mapping := config.PropertyMapping{
		Title:            "Name",
		Date:             "Date",
		Attendees:        "Members",
		AttendeesEnabled: true,
	}

	events := MapPagesToEvents(pages, mapping, loc)
	if len(events) != 1 {
		t.Fatalf("unexpected events length: got=%d want=1", len(events))
	}
	want := []string{"first@example.com", "second@example.com"}
	if !reflect.DeepEqual(events[0].Attendees, want) {
		t.Fatalf("unexpected attendees: got=%v want=%v", events[0].Attendees, want)
	}
}

func TestMapPagesToEvents_DisabledAttendees(t *testing.T) {
	loc := time.FixedZone("JST", 9*60*60)
	pages := []page{
		{
			ID:             "page-1",
			URL:            "https://notion.so/page-1",
			LastEditedTime: "2026-02-10T00:00:00Z",
			Properties: map[string]any{
				"Name": map[string]any{
					"type":  "title",
					"title": []any{map[string]any{"plain_text": "Team Meeting"}},
				},
				"Date": map[string]any{
					"type": "date",
					"date": map[string]any{
						"start": "2026-02-20T09:00:00+09:00",
					},
				},
				"Members": map[string]any{
					"type": "people",
					"people": []any{
						map[string]any{"person": map[string]any{"email": "first@example.com"}},
					},
				},
			},
		},
	}
	mapping := config.PropertyMapping{
		Title:            "Name",
		Date:             "Date",
		Attendees:        "Members",
		AttendeesEnabled: false,
	}

	events := MapPagesToEvents(pages, mapping, loc)
	if len(events) != 1 {
		t.Fatalf("unexpected events length: got=%d want=1", len(events))
	}
	if len(events[0].Attendees) != 0 {
		t.Fatalf("expected attendees to be empty when disabled, got=%v", events[0].Attendees)
	}
}
