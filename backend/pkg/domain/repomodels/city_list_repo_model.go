package repomodels

// CityListCriteria filters POST /cities/list.
type CityListCriteria struct {
	Keyword   string // matches Thai/English name or slug, case-insensitive
	CountryID string // empty = any
	Status    *bool  // nil = all
	Page      int
	PageSize  int
}
