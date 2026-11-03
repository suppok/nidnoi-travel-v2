package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: airline
[ 0] id           UUID         null: false  primary: true
[ 1] ref_id       INTEGER      null: true   unique: true  (legacy id)
[ 2] code         TEXT         null: false  unique: true  (IATA/ICAO, uppercase)
[ 3] name         TEXT         null: false
[ 4] logo         TEXT         null: true
[ 5] status       BOOLEAN      null: false  default: true
[ 6] create_date  TIMESTAMPTZ  null: false
[ 7] update_date  TIMESTAMPTZ  null: false
[ 8] create_by    UUID         null: false
[ 9] update_by    UUID         null: false
*/

// Airline is a row of the airline table.
type Airline struct {
	bun.BaseModel `bun:"table:airline,alias:al" json:"-" swaggerignore:"true"`

	ID         string      `bun:"id,pk"       json:"id"`                          // [0] UUID
	RefID      null.Int    `bun:"ref_id"      json:"refId" swaggertype:"integer"` // [1] INTEGER
	Code       string      `bun:"code"        json:"code"`                        // [2] TEXT
	Name       string      `bun:"name"        json:"name"`                        // [3] TEXT
	Logo       null.String `bun:"logo"        json:"logo" swaggertype:"string"`   // [4] TEXT
	Status     bool        `bun:"status"      json:"status"`                      // [5] BOOLEAN
	CreateDate time.Time   `bun:"create_date" json:"createDate"`                  // [6] TIMESTAMPTZ
	UpdateDate time.Time   `bun:"update_date" json:"updateDate"`                  // [7] TIMESTAMPTZ
	CreateBy   string      `bun:"create_by"   json:"createBy"`                    // [8] UUID
	UpdateBy   string      `bun:"update_by"   json:"updateBy"`                    // [9] UUID
}
