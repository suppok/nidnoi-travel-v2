package models

// BookingPublicCreateRequest is a booking request from the customer site. Prices are never taken
// from the request: the backend quotes them from the period.
type BookingPublicCreateRequest struct {
	TourCode     string `json:"tourCode" example:"NIDN000001"`
	PeriodID     string `json:"periodId"`
	AdultTwin    int    `json:"adultTwin"`    // adults sharing a room
	AdultSingle  int    `json:"adultSingle"`  // adults in a single room; only when the period has a single supplement
	ChildWithBed int    `json:"childWithBed"` // only when the period has a child-with-bed price
	ChildNoBed   int    `json:"childNoBed"`   // only when the period has a child-no-bed price
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Note         string `json:"note"` // special requests
	// Website is a honeypot: the form hides it, so only bots fill it in.
	Website string `json:"website"`
}

type BookingPublicCreateResponse struct {
	BookingNo  string `json:"bookingNo" example:"26X8K9AQ"`
	TotalPrice int    `json:"totalPrice"` // whole baht, as quoted
}
