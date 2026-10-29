package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: refresh_token
[ 0] id           UUID         null: false  primary: true
[ 1] user_id      UUID         null: false  (FK app_user.id)
[ 2] token_hash   TEXT         null: false  unique: true  (SHA-256 hex of the cookie value)
[ 3] persistent   BOOLEAN      null: false  ("remember me": persistent cookie)
[ 4] expires_at   TIMESTAMPTZ  null: false
[ 5] revoked_at   TIMESTAMPTZ  null: true
[ 6] replaced_by  UUID         null: true   (token issued by rotation)
[ 7] user_agent   TEXT         null: false  default: ''
[ 8] client_ip    TEXT         null: false  default: ''
[ 9] status       BOOLEAN      null: false  default: true
[10] create_date  TIMESTAMPTZ  null: false
[11] update_date  TIMESTAMPTZ  null: false
[12] create_by    UUID         null: false
[13] update_by    UUID         null: false
*/

// RefreshToken is a row of the refresh_token table.
type RefreshToken struct {
	bun.BaseModel `bun:"table:refresh_token,alias:rt" json:"-" swaggerignore:"true"`

	ID         string      `bun:"id,pk"       json:"id"`                              // [ 0] UUID
	UserID     string      `bun:"user_id"     json:"userId"`                          // [ 1] UUID
	TokenHash  string      `bun:"token_hash"  json:"-"`                               // [ 2] TEXT
	Persistent bool        `bun:"persistent"  json:"persistent"`                      // [ 3] BOOLEAN
	ExpiresAt  time.Time   `bun:"expires_at"  json:"expiresAt"`                       // [ 4] TIMESTAMPTZ
	RevokedAt  null.Time   `bun:"revoked_at"  json:"revokedAt" swaggertype:"string"`  // [ 5] TIMESTAMPTZ
	ReplacedBy null.String `bun:"replaced_by" json:"replacedBy" swaggertype:"string"` // [ 6] UUID
	UserAgent  string      `bun:"user_agent"  json:"userAgent"`                       // [ 7] TEXT
	ClientIP   string      `bun:"client_ip"   json:"clientIp"`                        // [ 8] TEXT
	Status     bool        `bun:"status"      json:"status"`                          // [ 9] BOOLEAN
	CreateDate time.Time   `bun:"create_date" json:"createDate"`                      // [10] TIMESTAMPTZ
	UpdateDate time.Time   `bun:"update_date" json:"updateDate"`                      // [11] TIMESTAMPTZ
	CreateBy   string      `bun:"create_by"   json:"createBy"`                        // [12] UUID
	UpdateBy   string      `bun:"update_by"   json:"updateBy"`                        // [13] UUID
}
