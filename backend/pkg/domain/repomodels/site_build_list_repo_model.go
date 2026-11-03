package repomodels

// SiteBuildListCriteria filters POST /site-builds/list.
type SiteBuildListCriteria struct {
	Trigger  string // "" = any
	Page     int
	PageSize int
}
