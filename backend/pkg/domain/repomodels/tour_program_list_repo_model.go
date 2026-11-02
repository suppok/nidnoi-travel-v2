package repomodels

// TourProgramListCriteria filters POST /tour-programs/list.
type TourProgramListCriteria struct {
	Keyword     string // matches name, code or wholesale tour code, case-insensitive
	CountryID   string // "" = any
	CityID      string // "" = any
	WholesaleID string // "" = any
	State       int    // 0 = any
	Status      *bool  // nil = all
	Page        int
	PageSize    int
}
