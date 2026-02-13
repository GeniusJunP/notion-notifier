package scheduler

import (
	"testing"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/models"
)

func TestMatchesDays(t *testing.T) {
	weekday := weekdayToConfig(time.Monday)

	if !matchesDays(nil, weekday) {
		t.Fatalf("expected empty days to match")
	}
	if !matchesDays([]int{}, weekday) {
		t.Fatalf("expected empty days to match")
	}
	if !matchesDays([]int{weekday}, weekday) {
		t.Fatalf("expected configured weekday to match")
	}
	if matchesDays([]int{weekdayToConfig(time.Tuesday)}, weekday) {
		t.Fatalf("expected non-matching weekday to fail")
	}
}

func TestMatchAdvanceConditions(t *testing.T) {
	start := time.Date(2026, 2, 16, 9, 0, 0, 0, time.UTC)
	event := models.Event{
		Title:        "Weekly Review",
		Location:     "Room A",
		StartDate:    start.Format("2006-01-02"),
		StartTime:    start.Format("15:04"),
		RawPropsJSON: "{}",
	}
	cfg := config.Config{}
	weekday := weekdayToConfig(start.Weekday())
	otherDay := weekdayToConfig(time.Sunday)
	if otherDay == weekday {
		otherDay = weekdayToConfig(time.Saturday)
	}

	tests := []struct {
		name string
		rule config.AdvanceNotification
		want bool
	}{
		{
			name: "days empty and filters empty",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek:      nil,
					PropertyFilters: nil,
				},
			},
			want: true,
		},
		{
			name: "days mismatch",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek: []int{otherDay},
				},
			},
			want: false,
		},
		{
			name: "one filter mismatched",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek: []int{weekday},
					PropertyFilters: []config.PropertyFilter{
						{Property: "location", Operator: "eq", Value: "Room A"},
						{Property: "title", Operator: "eq", Value: "Another Title"},
					},
				},
			},
			want: false,
		},
		{
			name: "all filters matched",
			rule: config.AdvanceNotification{
				Conditions: config.AdvanceConditions{
					DaysOfWeek: []int{weekday},
					PropertyFilters: []config.PropertyFilter{
						{Property: "location", Operator: "eq", Value: "Room A"},
						{Property: "title", Operator: "eq", Value: "Weekly Review"},
					},
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchAdvanceConditions(event, tc.rule, cfg)
			if got != tc.want {
				t.Fatalf("matchAdvanceConditions() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNotionOnOrAfterDate_JSTEarlyMorningUsesPreviousUTCDate(t *testing.T) {
	loc := time.FixedZone("JST", 9*60*60)
	now := time.Date(2026, 2, 13, 3, 0, 0, 0, loc)

	got := notionOnOrAfterDate(now, loc)
	want := "2026-02-12"
	if got != want {
		t.Fatalf("notionOnOrAfterDate() = %s, want %s", got, want)
	}
}

func TestNotionOnOrAfterDate_PSTUsesSameUTCDate(t *testing.T) {
	loc := time.FixedZone("PST", -8*60*60)
	now := time.Date(2026, 2, 13, 3, 0, 0, 0, loc)

	got := notionOnOrAfterDate(now, loc)
	want := "2026-02-13"
	if got != want {
		t.Fatalf("notionOnOrAfterDate() = %s, want %s", got, want)
	}
}

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
