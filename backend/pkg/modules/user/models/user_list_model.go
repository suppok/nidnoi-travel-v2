package models

import coremodels "backend/pkg/core/api/models"

type UserListRequest struct {
	Keyword  string `json:"keyword"`  // matches name or email
	Role     string `json:"role"`     // empty = any
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type UserListResponse struct {
	Items    []UserItem `json:"items"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}
