package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: country
[ 0] id           UUID         null: false  primary: true
[ 1] ref_id       INTEGER      null: true   unique: true  (legacy destination id)
[ 2] slug         TEXT         null: false  unique: true  (locked after create)
[ 3] name_th      TEXT         null: false
[ 4] name_en      TEXT         null: false
[ 5] description  TEXT         null: false  default: ''
[ 6] hero_image   TEXT         null: true
[ 7] sort_order   INTEGER      null: false  default: 0
[ 8] status       BOOLEAN      null: false  default: true
[ 9] create_date  TIMESTAMPTZ  null: false
[10] update_date  TIMESTAMPTZ  null: false
[11] create_by    UUID         null: false
[12] update_by    UUID         null: false
*/

// Country is a row of the country table (tour taxonomy; also continents and multi-country regions).
type Country struct {
	bun.BaseModel `bun:"table:country,alias:co" json:"-" swaggerignore:"true"`

	ID          string      `bun:"id,pk"       json:"id"`                             // [ 0] UUID
	RefID       null.Int    `bun:"ref_id"      json:"refId" swaggertype:"integer"`    // [ 1] INTEGER
	Slug        string      `bun:"slug"        json:"slug"`                           // [ 2] TEXT
	NameTh      string      `bun:"name_th"     json:"nameTh"`                         // [ 3] TEXT
	NameEn      string      `bun:"name_en"     json:"nameEn"`                         // [ 4] TEXT
	Description string      `bun:"description" json:"description"`                    // [ 5] TEXT
	HeroImage   null.String `bun:"hero_image"  json:"heroImage" swaggertype:"string"` // [ 6] TEXT
	SortOrder   int         `bun:"sort_order"  json:"sortOrder"`                      // [ 7] INTEGER
	Status      bool        `bun:"status"      json:"status"`                         // [ 8] BOOLEAN
	CreateDate  time.Time   `bun:"create_date" json:"createDate"`                     // [ 9] TIMESTAMPTZ
	UpdateDate  time.Time   `bun:"update_date" json:"updateDate"`                     // [10] TIMESTAMPTZ
	CreateBy    string      `bun:"create_by"   json:"createBy"`                       // [11] UUID
	UpdateBy    string      `bun:"update_by"   json:"updateBy"`                       // [12] UUID
}
