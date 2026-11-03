package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

type TourProgramWholesale struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

type TourProgramAirline struct {
	ID     string      `json:"id"`
	Code   string      `json:"code"`
	Name   string      `json:"name"`
	Logo   null.String `json:"logo" swaggertype:"string"`
	Status bool        `json:"status"`
}

type TourProgramCountry struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	NameTh string `json:"nameTh"`
	NameEn string `json:"nameEn"`
	Status bool   `json:"status"`
}

type TourProgramCity struct {
	ID        string `json:"id"`
	CountryID string `json:"countryId"`
	Slug      string `json:"slug"`
	NameTh    string `json:"nameTh"`
	NameEn    string `json:"nameEn"`
	Status    bool   `json:"status"`
}

type TourProgramImage struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// TourProgramItem is the full tour returned by get, create and update.
type TourProgramItem struct {
	ID                string               `json:"id"`
	RefID             null.Int             `json:"refId" swaggertype:"integer"` // legacy id
	Code              string               `json:"code"`                        // public, never changes
	Name              string               `json:"name"`
	Wholesale         TourProgramWholesale `json:"wholesale"`
	WholesaleTourCode string               `json:"wholesaleTourCode"`
	State             int                  `json:"state"` // 10 draft, 20 selling, 30 closed, 40 cancelled
	Days              int                  `json:"days"`
	Nights            int                  `json:"nights"`
	Airline           *TourProgramAirline  `json:"airline"` // null when the tour has no airline
	Highlight         string               `json:"highlight"`
	Description       string               `json:"description"`
	BannerImage       null.String          `json:"bannerImage" swaggertype:"string"` // tour detail page hero
	CardImage         null.String          `json:"cardImage" swaggertype:"string"`   // tour cards on home/listing pages
	PDFFile           null.String          `json:"pdfFile" swaggertype:"string"`
	DOCXFile          null.String          `json:"docxFile" swaggertype:"string"`
	InternalNote      string               `json:"internalNote"`
	Countries         []TourProgramCountry `json:"countries"` // display order; the first is the main country
	Cities            []TourProgramCity    `json:"cities"`    // display order
	Images            []TourProgramImage   `json:"images"`    // gallery, display order
	Status            bool                 `json:"status"`
	CreateDate        time.Time            `json:"createDate"`
	UpdateDate        time.Time            `json:"updateDate"`
}

// TourProgramListItem is the lighter row returned by list.
type TourProgramListItem struct {
	ID                string               `json:"id"`
	RefID             null.Int             `json:"refId" swaggertype:"integer"`
	Code              string               `json:"code"`
	Name              string               `json:"name"`
	Wholesale         TourProgramWholesale `json:"wholesale"`
	WholesaleTourCode string               `json:"wholesaleTourCode"`
	State             int                  `json:"state"`
	Days              int                  `json:"days"`
	Nights            int                  `json:"nights"`
	BannerImage       null.String          `json:"bannerImage" swaggertype:"string"`
	CardImage         null.String          `json:"cardImage" swaggertype:"string"`
	Countries         []TourProgramCountry `json:"countries"`
	Status            bool                 `json:"status"`
	CreateDate        time.Time            `json:"createDate"`
	UpdateDate        time.Time            `json:"updateDate"`
}

func newWholesale(w *domain.Wholesale) TourProgramWholesale {
	if w == nil {
		return TourProgramWholesale{}
	}
	return TourProgramWholesale{ID: w.ID, Code: w.Code, Name: w.Name, Status: w.Status}
}

func newAirline(a *domain.Airline) *TourProgramAirline {
	if a == nil {
		return nil
	}
	return &TourProgramAirline{ID: a.ID, Code: a.Code, Name: a.Name, Logo: a.Logo, Status: a.Status}
}

func newCountries(links []domain.TourProgramCountry) []TourProgramCountry {
	out := make([]TourProgramCountry, 0, len(links))
	for _, l := range links {
		if l.Country == nil {
			continue
		}
		c := l.Country
		out = append(out, TourProgramCountry{ID: c.ID, Slug: c.Slug, NameTh: c.NameTh, NameEn: c.NameEn, Status: c.Status})
	}
	return out
}

func newCities(links []domain.TourProgramCity) []TourProgramCity {
	out := make([]TourProgramCity, 0, len(links))
	for _, l := range links {
		if l.City == nil {
			continue
		}
		c := l.City
		out = append(out, TourProgramCity{ID: c.ID, CountryID: c.CountryID, Slug: c.Slug, NameTh: c.NameTh, NameEn: c.NameEn, Status: c.Status})
	}
	return out
}

// NewTourProgramItem expects the Wholesale, Airline, CountryLinks.Country, CityLinks.City and Images to be
// loaded in display order (TourProgramRepository.FindByID).
func NewTourProgramItem(t *domain.TourProgram) TourProgramItem {
	images := make([]TourProgramImage, len(t.Images))
	for i, img := range t.Images {
		images[i] = TourProgramImage{ID: img.ID, URL: img.URL}
	}
	return TourProgramItem{
		ID:                t.ID,
		RefID:             t.RefID,
		Code:              t.Code,
		Name:              t.Name,
		Wholesale:         newWholesale(t.Wholesale),
		WholesaleTourCode: t.WholesaleTourCode,
		State:             t.State,
		Days:              t.Days,
		Nights:            t.Nights,
		Airline:           newAirline(t.Airline),
		Highlight:         t.Highlight,
		Description:       t.Description,
		BannerImage:       t.BannerImage,
		CardImage:         t.CardImage,
		PDFFile:           t.PDFFile,
		DOCXFile:          t.DOCXFile,
		InternalNote:      t.InternalNote,
		Countries:         newCountries(t.CountryLinks),
		Cities:            newCities(t.CityLinks),
		Images:            images,
		Status:            t.Status,
		CreateDate:        t.CreateDate,
		UpdateDate:        t.UpdateDate,
	}
}

// NewTourProgramListItem expects the Wholesale and CountryLinks.Country to be loaded.
func NewTourProgramListItem(t *domain.TourProgram) TourProgramListItem {
	return TourProgramListItem{
		ID:                t.ID,
		RefID:             t.RefID,
		Code:              t.Code,
		Name:              t.Name,
		Wholesale:         newWholesale(t.Wholesale),
		WholesaleTourCode: t.WholesaleTourCode,
		State:             t.State,
		Days:              t.Days,
		Nights:            t.Nights,
		BannerImage:       t.BannerImage,
		CardImage:         t.CardImage,
		Countries:         newCountries(t.CountryLinks),
		Status:            t.Status,
		CreateDate:        t.CreateDate,
		UpdateDate:        t.UpdateDate,
	}
}
