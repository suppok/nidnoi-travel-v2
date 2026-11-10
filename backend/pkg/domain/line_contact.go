package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: line_contact
[ 0] id                  UUID         null: false  primary: true
[ 1] line_user_id        TEXT         null: false  unique: true  (Messaging API user ID)
[ 2] display_name        TEXT         null: false  default: ''
[ 3] picture_url         TEXT         null: true
[ 4] following           BOOLEAN      null: false  default: false
[ 5] last_message_at     TIMESTAMPTZ  null: true
[ 6] profile_updated_at  TIMESTAMPTZ  null: true
[ 7] status              BOOLEAN      null: false  default: true
[ 8] create_date         TIMESTAMPTZ  null: false
[ 9] update_date         TIMESTAMPTZ  null: false
[10] create_by           UUID         null: false
[11] update_by           UUID         null: false
*/

// LineContact is a customer known from LINE OA webhooks (chat, follow, unfollow).
type LineContact struct {
	bun.BaseModel `bun:"table:line_contact,alias:lc" json:"-" swaggerignore:"true"`

	ID               string      `bun:"id,pk"              json:"id"`                                    // [ 0] UUID
	LineUserID       string      `bun:"line_user_id"       json:"lineUserId"`                            // [ 1] TEXT
	DisplayName      string      `bun:"display_name"       json:"displayName"`                           // [ 2] TEXT
	PictureURL       null.String `bun:"picture_url"        json:"pictureUrl" swaggertype:"string"`       // [ 3] TEXT
	Following        bool        `bun:"following"          json:"following"`                             // [ 4] BOOLEAN
	LastMessageAt    null.Time   `bun:"last_message_at"    json:"lastMessageAt" swaggertype:"string"`    // [ 5] TIMESTAMPTZ
	ProfileUpdatedAt null.Time   `bun:"profile_updated_at" json:"profileUpdatedAt" swaggertype:"string"` // [ 6] TIMESTAMPTZ
	Status           bool        `bun:"status"             json:"status"`                                // [ 7] BOOLEAN
	CreateDate       time.Time   `bun:"create_date"        json:"createDate"`                            // [ 8] TIMESTAMPTZ
	UpdateDate       time.Time   `bun:"update_date"        json:"updateDate"`                            // [ 9] TIMESTAMPTZ
	CreateBy         string      `bun:"create_by"          json:"createBy"`                              // [10] UUID
	UpdateBy         string      `bun:"update_by"          json:"updateBy"`                              // [11] UUID
}
