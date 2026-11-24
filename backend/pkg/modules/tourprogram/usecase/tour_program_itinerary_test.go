package usecase

import (
	"strings"
	"testing"

	"backend/pkg/domain"
)

func TestCleanItinerary(t *testing.T) {
	got := cleanItinerary([]domain.TourItineraryDay{
		{Day: 7, Places: []string{"  สนามบิน   สุวรรณภูมิ ", "", "สนามบินปักกิ่ง"}},
		{Day: 8, Places: []string{" ", ""}}, // empty day: dropped
		{Day: 1, Places: []string{"พระราชวังต้องห้าม", "โรงแรม"}},
	})
	if len(got) != 2 || got[0].Day != 1 || got[1].Day != 2 {
		t.Fatalf("days = %+v, want 2 days renumbered 1, 2", got)
	}
	if strings.Join(got[0].Places, "|") != "สนามบิน สุวรรณภูมิ|สนามบินปักกิ่ง" {
		t.Fatalf("places = %q", got[0].Places)
	}
	if cleanItinerary(nil) == nil {
		t.Fatal("must never be nil (the column is NOT NULL)")
	}
}

func TestItineraryRule(t *testing.T) {
	many := make([]string, maxPlacesPerDay+1)
	for i := range many {
		many[i] = "x"
	}
	if itineraryRule([]domain.TourItineraryDay{{Day: 1, Places: many}}) == nil {
		t.Error("too many places must fail")
	}
	if itineraryRule([]domain.TourItineraryDay{{Day: 1, Places: []string{strings.Repeat("ก", maxPlaceRunes+1)}}}) == nil {
		t.Error("a place that is too long must fail")
	}
	if err := itineraryRule([]domain.TourItineraryDay{{Day: 1, Places: []string{"สนามบินสุวรรณภูมิ", "โรงแรม"}}}); err != nil {
		t.Errorf("valid day: %v", err)
	}
}
