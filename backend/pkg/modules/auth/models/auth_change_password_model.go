package models

import coremodels "backend/pkg/core/api/models"

type AuthChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
	RefreshToken    string                      `json:"-"` // current session; kept, all others are revoked
}

// AuthChangePasswordResponse is always nil (data: null).
type AuthChangePasswordResponse struct{}
