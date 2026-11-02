package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

// DateLayout is the wire format of startDate / endDate.
const DateLayout = "2006-01-02"

// TourProgramPeriodItem is the period shape returned by every staff period endpoint.
type TourProgramPeriodItem struct {
	ID                  string    `json:"id"`
	TourProgramID       string    `json:"tourProgramId"`
	StartDate           string    `json:"startDate" example:"2026-12-01"` // YYYY-MM-DD
	EndDate             string    `json:"endDate" example:"2026-12-05"`   // YYYY-MM-DD
	Price               int       `json:"price"`                          // adult, whole baht
	PriceBeforeDiscount null.Int  `json:"priceBeforeDiscount" swaggertype:"integer"`
	ChildWithBedPrice   null.Int  `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedPrice     null.Int  `json:"childNoBedPrice" swaggertype:"integer"`
	SingleSupplement    null.Int  `json:"singleSupplement" swaggertype:"integer"`
	Deposit             null.Int  `json:"deposit" swaggertype:"integer"`
	SeatsTotal          int       `json:"seatsTotal"`
	SeatsAvailable      int       `json:"seatsAvailable"`
	Note                string    `json:"note"` // staff only
	Status              bool      `json:"status"`
	CreateDate          time.Time `json:"createDate"`
	UpdateDate          time.Time `json:"updateDate"`
}

func NewTourProgramPeriodItem(p *domain.TourProgramPeriod) TourProgramPeriodItem {
	return TourProgramPeriodItem{
		ID:                  p.ID,
		TourProgramID:       p.TourProgramID,
		StartDate:           p.StartDate.Format(DateLayout),
		EndDate:             p.EndDate.Format(DateLayout),
		Price:               p.Price,
		PriceBeforeDiscount: p.PriceBeforeDiscount,
		ChildWithBedPrice:   p.ChildWithBedPrice,
		ChildNoBedPrice:     p.ChildNoBedPrice,
		SingleSupplement:    p.SingleSupplement,
		Deposit:             p.Deposit,
		SeatsTotal:          p.SeatsTotal,
		SeatsAvailable:      p.SeatsAvailable,
		Note:                p.Note,
		Status:              p.Status,
		CreateDate:          p.CreateDate,
		UpdateDate:          p.UpdateDate,
	}
}

// TourProgramPeriodFields are the editable fields shared by create and update.
type TourProgramPeriodFields struct {
	StartDate           string   `json:"startDate" example:"2026-12-01"`            // YYYY-MM-DD
	EndDate             string   `json:"endDate" example:"2026-12-05"`              // YYYY-MM-DD, on or after startDate
	Price               int      `json:"price"`                                     // adult, whole baht
	PriceBeforeDiscount null.Int `json:"priceBeforeDiscount" swaggertype:"integer"` // must be above price
	ChildWithBedPrice   null.Int `json:"childWithBedPrice" swaggertype:"integer"`
	ChildNoBedPrice     null.Int `json:"childNoBedPrice" swaggertype:"integer"`
	SingleSupplement    null.Int `json:"singleSupplement" swaggertype:"integer"`
	Deposit             null.Int `json:"deposit" swaggertype:"integer"`
	SeatsTotal          int      `json:"seatsTotal"`
	SeatsAvailable      int      `json:"seatsAvailable"` // 0..seatsTotal; 0 = sold out
	Note                string   `json:"note"`
}
