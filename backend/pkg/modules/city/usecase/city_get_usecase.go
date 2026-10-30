package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/city/models"
)

func (u *cityUseCase) GetValidate(request *models.CityGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *cityUseCase) Get(ctx context.Context, request *models.CityGetRequest) (*models.CityGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.CityRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get city")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return &models.CityGetResponse{CityItem: models.NewCityItem(found)}, nil
}
