package domain

import "github.com/uptrace/bun"

/*
DB Table Details
-------------------------------------
Table: tour_program_city
[ 0] tour_program_id  UUID     null: false  primary: true  (FK tour_program.id, on delete cascade)
[ 1] city_id          UUID     null: false  primary: true  (FK city.id)
[ 2] sort_order       INTEGER  null: false  default: 0
*/

// TourProgramCity links a tour to one of its cities. Rows are replaced as a whole when the tour is
// saved.
type TourProgramCity struct {
	bun.BaseModel `bun:"table:tour_program_city,alias:tpci" json:"-" swaggerignore:"true"`

	TourProgramID string `bun:"tour_program_id,pk" json:"tourProgramId"` // [0] UUID
	CityID        string `bun:"city_id,pk"         json:"cityId"`        // [1] UUID
	SortOrder     int    `bun:"sort_order"         json:"sortOrder"`     // [2] INTEGER

	City *City `bun:"rel:belongs-to,join:city_id=id" json:"city,omitempty"`
}
