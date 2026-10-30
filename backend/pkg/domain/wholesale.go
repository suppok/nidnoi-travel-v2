package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: wholesale
[ 0] id           UUID         null: false  primary: true
[ 1] ref_id       INTEGER      null: true   unique: true  (legacy id)
[ 2] code         TEXT         null: false  unique: true  (uppercase)
[ 3] name         TEXT         null: false
[ 4] state        INTEGER      null: false  default: 10  (pkg/constants/wholesalestate)
[ 5] status       BOOLEAN      null: false  default: true
[ 6] create_date  TIMESTAMPTZ  null: false
[ 7] update_date  TIMESTAMPTZ  null: false
[ 8] create_by    UUID         null: false
[ 9] update_by    UUID         null: false
*/

// Wholesale is a row of the wholesale table (tour operators we source tours from).
type Wholesale struct {
	bun.BaseModel `bun:"table:wholesale,alias:ws" json:"-" swaggerignore:"true"`

	ID         string    `bun:"id,pk"       json:"id"`                          // [0] UUID
	RefID      null.Int  `bun:"ref_id"      json:"refId" swaggertype:"integer"` // [1] INTEGER
	Code       string    `bun:"code"        json:"code"`                        // [2] TEXT
	Name       string    `bun:"name"        json:"name"`                        // [3] TEXT
	State      int       `bun:"state"       json:"state"`                       // [4] INTEGER
	Status     bool      `bun:"status"      json:"status"`                      // [5] BOOLEAN
	CreateDate time.Time `bun:"create_date" json:"createDate"`                  // [6] TIMESTAMPTZ
	UpdateDate time.Time `bun:"update_date" json:"updateDate"`                  // [7] TIMESTAMPTZ
	CreateBy   string    `bun:"create_by"   json:"createBy"`                    // [8] UUID
	UpdateBy   string    `bun:"update_by"   json:"updateBy"`                    // [9] UUID

	Contacts []WholesaleContact `bun:"rel:has-many,join:id=wholesale_id" json:"contacts,omitempty"`
}
