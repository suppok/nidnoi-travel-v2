package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/tourprogram/models"
)

func (u *tourProgramUseCase) UpdateValidate(request *models.TourProgramUpdateRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return u.fieldsValidate(&request.TourProgramFields)
}

// Update replaces every editable field, including the country, city and gallery lists.
func (u *tourProgramUseCase) Update(ctx context.Context, request *models.TourProgramUpdateRequest) (*models.TourProgramUpdateResponse, error) {
	normalize(&request.TourProgramFields)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.TourProgram
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.load(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if err := u.checkReferences(ctx, appDB, &request.TourProgramFields, target); err != nil {
			return err
		}

		actorID := request.UserRequestInfo.UserID
		stripRelations(target)
		applyFields(target, &request.TourProgramFields)
		target.UpdateDate = time.Now()
		target.UpdateBy = actorID
		if _, err := u.repositoryRegistry.TourProgramRepository.WithTx(appDB).Update(ctx, target, editableColumns...); err != nil {
			return uniqueConflict(err)
		}
		if err := u.replaceChildren(ctx, appDB, target.ID, actorID, &request.TourProgramFields); err != nil {
			return err
		}
		updated, err = u.load(ctx, appDB, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update tour program")
	}

	// TODO(revalidator): enqueue a customer-site rebuild once the deploy-hook revalidator exists.
	return &models.TourProgramUpdateResponse{TourProgramItem: models.NewTourProgramItem(updated)}, nil
}
