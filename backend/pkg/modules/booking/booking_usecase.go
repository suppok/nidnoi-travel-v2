package booking

import (
	"context"

	"backend/pkg/modules/booking/models"
)

type BookingUseCase interface {
	PublicCreate(ctx context.Context, request *models.BookingPublicCreateRequest) (*models.BookingPublicCreateResponse, error)
	List(ctx context.Context, request *models.BookingListRequest) (*models.BookingListResponse, error)
	Get(ctx context.Context, request *models.BookingGetRequest) (*models.BookingGetResponse, error)
	Update(ctx context.Context, request *models.BookingUpdateRequest) (*models.BookingUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.BookingUpdateStatusRequest) (*models.BookingUpdateStatusResponse, error)
}
