package repomodels

// BookingListCriteria filters POST /bookings/list.
type BookingListCriteria struct {
	Keyword  string // matches booking no, customer name, email, phone or tour code
	State    int    // 0 = any
	Status   *bool  // nil = any
	Page     int
	PageSize int
}
