package models

import coremodels "backend/pkg/core/api/models"

type BookingSourceListRequest struct {
	Keyword  string `json:"keyword"`  // matches name or short name
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 500 (dropdowns load every source)

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type BookingSourceListResponse struct {
	Items    []BookingSourceItem `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}
