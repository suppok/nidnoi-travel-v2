package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

type CityCreateRequest struct {
	RefID       null.Int    `json:"refId" swaggertype:"integer"` // optional legacy destination id
	CountryID   string      `json:"countryId"`                   // must be an active country
	Slug        string      `json:"slug"`                        // locked after create
	NameTh      string      `json:"nameTh"`
	NameEn      string      `json:"nameEn"`
	Description string      `json:"description"`
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CityCreateResponse struct {
	CityItem
}
