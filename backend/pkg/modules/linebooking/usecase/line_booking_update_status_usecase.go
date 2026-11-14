package usecase

import (
	"context"
	"time"

	"backend/pkg/constants/linebookingevent"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

func (u *lineBookingUseCase) UpdateStatusValidate(request *models.LineBookingUpdateStatusRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Status, validation.NotNil),
	)
}

func (u *lineBookingUseCase) UpdateStatus(ctx context.Context, request *models.LineBookingUpdateStatusRequest) (*models.LineBookingUpdateStatusResponse, error) {
	if err := u.UpdateStatusValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if target == nil || !visibleTo(target, request.UserRequestInfo) {
			return cerrors.NotFound()
		}
		if target.Status == *request.Status {
			return nil
		}

		target.Status = *request.Status
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "status", "update_date", "update_by"); err != nil {
			return err
		}
		action := linebookingevent.Deleted
		if target.Status {
			action = linebookingevent.Restored
		}
		return u.logEvent(ctx, appDB, target.ID, action, request.UserRequestInfo.UserID, domain.LineBookingEventDetail{})
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update line booking status")
	}
	return nil, nil
}
