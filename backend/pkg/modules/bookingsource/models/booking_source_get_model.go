package models

import coremodels "backend/pkg/core/api/models"

type BookingSourceGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type BookingSourceGetResponse struct {
	BookingSourceItem
}
