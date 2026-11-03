package models

import (
	"time"

	"github.com/guregu/null/v5"
)

// CountryPublicItem is a country for the customer site, with its public tour stats.
type CountryPublicItem struct {
	Slug        string      `json:"slug" example:"japan"`
	NameTh      string      `json:"nameTh" example:"ญี่ปุ่น"`
	NameEn      string      `json:"nameEn" example:"Japan"`
	Description string      `json:"description"` // landing-page intro
	HeroImage   null.String `json:"heroImage" swaggertype:"string"`
	SortOrder   int         `json:"sortOrder"`
	// TourCount counts public tours. The site generates no landing page when it is 0.
	TourCount int `json:"tourCount"`
	// StartingPrice is the lowest upcoming period price among those tours; null when there is none.
	StartingPrice null.Int  `json:"startingPrice" swaggertype:"integer"`
	UpdateDate    time.Time `json:"updateDate"`
}

// CountryPublicListResponse lists every active country in display order (not paged).
type CountryPublicListResponse struct {
	Items []CountryPublicItem `json:"items"`
}
