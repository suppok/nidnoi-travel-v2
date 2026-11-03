package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

type AirlineCreateRequest struct {
	RefID null.Int    `json:"refId" swaggertype:"integer"` // optional legacy id
	Code  string      `json:"code"`                        // IATA (2) or ICAO (3); stored uppercase
	Name  string      `json:"name"`
	Logo  null.String `json:"logo" swaggertype:"string"` // URL from POST /media/upload (airline-logo)

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AirlineCreateResponse struct {
	AirlineItem
}
