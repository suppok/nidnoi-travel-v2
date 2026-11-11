package models

import coremodels "backend/pkg/core/api/models"

type BookingSourceUpdateRequest struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	SortOrder int    `json:"sortOrder"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type BookingSourceUpdateResponse struct {
	BookingSourceItem
}
