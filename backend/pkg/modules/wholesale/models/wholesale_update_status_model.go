package models

import coremodels "backend/pkg/core/api/models"

type WholesaleUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// WholesaleUpdateStatusResponse is always nil (data: null).
type WholesaleUpdateStatusResponse struct{}
