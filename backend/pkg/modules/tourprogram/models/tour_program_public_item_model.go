package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/domain"
)

// publicDateLayout is the wire format of public period dates.
const publicDateLayout = "2006-01-02"

// Public DTOs for the customer site. Never add staff-only fields here (wholesale, docx, notes, audit).

type TaxonomyPublicRef struct {
	Slug   string `json:"slug" example:"japan"`
	NameTh string `json:"nameTh" example:"ญี่ปุ่น"`
	NameEn string `json:"nameEn" example:"Japan"`
}

type CityPublicRef struct {
	TaxonomyPublicRef
	CountrySlug string `json:"countrySlug" example:"japan"`
}

type AirlinePublic struct {
	Code string      `json:"code" example:"TG"`
	Name string      `json:"name" example:"Thai Airways"`
	Logo null.String `json:"logo" swaggertype:"string"`
}

// TourPeriodPublicItem is one upcoming departure. Prices are whole baht; dates are YYYY-MM-DD.
type TourPeriodPublicItem struct {
	ID                  string   `json:"id"`
	StartDate           string   `json:"startDate" example:"2026-12-01"`
	EndDate             string   `json:"endDate" example:"2026-12-05"`
	Price               int      `json:"price"`
	PriceBeforeDiscount null.Int `json:"priceBeforeDiscount" swaggertype:"integer"`
	ChildWithBedPrice   null.Int `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedPrice     null.Int `json:"childNoBedPrice" swaggertype:"integer"`
	SingleSupplement    null.Int `json:"singleSupplement" swaggertype:"integer"`
	Deposit             null.Int `json:"deposit" swaggertype:"integer"`
	SeatsAvailable      int      `json:"seatsAvailable"`
	// SoldOut is true when no seat is left or the tour is closed.
	SoldOut bool `json:"soldOut"`
}

// TourProgramPublicItem is the card / listing shape.
type TourProgramPublicItem struct {
	Code      string              `json:"code" example:"NIDN000001"`
	Name      string              `json:"name"`
	Highlight string              `json:"highlight"`
	Days      int                 `json:"days"`
	Nights    int                 `json:"nights"`
	Airline   *AirlinePublic      `json:"airline"`
	CardImage null.String         `json:"cardImage" swaggertype:"string"`
	Countries []TaxonomyPublicRef `json:"countries"`
	Cities    []CityPublicRef     `json:"cities"`
	// StartingPrice is the lowest upcoming period price ("ราคาเริ่มต้น"); null without upcoming periods.
	StartingPrice null.Int `json:"startingPrice" swaggertype:"integer"`
	// StartingPriceBeforeDiscount is the struck-through price of the period that sets StartingPrice.
	StartingPriceBeforeDiscount null.Int    `json:"startingPriceBeforeDiscount" swaggertype:"integer"`
	NextDepartureDate           null.String `json:"nextDepartureDate" swaggertype:"string" example:"2026-12-01"`
	PeriodCount                 int         `json:"periodCount"`
	UpdateDate                  time.Time   `json:"updateDate"`
}

// TourProgramPublicDetail is the tour page shape.
type TourProgramPublicDetail struct {
	TourProgramPublicItem
	Description string                 `json:"description"`
	BannerImage null.String            `json:"bannerImage" swaggertype:"string"`
	Images      []string               `json:"images"`
	PDFFile     null.String            `json:"pdfFile" swaggertype:"string"`
	Periods     []TourPeriodPublicItem `json:"periods"`
}

// NewTourPeriodPublicItems maps active upcoming periods (already filtered and ordered by the repository).
func NewTourPeriodPublicItems(state int, periods []domain.TourProgramPeriod) []TourPeriodPublicItem {
	closed := state == tourprogramstate.Closed
	items := make([]TourPeriodPublicItem, len(periods))
	for i, p := range periods {
		items[i] = TourPeriodPublicItem{
			ID:                  p.ID,
			StartDate:           p.StartDate.Format(publicDateLayout),
			EndDate:             p.EndDate.Format(publicDateLayout),
			Price:               p.Price,
			PriceBeforeDiscount: p.PriceBeforeDiscount,
			ChildWithBedPrice:   p.ChildWithBedPrice,
			ChildNoBedPrice:     p.ChildNoBedPrice,
			SingleSupplement:    p.SingleSupplement,
			Deposit:             p.Deposit,
			SeatsAvailable:      p.SeatsAvailable,
			SoldOut:             closed || p.SeatsAvailable == 0,
		}
	}
	return items
}

// NewTourProgramPublicItem builds the card shape. periods are the tour's active upcoming periods in
// departure order. Soft-deleted countries and cities (kept on older tours) are left out.
func NewTourProgramPublicItem(t *domain.TourProgram, periods []domain.TourProgramPeriod) TourProgramPublicItem {
	item := TourProgramPublicItem{
		Code:        t.Code,
		Name:        t.Name,
		Highlight:   t.Highlight,
		Days:        t.Days,
		Nights:      t.Nights,
		CardImage:   t.CardImage,
		Countries:   []TaxonomyPublicRef{},
		Cities:      []CityPublicRef{},
		PeriodCount: len(periods),
		UpdateDate:  t.UpdateDate,
	}
	if t.Airline != nil {
		item.Airline = &AirlinePublic{Code: t.Airline.Code, Name: t.Airline.Name, Logo: t.Airline.Logo}
	}
	for _, l := range t.CountryLinks {
		if c := l.Country; c != nil && c.Status {
			item.Countries = append(item.Countries, TaxonomyPublicRef{Slug: c.Slug, NameTh: c.NameTh, NameEn: c.NameEn})
		}
	}
	for _, l := range t.CityLinks {
		if c := l.City; c != nil && c.Status && c.Country != nil && c.Country.Status {
			item.Cities = append(item.Cities, CityPublicRef{
				TaxonomyPublicRef: TaxonomyPublicRef{Slug: c.Slug, NameTh: c.NameTh, NameEn: c.NameEn},
				CountrySlug:       c.Country.Slug,
			})
		}
	}

	var cheapest *domain.TourProgramPeriod
	for i := range periods {
		// Strictly lower keeps the earliest departure on a tie.
		if cheapest == nil || periods[i].Price < cheapest.Price {
			cheapest = &periods[i]
		}
	}
	if cheapest != nil {
		item.StartingPrice = null.IntFrom(int64(cheapest.Price))
		item.StartingPriceBeforeDiscount = cheapest.PriceBeforeDiscount
		item.NextDepartureDate = null.StringFrom(periods[0].StartDate.Format(publicDateLayout))
	}
	return item
}

func NewTourProgramPublicDetail(t *domain.TourProgram, periods []domain.TourProgramPeriod) TourProgramPublicDetail {
	images := make([]string, len(t.Images))
	for i, img := range t.Images {
		images[i] = img.URL
	}
	return TourProgramPublicDetail{
		TourProgramPublicItem: NewTourProgramPublicItem(t, periods),
		Description:           t.Description,
		BannerImage:           t.BannerImage,
		Images:                images,
		PDFFile:               t.PDFFile,
		Periods:               NewTourPeriodPublicItems(t.State, periods),
	}
}
