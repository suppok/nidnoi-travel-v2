package usecase

import (
	"testing"
	"time"
)

func TestNewSummaryRange(t *testing.T) {
	// 2026-11-15 10:00 in Bangkok, a Sunday.
	now := time.Date(2026, 11, 15, 3, 0, 0, 0, time.UTC)
	tests := []struct {
		days               int
		unit, from, prevFr string
		buckets            int
	}{
		{7, unitDay, "2026-11-09", "2026-11-02", 7},
		{30, unitDay, "2026-10-17", "2026-09-17", 30},
		{90, unitWeek, "2026-08-17", "2026-05-18", 13}, // 2026-08-18 is a Tuesday → Monday 08-17
		{365, unitMonth, "2025-11-01", "2024-10-17", 13},
	}
	for _, tt := range tests {
		r := newSummaryRange(now, tt.days)
		if r.Unit != tt.unit || r.From.Format(time.DateOnly) != tt.from || r.PreviousFrom.Format(time.DateOnly) != tt.prevFr {
			t.Errorf("days %d: got %s %s prev %s", tt.days, r.Unit, r.From.Format(time.DateOnly), r.PreviousFrom.Format(time.DateOnly))
		}
		if r.To.Format(time.DateOnly) != "2026-11-15" {
			t.Errorf("days %d: to %s", tt.days, r.To.Format(time.DateOnly))
		}
		if got := len(r.Buckets()); got != tt.buckets {
			t.Errorf("days %d: %d buckets, want %d", tt.days, got, tt.buckets)
		}
	}
}

func TestSummaryRangeInstants(t *testing.T) {
	// 23:30 on 2026-11-15 UTC is already 2026-11-16 in Bangkok.
	r := newSummaryRange(time.Date(2026, 11, 15, 23, 30, 0, 0, time.UTC), 7)
	cur := r.Current()
	if got := cur.From.UTC().Format(time.RFC3339); got != "2026-11-09T17:00:00Z" {
		t.Errorf("from %s", got)
	}
	if got := cur.To.UTC().Format(time.RFC3339); got != "2026-11-16T17:00:00Z" {
		t.Errorf("to %s", got)
	}
	if !r.Previous().To.Equal(cur.From) {
		t.Errorf("previous range must end where the current one starts")
	}
}
