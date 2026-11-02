package domain

import "github.com/uptrace/bun"

/*
DB Table Details
-------------------------------------
Table: tour_program_country
[ 0] tour_program_id  UUID     null: false  primary: true  (FK tour_program.id, on delete cascade)
[ 1] country_id       UUID     null: false  primary: true  (FK country.id)
[ 2] sort_order       INTEGER  null: false  default: 0
*/

// TourProgramCountry links a tour to one of its countries. Rows are replaced as a whole when the
// tour is saved.
type TourProgramCountry struct {
	bun.BaseModel `bun:"table:tour_program_country,alias:tpco" json:"-" swaggerignore:"true"`

	TourProgramID string `bun:"tour_program_id,pk" json:"tourProgramId"` // [0] UUID
	CountryID     string `bun:"country_id,pk"      json:"countryId"`     // [1] UUID
	SortOrder     int    `bun:"sort_order"         json:"sortOrder"`     // [2] INTEGER

	Country *Country `bun:"rel:belongs-to,join:country_id=id" json:"country,omitempty"`
}
