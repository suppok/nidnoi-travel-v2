package salesfinance

import (
	"context"

	"backend/pkg/modules/salesfinance/models"
)

// SalesFinanceUseCase manages the caller's own finance users: the finance users a sales user can
// send confirmed LINE bookings to, and their default.
type SalesFinanceUseCase interface {
	Get(ctx context.Context, request *models.SalesFinanceGetRequest) (*models.SalesFinanceGetResponse, error)
	Update(ctx context.Context, request *models.SalesFinanceUpdateRequest) (*models.SalesFinanceUpdateResponse, error)
}
