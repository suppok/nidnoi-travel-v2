package models

import coremodels "backend/pkg/core/api/models"

type BookingSourceCreateRequest struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"` // optional
	SortOrder int    `json:"sortOrder"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type BookingSourceCreateResponse struct {
	BookingSourceItem
}
