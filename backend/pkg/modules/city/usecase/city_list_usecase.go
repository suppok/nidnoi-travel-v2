package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/city/models"
)

func (u *cityUseCase) ListValidate(request *models.CityListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.CountryID, validation.UUID),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *cityUseCase) List(ctx context.Context, request *models.CityListRequest) (*models.CityListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	cities, total, err := u.repositoryRegistry.CityRepository.List(ctx, repomodels.CityListCriteria{
		Keyword:   request.Keyword,
		CountryID: request.CountryID,
		Status:    request.Status,
		Page:      request.Page,
		PageSize:  request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list cities")
	}

	items := make([]models.CityItem, len(cities))
	for i := range cities {
		items[i] = models.NewCityItem(&cities[i])
	}
	return &models.CityListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
