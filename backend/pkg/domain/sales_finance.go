package domain

import (
	"time"

	"github.com/uptrace/bun"
)

/*
DB Table Details
-------------------------------------
Table: sales_finance
[ 0] id               UUID         null: false  primary: true
[ 1] sales_user_id    UUID         null: false  (FK app_user.id)
[ 2] finance_user_id  UUID         null: false  (FK app_user.id; unique with sales_user_id)
[ 3] is_default       BOOLEAN      null: false  default: false  (at most one per sales user)
[ 4] create_date      TIMESTAMPTZ  null: false
[ 5] update_date      TIMESTAMPTZ  null: false
[ 6] create_by        UUID         null: false
[ 7] update_by        UUID         null: false
*/

// SalesFinance links a sales user to a finance user they can send confirmed LINE bookings to.
// Rows are replaced as a whole when the sales user saves their settings, so there is no status.
type SalesFinance struct {
	bun.BaseModel `bun:"table:sales_finance,alias:sf" json:"-" swaggerignore:"true"`

	ID            string    `bun:"id,pk"           json:"id"`            // [0] UUID
	SalesUserID   string    `bun:"sales_user_id"   json:"salesUserId"`   // [1] UUID
	FinanceUserID string    `bun:"finance_user_id" json:"financeUserId"` // [2] UUID
	IsDefault     bool      `bun:"is_default"      json:"isDefault"`     // [3] BOOLEAN
	CreateDate    time.Time `bun:"create_date"     json:"createDate"`    // [4] TIMESTAMPTZ
	UpdateDate    time.Time `bun:"update_date"     json:"updateDate"`    // [5] TIMESTAMPTZ
	CreateBy      string    `bun:"create_by"       json:"createBy"`      // [6] UUID
	UpdateBy      string    `bun:"update_by"       json:"updateBy"`      // [7] UUID

	FinanceUser *AppUser `bun:"rel:belongs-to,join:finance_user_id=id" json:"financeUser,omitempty"`
}
