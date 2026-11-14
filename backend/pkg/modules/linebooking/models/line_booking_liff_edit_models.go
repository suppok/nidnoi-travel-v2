package models

import (
	coremodels "backend/pkg/core/api/models"
)

// --- GET /liff/line-bookings/:bookingNo: reopen a paid booking in LIFF ---

type LineBookingLIFFGetRequest struct {
	BookingNo string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingLIFFGetResponse struct {
	LineBookingItem
}

// --- PUT /liff/line-bookings: edit the deal of a reopened booking and/or record the balance ---

// LineBookingLIFFUpdateRequest carries the whole deal again, like POST /line-bookings. The tour,
// period, first slip and sales owner stay as they are (tourCode, periodId, slipImage and
// salesUserId are ignored), and so does the payment plan (isDeposit is ignored).
type LineBookingLIFFUpdateRequest struct {
	BookingNo string `json:"bookingNo" example:"26X8K9AQ"`
	LineBookingCreateRequest
	// BalanceSlipImage is the slip of the balance (R2 URL from /media/upload/liff); "" = no balance
	// paid yet. Only for a deposit booking not paid in full (state 20) whose deposit
	// documents were approved; sending one makes the booking
	// wait for the sales owner to confirm the balance.
	BalanceSlipImage string `json:"balanceSlipImage"`
}

type LineBookingLIFFUpdateResponse struct {
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
