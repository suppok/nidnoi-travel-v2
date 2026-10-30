package repomodels

// AppUserListCriteria filters POST /users/list.
type AppUserListCriteria struct {
	Keyword     string // matches name or email, case-insensitive
	Role        string // empty = any
	ExcludeRole string // empty = none; hides users holding this role
	Status      *bool  // nil = all
	Page        int
	PageSize    int
}
