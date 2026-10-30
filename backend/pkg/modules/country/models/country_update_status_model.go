package models

import coremodels "backend/pkg/core/api/models"

type CountryUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// CountryUpdateStatusResponse is always nil (data: null).
type CountryUpdateStatusResponse struct{}
