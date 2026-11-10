package models

// LineBookingCreateRequest is submitted from the LIFF app (no JWT auth, LIFF token validated).
type LineBookingCreateRequest struct {
	LIFFAccessToken string `json:"liffAccessToken"`
	TourCode        string `json:"tourCode"`
	PeriodID        string `json:"periodId"`
	AdultTwin       int    `json:"adultTwin"`
	AdultSingle     int    `json:"adultSingle"`
	ChildWithBed    int    `json:"childWithBed"`
	ChildNoBed      int    `json:"childNoBed"`
	CustomerName    string `json:"customerName"`
	CustomerPhone   string `json:"customerPhone"`
	SlipImage       string `json:"slipImage"` // R2 URL from /media/upload/liff
	Note            string `json:"note"`
}

type LineBookingCreateResponse struct {
	BookingNo  string `json:"bookingNo" example:"26X8K9AQ"`
	TotalPrice int    `json:"totalPrice"`
}
