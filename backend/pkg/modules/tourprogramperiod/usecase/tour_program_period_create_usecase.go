package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/tourprogramperiod/models"
)

func (u *tourProgramPeriodUseCase) CreateValidate(request *models.TourProgramPeriodCreateRequest) error {
	if err := validation.Validate(request.TourProgramID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"tourProgramId": err}
	}
	return fieldsValidate(&request.TourProgramPeriodFields)
}

// Create adds a period to an active tour.
func (u *tourProgramPeriodUseCase) Create(ctx context.Context, request *models.TourProgramPeriodCreateRequest) (*models.TourProgramPeriodCreateResponse, error) {
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var created *domain.TourProgramPeriod
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		tour, err := u.repositoryRegistry.TourProgramRepository.WithTx(appDB).FindByID(ctx, request.TourProgramID)
		if err != nil {
			return err
		}
		if tour == nil || !tour.Status {
			return cerrors.FieldError("tourProgramId", "tour program not found or deleted")
		}

		actorID := request.UserRequestInfo.UserID
		now := time.Now()
		row := &domain.TourProgramPeriod{
			ID:            utils.NewID(),
			TourProgramID: tour.ID,
			Status:        true,
			CreateDate:    now,
			UpdateDate:    now,
			CreateBy:      actorID,
			UpdateBy:      actorID,
		}
		applyFields(row, &request.TourProgramPeriodFields)
		created, err = u.repositoryRegistry.TourProgramPeriodRepository.WithTx(appDB).Create(ctx, row)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "create tour program period")
	}
	return &models.TourProgramPeriodCreateResponse{TourProgramPeriodItem: models.NewTourProgramPeriodItem(created)}, nil
}
