package models

import coremodels "backend/pkg/core/api/models"

type AirlineListRequest struct {
	Keyword  string `json:"keyword"`  // matches name or code
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 500 (dropdowns load every airline)

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AirlineListResponse struct {
	Items    []AirlineItem `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
