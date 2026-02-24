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

func TestMatchUpcomingConditions(t *testing.T) {
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
		rule config.UpcomingNotification
		want bool
	}{
		{
			name: "days empty and filters empty",
			rule: config.UpcomingNotification{
				Conditions: config.UpcomingConditions{
					DaysOfWeek:      nil,
					PropertyFilters: nil,
				},
			},
			want: true,
		},
		{
			name: "days mismatch",
			rule: config.UpcomingNotification{
				Conditions: config.UpcomingConditions{
					DaysOfWeek: []int{otherDay},
				},
			},
			want: false,
		},
		{
			name: "one filter mismatched",
			rule: config.UpcomingNotification{
				Conditions: config.UpcomingConditions{
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
			rule: config.UpcomingNotification{
				Conditions: config.UpcomingConditions{
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
			got := matchUpcomingConditions(event, start, tc.rule, cfg)
			if got != tc.want {
				t.Fatalf("matchUpcomingConditions() = %v, want %v", got, tc.want)
			}
		})
	}
}
