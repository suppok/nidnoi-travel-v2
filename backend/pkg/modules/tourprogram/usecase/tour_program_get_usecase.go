package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/tourprogram/models"
)

func (u *tourProgramUseCase) GetValidate(request *models.TourProgramGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *tourProgramUseCase) Get(ctx context.Context, request *models.TourProgramGetRequest) (*models.TourProgramGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.load(ctx, u.coreRegistry.AppDB, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get tour program")
	}
	return &models.TourProgramGetResponse{TourProgramItem: models.NewTourProgramItem(found)}, nil
}
