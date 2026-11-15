package dashboard

import (
	"context"

	"backend/pkg/modules/dashboard/models"
)

type DashboardUseCase interface {
	Summary(ctx context.Context, request *models.DashboardSummaryRequest) (*models.DashboardSummaryResponse, error)
}
