package models

import coremodels "backend/pkg/core/api/models"

type TourProgramGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramGetResponse struct {
	TourProgramItem
}
