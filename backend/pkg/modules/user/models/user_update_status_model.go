package models

import coremodels "backend/pkg/core/api/models"

type UserUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = deactivate (signs the user out everywhere)

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// UserUpdateStatusResponse is always nil (data: null).
type UserUpdateStatusResponse struct{}
