package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/booking/models"
)

func (u *bookingUseCase) GetValidate(request *models.BookingGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *bookingUseCase) Get(ctx context.Context, request *models.BookingGetRequest) (*models.BookingGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}
	found, err := u.repositoryRegistry.BookingRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get booking")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return &models.BookingGetResponse{BookingItem: models.NewBookingItem(found)}, nil
}
