package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

const DateLayout = "2006-01-02"

// LineBookingItem is the shape returned by every staff LINE booking endpoint.
type LineBookingItem struct {
	ID                  string    `json:"id"`
	BookingNo           string    `json:"bookingNo" example:"26X8K9AQ"`
	LineUserID          string    `json:"lineUserId"`
	LineDisplayName     string    `json:"lineDisplayName"`
	TourProgramID       string    `json:"tourProgramId"`
	TourProgramPeriodID string    `json:"tourProgramPeriodId"`
	TourCode            string    `json:"tourCode"`
	TourName            string    `json:"tourName"`
	StartDate           string    `json:"startDate" example:"2026-12-01"`
	EndDate             string    `json:"endDate" example:"2026-12-05"`
	AdultTwinCount      int       `json:"adultTwinCount"`
	AdultTwinPrice      int       `json:"adultTwinPrice"`
	AdultSingleCount    int       `json:"adultSingleCount"`
	AdultSinglePrice    null.Int  `json:"adultSinglePrice" swaggertype:"integer"`
	ChildWithBedCount   int       `json:"childWithBedCount"`
	ChildWithBedPrice   null.Int  `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedCount     int       `json:"childNoBedCount"`
	ChildNoBedPrice     null.Int  `json:"childNoBedPrice" swaggertype:"integer"`
	TotalPrice          int       `json:"totalPrice"`
	CustomerName        string    `json:"customerName"`
	CustomerPhone       string    `json:"customerPhone"`
	SlipImage           string    `json:"slipImage"`
	Note                string    `json:"note"`
	State               int       `json:"state" enums:"10,20,30"`
	StaffNote           string    `json:"staffNote"`
	Status              bool      `json:"status"`
	CreateDate          time.Time `json:"createDate"`
	UpdateDate          time.Time `json:"updateDate"`
	// SalesUser owns the booking; null on bookings made before it was recorded.
	SalesUser *LineBookingUser `json:"salesUser"`
	// LineContact is the customer's LINE OA chat picked by staff; null until linked.
	LineContact *LineBookingContact `json:"lineContact"`
}

// LineBookingUser is a staff user referenced by a booking.
type LineBookingUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LineBookingContact is the customer's LINE OA chat linked to a booking.
type LineBookingContact struct {
	ID          string      `json:"id"`
	DisplayName string      `json:"displayName"`
	PictureURL  null.String `json:"pictureUrl" swaggertype:"string"`
	// Following is false once the customer blocks the OA: a push would not reach them.
	Following bool `json:"following"`
}

func NewLineBookingItem(b *domain.LineBooking) LineBookingItem {
	item := LineBookingItem{
		ID:                  b.ID,
		BookingNo:           b.BookingNo,
		LineUserID:          b.LineUserID,
		LineDisplayName:     b.LineDisplayName,
		TourProgramID:       b.TourProgramID,
		TourProgramPeriodID: b.TourProgramPeriodID,
		TourCode:            b.TourCode,
		TourName:            b.TourName,
		StartDate:           b.StartDate.Format(DateLayout),
		EndDate:             b.EndDate.Format(DateLayout),
		AdultTwinCount:      b.AdultTwinCount,
		AdultTwinPrice:      b.AdultTwinPrice,
		AdultSingleCount:    b.AdultSingleCount,
		AdultSinglePrice:    b.AdultSinglePrice,
		ChildWithBedCount:   b.ChildWithBedCount,
		ChildWithBedPrice:   b.ChildWithBedPrice,
		ChildNoBedCount:     b.ChildNoBedCount,
		ChildNoBedPrice:     b.ChildNoBedPrice,
		TotalPrice:          b.TotalPrice,
		CustomerName:        b.CustomerName,
		CustomerPhone:       b.CustomerPhone,
		SlipImage:           b.SlipImage,
		Note:                b.Note,
		State:               b.State,
		StaffNote:           b.StaffNote,
		Status:              b.Status,
		CreateDate:          b.CreateDate,
		UpdateDate:          b.UpdateDate,
	}
	if s := b.SalesUser; s != nil {
		item.SalesUser = &LineBookingUser{ID: s.ID, Name: s.Name}
	}
	if c := b.LineContact; c != nil {
		item.LineContact = &LineBookingContact{ID: c.ID, DisplayName: c.DisplayName, PictureURL: c.PictureURL, Following: c.Following}
	}
	return item
}
