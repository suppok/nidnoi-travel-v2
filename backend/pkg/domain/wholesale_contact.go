package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: wholesale_contact
[ 0] id            UUID         null: false  primary: true
[ 1] wholesale_id  UUID         null: false  (FK wholesale.id, on delete cascade)
[ 2] name          TEXT         null: false
[ 3] phone         TEXT         null: false
[ 4] sort_order    INTEGER      null: false  default: 0
[ 5] create_date   TIMESTAMPTZ  null: false
[ 6] update_date   TIMESTAMPTZ  null: false
[ 7] create_by     UUID         null: false
[ 8] update_by     UUID         null: false
*/

// WholesaleContact is a row of the wholesale_contact table. Rows are replaced as a whole when the
// wholesale is saved, so there is no status column.
type WholesaleContact struct {
	bun.BaseModel `bun:"table:wholesale_contact,alias:wc" json:"-" swaggerignore:"true"`

	ID          string    `bun:"id,pk"        json:"id"`          // [0] UUID
	WholesaleID string    `bun:"wholesale_id" json:"wholesaleId"` // [1] UUID
	Name        string    `bun:"name"         json:"name"`        // [2] TEXT
	Phone       string    `bun:"phone"        json:"phone"`       // [3] TEXT
	SortOrder   int       `bun:"sort_order"   json:"sortOrder"`   // [4] INTEGER
	CreateDate  time.Time `bun:"create_date"  json:"createDate"`  // [5] TIMESTAMPTZ
	UpdateDate  time.Time `bun:"update_date"  json:"updateDate"`  // [6] TIMESTAMPTZ
	CreateBy    string    `bun:"create_by"    json:"createBy"`    // [7] UUID
	UpdateBy    string    `bun:"update_by"    json:"updateBy"`    // [8] UUID
}
