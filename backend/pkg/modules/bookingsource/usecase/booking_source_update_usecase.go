package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/bookingsource/models"
)

func (u *bookingSourceUseCase) UpdateValidate(request *models.BookingSourceUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.ShortName, shortNameRules()...),
		validation.Field(&request.SortOrder, sortOrderRules()...),
	)
}

// Update replaces every editable field of the booking source.
func (u *bookingSourceUseCase) Update(ctx context.Context, request *models.BookingSourceUpdateRequest) (*models.BookingSourceUpdateResponse, error) {
	normalize(&request.Name, &request.ShortName)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.BookingSource
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}

		target.Name = request.Name
		target.ShortName = request.ShortName
		target.SortOrder = request.SortOrder
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		updated, err = u.repositoryRegistry.BookingSourceRepository.WithTx(appDB).Update(ctx, target,
			"name", "short_name", "sort_order", "update_date", "update_by")
		if err != nil {
			return uniqueConflict(err)
		}
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update booking source")
	}
	return &models.BookingSourceUpdateResponse{BookingSourceItem: models.NewBookingSourceItem(updated)}, nil
}
