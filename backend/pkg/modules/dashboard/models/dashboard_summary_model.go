package models

import (
	"time"

	coremodels "backend/pkg/core/api/models"
)

type DashboardSummaryRequest struct {
	// Days is the period to summarise, ending today (Thailand time): 7, 30, 90 or 365; default 30.
	Days int `json:"days"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// DashboardRange is the summarised period. From is aligned to the start of its first bucket, so
// it may begin a few days before today − days.
type DashboardRange struct {
	From string `json:"from"` // YYYY-MM-DD, inclusive
	To   string `json:"to"`   // YYYY-MM-DD, inclusive (today)
	Days int    `json:"days"` // as requested
	// Unit is the chart bucket: day (7, 30 days), week (90, weeks start Monday) or month (365).
	Unit string `json:"unit" enums:"day,week,month"`
	// PreviousFrom / PreviousTo is the period of the same length just before, for the deltas.
	PreviousFrom string `json:"previousFrom"`
	PreviousTo   string `json:"previousTo"`
}

type DashboardStateCount struct {
	State int `json:"state"`
	Count int `json:"count"`
}

// DashboardBucket is one chart point; every bucket of the range is present (zeros included).
type DashboardBucket struct {
	Date    string `json:"date"` // first day of the bucket, YYYY-MM-DD
	Count   int    `json:"count"`
	Revenue int    `json:"revenue"` // LINE bookings: paid bookings' total; customer bookings: 0
}

type DashboardSummaryResponse struct {
	Range       DashboardRange `json:"range"`
	GeneratedAt time.Time      `json:"generatedAt"`
	// Each section is null when the caller lacks its list permission.
	Tours        *DashboardTours        `json:"tours"`        // tour-program:list
	Bookings     *DashboardBookings     `json:"bookings"`     // booking:list
	LineBookings *DashboardLineBookings `json:"lineBookings"` // line-booking:list, scoped like the list
}

// --- Tours ----------------------------------------------------------------------------------------

type DashboardTours struct {
	ByState   []DashboardStateCount  `json:"byState"`   // active tours, any update date
	ByCountry []DashboardCountryTour `json:"byCountry"` // public tours per country, top 8
	// Seats covers the departures of public tours in the next Seats.Days days.
	Seats DashboardSeats `json:"seats"`
	// LowSeatPeriods are those departures with seats left, fewest first (max 8).
	LowSeatPeriods []DashboardPeriod `json:"lowSeatPeriods"`
}

type DashboardCountryTour struct {
	CountryID string `json:"countryId"`
	NameTh    string `json:"nameTh"`
	Count     int    `json:"count"`
}

type DashboardSeats struct {
	Days              int `json:"days"`
	LowSeatsThreshold int `json:"lowSeatsThreshold"`
	Periods           int `json:"periods"`
	SeatsTotal        int `json:"seatsTotal"`
	SeatsAvailable    int `json:"seatsAvailable"`
	SoldOut           int `json:"soldOut"`
	LowSeats          int `json:"lowSeats"`
}

type DashboardPeriod struct {
	PeriodID       string `json:"periodId"`
	TourProgramID  string `json:"tourProgramId"`
	TourCode       string `json:"tourCode"`
	TourName       string `json:"tourName"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	Price          int    `json:"price"`
	SeatsTotal     int    `json:"seatsTotal"`
	SeatsAvailable int    `json:"seatsAvailable"`
}

// --- Customer booking requests -------------------------------------------------------------------

type DashboardBookings struct {
	Total         int                   `json:"total"`         // created in the range
	PreviousTotal int                   `json:"previousTotal"` // created in the previous range
	NewCount      int                   `json:"newCount"`      // still new, any create date
	ByState       []DashboardStateCount `json:"byState"`       // created in the range
	Buckets       []DashboardBucket     `json:"buckets"`
}

// --- LINE bookings -------------------------------------------------------------------------------

type DashboardLineBookings struct {
	// ScopeAll: the caller sees every booking (admins); otherwise their own (sales) / sent to them (finance).
	ScopeAll bool                  `json:"scopeAll"`
	Totals   DashboardLineTotals   `json:"totals"`
	Previous DashboardLineTotals   `json:"previous"`
	Buckets  []DashboardBucket     `json:"buckets"`
	ByState  []DashboardStateCount `json:"byState"`
	BySource []DashboardLineGroup  `json:"bySource"` // top 8
	ByTour   []DashboardLineGroup  `json:"byTour"`   // top 8
	BySales  []DashboardLineGroup  `json:"bySales"`  // top 8; [] unless ScopeAll
	Pipeline DashboardLinePipeline `json:"pipeline"` // any create date
	// Departures are the paid bookings departing in the next DeparturesDays days (max 10).
	DeparturesDays int                  `json:"departuresDays"`
	Departures     []DashboardDeparture `json:"departures"`
}

// DashboardLineTotals: Revenue, Travellers and AverageValue count the paid bookings only.
type DashboardLineTotals struct {
	Count        int `json:"count"`
	PaidCount    int `json:"paidCount"`
	Cancelled    int `json:"cancelled"`
	Revenue      int `json:"revenue"`
	Travellers   int `json:"travellers"`
	AverageValue int `json:"averageValue"` // Revenue / PaidCount, rounded
}

type DashboardLineGroup struct {
	Key        string `json:"key"`   // booking source id, tour code or user id; "" = not set
	Label      string `json:"label"` // "" = not set
	Count      int    `json:"count"`
	PaidCount  int    `json:"paidCount"`
	Revenue    int    `json:"revenue"`
	Travellers int    `json:"travellers"`
}

type DashboardLinePipeline struct {
	WaitingConfirm   int `json:"waitingConfirm"`
	WaitingDocuments int `json:"waitingDocuments"`
	WaitingReview    int `json:"waitingReview"`
	BalancePending   int `json:"balancePending"`
}

type DashboardUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DashboardDeparture struct {
	ID           string         `json:"id"`
	BookingNo    string         `json:"bookingNo"`
	TourCode     string         `json:"tourCode"`
	TourName     string         `json:"tourName"`
	StartDate    string         `json:"startDate"`
	EndDate      string         `json:"endDate"`
	CustomerName string         `json:"customerName"`
	Travellers   int            `json:"travellers"`
	State        int            `json:"state"`
	TotalPrice   int            `json:"totalPrice"`
	SalesUser    *DashboardUser `json:"salesUser"`
}
