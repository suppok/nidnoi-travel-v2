package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/country/models"
)

func (u *countryUseCase) GetValidate(request *models.CountryGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *countryUseCase) Get(ctx context.Context, request *models.CountryGetRequest) (*models.CountryGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.CountryRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get country")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return &models.CountryGetResponse{CountryItem: models.NewCountryItem(found)}, nil
}
