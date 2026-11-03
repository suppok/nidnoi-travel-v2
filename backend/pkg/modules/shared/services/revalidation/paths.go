package revalidation

import (
	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/domain"
)

// Customer-site paths a mutation affects. Today they are only logged (Cloudflare rebuilds the whole
// site); they are kept so a per-path strategy can come back without touching use cases.
// Use these helpers — never hand-rolled path lists.

// TourVisible reports whether the customer site shows the tour.
func TourVisible(t *domain.TourProgram) bool {
	return t != nil && t.Status && tourprogramstate.IsPublic(t.State)
}

// PathsForTour lists the home page, the tour page and the landings it appears on. The tour must be
// loaded with CountryLinks.Country and CityLinks.City(.Country) for landing paths.
func PathsForTour(t *domain.TourProgram) []string {
	if t == nil {
		return nil
	}
	paths := []string{"/", "/tour/" + t.Code + "/"}
	for _, l := range t.CountryLinks {
		if l.Country != nil {
			paths = append(paths, "/tours/"+l.Country.Slug+"/")
		}
	}
	for _, l := range t.CityLinks {
		if l.City != nil && l.City.Country != nil {
			paths = append(paths, "/tours/"+l.City.Country.Slug+"/"+l.City.Slug+"/")
		}
	}
	return paths
}

// PathsForCountry lists the pages showing a country (home, its landing page).
func PathsForCountry(c *domain.Country) []string {
	return []string{"/", "/tours/" + c.Slug + "/"}
}

// PathsForCity lists the pages showing a city (its country landing and its own landing).
func PathsForCity(countrySlug, citySlug string) []string {
	return []string{"/tours/" + countrySlug + "/", "/tours/" + countrySlug + "/" + citySlug + "/"}
}
