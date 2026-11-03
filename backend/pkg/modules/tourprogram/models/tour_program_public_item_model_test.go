package models

import (
	"testing"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/domain"
)

func day(s string) time.Time {
	t, _ := time.Parse(time.DateOnly, s)
	return t
}

func period(id, start string, price int, before null.Int, seats int) domain.TourProgramPeriod {
	return domain.TourProgramPeriod{
		ID: id, StartDate: day(start), EndDate: day(start).AddDate(0, 0, 4),
		Price: price, PriceBeforeDiscount: before, SeatsAvailable: seats, Status: true,
	}
}

func tour(state int) *domain.TourProgram {
	japan := &domain.Country{Slug: "japan", NameTh: "ญี่ปุ่น", NameEn: "Japan", Status: true}
	gone := &domain.Country{Slug: "old", Status: false}
	return &domain.TourProgram{
		Code: "NIDN000001", Name: "Tokyo", State: state, Days: 5, Nights: 3,
		Airline: &domain.Airline{Code: "TG", Name: "Thai Airways"},
		CountryLinks: []domain.TourProgramCountry{
			{Country: japan},
			{Country: gone},
		},
		CityLinks: []domain.TourProgramCity{
			{City: &domain.City{Slug: "tokyo", NameTh: "โตเกียว", Status: true, Country: japan}},
			{City: &domain.City{Slug: "deleted", Status: false, Country: japan}},
			{City: &domain.City{Slug: "orphan", Status: true, Country: gone}},
		},
	}
}

func TestPublicItemStartingPrice(t *testing.T) {
	periods := []domain.TourProgramPeriod{
		period("a", "2026-12-01", 49900, null.IntFrom(56900), 3),
		period("b", "2026-12-10", 47900, null.IntFrom(54900), 0), // sold out still counts
		period("c", "2027-01-10", 47900, null.Int{}, 20),         // tie: earliest wins
	}
	item := NewTourProgramPublicItem(tour(tourprogramstate.Selling), periods)

	if item.StartingPrice.Int64 != 47900 || item.StartingPriceBeforeDiscount.Int64 != 54900 {
		t.Errorf("starting price = %v / %v, want 47900 / 54900", item.StartingPrice, item.StartingPriceBeforeDiscount)
	}
	if item.NextDepartureDate.String != "2026-12-01" || item.PeriodCount != 3 {
		t.Errorf("next departure %q, count %d", item.NextDepartureDate.String, item.PeriodCount)
	}
	if item.Airline == nil || item.Airline.Code != "TG" {
		t.Errorf("airline = %+v", item.Airline)
	}
}

func TestPublicItemHidesDeletedTaxonomy(t *testing.T) {
	item := NewTourProgramPublicItem(tour(tourprogramstate.Selling), nil)
	if len(item.Countries) != 1 || item.Countries[0].Slug != "japan" {
		t.Errorf("countries = %+v, want only japan", item.Countries)
	}
	if len(item.Cities) != 1 || item.Cities[0].Slug != "tokyo" || item.Cities[0].CountrySlug != "japan" {
		t.Errorf("cities = %+v, want only tokyo/japan", item.Cities)
	}
	if item.StartingPrice.Valid || item.NextDepartureDate.Valid {
		t.Errorf("no periods must leave price and next departure null")
	}
}

func TestPublicPeriodsSoldOut(t *testing.T) {
	periods := []domain.TourProgramPeriod{
		period("a", "2026-12-01", 100, null.Int{}, 5),
		period("b", "2026-12-10", 100, null.Int{}, 0),
	}
	selling := NewTourPeriodPublicItems(tourprogramstate.Selling, periods)
	if selling[0].SoldOut || !selling[1].SoldOut {
		t.Errorf("selling: soldOut = %v, %v; want false, true", selling[0].SoldOut, selling[1].SoldOut)
	}
	if selling[0].StartDate != "2026-12-01" || selling[0].EndDate != "2026-12-05" {
		t.Errorf("dates = %s – %s", selling[0].StartDate, selling[0].EndDate)
	}
	for _, p := range NewTourPeriodPublicItems(tourprogramstate.Closed, periods) {
		if !p.SoldOut {
			t.Errorf("closed tour: period %s not sold out", p.ID)
		}
	}
}
