package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/salesfinance/models"
)

func (u *salesFinanceUseCase) Get(ctx context.Context, request *models.SalesFinanceGetRequest) (*models.SalesFinanceGetResponse, error) {
	s, err := u.settings(ctx, request.UserRequestInfo.UserID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get sales finance")
	}
	return &models.SalesFinanceGetResponse{SalesFinanceSettings: *s}, nil
}
