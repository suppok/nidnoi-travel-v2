package models

import (
	coremodels "backend/pkg/core/api/models"
)

type SalesFinanceUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SalesFinanceSettings is the caller's finance users.
type SalesFinanceSettings struct {
	// Options are every active finance user (except the caller), by name.
	Options []SalesFinanceUser `json:"options"`
	// FinanceUserIDs are the caller's finance users (active ones), by name.
	FinanceUserIDs []string `json:"financeUserIds"`
	// DefaultFinanceUserID is preselected in the confirm dialog; "" when none.
	DefaultFinanceUserID string `json:"defaultFinanceUserId"`
}

// --- get ---

type SalesFinanceGetRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type SalesFinanceGetResponse struct {
	SalesFinanceSettings
}

// --- update (replaces the caller's list) ---

type SalesFinanceUpdateRequest struct {
	FinanceUserIDs []string `json:"financeUserIds"` // active finance users, max 20
	// DefaultFinanceUserID must be in FinanceUserIDs; "" = the only one when there is one, else none.
	DefaultFinanceUserID string `json:"defaultFinanceUserId"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type SalesFinanceUpdateResponse struct {
	SalesFinanceSettings
}
