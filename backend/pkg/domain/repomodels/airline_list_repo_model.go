package repomodels

// AirlineListCriteria filters POST /airlines/list.
type AirlineListCriteria struct {
	Keyword  string // matches name or code, case-insensitive
	Status   *bool  // nil = all
	Page     int
	PageSize int
}
