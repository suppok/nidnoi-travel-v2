package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/airline/models"
)

func (u *airlineUseCase) GetValidate(request *models.AirlineGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *airlineUseCase) Get(ctx context.Context, request *models.AirlineGetRequest) (*models.AirlineGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.AirlineRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get airline")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return &models.AirlineGetResponse{AirlineItem: models.NewAirlineItem(found)}, nil
}
