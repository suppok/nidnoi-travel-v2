package tourprogramperiod

import (
	"context"

	"backend/pkg/modules/tourprogramperiod/models"
)

type TourProgramPeriodUseCase interface {
	List(ctx context.Context, request *models.TourProgramPeriodListRequest) (*models.TourProgramPeriodListResponse, error)
	Create(ctx context.Context, request *models.TourProgramPeriodCreateRequest) (*models.TourProgramPeriodCreateResponse, error)
	Update(ctx context.Context, request *models.TourProgramPeriodUpdateRequest) (*models.TourProgramPeriodUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.TourProgramPeriodUpdateStatusRequest) (*models.TourProgramPeriodUpdateStatusResponse, error)
}
