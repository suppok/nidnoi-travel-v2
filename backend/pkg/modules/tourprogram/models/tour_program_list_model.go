package models

import coremodels "backend/pkg/core/api/models"

type TourProgramListRequest struct {
	Keyword     string `json:"keyword"`     // matches name, code or wholesale tour code
	CountryID   string `json:"countryId"`   // optional
	CityID      string `json:"cityId"`      // optional
	WholesaleID string `json:"wholesaleId"` // optional
	State       int    `json:"state"`       // 0 or omitted = any
	Status      *bool  `json:"status"`      // omitted = all
	Page        int    `json:"page"`        // default 1
	PageSize    int    `json:"pageSize"`    // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramListResponse struct {
	Items    []TourProgramListItem `json:"items"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}
