package models

import (
	"time"

	"github.com/guregu/null/v5"
)

// CityPublicItem is a city for the customer site, with its public tour stats.
type CityPublicItem struct {
	Slug        string      `json:"slug" example:"tokyo"`
	CountrySlug string      `json:"countrySlug" example:"japan"`
	NameTh      string      `json:"nameTh" example:"โตเกียว"`
	NameEn      string      `json:"nameEn" example:"Tokyo"`
	Description string      `json:"description"` // landing-page intro
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`
	// TourCount counts public tours. The site generates no landing page when it is 0.
	TourCount int `json:"tourCount"`
	// StartingPrice is the lowest upcoming period price among those tours; null when there is none.
	StartingPrice null.Int  `json:"startingPrice" swaggertype:"integer"`
	UpdateDate    time.Time `json:"updateDate"`
}

// CityPublicListResponse lists every active city of an active country in display order (not paged).
type CityPublicListResponse struct {
	Items []CityPublicItem `json:"items"`
}
