package models

import (
	"time"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
)

// --- GET /line-bookings/:id/events: the booking's history ---

type LineBookingEventListRequest struct {
	ID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingEventListResponse struct {
	// Items are oldest first.
	Items []LineBookingEventItem `json:"items"`
}

// LineBookingEventItem is one action in a booking's history.
type LineBookingEventItem struct {
	ID string `json:"id"`
	// Action is one of pkg/constants/linebookingevent.
	Action string `json:"action" example:"edited"`
	// Changes are the fields the action changed, as display strings (edits, LINE chat, staff note).
	Changes []domain.LineBookingEventChange `json:"changes"`
	// Target names what the action pointed at: a finance user, a LINE chat, document codes ("" = none).
	Target string `json:"target"`
	// Note is the fix request of documents_rejected or the text of a comment ("" = none).
	Note string `json:"note"`
	// Mentions are the users @mentioned in a comment ([] otherwise).
	Mentions   []LineBookingUser `json:"mentions"`
	CreateDate time.Time         `json:"createDate"`
	// User did it; null if the user row is gone.
	User *LineBookingUser `json:"user"`
}
