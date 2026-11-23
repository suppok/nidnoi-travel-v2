package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: app_config
[ 0] id           UUID         null: false  primary: true
[ 1] code         TEXT         null: false  unique: true
[ 2] value        JSONB        null: false
[ 3] model_type   TEXT         null: false  default: ''
[ 4] description  TEXT         null: false  default: ''
[ 5] status       BOOLEAN      null: false  default: true
[ 6] create_date  TIMESTAMPTZ  null: false
[ 7] update_date  TIMESTAMPTZ  null: false
[ 8] create_by    UUID         null: false
[ 9] update_by    UUID         null: false
*/

// AppConfig is a row of the app_config table. Read typed values through
// appconfigload.GetAppConfig with a code from pkg/constants/appconfigcode.
type AppConfig struct {
	bun.BaseModel `bun:"table:app_config,alias:ac" json:"-" swaggerignore:"true"`

	ID   string `bun:"id,pk"       json:"id"`   // [0] UUID
	Code string `bun:"code,unique" json:"code"` // [1] TEXT
	// Value is the raw JSON text. No `type:jsonb` tag on purpose: with it Bun would encode the
	// string as a JSON string literal instead of storing the document it holds.
	Value       string    `bun:"value"       json:"value"`       // [2] JSONB
	ModelType   string    `bun:"model_type"  json:"modelType"`   // [3] TEXT
	Description string    `bun:"description" json:"description"` // [4] TEXT
	Status      bool      `bun:"status"      json:"status"`      // [5] BOOLEAN
	CreateDate  time.Time `bun:"create_date" json:"createDate"`  // [6] TIMESTAMPTZ
	UpdateDate  time.Time `bun:"update_date" json:"updateDate"`  // [7] TIMESTAMPTZ
	CreateBy    string    `bun:"create_by"   json:"createBy"`    // [8] UUID
	UpdateBy    string    `bun:"update_by"   json:"updateBy"`    // [9] UUID
}
