package models

import coremodels "backend/pkg/core/api/models"

type UserResetPasswordRequest struct {
	ID                string `json:"id"`
	TemporaryPassword string `json:"temporaryPassword"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// UserResetPasswordResponse is always nil (data: null).
type UserResetPasswordResponse struct{}
