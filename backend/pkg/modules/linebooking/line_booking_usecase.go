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
	List(ctx context.Context, request *models.LineBookingListRequest) (*models.LineBookingListResponse, error)
	Get(ctx context.Context, request *models.LineBookingGetRequest) (*models.LineBookingGetResponse, error)
	Update(ctx context.Context, request *models.LineBookingUpdateRequest) (*models.LineBookingUpdateResponse, error)
	// UpdateLineContact links the booking to the customer's LINE OA chat, or unlinks it.
	UpdateLineContact(ctx context.Context, request *models.LineBookingUpdateLineContactRequest) (*models.LineBookingUpdateLineContactResponse, error)
	UpdateStatus(ctx context.Context, request *models.LineBookingUpdateStatusRequest) (*models.LineBookingUpdateStatusResponse, error)
}
