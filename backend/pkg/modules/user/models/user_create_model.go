package models

import coremodels "backend/pkg/core/api/models"

type UserCreateRequest struct {
	Email             string   `json:"email"`
	Name              string   `json:"name"`
	Roles             []string `json:"roles"`             // at least one
	TemporaryPassword string   `json:"temporaryPassword"` // user must change it at first login

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type UserCreateResponse struct {
	UserItem
}
