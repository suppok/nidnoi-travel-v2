package models

import coremodels "backend/pkg/core/api/models"

type TourProgramUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// TourProgramUpdateStatusResponse is always nil (data: null).
type TourProgramUpdateStatusResponse struct{}
