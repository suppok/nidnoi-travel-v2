package models

import coremodels "backend/pkg/core/api/models"

// --- list ---

type BookingListRequest struct {
	Keyword  string `json:"keyword"`  // booking no, customer name, email, phone or tour code
	State    int    `json:"state"`    // 0 = all
	Status   *bool  `json:"status"`   // omitted = all
	Page     int    `json:"page"`     // default 1
	PageSize int    `json:"pageSize"` // default 20, max 100

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// BookingListResponse lists bookings newest first. The backoffice polls it with
// {state: 10, status: true, pageSize: 1} for the new-booking badge.
type BookingListResponse struct {
	Items    []BookingItem `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// --- get ---

type BookingGetRequest struct {
	ID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type BookingGetResponse struct {
	BookingItem
}

// --- update (sales workflow) ---

type BookingUpdateRequest struct {
	ID        string `json:"id"`
	State     int    `json:"state"` // 10 new, 20 contacted, 30 confirmed, 40 cancelled
	StaffNote string `json:"staffNote"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type BookingUpdateResponse struct {
	BookingItem
}

// --- status ---

type BookingUpdateStatusRequest struct {
	ID     string `json:"id"`
	Status *bool  `json:"status"` // false = soft delete (spam, duplicates)

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// BookingUpdateStatusResponse is always nil (data: null).
type BookingUpdateStatusResponse struct{}
