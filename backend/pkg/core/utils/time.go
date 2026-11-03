package utils

import "time"

// thailand is UTC+7 with no daylight saving. A fixed zone avoids depending on tzdata in the image.
var thailand = time.FixedZone("ICT", 7*60*60)

// TodayInThailand returns today's date in Thailand as midnight UTC (the form DATE columns compare as).
func TodayInThailand(now time.Time) time.Time {
	y, m, d := now.In(thailand).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
