package models

import coremodels "backend/pkg/core/api/models"

type WholesaleGetRequest struct {
	ID string `json:"-"` // path param

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type WholesaleGetResponse struct {
	WholesaleItem
}
