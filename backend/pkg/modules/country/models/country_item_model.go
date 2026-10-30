package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// CountryItem is the country shape returned by every staff country endpoint.
type CountryItem struct {
	ID          string      `json:"id"`
	RefID       null.Int    `json:"refId" swaggertype:"integer"` // legacy destination id
	Slug        string      `json:"slug"`
	NameTh      string      `json:"nameTh"`
	NameEn      string      `json:"nameEn"`
	Description string      `json:"description"`
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`
	Status      bool        `json:"status"`
	CreateDate  time.Time   `json:"createDate"`
	UpdateDate  time.Time   `json:"updateDate"`
}

func NewCountryItem(c *domain.Country) CountryItem {
	return CountryItem{
		ID:          c.ID,
		RefID:       c.RefID,
		Slug:        c.Slug,
		NameTh:      c.NameTh,
		NameEn:      c.NameEn,
		Description: c.Description,
		HeroImage:   c.HeroImage,
		SortOrder:   c.SortOrder,
		Status:      c.Status,
		CreateDate:  c.CreateDate,
		UpdateDate:  c.UpdateDate,
	}
}
