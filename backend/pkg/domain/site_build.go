package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: site_build
[ 0] id           UUID         null: false  primary: true
[ 1] trigger      TEXT         null: false
[ 2] outcome      TEXT         null: false
[ 3] paths        TEXT[]       null: false  default: '{}'
[ 4] attempts     INTEGER      null: false  default: 0
[ 5] error        TEXT         null: false  default: ''
[ 6] status       BOOLEAN      null: false  default: true
[ 7] create_date  TIMESTAMPTZ  null: false
[ 8] update_date  TIMESTAMPTZ  null: false
[ 9] create_by    UUID         null: false
[10] update_by    UUID         null: false
*/

// SiteBuild is a row of the site_build table: one customer-site build the backend triggered.
type SiteBuild struct {
	bun.BaseModel `bun:"table:site_build,alias:sb" json:"-" swaggerignore:"true"`

	ID         string    `bun:"id,pk"              json:"id"`         // [ 0] UUID
	Trigger    string    `bun:"trigger"            json:"trigger"`    // [ 1] TEXT
	Outcome    string    `bun:"outcome"            json:"outcome"`    // [ 2] TEXT
	Paths      []string  `bun:"paths,array"        json:"paths"`      // [ 3] TEXT[]
	Attempts   int       `bun:"attempts"           json:"attempts"`   // [ 4] INTEGER
	Error      string    `bun:"error"              json:"error"`      // [ 5] TEXT
	Status     bool      `bun:"status"             json:"status"`     // [ 6] BOOLEAN
	CreateDate time.Time `bun:"create_date"        json:"createDate"` // [ 7] TIMESTAMPTZ
	UpdateDate time.Time `bun:"update_date"        json:"updateDate"` // [ 8] TIMESTAMPTZ
	CreateBy   string    `bun:"create_by"          json:"createBy"`   // [ 9] UUID
	UpdateBy   string    `bun:"update_by"          json:"updateBy"`   // [10] UUID

	// Creator is the user who pressed "rebuild"; nil for system builds.
	Creator *AppUser `bun:"rel:belongs-to,join:create_by=id" json:"creator,omitempty"`
}
