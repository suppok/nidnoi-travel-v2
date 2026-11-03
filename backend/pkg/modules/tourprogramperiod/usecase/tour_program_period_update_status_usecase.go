package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/tourprogramperiod/models"
)

func (u *tourProgramPeriodUseCase) UpdateStatusValidate(request *models.TourProgramPeriodUpdateStatusRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Status, validation.NotNil),
	)
}

// UpdateStatus soft-deletes or restores a period.
func (u *tourProgramPeriodUseCase) UpdateStatus(ctx context.Context, request *models.TourProgramPeriodUpdateStatusRequest) (*models.TourProgramPeriodUpdateStatusResponse, error) {
	if err := u.UpdateStatusValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var rebuild []string
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.TourProgramPeriodRepository.WithTx(appDB)
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if target == nil {
			return cerrors.NotFound()
		}
		if target.Status == *request.Status {
			return nil
		}
		tour, err := u.loadTour(ctx, appDB, target.TourProgramID)
		if err != nil {
			return err
		}
		w, err := u.watch(ctx, appDB, tour)
		if err != nil {
			return err
		}

		target.Status = *request.Status
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "status", "update_date", "update_by"); err != nil {
			return err
		}
		rebuild, err = u.changedPaths(ctx, appDB, w)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update tour program period status")
	}
	u.enqueue(ctx, rebuild)
	return nil, nil
}
