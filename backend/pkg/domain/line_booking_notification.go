package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: line_booking_notification  (no status / update columns: only read_at changes)
[ 0] id               UUID         null: false  primary: true
[ 1] user_id          UUID         null: false  (FK app_user.id, the recipient)
[ 2] line_booking_id  UUID         null: false  (FK line_booking.id)
[ 3] event_id         UUID         null: false  (FK line_booking_event.id, the comment)
[ 4] kind             TEXT         null: false  (linebookingevent.Notify*)
[ 5] read_at          TIMESTAMPTZ  null: true   (null = unread; set when the recipient opens the booking)
[ 6] create_date      TIMESTAMPTZ  null: false
*/

// LineBookingNotification tells a staff user about a comment on a LINE booking.
type LineBookingNotification struct {
	bun.BaseModel `bun:"table:line_booking_notification,alias:lbn" json:"-" swaggerignore:"true"`

	ID            string    `bun:"id,pk"           json:"id"`            // [0] UUID
	UserID        string    `bun:"user_id"         json:"userId"`        // [1] UUID
	LineBookingID string    `bun:"line_booking_id" json:"lineBookingId"` // [2] UUID
	EventID       string    `bun:"event_id"        json:"eventId"`       // [3] UUID
	Kind          string    `bun:"kind"            json:"kind"`          // [4] TEXT
	ReadAt        null.Time `bun:"read_at"         json:"readAt"`        // [5] TIMESTAMPTZ
	CreateDate    time.Time `bun:"create_date"     json:"createDate"`    // [6] TIMESTAMPTZ

	Event       *LineBookingEvent `bun:"rel:belongs-to,join:event_id=id"        json:"event,omitempty"`
	LineBooking *LineBooking      `bun:"rel:belongs-to,join:line_booking_id=id" json:"lineBooking,omitempty"`
}
