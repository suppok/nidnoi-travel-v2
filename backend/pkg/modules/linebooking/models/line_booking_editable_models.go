package models

import (
	coremodels "backend/pkg/core/api/models"
)

// --- GET /line-bookings/editable/:bookingNo: reopen a paid booking ---

type LineBookingEditableGetRequest struct {
	BookingNo string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingEditableGetResponse struct {
	LineBookingItem
}

// --- POST /line-bookings/editable/list: the bookings of one customer chat that can be reopened ---

type LineBookingEditableListRequest struct {
	// LineContactID is the customer's chat (POST /line-contacts/list).
	LineContactID string `json:"lineContactId"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

// LineBookingEditableListResponse lists the chat's active paid bookings (deposit or in full), newest
// confirmation first, at most editableListLimit.
type LineBookingEditableListResponse struct {
	Items []LineBookingItem `json:"items"`
}

// --- PUT /line-bookings/editable: edit the deal of a reopened booking and/or record the balance ---

// LineBookingDealUpdateRequest carries the whole deal again, like POST /line-bookings. The tour,
// period, first slip and sales owner stay as they are (tourCode, periodId, slipImage and
// salesUserId are ignored), and so does the payment plan (isDeposit is ignored).
type LineBookingDealUpdateRequest struct {
	BookingNo string `json:"bookingNo" example:"26X8K9AQ"`
	LineBookingCreateRequest
	// BalanceSlipImage is the slip of the balance (R2 URL from /media/upload/line-booking-slip); "" = no balance
	// paid yet. Only for a deposit booking not paid in full (state 20) whose deposit
	// documents were approved; sending one makes the booking
	// wait for the sales owner to confirm the balance.
	BalanceSlipImage string `json:"balanceSlipImage"`
	// BalanceSlipImages are several balance slips (1 to 10); takes over BalanceSlipImage when set.
	BalanceSlipImages []string `json:"balanceSlipImages"`
}

type LineBookingDealUpdateResponse struct {
	BookingNo string `json:"bookingNo" example:"26X8K9AQ"`
	// TotalPrice is the new grand total, card fee included.
	TotalPrice int `json:"totalPrice"`
	// BalancePending: the balance waits for the sales owner's confirm.
	BalancePending bool `json:"balancePending"`
}

// --- PUT /line-bookings/confirm-balance: the sales owner confirms the balance ---

type LineBookingConfirmBalanceRequest struct {
	ID string `json:"id"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingConfirmBalanceResponse struct {
	LineBookingItem
}
