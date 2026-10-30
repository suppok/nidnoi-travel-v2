package city

import (
	"context"

	"backend/pkg/modules/city/models"
)

type CityUseCase interface {
	List(ctx context.Context, request *models.CityListRequest) (*models.CityListResponse, error)
	Get(ctx context.Context, request *models.CityGetRequest) (*models.CityGetResponse, error)
	Create(ctx context.Context, request *models.CityCreateRequest) (*models.CityCreateResponse, error)
	Update(ctx context.Context, request *models.CityUpdateRequest) (*models.CityUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.CityUpdateStatusRequest) (*models.CityUpdateStatusResponse, error)
}
