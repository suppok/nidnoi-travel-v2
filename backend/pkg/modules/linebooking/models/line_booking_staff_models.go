package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// --- list ---

type LineBookingListRequest struct {
	Keyword  string `json:"keyword"`  // booking no, customer name, phone, tour code or LINE display name
	State    int    `json:"state"`    // 0 = all
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingListResponse struct {
	Items    []LineBookingItem `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// --- get ---

type LineBookingGetRequest struct {
	ID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingGetResponse struct {
	LineBookingItem
}

// --- update (staff workflow) ---

type LineBookingUpdateRequest struct {
	ID        string `json:"id"`
	State     int    `json:"state"` // 10 new, 20 confirmed, 30 cancelled
	StaffNote string `json:"staffNote"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingUpdateResponse struct {
	LineBookingItem
}

// --- link the customer's LINE OA chat ---

type LineBookingUpdateLineContactRequest struct {
	ID string `json:"id"`
	// LineContactID is a line_contact id from POST /line-contacts/list; null unlinks.
	LineContactID null.String `json:"lineContactId" swaggertype:"string"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingUpdateLineContactResponse struct {
	LineBookingItem
}

// --- status ---

type LineBookingUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingUpdateStatusResponse struct{}
