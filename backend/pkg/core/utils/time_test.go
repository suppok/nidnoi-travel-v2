package utils

import (
	"testing"
	"time"
)

func TestTodayInThailand(t *testing.T) {
	cases := map[string]string{
		"2026-11-03T16:59:59Z": "2026-11-03", // 23:59:59 in Bangkok
		"2026-11-03T17:00:00Z": "2026-11-04", // midnight in Bangkok
		"2026-12-31T20:00:00Z": "2027-01-01",
	}
	for in, want := range cases {
		now, _ := time.Parse(time.RFC3339, in)
		got := TodayInThailand(now)
		if got.Format(time.DateOnly) != want || got.Location() != time.UTC || got.Hour() != 0 {
			t.Errorf("TodayInThailand(%s) = %v, want %s 00:00 UTC", in, got, want)
		}
	}
}
