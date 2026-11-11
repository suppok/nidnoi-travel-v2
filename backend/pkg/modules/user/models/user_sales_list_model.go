package models

import coremodels "backend/pkg/core/api/models"

// --- GET /users/sales: sales person options for the backoffice sales tools ---

type UserSalesListRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type UserSalesItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"` // digits only; "" = none
}

type UserSalesListResponse struct {
	// Items are active users with the sales role, by name.
	Items []UserSalesItem `json:"items"`
	// DefaultID is the caller when they are in Items, otherwise "".
	DefaultID string `json:"defaultId"`
}
