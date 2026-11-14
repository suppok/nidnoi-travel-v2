package linebooking

import (
	"context"

	"backend/pkg/modules/linebooking/models"
)

type LineBookingUseCase interface {
	Create(ctx context.Context, request *models.LineBookingCreateRequest) (*models.LineBookingCreateResponse, error)
	// SalesUserList returns the sales owner options for the LIFF booking form.
	SalesUserList(ctx context.Context, request *models.LineBookingSalesUserListRequest) (*models.LineBookingSalesUserListResponse, error)
	// SourceList returns the booking source options for the LIFF booking form.
	SourceList(ctx context.Context, request *models.LineBookingSourceListRequest) (*models.LineBookingSourceListResponse, error)
	// LIFFGet reopens a deposit booking in LIFF by its booking no.
	LIFFGet(ctx context.Context, request *models.LineBookingLIFFGetRequest) (*models.LineBookingLIFFGetResponse, error)
	// LIFFUpdate edits a reopened booking's deal and/or records the balance slip.
	LIFFUpdate(ctx context.Context, request *models.LineBookingLIFFUpdateRequest) (*models.LineBookingLIFFUpdateResponse, error)
	// ConfirmBalance confirms the balance recorded from LIFF: paid in full, finance issues the RE.
	ConfirmBalance(ctx context.Context, request *models.LineBookingConfirmBalanceRequest) (*models.LineBookingConfirmBalanceResponse, error)
	// LIFFEvents returns the history of a booking reopened in LIFF, oldest first.
	LIFFEvents(ctx context.Context, request *models.LineBookingLIFFGetRequest) (*models.LineBookingEventListResponse, error)
	// Events returns the booking's history, oldest first.
	Events(ctx context.Context, request *models.LineBookingEventListRequest) (*models.LineBookingEventListResponse, error)
	List(ctx context.Context, request *models.LineBookingListRequest) (*models.LineBookingListResponse, error)
	Get(ctx context.Context, request *models.LineBookingGetRequest) (*models.LineBookingGetResponse, error)
	// Update saves the staff note.
	Update(ctx context.Context, request *models.LineBookingUpdateRequest) (*models.LineBookingUpdateResponse, error)
	// Confirm marks a new booking deposit paid or paid in full and sends it to a finance user.
	Confirm(ctx context.Context, request *models.LineBookingConfirmRequest) (*models.LineBookingConfirmResponse, error)
	// Cancel cancels a new booking.
	Cancel(ctx context.Context, request *models.LineBookingCancelRequest) (*models.LineBookingCancelResponse, error)
	// Documents saves the finance documents (AI + IV, or RE) of a confirmed booking.
	Documents(ctx context.Context, request *models.LineBookingDocumentsRequest) (*models.LineBookingDocumentsResponse, error)
	// DocumentsApprove approves the finance documents and pushes them to the customer's LINE chat.
	DocumentsApprove(ctx context.Context, request *models.LineBookingDocumentsApproveRequest) (*models.LineBookingDocumentsApproveResponse, error)
	// DocumentsReject sends the finance documents back to finance with a note.
	DocumentsReject(ctx context.Context, request *models.LineBookingDocumentsRejectRequest) (*models.LineBookingDocumentsRejectResponse, error)
	// FinanceUserList returns the finance users a booking can be sent to on confirm.
	FinanceUserList(ctx context.Context, request *models.LineBookingFinanceUserListRequest) (*models.LineBookingFinanceUserListResponse, error)
	// UpdateLineContact links the booking to the customer's LINE OA chat, or unlinks it.
	UpdateLineContact(ctx context.Context, request *models.LineBookingUpdateLineContactRequest) (*models.LineBookingUpdateLineContactResponse, error)
	UpdateStatus(ctx context.Context, request *models.LineBookingUpdateStatusRequest) (*models.LineBookingUpdateStatusResponse, error)
}
