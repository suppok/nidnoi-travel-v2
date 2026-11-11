package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: booking_source
[ 0] id           UUID         null: false  primary: true
[ 1] name         TEXT         null: false  unique: true  (case-insensitive)
[ 2] short_name   TEXT         null: false  default: ''
[ 3] sort_order   INTEGER      null: false  default: 0
[ 4] status       BOOLEAN      null: false  default: true
[ 5] create_date  TIMESTAMPTZ  null: false
[ 6] update_date  TIMESTAMPTZ  null: false
[ 7] create_by    UUID         null: false
[ 8] update_by    UUID         null: false
*/

// BookingSource is a row of the booking_source table.
type BookingSource struct {
	bun.BaseModel `bun:"table:booking_source,alias:bs" json:"-" swaggerignore:"true"`

	ID         string    `bun:"id,pk"       json:"id"`         // [0] UUID
	Name       string    `bun:"name"        json:"name"`       // [1] TEXT
	ShortName  string    `bun:"short_name"  json:"shortName"`  // [2] TEXT
	SortOrder  int       `bun:"sort_order"  json:"sortOrder"`  // [3] INTEGER
	Status     bool      `bun:"status"      json:"status"`     // [4] BOOLEAN
	CreateDate time.Time `bun:"create_date" json:"createDate"` // [5] TIMESTAMPTZ
	UpdateDate time.Time `bun:"update_date" json:"updateDate"` // [6] TIMESTAMPTZ
	CreateBy   string    `bun:"create_by"   json:"createBy"`   // [7] UUID
	UpdateBy   string    `bun:"update_by"   json:"updateBy"`   // [8] UUID
}
