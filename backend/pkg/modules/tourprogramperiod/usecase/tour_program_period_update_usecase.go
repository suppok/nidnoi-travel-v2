package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/tourprogramperiod/models"
)

func (u *tourProgramPeriodUseCase) UpdateValidate(request *models.TourProgramPeriodUpdateRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return fieldsValidate(&request.TourProgramPeriodFields)
}

// Update replaces every editable field of a period (dates, prices, seats, note).
func (u *tourProgramPeriodUseCase) Update(ctx context.Context, request *models.TourProgramPeriodUpdateRequest) (*models.TourProgramPeriodUpdateResponse, error) {
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.TourProgramPeriod
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.TourProgramPeriodRepository.WithTx(appDB)
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if target == nil {
			return cerrors.NotFound()
		}

		applyFields(target, &request.TourProgramPeriodFields)
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		updated, err = repo.Update(ctx, target, editableColumns...)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update tour program period")
	}
	return &models.TourProgramPeriodUpdateResponse{TourProgramPeriodItem: models.NewTourProgramPeriodItem(updated)}, nil
}
