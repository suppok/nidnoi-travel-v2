package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: line_booking_event  (append-only: no status / update columns)
[ 0] id               UUID         null: false  primary: true
[ 1] line_booking_id  UUID         null: false  (FK line_booking.id)
[ 2] action           TEXT         null: false  (linebookingevent.*)
[ 3] detail           JSONB        null: false  default: '{}'
[ 4] create_date      TIMESTAMPTZ  null: false
[ 5] create_by        UUID         null: false  (FK app_user.id, who did it)
*/

// LineBookingEventChange is one field changed by an action, as display strings ("" = empty).
type LineBookingEventChange struct {
	Field string `json:"field"` // a key the backoffice labels (adultTwin, totalPrice, customerName, ...)
	From  string `json:"from"`
	To    string `json:"to"`
}

// LineBookingEventDetail is what an action carries; every part is optional.
type LineBookingEventDetail struct {
	Changes []LineBookingEventChange `json:"changes,omitempty"`
	// Target names what the action pointed at: a finance user, a LINE chat, document codes.
	Target string `json:"target,omitempty"`
	Note   string `json:"note,omitempty"`
	// Mentions are the users @mentioned in a comment, with their names at the time.
	Mentions []LineBookingEventMention `json:"mentions,omitempty"`
}

// LineBookingEventMention is a user @mentioned in a comment.
type LineBookingEventMention struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LineBookingEvent is a row of line_booking_event: one action in a LINE booking's history.
type LineBookingEvent struct {
	bun.BaseModel `bun:"table:line_booking_event,alias:lbe" json:"-" swaggerignore:"true"`

	ID            string                 `bun:"id,pk"                 json:"id"`            // [0] UUID
	LineBookingID string                 `bun:"line_booking_id"       json:"lineBookingId"` // [1] UUID
	Action        string                 `bun:"action"                json:"action"`        // [2] TEXT
	Detail        LineBookingEventDetail `bun:"detail,type:jsonb"     json:"detail"`        // [3] JSONB
	CreateDate    time.Time              `bun:"create_date"           json:"createDate"`    // [4] TIMESTAMPTZ
	CreateBy      string                 `bun:"create_by"             json:"createBy"`      // [5] UUID

	CreateUser *AppUser `bun:"rel:belongs-to,join:create_by=id" json:"createUser,omitempty"`
}
