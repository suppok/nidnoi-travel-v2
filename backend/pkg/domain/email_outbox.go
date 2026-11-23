package domain

import (
	"strconv"
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: email_outbox
[ 0] id                   UUID         null: false  primary: true
[ 1] line_booking_id      UUID         null: false  (FK line_booking.id; unique with type)
[ 2] type                 TEXT         null: false  (pkg/constants/emailoutbox Type*)
[ 3] from_email           TEXT         null: false
[ 4] to_emails            TEXT[]       null: false  (at least one)
[ 5] cc_emails            TEXT[]       null: false  default: '{}'
[ 6] reply_to             TEXT         null: false  default: ''
[ 7] subject              TEXT         null: false
[ 8] body_html            TEXT         null: false
[ 9] body_text            TEXT         null: false
[10] state                INTEGER      null: false  default: 10  (pkg/constants/emailoutbox State*)
[11] attempts             INTEGER      null: false  default: 0   (in the current round)
[12] send_round           INTEGER      null: false  default: 1   (+1 per manual resend)
[13] next_attempt_at      TIMESTAMPTZ  null: false
[14] last_error           TEXT         null: false  default: ''
[15] provider_message_id  TEXT         null: false  default: ''
[16] sent_at              TIMESTAMPTZ  null: true
[17] status               BOOLEAN      null: false  default: true
[18] create_date          TIMESTAMPTZ  null: false
[19] update_date          TIMESTAMPTZ  null: false
[20] create_by            UUID         null: false
[21] update_by            UUID         null: false
*/

// EmailOutbox is a row of the email_outbox table: one queued email.
type EmailOutbox struct {
	bun.BaseModel `bun:"table:email_outbox,alias:eo" json:"-" swaggerignore:"true"`

	ID                string    `bun:"id,pk"               json:"id"`                          // [ 0] UUID
	LineBookingID     string    `bun:"line_booking_id"     json:"lineBookingId"`               // [ 1] UUID
	Type              string    `bun:"type"                json:"type"`                        // [ 2] TEXT
	FromEmail         string    `bun:"from_email"          json:"fromEmail"`                   // [ 3] TEXT
	ToEmails          []string  `bun:"to_emails,array"     json:"toEmails"`                    // [ 4] TEXT[]
	CcEmails          []string  `bun:"cc_emails,array"     json:"ccEmails"`                    // [ 5] TEXT[]
	ReplyTo           string    `bun:"reply_to"            json:"replyTo"`                     // [ 6] TEXT
	Subject           string    `bun:"subject"             json:"subject"`                     // [ 7] TEXT
	BodyHTML          string    `bun:"body_html"           json:"bodyHtml"`                    // [ 8] TEXT
	BodyText          string    `bun:"body_text"           json:"bodyText"`                    // [ 9] TEXT
	State             int       `bun:"state"               json:"state"`                       // [10] INTEGER
	Attempts          int       `bun:"attempts"            json:"attempts"`                    // [11] INTEGER
	SendRound         int       `bun:"send_round"          json:"sendRound"`                   // [12] INTEGER
	NextAttemptAt     time.Time `bun:"next_attempt_at"     json:"nextAttemptAt"`               // [13] TIMESTAMPTZ
	LastError         string    `bun:"last_error"          json:"lastError"`                   // [14] TEXT
	ProviderMessageID string    `bun:"provider_message_id" json:"providerMessageId"`           // [15] TEXT
	SentAt            null.Time `bun:"sent_at"             json:"sentAt" swaggertype:"string"` // [16] TIMESTAMPTZ
	Status            bool      `bun:"status"              json:"status"`                      // [17] BOOLEAN
	CreateDate        time.Time `bun:"create_date"         json:"createDate"`                  // [18] TIMESTAMPTZ
	UpdateDate        time.Time `bun:"update_date"         json:"updateDate"`                  // [19] TIMESTAMPTZ
	CreateBy          string    `bun:"create_by"           json:"createBy"`                    // [20] UUID
	UpdateBy          string    `bun:"update_by"           json:"updateBy"`                    // [21] UUID
}

// IdempotencyKey is the provider key of the current send round: retries of a round are deduplicated
// by the provider, a manual resend (new round) is sent again.
func (e *EmailOutbox) IdempotencyKey() string {
	return e.ID + "-" + strconv.Itoa(e.SendRound)
}
