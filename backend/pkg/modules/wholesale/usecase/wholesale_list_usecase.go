package usecase

import (
	"context"

	"backend/pkg/constants/wholesalestate"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/wholesale/models"
)

func (u *wholesaleUseCase) ListValidate(request *models.WholesaleListRequest) error {
	known := make([]any, len(wholesalestate.All))
	for i, s := range wholesalestate.All {
		known[i] = s
	}
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.State, validation.In(known...)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *wholesaleUseCase) List(ctx context.Context, request *models.WholesaleListRequest) (*models.WholesaleListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	wholesales, total, err := u.repositoryRegistry.WholesaleRepository.List(ctx, repomodels.WholesaleListCriteria{
		Keyword:  request.Keyword,
		State:    request.State,
		Status:   request.Status,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list wholesales")
	}

	items := make([]models.WholesaleItem, len(wholesales))
	for i := range wholesales {
		items[i] = models.NewWholesaleItem(&wholesales[i])
	}
	return &models.WholesaleListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
