package models

// TourProgramPublicListRequest is read from the query string of GET /tour-programs/public.
// The json tags name the query parameters in VALIDATION errors.
type TourProgramPublicListRequest struct {
	Country  string `json:"country"`  // country slug, optional
	City     string `json:"city"`     // city slug, optional
	DateFrom string `json:"dateFrom"` // YYYY-MM-DD, optional: has an upcoming departure on/after
	DateTo   string `json:"dateTo"`   // YYYY-MM-DD, optional: has an upcoming departure on/before
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 100
}

type TourProgramPublicListResponse struct {
	Items    []TourProgramPublicItem `json:"items"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
}
