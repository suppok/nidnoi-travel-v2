package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

type AirlineUpdateRequest struct {
	ID    string      `json:"id"`
	RefID null.Int    `json:"refId" swaggertype:"integer"`
	Code  string      `json:"code"`
	Name  string      `json:"name"`
	Logo  null.String `json:"logo" swaggertype:"string"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AirlineUpdateResponse struct {
	AirlineItem
}
