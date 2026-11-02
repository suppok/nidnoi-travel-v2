package models

import coremodels "backend/pkg/core/api/models"

// TourProgramUpdateRequest replaces every editable field. code is never changed.
type TourProgramUpdateRequest struct {
	ID string `json:"id"`
	TourProgramFields

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramUpdateResponse struct {
	TourProgramItem
}
