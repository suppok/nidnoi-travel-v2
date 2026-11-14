package models

import (
	"time"

	coremodels "backend/pkg/core/api/models"
)

// --- POST /line-bookings/comments: add a comment to a booking's history ---

type LineBookingCommentRequest struct {
	ID   string `json:"id"`   // the booking
	Text string `json:"text"` // 1 to 2000 characters
	// MentionUserIDs are users @mentioned in the text; each must be one of
	// GET /line-bookings/:id/mentionable. They get a "mention" notification.
	MentionUserIDs []string `json:"mentionUserIds"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingCommentResponse struct {
	LineBookingEventItem
}

// --- GET /line-bookings/:id/mentionable: who may be @mentioned on a booking ---

type LineBookingMentionableRequest struct {
	ID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingMentionableResponse struct {
	// Items are the users who can see the booking: its sales owner, its finance user and admins.
	Items []LineBookingUser `json:"items"`
}

// --- POST /line-booking-notifications/list: the caller's notifications ---

type LineBookingNotificationListRequest struct {
	Unread   bool `json:"unread"`   // only unread ones
	Page     int  `json:"page"`     // default 1
	PageSize int  `json:"pageSize"` // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingNotificationListResponse struct {
	Items    []LineBookingNotificationItem `json:"items"`
	Total    int                           `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"pageSize"`
}

// LineBookingNotificationItem is a comment the caller is told about.
type LineBookingNotificationItem struct {
	ID string `json:"id"`
	// Kind is "mention" (the caller was @mentioned) or "comment" (on a booking they own / handle).
	Kind          string           `json:"kind" enums:"mention,comment"`
	LineBookingID string           `json:"lineBookingId"`
	BookingNo     string           `json:"bookingNo"`
	Text          string           `json:"text"`
	Author        *LineBookingUser `json:"author"`
	CreateDate    time.Time        `json:"createDate"`
	ReadAt        *time.Time       `json:"readAt"`
}

// --- PUT /line-booking-notifications/read: mark a booking's notifications read ---

type LineBookingNotificationReadRequest struct {
	LineBookingID string `json:"lineBookingId"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}
