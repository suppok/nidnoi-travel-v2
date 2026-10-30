package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

type CountryCreateRequest struct {
	RefID       null.Int    `json:"refId" swaggertype:"integer"` // optional legacy destination id
	Slug        string      `json:"slug"`                        // locked after create
	NameTh      string      `json:"nameTh"`
	NameEn      string      `json:"nameEn"`
	Description string      `json:"description"`
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CountryCreateResponse struct {
	CountryItem
}
