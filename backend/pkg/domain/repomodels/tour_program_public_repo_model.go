package repomodels

import (
	"time"

	"github.com/guregu/null/v5"
)

// TourProgramPublicListCriteria filters GET /tour-programs/public. Only active tours in a public
// state are ever returned; that rule lives in the repository, not here.
type TourProgramPublicListCriteria struct {
	CountrySlug string    // "" = any
	CitySlug    string    // "" = any
	DepartAfter time.Time // only upcoming periods count: start_date > DepartAfter (date)
	DateFrom    null.Time // tours with an upcoming period starting on/after this date
	DateTo      null.Time // ... and on/before this date
	Page        int
	PageSize    int
}

// TaxonomyPublicStats is one country's or city's public tour count and lowest upcoming price.
type TaxonomyPublicStats struct {
	ID            string   `bun:"id"`
	TourCount     int      `bun:"tour_count"`
	StartingPrice null.Int `bun:"starting_price"`
}
