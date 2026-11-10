package models

import coremodels "backend/pkg/core/api/models"

// --- GET /liff/sales-users: sales owner options for the LIFF booking form ---

type LineBookingSalesUserListRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingSalesUserListResponse struct {
	// Items are active users with the sales role, by name.
	Items []LineBookingUser `json:"items"`
	// DefaultID is the caller when they are in Items, otherwise "".
	DefaultID string `json:"defaultId"`
}
