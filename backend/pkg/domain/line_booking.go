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
[32] booking_source_id       UUID         null: true   (FK booking_source.id)
[33] ref_name                TEXT         null: false  default: ''  (customer's name on the source)
[34] ref_booking_no          TEXT         null: false  default: ''  (booking number on the source)
[35] customer_address        TEXT         null: false  default: ''
[36] infant_count            INTEGER      null: false  default: 0
[37] infant_price            INTEGER      null: true
[38] adult_price_before_discount INTEGER  null: true
[39] discount_per_person     INTEGER      null: false  default: 0
[40] charges                 JSONB        null: false  default: '[]'
[41] payment_method          TEXT         null: false  default: 'transfer'  (transfer | credit)
[42] credit_fee_bps          INTEGER      null: false  default: 0  (300 = 3%)
[43] card_fee                INTEGER      null: false  default: 0
[44] is_deposit              BOOLEAN      null: false  default: false
[45] deposit_per_person      INTEGER      null: false  default: 0
[46] deposit_due_date        DATE         null: true
[47] deposit_due_time        TEXT         null: false  default: ''  (HH:mm)
[48] remaining_due_date      DATE         null: true
[49] full_due_date           DATE         null: true
[50] full_due_time           TEXT         null: false  default: ''  (HH:mm)
[51] finance_user_id         UUID         null: true   (FK app_user.id, the finance user who issues the receipt; set on confirm)
[52] confirmed_at            TIMESTAMPTZ  null: true   (when the sales owner confirmed the payment)
[53] ai_file                 TEXT         null: false  default: ''  (PDF URL, AI ใบรับเงินมัดจำ; deposit)
[54] iv_file                 TEXT         null: false  default: ''  (PDF URL, IV ใบแจ้งหนี้; deposit)
[55] re_file                 TEXT         null: false  default: ''  (PDF URL, RE ใบเสร็จรับเงิน; full payment)
[56] documents_at            TIMESTAMPTZ  null: true   (when finance first sent the documents; null = waiting for finance)
[57] documents_by            UUID         null: true   (FK app_user.id, the finance user who sent them)
[58] documents_approved_at   TIMESTAMPTZ  null: true   (the sales owner approved the documents)
[59] documents_approved_by   UUID         null: true   (FK app_user.id)
[60] customer_sent_at        TIMESTAMPTZ  null: true   (documents card pushed to the customer's LINE chat; null = not pushed)
[61] customer_sent_contact_id UUID        null: true   (FK line_contact.id, the chat it was pushed to)
[62] fix_note                TEXT         null: false  default: ''  (latest fix request from sales to finance)
[63] fix_requested_at        TIMESTAMPTZ  null: true
*/

// LineBookingCharge is one extra charge line (visa, taxes, options, ...): count x price.
type LineBookingCharge struct {
	Kind  string `json:"kind"` // linebookingpayment.Charge*
	Name  string `json:"name"` // only for kind "option"
	Count int    `json:"count"`
	Price int    `json:"price"`
}

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

	BookingSourceID          null.String         `bun:"booking_source_id"            json:"bookingSourceId" swaggertype:"string"`           // [32] UUID
	RefName                  string              `bun:"ref_name"                     json:"refName"`                                        // [33] TEXT
	RefBookingNo             string              `bun:"ref_booking_no"               json:"refBookingNo"`                                   // [34] TEXT
	CustomerAddress          string              `bun:"customer_address"             json:"customerAddress"`                                // [35] TEXT
	InfantCount              int                 `bun:"infant_count"                 json:"infantCount"`                                    // [36] INTEGER
	InfantPrice              null.Int            `bun:"infant_price"                 json:"infantPrice" swaggertype:"integer"`              // [37] INTEGER
	AdultPriceBeforeDiscount null.Int            `bun:"adult_price_before_discount"  json:"adultPriceBeforeDiscount" swaggertype:"integer"` // [38] INTEGER
	DiscountPerPerson        int                 `bun:"discount_per_person"          json:"discountPerPerson"`                              // [39] INTEGER
	Charges                  []LineBookingCharge `bun:"charges,type:jsonb"           json:"charges"`                                        // [40] JSONB
	PaymentMethod            string              `bun:"payment_method"               json:"paymentMethod"`                                  // [41] TEXT
	CreditFeeBps             int                 `bun:"credit_fee_bps"               json:"creditFeeBps"`                                   // [42] INTEGER
	CardFee                  int                 `bun:"card_fee"                     json:"cardFee"`                                        // [43] INTEGER
	IsDeposit                bool                `bun:"is_deposit"                   json:"isDeposit"`                                      // [44] BOOLEAN
	DepositPerPerson         int                 `bun:"deposit_per_person"           json:"depositPerPerson"`                               // [45] INTEGER
	DepositDueDate           *time.Time          `bun:"deposit_due_date,type:date"   json:"depositDueDate" swaggertype:"string"`            // [46] DATE
	DepositDueTime           string              `bun:"deposit_due_time"             json:"depositDueTime"`                                 // [47] TEXT
	RemainingDueDate         *time.Time          `bun:"remaining_due_date,type:date" json:"remainingDueDate" swaggertype:"string"`          // [48] DATE
	FullDueDate              *time.Time          `bun:"full_due_date,type:date"      json:"fullDueDate" swaggertype:"string"`               // [49] DATE
	FullDueTime              string              `bun:"full_due_time"                json:"fullDueTime"`                                    // [50] TEXT
	FinanceUserID            null.String         `bun:"finance_user_id"              json:"financeUserId" swaggertype:"string"`             // [51] UUID
	ConfirmedAt              null.Time           `bun:"confirmed_at"                 json:"confirmedAt" swaggertype:"string"`               // [52] TIMESTAMPTZ
	AIFile                   string              `bun:"ai_file"                      json:"aiFile"`                                         // [53] TEXT
	IVFile                   string              `bun:"iv_file"                      json:"ivFile"`                                         // [54] TEXT
	REFile                   string              `bun:"re_file"                      json:"reFile"`                                         // [55] TEXT
	DocumentsAt              null.Time           `bun:"documents_at"                 json:"documentsAt" swaggertype:"string"`               // [56] TIMESTAMPTZ
	DocumentsBy              null.String         `bun:"documents_by"                 json:"documentsBy" swaggertype:"string"`               // [57] UUID
	DocumentsApprovedAt      null.Time           `bun:"documents_approved_at"        json:"documentsApprovedAt" swaggertype:"string"`       // [58] TIMESTAMPTZ
	DocumentsApprovedBy      null.String         `bun:"documents_approved_by"        json:"documentsApprovedBy" swaggertype:"string"`       // [59] UUID
	CustomerSentAt           null.Time           `bun:"customer_sent_at"             json:"customerSentAt" swaggertype:"string"`            // [60] TIMESTAMPTZ
	CustomerSentContactID    null.String         `bun:"customer_sent_contact_id"     json:"customerSentContactId" swaggertype:"string"`     // [61] UUID
	FixNote                  string              `bun:"fix_note"                     json:"fixNote"`                                        // [62] TEXT
	FixRequestedAt           null.Time           `bun:"fix_requested_at"             json:"fixRequestedAt" swaggertype:"string"`            // [63] TIMESTAMPTZ

	LineContact *LineContact `bun:"rel:belongs-to,join:line_contact_id=id" json:"lineContact,omitempty"`
	SalesUser   *AppUser     `bun:"rel:belongs-to,join:sales_user_id=id"   json:"salesUser,omitempty"`
	FinanceUser *AppUser     `bun:"rel:belongs-to,join:finance_user_id=id" json:"financeUser,omitempty"`
	// BookingSource is where the customer booked (LINE OA, Facebook, ...).
	BookingSource *BookingSource `bun:"rel:belongs-to,join:booking_source_id=id" json:"bookingSource,omitempty"`
}

// Travellers counts everyone on the booking, infants included.
func (b *LineBooking) Travellers() int {
	return b.AdultTwinCount + b.AdultSingleCount + b.ChildWithBedCount + b.ChildNoBedCount + b.InfantCount
}
