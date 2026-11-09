package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// DateLayout is the wire format of startDate / endDate.
const DateLayout = "2006-01-02"

type BookingUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BookingItem is the booking shape returned by every staff booking endpoint.
type BookingItem struct {
	ID                  string   `json:"id"`
	BookingNo           string   `json:"bookingNo" example:"26X8K9AQ"`
	TourProgramID       string   `json:"tourProgramId"`
	TourProgramPeriodID string   `json:"tourProgramPeriodId"`
	TourCode            string   `json:"tourCode"` // at booking time
	TourName            string   `json:"tourName"` // at booking time
	StartDate           string   `json:"startDate" example:"2026-12-01"`
	EndDate             string   `json:"endDate" example:"2026-12-05"`
	AdultTwinCount      int      `json:"adultTwinCount"`
	AdultTwinPrice      int      `json:"adultTwinPrice"`
	AdultSingleCount    int      `json:"adultSingleCount"`
	AdultSinglePrice    null.Int `json:"adultSinglePrice" swaggertype:"integer"`
	ChildWithBedCount   int      `json:"childWithBedCount"`
	ChildWithBedPrice   null.Int `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedCount     int      `json:"childNoBedCount"`
	ChildNoBedPrice     null.Int `json:"childNoBedPrice" swaggertype:"integer"`
	TotalPrice          int      `json:"totalPrice"`
	FirstName           string   `json:"firstName"`
	LastName            string   `json:"lastName"`
	Email               string   `json:"email"`
	Phone               string   `json:"phone"`
	CustomerNote        string   `json:"customerNote"`
	State               int      `json:"state" enums:"10,20,30,40"` // 10 new, 20 contacted, 30 confirmed, 40 cancelled
	// Assignee is the staff user who first took the booking out of "new"; null until then.
	Assignee   *BookingUser `json:"assignee"`
	StaffNote  string       `json:"staffNote"`
	Status     bool         `json:"status"`
	CreateDate time.Time    `json:"createDate"`
	UpdateDate time.Time    `json:"updateDate"`
}

func NewBookingItem(b *domain.Booking) BookingItem {
	item := BookingItem{
		ID:                  b.ID,
		BookingNo:           b.BookingNo,
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
		FirstName:           b.FirstName,
		LastName:            b.LastName,
		Email:               b.Email,
		Phone:               b.Phone,
		CustomerNote:        b.CustomerNote,
		State:               b.State,
		StaffNote:           b.StaffNote,
		Status:              b.Status,
		CreateDate:          b.CreateDate,
		UpdateDate:          b.UpdateDate,
	}
	if b.Assignee != nil {
		item.Assignee = &BookingUser{ID: b.Assignee.ID, Name: b.Assignee.Name}
	}
	return item
}
