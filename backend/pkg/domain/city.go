package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: city
[ 0] id           UUID         null: false  primary: true
[ 1] ref_id       INTEGER      null: true   unique: true  (legacy destination id)
[ 2] country_id   UUID         null: false  (FK country.id)
[ 3] slug         TEXT         null: false  unique: true  (locked after create)
[ 4] name_th      TEXT         null: false
[ 5] name_en      TEXT         null: false
[ 6] description  TEXT         null: false  default: ''
[ 7] hero_image   TEXT         null: true
[ 8] sort_order   INTEGER      null: false  default: 0
[ 9] status       BOOLEAN      null: false  default: true
[10] create_date  TIMESTAMPTZ  null: false
[11] update_date  TIMESTAMPTZ  null: false
[12] create_by    UUID         null: false
[13] update_by    UUID         null: false
*/

// City is a row of the city table (tour taxonomy; also provinces and sub-country regions).
type City struct {
	bun.BaseModel `bun:"table:city,alias:ci" json:"-" swaggerignore:"true"`

	ID          string      `bun:"id,pk"       json:"id"`                             // [ 0] UUID
	RefID       null.Int    `bun:"ref_id"      json:"refId" swaggertype:"integer"`    // [ 1] INTEGER
	CountryID   string      `bun:"country_id"  json:"countryId"`                      // [ 2] UUID
	Slug        string      `bun:"slug"        json:"slug"`                           // [ 3] TEXT
	NameTh      string      `bun:"name_th"     json:"nameTh"`                         // [ 4] TEXT
	NameEn      string      `bun:"name_en"     json:"nameEn"`                         // [ 5] TEXT
	Description string      `bun:"description" json:"description"`                    // [ 6] TEXT
	HeroImage   null.String `bun:"hero_image"  json:"heroImage" swaggertype:"string"` // [ 7] TEXT
	SortOrder   int         `bun:"sort_order"  json:"sortOrder"`                      // [ 8] INTEGER
	Status      bool        `bun:"status"      json:"status"`                         // [ 9] BOOLEAN
	CreateDate  time.Time   `bun:"create_date" json:"createDate"`                     // [10] TIMESTAMPTZ
	UpdateDate  time.Time   `bun:"update_date" json:"updateDate"`                     // [11] TIMESTAMPTZ
	CreateBy    string      `bun:"create_by"   json:"createBy"`                       // [12] UUID
	UpdateBy    string      `bun:"update_by"   json:"updateBy"`                       // [13] UUID

	Country *Country `bun:"rel:belongs-to,join:country_id=id" json:"country,omitempty"`
}
