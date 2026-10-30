package repomodels

// WholesaleListCriteria filters POST /wholesales/list.
type WholesaleListCriteria struct {
	Keyword  string // matches name, code or a contact's name/phone, case-insensitive
	State    int    // 0 = any
	Status   *bool  // nil = all
	Page     int
	PageSize int
}
