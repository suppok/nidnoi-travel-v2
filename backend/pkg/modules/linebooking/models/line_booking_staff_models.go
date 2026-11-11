package models

import (
	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// --- list ---

type LineBookingListRequest struct {
	Keyword string `json:"keyword"` // booking no, customer name, phone, tour code or LINE display name
	State   int    `json:"state"`   // 0 = all
	// Paid lists confirmed bookings only (states 20 and 25), newest confirmation first.
	Paid bool `json:"paid"`
	// AssignedToMe lists only bookings sent to the caller as finance (the finance alert).
	AssignedToMe bool `json:"assignedToMe"`
	// OwnedByMe lists only bookings the caller owns as sales (the documents-ready alert).
	OwnedByMe bool `json:"ownedByMe"`
	// DocumentsDone: true = finance sent the documents (newest first), false = waiting for finance, omitted = any.
	DocumentsDone *bool `json:"documentsDone"`
	// DocumentsApproved: true = the sales owner approved the documents, false = not yet, omitted = any.
	DocumentsApproved *bool `json:"documentsApproved"`
	Status            *bool `json:"status"`   // omitted = all
	Page              int   `json:"page"`     // default 1
	PageSize          int   `json:"pageSize"` // default 20, max 100

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

// --- update the staff note (state changes go through confirm / cancel) ---

type LineBookingUpdateRequest struct {
	ID        string `json:"id"`
	StaffNote string `json:"staffNote"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingUpdateResponse struct {
	LineBookingItem
}

// --- confirm: the sales owner checked the slip; the booking goes to a finance user ---

type LineBookingConfirmRequest struct {
	ID string `json:"id"`
	// FinanceUserID is one of GET /line-bookings/:id/finance-users. The state becomes 20 (deposit
	// paid) or 25 (paid in full) by the booking's payment plan.
	FinanceUserID string `json:"financeUserId"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingConfirmResponse struct {
	LineBookingItem
}

// --- finance documents ---

// LineBookingDocumentsRequest: PDFs from POST /media/upload/line-booking. A deposit booking needs
// aiFile (ใบรับเงินมัดจำ) + ivFile (ใบแจ้งหนี้); a full payment needs reFile (ใบเสร็จรับเงิน).
type LineBookingDocumentsRequest struct {
	ID     string `json:"id"`
	AIFile string `json:"aiFile"`
	IVFile string `json:"ivFile"`
	REFile string `json:"reFile"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingDocumentsResponse struct {
	LineBookingItem
}

// --- sales review of the finance documents ---

// LineBookingDocumentsApproveRequest: the documents are right; push them to the customer's LINE chat
// when one is linked and still follows the OA.
type LineBookingDocumentsApproveRequest struct {
	ID string `json:"id"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingDocumentsApproveResponse struct {
	LineBookingItem
}

// LineBookingDocumentsRejectRequest sends the documents back to finance with a note.
type LineBookingDocumentsRejectRequest struct {
	ID   string `json:"id"`
	Note string `json:"note"` // required, max 1000

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingDocumentsRejectResponse struct {
	LineBookingItem
}

// --- cancel ---

type LineBookingCancelRequest struct {
	ID string `json:"id"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingCancelResponse struct {
	LineBookingItem
}

// --- finance users the booking can be sent to ---

type LineBookingFinanceUserListRequest struct {
	ID string `json:"-"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LineBookingFinanceUserListResponse struct {
	// Items are the sales owner's finance users (active ones); every active finance user when the
	// booking has no owner.
	Items []LineBookingUser `json:"items"`
	// DefaultID is the sales owner's default finance user, "" when none.
	DefaultID string `json:"defaultId"`
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
