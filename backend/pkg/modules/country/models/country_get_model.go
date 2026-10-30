package models

import coremodels "backend/pkg/core/api/models"

type CountryGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CountryGetResponse struct {
	CountryItem
}
