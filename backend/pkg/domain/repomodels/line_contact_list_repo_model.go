package repomodels

// LineContactListCriteria filters POST /line-contacts/list.
type LineContactListCriteria struct {
	Keyword   string // matches display name
	Following *bool  // nil = any
	Page      int
	PageSize  int
}
