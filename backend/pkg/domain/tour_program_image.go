package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: tour_program_image
[ 0] id               UUID         null: false  primary: true
[ 1] tour_program_id  UUID         null: false  (FK tour_program.id, on delete cascade)
[ 2] url              TEXT         null: false
[ 3] sort_order       INTEGER      null: false  default: 0
[ 4] create_date      TIMESTAMPTZ  null: false
[ 5] update_date      TIMESTAMPTZ  null: false
[ 6] create_by        UUID         null: false
[ 7] update_by        UUID         null: false
*/

// TourProgramImage is a gallery image of a tour. Rows are replaced as a whole when the tour is
// saved, so there is no status column.
type TourProgramImage struct {
	bun.BaseModel `bun:"table:tour_program_image,alias:tpi" json:"-" swaggerignore:"true"`

	ID            string    `bun:"id,pk"           json:"id"`            // [0] UUID
	TourProgramID string    `bun:"tour_program_id" json:"tourProgramId"` // [1] UUID
	URL           string    `bun:"url"             json:"url"`           // [2] TEXT
	SortOrder     int       `bun:"sort_order"      json:"sortOrder"`     // [3] INTEGER
	CreateDate    time.Time `bun:"create_date"     json:"createDate"`    // [4] TIMESTAMPTZ
	UpdateDate    time.Time `bun:"update_date"     json:"updateDate"`    // [5] TIMESTAMPTZ
	CreateBy      string    `bun:"create_by"       json:"createBy"`      // [6] UUID
	UpdateBy      string    `bun:"update_by"       json:"updateBy"`      // [7] UUID
}
