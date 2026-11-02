package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	"backend/pkg/modules/tourprogram/models"
)

// FormatTourCode builds the public code: prefix + 6-digit zero-padded number (NIDN000001).
// tour_code_seq has MAXVALUE 999999, so the number never needs a seventh digit.
func FormatTourCode(prefix string, number int64) string {
	return fmt.Sprintf("%s%06d", prefix, number)
}

func (u *tourProgramUseCase) CreateValidate(request *models.TourProgramCreateRequest) error {
	return u.fieldsValidate(&request.TourProgramFields)
}

// Create inserts the tour with a new public code from tour_code_seq.
func (u *tourProgramUseCase) Create(ctx context.Context, request *models.TourProgramCreateRequest) (*models.TourProgramCreateResponse, error) {
	normalize(&request.TourProgramFields)
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var created *domain.TourProgram
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		if err := u.checkReferences(ctx, appDB, &request.TourProgramFields, nil); err != nil {
			return err
		}

		repo := u.repositoryRegistry.TourProgramRepository.WithTx(appDB)
		number, err := repo.NextCodeNumber(ctx)
		if err != nil {
			return err
		}

		actorID := request.UserRequestInfo.UserID
		now := time.Now()
		row := &domain.TourProgram{
			ID:         utils.NewID(),
			Code:       FormatTourCode(u.coreRegistry.Config.TourCode.Prefix, number),
			Status:     true,
			CreateDate: now,
			UpdateDate: now,
			CreateBy:   actorID,
			UpdateBy:   actorID,
		}
		applyFields(row, &request.TourProgramFields)
		if _, err := repo.Create(ctx, row); err != nil {
			return uniqueConflict(err)
		}
		if err := u.replaceChildren(ctx, appDB, row.ID, actorID, &request.TourProgramFields); err != nil {
			return err
		}
		created, err = u.load(ctx, appDB, row.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "create tour program")
	}

	// TODO(revalidator): enqueue a customer-site rebuild once the deploy-hook revalidator exists.
	return &models.TourProgramCreateResponse{TourProgramItem: models.NewTourProgramItem(created)}, nil
}
