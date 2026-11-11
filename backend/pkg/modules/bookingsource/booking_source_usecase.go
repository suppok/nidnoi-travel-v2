package bookingsource

import (
	"context"

	"backend/pkg/modules/bookingsource/models"
)

type BookingSourceUseCase interface {
	List(ctx context.Context, request *models.BookingSourceListRequest) (*models.BookingSourceListResponse, error)
	Get(ctx context.Context, request *models.BookingSourceGetRequest) (*models.BookingSourceGetResponse, error)
	Create(ctx context.Context, request *models.BookingSourceCreateRequest) (*models.BookingSourceCreateResponse, error)
	Update(ctx context.Context, request *models.BookingSourceUpdateRequest) (*models.BookingSourceUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.BookingSourceUpdateStatusRequest) (*models.BookingSourceUpdateStatusResponse, error)
}
