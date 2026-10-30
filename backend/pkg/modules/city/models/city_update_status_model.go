package models

import coremodels "backend/pkg/core/api/models"

type CityUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// CityUpdateStatusResponse is always nil (data: null).
type CityUpdateStatusResponse struct{}
