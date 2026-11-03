package models

import coremodels "backend/pkg/core/api/models"

type AirlineGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AirlineGetResponse struct {
	AirlineItem
}
