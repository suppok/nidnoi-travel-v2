package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

const DateLayout = "2006-01-02"

// LineBookingItem is the shape returned by every staff LINE booking endpoint.
type LineBookingItem struct {
	ID                       string    `json:"id"`
	BookingNo                string    `json:"bookingNo" example:"26X8K9AQ"`
	LineUserID               string    `json:"lineUserId"`
	LineDisplayName          string    `json:"lineDisplayName"`
	TourProgramID            string    `json:"tourProgramId"`
	TourProgramPeriodID      string    `json:"tourProgramPeriodId"`
	TourCode                 string    `json:"tourCode"`
	TourName                 string    `json:"tourName"`
	StartDate                string    `json:"startDate" example:"2026-12-01"`
	EndDate                  string    `json:"endDate" example:"2026-12-05"`
	AdultTwinCount           int       `json:"adultTwinCount"`
	AdultTwinPrice           int       `json:"adultTwinPrice"`
	AdultSingleCount         int       `json:"adultSingleCount"`
	AdultSinglePrice         null.Int  `json:"adultSinglePrice" swaggertype:"integer"`
	ChildWithBedCount        int       `json:"childWithBedCount"`
	ChildWithBedPrice        null.Int  `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedCount          int       `json:"childNoBedCount"`
	ChildNoBedPrice          null.Int  `json:"childNoBedPrice" swaggertype:"integer"`
	TotalPrice               int       `json:"totalPrice"`
	CustomerName             string    `json:"customerName"`
	CustomerPhone            string    `json:"customerPhone"`
	SlipImage                string    `json:"slipImage"`
	Note                     string    `json:"note"`
	State                    int       `json:"state" enums:"10,20,25,30"`
	StaffNote                string    `json:"staffNote"`
	Status                   bool      `json:"status"`
	CreateDate               time.Time `json:"createDate"`
	UpdateDate               time.Time `json:"updateDate"`
	InfantCount              int       `json:"infantCount"`
	InfantPrice              null.Int  `json:"infantPrice" swaggertype:"integer"`
	AdultPriceBeforeDiscount null.Int  `json:"adultPriceBeforeDiscount" swaggertype:"integer"`
	DiscountPerPerson        int       `json:"discountPerPerson"`
	// Charges are the extra lines (visa, taxes, options, ...); [] on older bookings.
	Charges          []domain.LineBookingCharge `json:"charges"`
	PaymentMethod    string                     `json:"paymentMethod" enums:"transfer,credit"`
	CreditFeePercent float64                    `json:"creditFeePercent"`
	CardFee          int                        `json:"cardFee"`
	IsDeposit        bool                       `json:"isDeposit"`
	DepositPerPerson int                        `json:"depositPerPerson"`
	DepositDueDate   null.String                `json:"depositDueDate" swaggertype:"string" example:"2026-12-01"`
	DepositDueTime   string                     `json:"depositDueTime" example:"18:00"`
	RemainingDueDate null.String                `json:"remainingDueDate" swaggertype:"string"`
	FullDueDate      null.String                `json:"fullDueDate" swaggertype:"string"`
	FullDueTime      string                     `json:"fullDueTime"`
	CustomerAddress  string                     `json:"customerAddress"`
	RefName          string                     `json:"refName"`
	RefBookingNo     string                     `json:"refBookingNo"`
	// BookingSource is where the customer booked; null when not given.
	BookingSource *LineBookingSource `json:"bookingSource"`
	// SalesUser owns the booking; null on bookings made before it was recorded.
	SalesUser *LineBookingUser `json:"salesUser"`
	// FinanceUser issues the receipt; set when the sales owner confirms (null before).
	FinanceUser *LineBookingUser `json:"financeUser"`
	ConfirmedAt null.Time        `json:"confirmedAt" swaggertype:"string"`
	// Finance documents (PDF URLs, "" = none): deposit → aiFile + ivFile, full payment → reFile.
	AIFile string `json:"aiFile"`
	IVFile string `json:"ivFile"`
	REFile string `json:"reFile"`
	// DocumentsAt is when finance sent them; null on a confirmed booking = waiting for finance.
	DocumentsAt null.Time `json:"documentsAt" swaggertype:"string"`
	// DocumentsApprovedAt: the sales owner approved them; CustomerSentAt: the card was pushed to the
	// customer's LINE chat (null = approved without sending).
	DocumentsApprovedAt null.Time `json:"documentsApprovedAt" swaggertype:"string"`
	CustomerSentAt      null.Time `json:"customerSentAt" swaggertype:"string"`
	// FixNote is the latest fix request from sales to finance ("" = none).
	FixNote        string    `json:"fixNote"`
	FixRequestedAt null.Time `json:"fixRequestedAt" swaggertype:"string"`
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

		InfantCount:              b.InfantCount,
		InfantPrice:              b.InfantPrice,
		AdultPriceBeforeDiscount: b.AdultPriceBeforeDiscount,
		DiscountPerPerson:        b.DiscountPerPerson,
		Charges:                  b.Charges,
		PaymentMethod:            b.PaymentMethod,
		CreditFeePercent:         float64(b.CreditFeeBps) / 100,
		CardFee:                  b.CardFee,
		IsDeposit:                b.IsDeposit,
		DepositPerPerson:         b.DepositPerPerson,
		DepositDueDate:           formatDate(b.DepositDueDate),
		DepositDueTime:           b.DepositDueTime,
		RemainingDueDate:         formatDate(b.RemainingDueDate),
		FullDueDate:              formatDate(b.FullDueDate),
		FullDueTime:              b.FullDueTime,
		CustomerAddress:          b.CustomerAddress,
		RefName:                  b.RefName,
		RefBookingNo:             b.RefBookingNo,
	}
	if item.Charges == nil {
		item.Charges = []domain.LineBookingCharge{}
	}
	if s := b.BookingSource; s != nil {
		item.BookingSource = &LineBookingSource{ID: s.ID, Name: s.Name, ShortName: s.ShortName}
	}
	item.ConfirmedAt = b.ConfirmedAt
	item.AIFile, item.IVFile, item.REFile, item.DocumentsAt = b.AIFile, b.IVFile, b.REFile, b.DocumentsAt
	item.DocumentsApprovedAt, item.CustomerSentAt = b.DocumentsApprovedAt, b.CustomerSentAt
	item.FixNote, item.FixRequestedAt = b.FixNote, b.FixRequestedAt
	if f := b.FinanceUser; f != nil {
		item.FinanceUser = &LineBookingUser{ID: f.ID, Name: f.Name}
	}
	if s := b.SalesUser; s != nil {
		item.SalesUser = &LineBookingUser{ID: s.ID, Name: s.Name}
	}
	if c := b.LineContact; c != nil {
		item.LineContact = &LineBookingContact{ID: c.ID, DisplayName: c.DisplayName, PictureURL: c.PictureURL, Following: c.Following}
	}
	return item
}

// formatDate writes a DATE column. Nullable DATE columns are *time.Time, not null.Time:
// null.Time scans through database/sql, which can't read the driver's text DATE value.
func formatDate(t *time.Time) null.String {
	if t == nil {
		return null.String{}
	}
	return null.StringFrom(t.Format(DateLayout))
}
