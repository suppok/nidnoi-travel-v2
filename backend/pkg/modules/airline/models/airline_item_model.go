package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// AirlineItem is the airline shape returned by every staff airline endpoint.
type AirlineItem struct {
	ID         string      `json:"id"`
	RefID      null.Int    `json:"refId" swaggertype:"integer"` // legacy id
	Code       string      `json:"code"`                        // IATA/ICAO, uppercase
	Name       string      `json:"name"`
	Logo       null.String `json:"logo" swaggertype:"string"`
	Status     bool        `json:"status"`
	CreateDate time.Time   `json:"createDate"`
	UpdateDate time.Time   `json:"updateDate"`
}

func NewAirlineItem(a *domain.Airline) AirlineItem {
	return AirlineItem{
		ID:         a.ID,
		RefID:      a.RefID,
		Code:       a.Code,
		Name:       a.Name,
		Logo:       a.Logo,
		Status:     a.Status,
		CreateDate: a.CreateDate,
		UpdateDate: a.UpdateDate,
	}
}
