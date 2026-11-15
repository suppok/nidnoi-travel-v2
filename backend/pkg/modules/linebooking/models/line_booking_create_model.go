package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// LineBookingCreateRequest is entered by a staff user on the backoffice booking entry page. It
// carries the same deal as the "สรุปยอดโอน" summary. Fields added with the summary are optional, so
// a request with counts only still gets the period quote.
type LineBookingCreateRequest struct {
	TourCode     string `json:"tourCode"`
	PeriodID     string `json:"periodId"`
	AdultTwin    int    `json:"adultTwin"`
	AdultSingle  int    `json:"adultSingle"`
	ChildWithBed int    `json:"childWithBed"`
	ChildNoBed   int    `json:"childNoBed"`
	Infant       int    `json:"infant"`

	// Unit prices agreed with the customer (whole baht). Null = the period's price. AdultPrice is per
	// adult; an adult in a single room pays AdultPrice + SingleSupplement.
	AdultPrice        null.Int `json:"adultPrice" swaggertype:"integer"`
	SingleSupplement  null.Int `json:"singleSupplement" swaggertype:"integer"`
	ChildWithBedPrice null.Int `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedPrice   null.Int `json:"childNoBedPrice" swaggertype:"integer"`
	InfantPrice       null.Int `json:"infantPrice" swaggertype:"integer"` // required when Infant > 0
	// AdultPriceBeforeDiscount is shown as "(จาก …)"; display only.
	AdultPriceBeforeDiscount null.Int `json:"adultPriceBeforeDiscount" swaggertype:"integer"`
	// DiscountPerPerson comes off every traveller's price, infants included.
	DiscountPerPerson int                    `json:"discountPerPerson"`
	Charges           []LineBookingChargeReq `json:"charges"`

	PaymentMethod    string  `json:"paymentMethod" enums:"transfer,credit"` // "" = transfer
	CreditFeePercent float64 `json:"creditFeePercent" example:"3"`          // credit only, up to 2 decimals
	IsDeposit        bool    `json:"isDeposit"`
	DepositPerPerson int     `json:"depositPerPerson"`
	// Dates are YYYY-MM-DD, times HH:mm; all optional.
	DepositDueDate   null.String `json:"depositDueDate" swaggertype:"string"`
	DepositDueTime   string      `json:"depositDueTime"`
	RemainingDueDate null.String `json:"remainingDueDate" swaggertype:"string"`
	FullDueDate      null.String `json:"fullDueDate" swaggertype:"string"`
	FullDueTime      string      `json:"fullDueTime"`

	// BookingSourceID is where the customer booked (POST /booking-sources/list); optional.
	BookingSourceID null.String `json:"bookingSourceId" swaggertype:"string"`
	RefName         string      `json:"refName"`      // the customer's name on that source
	RefBookingNo    string      `json:"refBookingNo"` // the booking number on that source

	CustomerName    string `json:"customerName"`
	CustomerPhone   string `json:"customerPhone"`
	CustomerAddress string `json:"customerAddress"`
	// SlipImages are the payment's slips (R2 URLs from /media/upload/line-booking-slip), 1 to 10. SlipImage is the
	// older single-slip field, still accepted when SlipImages is empty.
	SlipImages []string `json:"slipImages"`
	SlipImage  string   `json:"slipImage"`
	Note       string   `json:"note"`
	// SalesUserID is the sales person who owns the booking (GET /users/sales).
	SalesUserID string `json:"salesUserId"`
	// LineContactID is the customer's LINE OA chat (POST /line-contacts/list); optional.
	LineContactID null.String `json:"lineContactId" swaggertype:"string"`
	// ChatURL is that chat in LINE OA Manager (https://chat.line.biz/…), saved on the contact so
	// staff can open it; optional, needs lineContactId. "" keeps the one already saved.
	ChatURL string `json:"chatUrl"`

	// The signed-in staff user (and the LINE account they linked, if any).
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// LineBookingChargeReq is an extra charge line: count x price.
type LineBookingChargeReq struct {
	Kind  string `json:"kind" enums:"visa,foreignPassport,joinLand,fuelTax,airportTax,option"`
	Name  string `json:"name"` // required for "option", ignored otherwise
	Count int    `json:"count"`
	Price int    `json:"price"`
}

type LineBookingCreateResponse struct {
	BookingNo string `json:"bookingNo" example:"26X8K9AQ"`
	// TotalPrice is the grand total, card fee included.
	TotalPrice int `json:"totalPrice"`
}

// LineBookingSource is a booking source referenced by a booking.
type LineBookingSource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"` // label of the reference name ("ชื่อ <shortName>"); "" = none
}
