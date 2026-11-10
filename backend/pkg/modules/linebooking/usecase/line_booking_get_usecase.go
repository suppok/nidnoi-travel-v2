package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/linebooking/models"
)

func (u *lineBookingUseCase) GetValidate(request *models.LineBookingGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *lineBookingUseCase) Get(ctx context.Context, request *models.LineBookingGetRequest) (*models.LineBookingGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}
	found, err := u.repositoryRegistry.LineBookingRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking")
	}
	if found == nil || !visibleTo(found, request.UserRequestInfo) {
		return nil, cerrors.NotFound()
	}
	return &models.LineBookingGetResponse{LineBookingItem: models.NewLineBookingItem(found)}, nil
}
