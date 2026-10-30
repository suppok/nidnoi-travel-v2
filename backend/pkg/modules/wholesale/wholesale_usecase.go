package wholesale

import (
	"context"

	"backend/pkg/modules/wholesale/models"
)

type WholesaleUseCase interface {
	List(ctx context.Context, request *models.WholesaleListRequest) (*models.WholesaleListResponse, error)
	Get(ctx context.Context, request *models.WholesaleGetRequest) (*models.WholesaleGetResponse, error)
	Create(ctx context.Context, request *models.WholesaleCreateRequest) (*models.WholesaleCreateResponse, error)
	Update(ctx context.Context, request *models.WholesaleUpdateRequest) (*models.WholesaleUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.WholesaleUpdateStatusRequest) (*models.WholesaleUpdateStatusResponse, error)
}
