package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/city/models"
	"backend/pkg/modules/shared/services/revalidation"
)

func (u *cityUseCase) UpdateStatusValidate(request *models.CityUpdateStatusRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Status, validation.NotNil),
	)
}

// UpdateStatus soft-deletes or restores a city. Restoring requires its country to be active.
// TAXONOMY_IN_USE (active tours referencing it) is checked once tour programs exist.
func (u *cityUseCase) UpdateStatus(ctx context.Context, request *models.CityUpdateStatusRequest) (*models.CityUpdateStatusResponse, error) {
	if err := u.UpdateStatusValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var rebuild []string // set when the status actually changes
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if target.Status == *request.Status {
			return nil
		}
		if *request.Status {
			if _, err := u.activeCountry(ctx, appDB, target.CountryID); err != nil {
				return err
			}
		}

		if target.Country != nil {
			rebuild = revalidation.PathsForCity(target.Country.Slug, target.Slug)
		}
		target.Status = *request.Status
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		target.Country = nil
		_, err = u.repositoryRegistry.CityRepository.WithTx(appDB).Update(ctx, target, "status", "update_date", "update_by")
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update city status")
	}
	if rebuild != nil {
		u.enqueueIfShown(ctx, request.ID, rebuild)
	}
	return nil, nil
}
