package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/shared/services/revalidation"
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
	var wasVisible bool
	var pathsBefore []string // the tour may leave countries/cities: rebuild those landings too
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.load(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if err := u.checkReferences(ctx, appDB, &request.TourProgramFields, target); err != nil {
			return err
		}

		wasVisible, pathsBefore = revalidation.TourVisible(target), revalidation.PathsForTour(target)
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

	if wasVisible || revalidation.TourVisible(updated) {
		u.shareRegistry.ServiceRegistry.Revalidator.Enqueue(ctx, append(pathsBefore, revalidation.PathsForTour(updated)...))
	}
	return &models.TourProgramUpdateResponse{TourProgramItem: models.NewTourProgramItem(updated)}, nil
}
