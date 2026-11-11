package models

import coremodels "backend/pkg/core/api/models"

type UserUpdateRequest struct {
	ID    string   `json:"id"`
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Phone string   `json:"phone"` // optional; spaces and dashes are removed
	Roles []string `json:"roles"` // at least one

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type UserUpdateResponse struct {
	UserItem
}
