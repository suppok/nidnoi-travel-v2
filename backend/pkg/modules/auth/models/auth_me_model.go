package models

import coremodels "backend/pkg/core/api/models"

type AuthMeRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AuthMeResponse struct {
	ID                 string   `json:"id"`
	Email              string   `json:"email"`
	Name               string   `json:"name"`
	Roles              []string `json:"roles"`
	Permissions        []string `json:"permissions"` // union of the roles' permission codes
	MustChangePassword bool     `json:"mustChangePassword"`
}
