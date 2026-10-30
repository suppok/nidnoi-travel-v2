package models

import coremodels "backend/pkg/core/api/models"

type CityGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CityGetResponse struct {
	CityItem
}
