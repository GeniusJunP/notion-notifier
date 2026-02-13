package notion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/retry"
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

func TestQueryDatabaseOnOrAfter_SendsDateFilter(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: got=%s want=%s", r.Method, http.MethodPost)
		}
		if !strings.Contains(r.URL.Path, "/databases/db-123/query") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[],"has_more":false}`))
	}))
	defer srv.Close()

	client := New(srv.Client(), "secret", retry.Config{})
	client.baseURL = srv.URL

	_, err := client.QueryDatabaseOnOrAfter(context.Background(), "db-123", "日付", "2026-02-12")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if gotBody["page_size"] != float64(100) {
		t.Fatalf("unexpected page_size: %#v", gotBody["page_size"])
	}
	filter, ok := gotBody["filter"].(map[string]any)
	if !ok {
		t.Fatalf("missing filter in request body: %#v", gotBody)
	}
	if filter["property"] != "日付" {
		t.Fatalf("unexpected filter.property: got=%v want=%v", filter["property"], "日付")
	}
	dateFilter, ok := filter["date"].(map[string]any)
	if !ok {
		t.Fatalf("missing filter.date: %#v", filter)
	}
	if dateFilter["on_or_after"] != "2026-02-12" {
		t.Fatalf("unexpected on_or_after: got=%v want=%v", dateFilter["on_or_after"], "2026-02-12")
	}
}

func TestQueryDatabase_NoFilterWhenDateConfigMissing(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[],"has_more":false}`))
	}))
	defer srv.Close()

	client := New(srv.Client(), "secret", retry.Config{})
	client.baseURL = srv.URL

	_, err := client.QueryDatabase(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if _, exists := gotBody["filter"]; exists {
		t.Fatalf("filter should be omitted for QueryDatabase(): %#v", gotBody)
	}
}
