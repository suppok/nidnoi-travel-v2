package models

import (
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/domain"
)

type WholesaleContactItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// WholesaleItem is the wholesale shape returned by every staff wholesale endpoint.
type WholesaleItem struct {
	ID         string                 `json:"id"`
	RefID      null.Int               `json:"refId" swaggertype:"integer"` // legacy id
	Code       string                 `json:"code"`
	Name       string                 `json:"name"`
	State      int                    `json:"state"` // 10 on request, 20 connected, 30 blocked, 40 declined
	Contacts   []WholesaleContactItem `json:"contacts"`
	Status     bool                   `json:"status"`
	CreateDate time.Time              `json:"createDate"`
	UpdateDate time.Time              `json:"updateDate"`
}

// NewWholesaleItem expects w.Contacts to be loaded in display order.
func NewWholesaleItem(w *domain.Wholesale) WholesaleItem {
	contacts := make([]WholesaleContactItem, len(w.Contacts))
	for i, c := range w.Contacts {
		contacts[i] = WholesaleContactItem{ID: c.ID, Name: c.Name, Phone: c.Phone}
	}
	return WholesaleItem{
		ID:         w.ID,
		RefID:      w.RefID,
		Code:       w.Code,
		Name:       w.Name,
		State:      w.State,
		Contacts:   contacts,
		Status:     w.Status,
		CreateDate: w.CreateDate,
		UpdateDate: w.UpdateDate,
	}
}
