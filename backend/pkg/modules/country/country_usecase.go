package country

import (
	"context"

	"backend/pkg/modules/country/models"
)

type CountryUseCase interface {
	List(ctx context.Context, request *models.CountryListRequest) (*models.CountryListResponse, error)
	Get(ctx context.Context, request *models.CountryGetRequest) (*models.CountryGetResponse, error)
	Create(ctx context.Context, request *models.CountryCreateRequest) (*models.CountryCreateResponse, error)
	Update(ctx context.Context, request *models.CountryUpdateRequest) (*models.CountryUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.CountryUpdateStatusRequest) (*models.CountryUpdateStatusResponse, error)

	// Public (customer site)
	PublicList(ctx context.Context) (*models.CountryPublicListResponse, error)
}
