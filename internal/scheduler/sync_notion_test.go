package scheduler

import (
	"testing"
	"time"
)

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
