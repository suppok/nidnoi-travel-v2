package utils

import (
	"fmt"
	"time"
)

// Thailand is UTC+7 with no daylight saving. A fixed zone avoids depending on tzdata in the image.
var Thailand = time.FixedZone("ICT", 7*60*60)

// TodayInThailand returns today's date in Thailand as midnight UTC (the form DATE columns compare as).
func TodayInThailand(now time.Time) time.Time {
	y, m, d := now.In(Thailand).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ThaiDate is d/m/yyyy in the Buddhist era, as Thai readers expect dates.
func ThaiDate(t time.Time) string {
	return fmt.Sprintf("%d/%d/%d", t.Day(), int(t.Month()), t.Year()+543)
}
