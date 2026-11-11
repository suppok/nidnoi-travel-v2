package models

import (
	"time"

	"backend/pkg/domain"
)

// BookingSourceItem is the booking source shape returned by every booking source endpoint.
type BookingSourceItem struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ShortName  string    `json:"shortName"` // label for the source's reference name ("ชื่อ <shortName>"); "" = none
	SortOrder  int       `json:"sortOrder"`
	Status     bool      `json:"status"`
	CreateDate time.Time `json:"createDate"`
	UpdateDate time.Time `json:"updateDate"`
}

func NewBookingSourceItem(s *domain.BookingSource) BookingSourceItem {
	return BookingSourceItem{
		ID:         s.ID,
		Name:       s.Name,
		ShortName:  s.ShortName,
		SortOrder:  s.SortOrder,
		Status:     s.Status,
		CreateDate: s.CreateDate,
		UpdateDate: s.UpdateDate,
	}
}
