package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// TourProgramFields are the editable fields shared by create and update. Countries, cities and
// images are replaced as a whole on every save.
type TourProgramFields struct {
	RefID             null.Int    `json:"refId" swaggertype:"integer"` // optional legacy id
	Name              string      `json:"name"`
	WholesaleID       string      `json:"wholesaleId"`
	WholesaleTourCode string      `json:"wholesaleTourCode"` // the wholesale's own code, staff only
	State             int         `json:"state"`             // default 10 (draft)
	Days              int         `json:"days"`
	Nights            int         `json:"nights"`
	AirlineID         null.String `json:"airlineId" swaggertype:"string"`   // optional airline master id
	Highlight         string      `json:"highlight"`                        // short summary for tour cards
	Description       string      `json:"description"`                      // plain text
	BannerImage       null.String `json:"bannerImage" swaggertype:"string"` // URL from POST /media/upload (tour-banner), detail page hero
	CardImage         null.String `json:"cardImage" swaggertype:"string"`   // URL from POST /media/upload (tour-card), tour cards
	PDFFile           null.String `json:"pdfFile" swaggertype:"string"`     // URL from POST /media/upload (tour-pdf), public
	DOCXFile          null.String `json:"docxFile" swaggertype:"string"`    // URL from POST /media/upload (tour-docx), staff only
	InternalNote      string      `json:"internalNote"`                     // staff only
	CountryIDs        []string    `json:"countryIds"`                       // one or more, display order
	CityIDs           []string    `json:"cityIds"`                          // each must belong to one of countryIds
	Images            []string    `json:"images"`                           // gallery URLs (tour-gallery), display order
}

type TourProgramCreateRequest struct {
	TourProgramFields

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type TourProgramCreateResponse struct {
	TourProgramItem
}
