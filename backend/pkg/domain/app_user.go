package domain

import (
	"time"

	"github.com/guregu/null/v5"
	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: app_user
[ 0] id                    UUID         null: false  primary: true
[ 1] email                 TEXT         null: false  unique: true  (lowercase)
[ 2] name                  TEXT         null: false
[ 3] password_hash         TEXT         null: false  (bcrypt)
[ 4] roles                 TEXT[]       null: false  (1+ of admin, sales, finance)
[ 5] must_change_password  BOOLEAN      null: false  default: true
[ 6] last_login_at         TIMESTAMPTZ  null: true
[ 7] status                BOOLEAN      null: false  default: true
[ 8] create_date           TIMESTAMPTZ  null: false
[ 9] update_date           TIMESTAMPTZ  null: false
[10] create_by             UUID         null: false
[11] update_by             UUID         null: false
*/

// AppUser is a row of the app_user table (backoffice staff).
type AppUser struct {
	bun.BaseModel `bun:"table:app_user,alias:au" json:"-" swaggerignore:"true"`

	ID                 string    `bun:"id,pk"                json:"id"`                               // [ 0] UUID
	Email              string    `bun:"email"                json:"email"`                            // [ 1] TEXT
	Name               string    `bun:"name"                 json:"name"`                             // [ 2] TEXT
	PasswordHash       string    `bun:"password_hash"        json:"-"`                                // [ 3] TEXT
	Roles              []string  `bun:"roles,array"          json:"roles"`                            // [ 4] TEXT[]
	MustChangePassword bool      `bun:"must_change_password" json:"mustChangePassword"`               // [ 5] BOOLEAN
	LastLoginAt        null.Time `bun:"last_login_at"        json:"lastLoginAt" swaggertype:"string"` // [ 6] TIMESTAMPTZ
	Status             bool      `bun:"status"               json:"status"`                           // [ 7] BOOLEAN
	CreateDate         time.Time `bun:"create_date"          json:"createDate"`                       // [ 8] TIMESTAMPTZ
	UpdateDate         time.Time `bun:"update_date"          json:"updateDate"`                       // [ 9] TIMESTAMPTZ
	CreateBy           string    `bun:"create_by"            json:"createBy"`                         // [10] UUID
	UpdateBy           string    `bun:"update_by"            json:"updateBy"`                         // [11] UUID
}
