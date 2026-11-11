package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/bookingsource/models"
)

func (u *bookingSourceUseCase) GetValidate(request *models.BookingSourceGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *bookingSourceUseCase) Get(ctx context.Context, request *models.BookingSourceGetRequest) (*models.BookingSourceGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.BookingSourceRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get booking source")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return &models.BookingSourceGetResponse{BookingSourceItem: models.NewBookingSourceItem(found)}, nil
}
