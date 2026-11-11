package models

import coremodels "backend/pkg/core/api/models"

type BookingSourceUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// BookingSourceUpdateStatusResponse is always nil (data: null).
type BookingSourceUpdateStatusResponse struct{}
