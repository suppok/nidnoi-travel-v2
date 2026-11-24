package tourprogram

import (
	"context"

	"backend/pkg/modules/tourprogram/models"
)

type TourProgramUseCase interface {
	List(ctx context.Context, request *models.TourProgramListRequest) (*models.TourProgramListResponse, error)
	Get(ctx context.Context, request *models.TourProgramGetRequest) (*models.TourProgramGetResponse, error)
	Create(ctx context.Context, request *models.TourProgramCreateRequest) (*models.TourProgramCreateResponse, error)
	Update(ctx context.Context, request *models.TourProgramUpdateRequest) (*models.TourProgramUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.TourProgramUpdateStatusRequest) (*models.TourProgramUpdateStatusResponse, error)

	// AI prefill from a PDF (nothing is saved): AIExtract starts a background job, AIExtractResult polls it.
	AIExtractConfig(ctx context.Context, request *models.TourProgramAIExtractConfigRequest) (*models.TourProgramAIExtractConfigResponse, error)
	AIExtract(ctx context.Context, request *models.TourProgramAIExtractRequest) (*models.TourProgramAIExtractStartResponse, error)
	AIExtractResult(ctx context.Context, request *models.TourProgramAIExtractResultRequest) (*models.TourProgramAIExtractResultResponse, error)

	// Public (customer site)
	PublicList(ctx context.Context, request *models.TourProgramPublicListRequest) (*models.TourProgramPublicListResponse, error)
	PublicGet(ctx context.Context, request *models.TourProgramPublicGetRequest) (*models.TourProgramPublicGetResponse, error)
	PublicPeriods(ctx context.Context, request *models.TourProgramPublicPeriodsRequest) (*models.TourProgramPublicPeriodsResponse, error)
}
