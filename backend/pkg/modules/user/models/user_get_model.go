package models

import coremodels "backend/pkg/core/api/models"

type UserGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type UserGetResponse struct {
	UserItem
}
