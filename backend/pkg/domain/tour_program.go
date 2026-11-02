package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: tour_program
[ 0] id                   UUID         null: false  primary: true
[ 1] ref_id               INTEGER      null: true   unique: true  (legacy id)
[ 2] code                 TEXT         null: false  unique: true  (from tour_code_seq, never changes)
[ 3] name                 TEXT         null: false
[ 4] wholesale_id         UUID         null: false  (FK wholesale.id)
[ 5] wholesale_tour_code  TEXT         null: false  default: ''
[ 6] state                INTEGER      null: false  default: 10  (pkg/constants/tourprogramstate)
[ 7] days                 INTEGER      null: false
[ 8] nights               INTEGER      null: false
[ 9] airline              TEXT         null: false  default: ''
[10] highlight            TEXT         null: false  default: ''
[11] description          TEXT         null: false  default: ''
[12] banner_image         TEXT         null: true   (tour detail page hero)
[13] pdf_file             TEXT         null: true   (public)
[14] docx_file            TEXT         null: true   (staff only)
[15] internal_note        TEXT         null: false  default: ''
[16] status               BOOLEAN      null: false  default: true
[17] create_date          TIMESTAMPTZ  null: false
[18] update_date          TIMESTAMPTZ  null: false
[19] create_by            UUID         null: false
[20] update_by            UUID         null: false
[21] card_image           TEXT         null: true   (tour cards on home/listing pages)
*/

// TourProgram is a row of the tour_program table.
type TourProgram struct {
	bun.BaseModel `bun:"table:tour_program,alias:tp" json:"-" swaggerignore:"true"`

	ID                string      `bun:"id,pk"               json:"id"`                               // [ 0] UUID
	RefID             null.Int    `bun:"ref_id"              json:"refId" swaggertype:"integer"`      // [ 1] INTEGER
	Code              string      `bun:"code"                json:"code"`                             // [ 2] TEXT
	Name              string      `bun:"name"                json:"name"`                             // [ 3] TEXT
	WholesaleID       string      `bun:"wholesale_id"        json:"wholesaleId"`                      // [ 4] UUID
	WholesaleTourCode string      `bun:"wholesale_tour_code" json:"wholesaleTourCode"`                // [ 5] TEXT
	State             int         `bun:"state"               json:"state"`                            // [ 6] INTEGER
	Days              int         `bun:"days"                json:"days"`                             // [ 7] INTEGER
	Nights            int         `bun:"nights"              json:"nights"`                           // [ 8] INTEGER
	Airline           string      `bun:"airline"             json:"airline"`                          // [ 9] TEXT
	Highlight         string      `bun:"highlight"           json:"highlight"`                        // [10] TEXT
	Description       string      `bun:"description"         json:"description"`                      // [11] TEXT
	BannerImage       null.String `bun:"banner_image"        json:"bannerImage" swaggertype:"string"` // [12] TEXT
	CardImage         null.String `bun:"card_image"          json:"cardImage" swaggertype:"string"`   // [21] TEXT
	PDFFile           null.String `bun:"pdf_file"            json:"pdfFile" swaggertype:"string"`     // [13] TEXT
	DOCXFile          null.String `bun:"docx_file"           json:"docxFile" swaggertype:"string"`    // [14] TEXT
	InternalNote      string      `bun:"internal_note"       json:"internalNote"`                     // [15] TEXT
	Status            bool        `bun:"status"              json:"status"`                           // [16] BOOLEAN
	CreateDate        time.Time   `bun:"create_date"         json:"createDate"`                       // [17] TIMESTAMPTZ
	UpdateDate        time.Time   `bun:"update_date"         json:"updateDate"`                       // [18] TIMESTAMPTZ
	CreateBy          string      `bun:"create_by"           json:"createBy"`                         // [19] UUID
	UpdateBy          string      `bun:"update_by"           json:"updateBy"`                         // [20] UUID

	Wholesale    *Wholesale           `bun:"rel:belongs-to,join:wholesale_id=id"   json:"wholesale,omitempty"`
	CountryLinks []TourProgramCountry `bun:"rel:has-many,join:id=tour_program_id" json:"countryLinks,omitempty"`
	CityLinks    []TourProgramCity    `bun:"rel:has-many,join:id=tour_program_id" json:"cityLinks,omitempty"`
	Images       []TourProgramImage   `bun:"rel:has-many,join:id=tour_program_id" json:"images,omitempty"`
}
