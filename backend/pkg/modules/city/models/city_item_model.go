package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// CityCountry is the parent country summary embedded in every city.
type CityCountry struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	NameTh string `json:"nameTh"`
	NameEn string `json:"nameEn"`
	Status bool   `json:"status"`
}

// CityItem is the city shape returned by every staff city endpoint.
type CityItem struct {
	ID          string      `json:"id"`
	RefID       null.Int    `json:"refId" swaggertype:"integer"` // legacy destination id
	CountryID   string      `json:"countryId"`
	Country     CityCountry `json:"country"`
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

// NewCityItem expects c.Country to be loaded.
func NewCityItem(c *domain.City) CityItem {
	item := CityItem{
		ID:          c.ID,
		RefID:       c.RefID,
		CountryID:   c.CountryID,
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
	if c.Country != nil {
		item.Country = CityCountry{ID: c.Country.ID, Slug: c.Country.Slug, NameTh: c.Country.NameTh, NameEn: c.Country.NameEn, Status: c.Country.Status}
	}
	return item
}
