package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// CityUpdateRequest has no slug: it is locked after create.
type CityUpdateRequest struct {
	ID          string      `json:"id"`
	RefID       null.Int    `json:"refId" swaggertype:"integer"`
	CountryID   string      `json:"countryId"` // must be an active country when changed
	NameTh      string      `json:"nameTh"`
	NameEn      string      `json:"nameEn"`
	Description string      `json:"description"`
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CityUpdateResponse struct {
	CityItem
}
