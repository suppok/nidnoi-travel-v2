package models

import coremodels "backend/pkg/core/api/models"

type AirlineUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// AirlineUpdateStatusResponse is always nil (data: null).
type AirlineUpdateStatusResponse struct{}
