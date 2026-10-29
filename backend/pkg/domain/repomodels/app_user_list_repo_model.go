package repomodels

// AppUserListCriteria filters POST /users/list.
type AppUserListCriteria struct {
	Keyword  string // matches name or email, case-insensitive
	Role     string // empty = any
	Status   *bool  // nil = all
	Page     int
	PageSize int
}
