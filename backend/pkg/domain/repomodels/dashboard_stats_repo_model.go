package repomodels

import "time"

// Aggregate rows for the backoffice dashboard (POST /dashboard/summary). Each repository answers its
// own table; the dashboard use case combines them.

// StateCount is the number of rows in one workflow state.
type StateCount struct {
	State int `bun:"state"`
	Count int `bun:"count"`
}

// DateRange is [From, To): instants, usually Thailand midnights.
type DateRange struct {
	From time.Time
	To   time.Time
}

// BucketCount is the rows created in one time bucket (bucket = its first day in Thailand time).
type BucketCount struct {
	Bucket  time.Time `bun:"bucket"`
	Count   int       `bun:"count"`
	Revenue int       `bun:"revenue"` // sum of total_price of the paid rows; 0 where it doesn't apply
}

// LineBookingStatsCriteria limits the LINE booking aggregates: active rows created in Range, that
// the caller may see (Scope nil = every booking).
type LineBookingStatsCriteria struct {
	Range DateRange
	Scope *LineBookingScope
}

// LineBookingTotals sums the bookings matching a LineBookingStatsCriteria. Revenue and Travellers
// count the paid bookings (linebookingstate.Paid) only.
type LineBookingTotals struct {
	Count      int `bun:"count"`
	PaidCount  int `bun:"paid_count"`
	Cancelled  int `bun:"cancelled"`
	Revenue    int `bun:"revenue"`
	Travellers int `bun:"travellers"`
}

// LineBookingGroup is one row of a LINE booking breakdown (by source, tour or sales user). Revenue
// and Travellers count the paid bookings only.
type LineBookingGroup struct {
	Key        string `bun:"key"`   // booking_source.id, tour_code or app_user.id; '' = none
	Label      string `bun:"label"` // its name ('' = none)
	Count      int    `bun:"count"`
	PaidCount  int    `bun:"paid_count"`
	Revenue    int    `bun:"revenue"`
	Travellers int    `bun:"travellers"`
}

// LineBookingPipeline counts active bookings (any create date) waiting on someone.
type LineBookingPipeline struct {
	WaitingConfirm   int `bun:"waiting_confirm"`   // state new: the sales owner has to confirm
	WaitingDocuments int `bun:"waiting_documents"` // paid, finance hasn't sent the documents of the current round
	WaitingReview    int `bun:"waiting_review"`    // finance sent them, the sales owner hasn't approved
	BalancePending   int `bun:"balance_pending"`   // balance recorded, waits for the owner's confirm
}

// PeriodSeatStats sums the seats of the active periods of active, public tours departing in a range.
type PeriodSeatStats struct {
	Periods        int `bun:"periods"`
	SeatsTotal     int `bun:"seats_total"`
	SeatsAvailable int `bun:"seats_available"`
	SoldOut        int `bun:"sold_out"`  // seats_available = 0
	LowSeats       int `bun:"low_seats"` // 1..LowSeatsThreshold seats left
}

// UpcomingPeriod is one departure of an active, public tour.
type UpcomingPeriod struct {
	PeriodID       string    `bun:"period_id"`
	TourProgramID  string    `bun:"tour_program_id"`
	TourCode       string    `bun:"tour_code"`
	TourName       string    `bun:"tour_name"`
	StartDate      time.Time `bun:"start_date"`
	EndDate        time.Time `bun:"end_date"`
	Price          int       `bun:"price"`
	SeatsTotal     int       `bun:"seats_total"`
	SeatsAvailable int       `bun:"seats_available"`
}

// CountryTourCount is the number of active, public tours that visit a country.
type CountryTourCount struct {
	CountryID string `bun:"country_id"`
	NameTh    string `bun:"name_th"`
	Count     int    `bun:"count"`
}
