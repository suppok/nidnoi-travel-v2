package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: booking
[ 0] id                      UUID         null: false  primary: true
[ 1] booking_no              TEXT         null: false  (unique, year + 6 random chars, e.g. 26X8K9AQ)
[ 2] tour_program_id         UUID         null: false  (FK tour_program.id)
[ 3] tour_program_period_id  UUID         null: false  (FK tour_program_period.id)
[ 4] tour_code               TEXT         null: false  (snapshot)
[ 5] tour_name               TEXT         null: false  (snapshot)
[ 6] start_date              DATE         null: false  (snapshot)
[ 7] end_date                DATE         null: false  (snapshot)
[ 8] adult_twin_count        INTEGER      null: false  default: 0
[ 9] adult_twin_price        INTEGER      null: false
[10] adult_single_count      INTEGER      null: false  default: 0
[11] adult_single_price      INTEGER      null: true
[12] child_with_bed_count    INTEGER      null: false  default: 0
[13] child_with_bed_price    INTEGER      null: true
[14] child_no_bed_count      INTEGER      null: false  default: 0
[15] child_no_bed_price      INTEGER      null: true
[16] total_price             INTEGER      null: false
[17] first_name              TEXT         null: false
[18] last_name               TEXT         null: false  default: ''
[19] email                   TEXT         null: false
[20] phone                   TEXT         null: false
[21] customer_note           TEXT         null: false  default: ''
[22] state                   INTEGER      null: false  default: 10
[23] assignee_id             UUID         null: true   (FK app_user.id)
[24] staff_note              TEXT         null: false  default: ''
[25] status                  BOOLEAN      null: false  default: true
[26] create_date             TIMESTAMPTZ  null: false
[27] update_date             TIMESTAMPTZ  null: false
[28] create_by               UUID         null: false
[29] update_by               UUID         null: false
*/

// Booking is a booking request (sales lead) submitted on the customer site.
type Booking struct {
	bun.BaseModel `bun:"table:booking,alias:bk" json:"-" swaggerignore:"true"`

	ID                  string      `bun:"id,pk"                  json:"id"`                                      // [ 0] UUID
	BookingNo           string      `bun:"booking_no"             json:"bookingNo"`                               // [ 1] TEXT
	TourProgramID       string      `bun:"tour_program_id"        json:"tourProgramId"`                           // [ 2] UUID
	TourProgramPeriodID string      `bun:"tour_program_period_id" json:"tourProgramPeriodId"`                     // [ 3] UUID
	TourCode            string      `bun:"tour_code"              json:"tourCode"`                                // [ 4] TEXT
	TourName            string      `bun:"tour_name"              json:"tourName"`                                // [ 5] TEXT
	StartDate           time.Time   `bun:"start_date,type:date"   json:"startDate"`                               // [ 6] DATE
	EndDate             time.Time   `bun:"end_date,type:date"     json:"endDate"`                                 // [ 7] DATE
	AdultTwinCount      int         `bun:"adult_twin_count"       json:"adultTwinCount"`                          // [ 8] INTEGER
	AdultTwinPrice      int         `bun:"adult_twin_price"       json:"adultTwinPrice"`                          // [ 9] INTEGER
	AdultSingleCount    int         `bun:"adult_single_count"     json:"adultSingleCount"`                        // [10] INTEGER
	AdultSinglePrice    null.Int    `bun:"adult_single_price"     json:"adultSinglePrice" swaggertype:"integer"`  // [11] INTEGER
	ChildWithBedCount   int         `bun:"child_with_bed_count"   json:"childWithBedCount"`                       // [12] INTEGER
	ChildWithBedPrice   null.Int    `bun:"child_with_bed_price"   json:"childWithBedPrice" swaggertype:"integer"` // [13] INTEGER
	ChildNoBedCount     int         `bun:"child_no_bed_count"     json:"childNoBedCount"`                         // [14] INTEGER
	ChildNoBedPrice     null.Int    `bun:"child_no_bed_price"     json:"childNoBedPrice" swaggertype:"integer"`   // [15] INTEGER
	TotalPrice          int         `bun:"total_price"            json:"totalPrice"`                              // [16] INTEGER
	FirstName           string      `bun:"first_name"             json:"firstName"`                               // [17] TEXT
	LastName            string      `bun:"last_name"              json:"lastName"`                                // [18] TEXT
	Email               string      `bun:"email"                  json:"email"`                                   // [19] TEXT
	Phone               string      `bun:"phone"                  json:"phone"`                                   // [20] TEXT
	CustomerNote        string      `bun:"customer_note"          json:"customerNote"`                            // [21] TEXT
	State               int         `bun:"state"                  json:"state"`                                   // [22] INTEGER
	AssigneeID          null.String `bun:"assignee_id"            json:"assigneeId" swaggertype:"string"`         // [23] UUID
	StaffNote           string      `bun:"staff_note"             json:"staffNote"`                               // [24] TEXT
	Status              bool        `bun:"status"                 json:"status"`                                  // [25] BOOLEAN
	CreateDate          time.Time   `bun:"create_date"            json:"createDate"`                              // [26] TIMESTAMPTZ
	UpdateDate          time.Time   `bun:"update_date"            json:"updateDate"`                              // [27] TIMESTAMPTZ
	CreateBy            string      `bun:"create_by"              json:"createBy"`                                // [28] UUID
	UpdateBy            string      `bun:"update_by"              json:"updateBy"`                                // [29] UUID

	Assignee *AppUser `bun:"rel:belongs-to,join:assignee_id=id" json:"assignee,omitempty"`
}
