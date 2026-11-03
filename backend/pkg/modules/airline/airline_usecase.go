package airline

import (
	"context"

	"backend/pkg/modules/airline/models"
)

type AirlineUseCase interface {
	List(ctx context.Context, request *models.AirlineListRequest) (*models.AirlineListResponse, error)
	Get(ctx context.Context, request *models.AirlineGetRequest) (*models.AirlineGetResponse, error)
	Create(ctx context.Context, request *models.AirlineCreateRequest) (*models.AirlineCreateResponse, error)
	Update(ctx context.Context, request *models.AirlineUpdateRequest) (*models.AirlineUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.AirlineUpdateStatusRequest) (*models.AirlineUpdateStatusResponse, error)
}
