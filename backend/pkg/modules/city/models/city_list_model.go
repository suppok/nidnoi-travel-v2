package models

import coremodels "backend/pkg/core/api/models"

type CityListRequest struct {
	Keyword   string `json:"keyword"`   // matches Thai/English name or slug
	CountryID string `json:"countryId"` // empty = any
	Status    *bool  `json:"status"`    // omitted = all
	Page      int    `json:"page"`      // default 1
	PageSize  int    `json:"pageSize"`  // default 20, max 500

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CityListResponse struct {
	Items    []CityItem `json:"items"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}
