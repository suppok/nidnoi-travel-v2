package repomodels

// LineBookingListCriteria filters POST /line-bookings/list.
type LineBookingListCriteria struct {
	Keyword  string // matches booking no, customer name, customer phone, tour code or LINE display name
	State    int    // 0 = any
	Status   *bool  // nil = any
	Page     int
	PageSize int
}
