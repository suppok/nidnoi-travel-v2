package repomodels

// CountryListCriteria filters POST /countries/list.
type CountryListCriteria struct {
	Keyword  string // matches Thai/English name or slug, case-insensitive
	Status   *bool  // nil = all
	Page     int
	PageSize int
}
