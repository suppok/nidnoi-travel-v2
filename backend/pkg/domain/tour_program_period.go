package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: tour_program_period
[ 0] id                     UUID         null: false  primary: true
[ 1] tour_program_id        UUID         null: false  (FK tour_program.id)
[ 2] start_date             DATE         null: false
[ 3] end_date               DATE         null: false
[ 4] price                  INTEGER      null: false  (adult, whole baht)
[ 5] price_before_discount  INTEGER      null: true
[ 6] child_with_bed_price   INTEGER      null: true
[ 7] child_no_bed_price     INTEGER      null: true
[ 8] single_supplement      INTEGER      null: true
[ 9] deposit                INTEGER      null: true
[10] seats_total            INTEGER      null: false
[11] seats_available        INTEGER      null: false
[12] note                   TEXT         null: false  default: ''  (staff only)
[13] status                 BOOLEAN      null: false  default: true
[14] create_date            TIMESTAMPTZ  null: false
[15] update_date            TIMESTAMPTZ  null: false
[16] create_by              UUID         null: false
[17] update_by              UUID         null: false
*/

// TourProgramPeriod is a departure of a tour with its prices and seats.
type TourProgramPeriod struct {
	bun.BaseModel `bun:"table:tour_program_period,alias:tpp" json:"-" swaggerignore:"true"`

	ID                  string    `bun:"id,pk"                 json:"id"`                                        // [ 0] UUID
	TourProgramID       string    `bun:"tour_program_id"       json:"tourProgramId"`                             // [ 1] UUID
	StartDate           time.Time `bun:"start_date,type:date"  json:"startDate"`                                 // [ 2] DATE
	EndDate             time.Time `bun:"end_date,type:date"    json:"endDate"`                                   // [ 3] DATE
	Price               int       `bun:"price"                 json:"price"`                                     // [ 4] INTEGER
	PriceBeforeDiscount null.Int  `bun:"price_before_discount" json:"priceBeforeDiscount" swaggertype:"integer"` // [ 5] INTEGER
	ChildWithBedPrice   null.Int  `bun:"child_with_bed_price"  json:"childWithBedPrice" swaggertype:"integer"`   // [ 6] INTEGER
	ChildNoBedPrice     null.Int  `bun:"child_no_bed_price"    json:"childNoBedPrice" swaggertype:"integer"`     // [ 7] INTEGER
	SingleSupplement    null.Int  `bun:"single_supplement"     json:"singleSupplement" swaggertype:"integer"`    // [ 8] INTEGER
	Deposit             null.Int  `bun:"deposit"               json:"deposit" swaggertype:"integer"`             // [ 9] INTEGER
	SeatsTotal          int       `bun:"seats_total"           json:"seatsTotal"`                                // [10] INTEGER
	SeatsAvailable      int       `bun:"seats_available"       json:"seatsAvailable"`                            // [11] INTEGER
	Note                string    `bun:"note"                  json:"note"`                                      // [12] TEXT
	Status              bool      `bun:"status"                json:"status"`                                    // [13] BOOLEAN
	CreateDate          time.Time `bun:"create_date"           json:"createDate"`                                // [14] TIMESTAMPTZ
	UpdateDate          time.Time `bun:"update_date"           json:"updateDate"`                                // [15] TIMESTAMPTZ
	CreateBy            string    `bun:"create_by"             json:"createBy"`                                  // [16] UUID
	UpdateBy            string    `bun:"update_by"             json:"updateBy"`                                  // [17] UUID
}
