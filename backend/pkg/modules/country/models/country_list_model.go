package models

import coremodels "backend/pkg/core/api/models"

type CountryListRequest struct {
	Keyword  string `json:"keyword"`  // matches Thai/English name or slug
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 500 (dropdowns load every country)

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CountryListResponse struct {
	Items    []CountryItem `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
