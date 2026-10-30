package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/country/models"
)

func (u *countryUseCase) ListValidate(request *models.CountryListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *countryUseCase) List(ctx context.Context, request *models.CountryListRequest) (*models.CountryListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	countries, total, err := u.repositoryRegistry.CountryRepository.List(ctx, repomodels.CountryListCriteria{
		Keyword:  request.Keyword,
		Status:   request.Status,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list countries")
	}

	items := make([]models.CountryItem, len(countries))
	for i := range countries {
		items[i] = models.NewCountryItem(&countries[i])
	}
	return &models.CountryListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
