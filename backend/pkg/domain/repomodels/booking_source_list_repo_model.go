package repomodels

// BookingSourceListCriteria filters POST /booking-sources/list.
type BookingSourceListCriteria struct {
	Keyword  string // matches name or short name, case-insensitive
	Status   *bool  // nil = all
	Page     int
	PageSize int
}
