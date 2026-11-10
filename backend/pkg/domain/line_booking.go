package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: line_booking
[ 0] id                      UUID         null: false  primary: true
[ 1] booking_no              TEXT         null: false  (unique, year + 6 random chars, e.g. 26X8K9AQ)
[ 2] line_user_id            TEXT         null: false  (LINE user ID of sales who submitted)
[ 3] line_display_name       TEXT         null: false  (LINE display name of sales)
[ 4] tour_program_id         UUID         null: false  (FK tour_program.id)
[ 5] tour_code               TEXT         null: false  (snapshot)
[ 6] tour_name               TEXT         null: false  (snapshot)
[ 7] tour_program_period_id  UUID         null: false  (FK tour_program_period.id)
[ 8] start_date              DATE         null: false  (snapshot)
[ 9] end_date                DATE         null: false  (snapshot)
[10] adult_twin_count        INTEGER      null: false  default: 0
[11] adult_twin_price        INTEGER      null: false
[12] adult_single_count      INTEGER      null: false  default: 0
[13] adult_single_price      INTEGER      null: true
[14] child_with_bed_count    INTEGER      null: false  default: 0
[15] child_with_bed_price    INTEGER      null: true
[16] child_no_bed_count      INTEGER      null: false  default: 0
[17] child_no_bed_price      INTEGER      null: true
[18] total_price             INTEGER      null: false
[19] customer_name           TEXT         null: false
[20] customer_phone          TEXT         null: false
[21] slip_image              TEXT         null: false  (R2 URL)
[22] note                    TEXT         null: false  default: ''
[23] state                   INTEGER      null: false  default: 10
[24] staff_note              TEXT         null: false  default: ''
[25] status                  BOOLEAN      null: false  default: true
[26] create_date             TIMESTAMPTZ  null: false
[27] update_date             TIMESTAMPTZ  null: false
[28] create_by               UUID         null: false
[29] update_by               UUID         null: false
[30] line_contact_id         UUID         null: true   (FK line_contact.id, the customer's LINE OA chat)
[31] sales_user_id           UUID         null: true   (FK app_user.id, the sales person who owns the booking)
*/

// LineBooking is a booking submitted by sales staff from the LIFF app inside LINE OA.
type LineBooking struct {
	bun.BaseModel `bun:"table:line_booking,alias:lb" json:"-" swaggerignore:"true"`

	ID                  string      `bun:"id,pk"                  json:"id"`                                      // [ 0] UUID
	BookingNo           string      `bun:"booking_no"             json:"bookingNo"`                               // [ 1] TEXT
	LineUserID          string      `bun:"line_user_id"           json:"lineUserId"`                              // [ 2] TEXT
	LineDisplayName     string      `bun:"line_display_name"      json:"lineDisplayName"`                         // [ 3] TEXT
	TourProgramID       string      `bun:"tour_program_id"        json:"tourProgramId"`                           // [ 4] UUID
	TourCode            string      `bun:"tour_code"              json:"tourCode"`                                // [ 5] TEXT
	TourName            string      `bun:"tour_name"              json:"tourName"`                                // [ 6] TEXT
	TourProgramPeriodID string      `bun:"tour_program_period_id" json:"tourProgramPeriodId"`                     // [ 7] UUID
	StartDate           time.Time   `bun:"start_date,type:date"   json:"startDate"`                               // [ 8] DATE
	EndDate             time.Time   `bun:"end_date,type:date"     json:"endDate"`                                 // [ 9] DATE
	AdultTwinCount      int         `bun:"adult_twin_count"       json:"adultTwinCount"`                          // [10] INTEGER
	AdultTwinPrice      int         `bun:"adult_twin_price"       json:"adultTwinPrice"`                          // [11] INTEGER
	AdultSingleCount    int         `bun:"adult_single_count"     json:"adultSingleCount"`                        // [12] INTEGER
	AdultSinglePrice    null.Int    `bun:"adult_single_price"     json:"adultSinglePrice" swaggertype:"integer"`  // [13] INTEGER
	ChildWithBedCount   int         `bun:"child_with_bed_count"   json:"childWithBedCount"`                       // [14] INTEGER
	ChildWithBedPrice   null.Int    `bun:"child_with_bed_price"   json:"childWithBedPrice" swaggertype:"integer"` // [15] INTEGER
	ChildNoBedCount     int         `bun:"child_no_bed_count"     json:"childNoBedCount"`                         // [16] INTEGER
	ChildNoBedPrice     null.Int    `bun:"child_no_bed_price"     json:"childNoBedPrice" swaggertype:"integer"`   // [17] INTEGER
	TotalPrice          int         `bun:"total_price"            json:"totalPrice"`                              // [18] INTEGER
	CustomerName        string      `bun:"customer_name"          json:"customerName"`                            // [19] TEXT
	CustomerPhone       string      `bun:"customer_phone"         json:"customerPhone"`                           // [20] TEXT
	SlipImage           string      `bun:"slip_image"             json:"slipImage"`                               // [21] TEXT
	Note                string      `bun:"note"                   json:"note"`                                    // [22] TEXT
	State               int         `bun:"state"                  json:"state"`                                   // [23] INTEGER
	StaffNote           string      `bun:"staff_note"             json:"staffNote"`                               // [24] TEXT
	Status              bool        `bun:"status"                 json:"status"`                                  // [25] BOOLEAN
	CreateDate          time.Time   `bun:"create_date"            json:"createDate"`                              // [26] TIMESTAMPTZ
	UpdateDate          time.Time   `bun:"update_date"            json:"updateDate"`                              // [27] TIMESTAMPTZ
	CreateBy            string      `bun:"create_by"              json:"createBy"`                                // [28] UUID
	UpdateBy            string      `bun:"update_by"              json:"updateBy"`                                // [29] UUID
	LineContactID       null.String `bun:"line_contact_id"      json:"lineContactId" swaggertype:"string"`        // [30] UUID

	SalesUserID null.String `bun:"sales_user_id"          json:"salesUserId" swaggertype:"string"` // [31] UUID

	LineContact *LineContact `bun:"rel:belongs-to,join:line_contact_id=id" json:"lineContact,omitempty"`
	SalesUser   *AppUser     `bun:"rel:belongs-to,join:sales_user_id=id"   json:"salesUser,omitempty"`
}
