package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/airline/models"
)

func (u *airlineUseCase) UpdateValidate(request *models.AirlineUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.Code, codeRules()...),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.Logo, u.logoRules()...),
	)
}

// Update replaces every editable field of the airline.
func (u *airlineUseCase) Update(ctx context.Context, request *models.AirlineUpdateRequest) (*models.AirlineUpdateResponse, error) {
	normalize(&request.Code, &request.Name, &request.Logo)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.Airline
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}

		target.RefID = request.RefID
		target.Code = request.Code
		target.Name = request.Name
		target.Logo = request.Logo
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		updated, err = u.repositoryRegistry.AirlineRepository.WithTx(appDB).Update(ctx, target,
			"ref_id", "code", "name", "logo", "update_date", "update_by")
		if err != nil {
			return uniqueConflict(err)
		}
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update airline")
	}
	return &models.AirlineUpdateResponse{AirlineItem: models.NewAirlineItem(updated)}, nil
}
