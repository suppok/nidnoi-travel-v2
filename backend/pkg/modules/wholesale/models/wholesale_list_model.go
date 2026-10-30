package models

import coremodels "backend/pkg/core/api/models"

type WholesaleListRequest struct {
	Keyword  string `json:"keyword"`  // matches name, code, or a contact's name/phone
	State    int    `json:"state"`    // 0 or omitted = any
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 500

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type WholesaleListResponse struct {
	Items    []WholesaleItem `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}
