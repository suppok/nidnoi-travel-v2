package usecase

import (
	"time"

	"backend/pkg/core/utils"
	"backend/pkg/domain/repomodels"
)

// Bucket units, also the date_trunc field names the repositories group by.
const (
	unitDay   = "day"
	unitWeek  = "week"
	unitMonth = "month"
)

// summaryRange is the summarised period in Thailand dates. Dates are midnight UTC (the form DATE
// columns compare as); Instants turns them into the Thailand-midnight instants timestamps compare to.
type summaryRange struct {
	Unit         string
	From, To     time.Time // To is today, inclusive
	PreviousFrom time.Time // the previous period is [PreviousFrom, From)
}

// unitFor picks the chart bucket for a period length.
func unitFor(days int) string {
	switch {
	case days > 120:
		return unitMonth
	case days > 31:
		return unitWeek
	default:
		return unitDay
	}
}

// bucketStart is the first day of d's bucket (weeks start Monday, like Postgres date_trunc).
func bucketStart(d time.Time, unit string) time.Time {
	switch unit {
	case unitWeek:
		return d.AddDate(0, 0, -((int(d.Weekday()) + 6) % 7))
	case unitMonth:
		return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return d
	}
}

func nextBucket(d time.Time, unit string) time.Time {
	switch unit {
	case unitWeek:
		return d.AddDate(0, 0, 7)
	case unitMonth:
		return d.AddDate(0, 1, 0)
	default:
		return d.AddDate(0, 0, 1)
	}
}

// newSummaryRange covers the last days days up to today, widened back to the start of the first
// bucket so every bucket is whole (except the current one).
func newSummaryRange(now time.Time, days int) summaryRange {
	unit := unitFor(days)
	today := utils.TodayInThailand(now)
	from := bucketStart(today.AddDate(0, 0, -(days-1)), unit)
	length := int(today.Sub(from).Hours()/24) + 1
	return summaryRange{Unit: unit, From: from, To: today, PreviousFrom: from.AddDate(0, 0, -length)}
}

// thaiMidnight is the instant a Thailand date starts.
func thaiMidnight(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, utils.Thailand)
}

func (r summaryRange) Current() repomodels.DateRange {
	return repomodels.DateRange{From: thaiMidnight(r.From), To: thaiMidnight(r.To.AddDate(0, 0, 1))}
}

func (r summaryRange) Previous() repomodels.DateRange {
	return repomodels.DateRange{From: thaiMidnight(r.PreviousFrom), To: thaiMidnight(r.From)}
}

// Buckets lists the first day of every bucket in the range.
func (r summaryRange) Buckets() []time.Time {
	var out []time.Time
	for d := r.From; !d.After(r.To); d = nextBucket(d, r.Unit) {
		out = append(out, d)
	}
	return out
}
