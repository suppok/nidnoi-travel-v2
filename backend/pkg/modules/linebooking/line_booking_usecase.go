package linebooking

import (
	"context"

	"backend/pkg/modules/linebooking/models"
)

type LineBookingUseCase interface {
	Create(ctx context.Context, request *models.LineBookingCreateRequest) (*models.LineBookingCreateResponse, error)
	// EditableGet reopens a paid booking by its booking no.
	EditableGet(ctx context.Context, request *models.LineBookingEditableGetRequest) (*models.LineBookingEditableGetResponse, error)
	// EditableList lists the bookings of one customer chat that can be reopened.
	EditableList(ctx context.Context, request *models.LineBookingEditableListRequest) (*models.LineBookingEditableListResponse, error)
	// UpdateDeal edits a reopened booking's deal and/or records the balance slip.
	UpdateDeal(ctx context.Context, request *models.LineBookingDealUpdateRequest) (*models.LineBookingDealUpdateResponse, error)
	// ConfirmBalance confirms the balance sales recorded: paid in full, finance issues the RE.
	ConfirmBalance(ctx context.Context, request *models.LineBookingConfirmBalanceRequest) (*models.LineBookingConfirmBalanceResponse, error)
	// EditableEvents returns the history of a reopened booking, oldest first.
	EditableEvents(ctx context.Context, request *models.LineBookingEditableGetRequest) (*models.LineBookingEventListResponse, error)
	// Comment adds a comment (with @mentions) to the booking's history and notifies.
	Comment(ctx context.Context, request *models.LineBookingCommentRequest) (*models.LineBookingCommentResponse, error)
	// Mentionable lists who may be @mentioned on a booking.
	Mentionable(ctx context.Context, request *models.LineBookingMentionableRequest) (*models.LineBookingMentionableResponse, error)
	// Notifications lists the caller's comment notifications, newest first.
	Notifications(ctx context.Context, request *models.LineBookingNotificationListRequest) (*models.LineBookingNotificationListResponse, error)
	// NotificationsRead marks the caller's notifications on one booking read.
	NotificationsRead(ctx context.Context, request *models.LineBookingNotificationReadRequest) error
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
