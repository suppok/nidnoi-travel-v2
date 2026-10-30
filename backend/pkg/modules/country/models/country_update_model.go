package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// CountryUpdateRequest has no slug: it is locked after create.
type CountryUpdateRequest struct {
	ID          string      `json:"id"`
	RefID       null.Int    `json:"refId" swaggertype:"integer"`
	NameTh      string      `json:"nameTh"`
	NameEn      string      `json:"nameEn"`
	Description string      `json:"description"`
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type CountryUpdateResponse struct {
	CountryItem
}
