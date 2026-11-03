package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/airline/models"
)

func (u *airlineUseCase) UpdateStatusValidate(request *models.AirlineUpdateStatusRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Status, validation.NotNil),
	)
}

// UpdateStatus soft-deletes or restores an airline.
func (u *airlineUseCase) UpdateStatus(ctx context.Context, request *models.AirlineUpdateStatusRequest) (*models.AirlineUpdateStatusResponse, error) {
	if err := u.UpdateStatusValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if target.Status == *request.Status {
			return nil
		}

		target.Status = *request.Status
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		_, err = u.repositoryRegistry.AirlineRepository.WithTx(appDB).Update(ctx, target, "status", "update_date", "update_by")
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update airline status")
	}
	return nil, nil
}
