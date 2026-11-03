package usecase

import (
	"context"
	"time"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/shared/services/revalidation"
	"backend/pkg/modules/tourprogram/models"
)

func (u *tourProgramUseCase) UpdateStatusValidate(request *models.TourProgramUpdateStatusRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Status, validation.NotNil),
	)
}

// UpdateStatus soft-deletes or restores a tour. Its periods, links and gallery are kept.
func (u *tourProgramUseCase) UpdateStatus(ctx context.Context, request *models.TourProgramUpdateStatusRequest) (*models.TourProgramUpdateStatusResponse, error) {
	if err := u.UpdateStatusValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var affected []string // set when the change shows or hides a public tour
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.load(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if target.Status == *request.Status {
			return nil
		}

		if tourprogramstate.IsPublic(target.State) {
			affected = revalidation.PathsForTour(target)
		}
		stripRelations(target)
		target.Status = *request.Status
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		_, err = u.repositoryRegistry.TourProgramRepository.WithTx(appDB).Update(ctx, target, "status", "update_date", "update_by")
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update tour program status")
	}

	if affected != nil {
		u.shareRegistry.ServiceRegistry.Revalidator.Enqueue(ctx, affected)
	}
	return nil, nil
}
