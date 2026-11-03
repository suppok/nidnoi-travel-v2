package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/airline/models"
)

func (u *airlineUseCase) ListValidate(request *models.AirlineListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *airlineUseCase) List(ctx context.Context, request *models.AirlineListRequest) (*models.AirlineListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	airlines, total, err := u.repositoryRegistry.AirlineRepository.List(ctx, repomodels.AirlineListCriteria{
		Keyword:  request.Keyword,
		Status:   request.Status,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list airlines")
	}

	items := make([]models.AirlineItem, len(airlines))
	for i := range airlines {
		items[i] = models.NewAirlineItem(&airlines[i])
	}
	return &models.AirlineListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
