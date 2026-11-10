package models

import coremodels "backend/pkg/core/api/models"

// LineBookingCreateRequest is submitted from the LIFF app by a staff user linked to the LINE account.
type LineBookingCreateRequest struct {
	TourCode      string `json:"tourCode"`
	PeriodID      string `json:"periodId"`
	AdultTwin     int    `json:"adultTwin"`
	AdultSingle   int    `json:"adultSingle"`
	ChildWithBed  int    `json:"childWithBed"`
	ChildNoBed    int    `json:"childNoBed"`
	CustomerName  string `json:"customerName"`
	CustomerPhone string `json:"customerPhone"`
	SlipImage     string `json:"slipImage"` // R2 URL from /media/upload/liff
	Note          string `json:"note"`

	// The linked staff user, with the LINE account from the verified LIFF token.
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingCreateResponse struct {
	BookingNo  string `json:"bookingNo" example:"26X8K9AQ"`
	TotalPrice int    `json:"totalPrice"`
}
