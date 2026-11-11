package usecase

import (
	"context"
	"strings"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

const maxStaffNoteRunes = 2000

func (u *lineBookingUseCase) UpdateValidate(request *models.LineBookingUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.StaffNote, validation.RuneLength(0, maxStaffNoteRunes)),
	)
}

func (u *lineBookingUseCase) Update(ctx context.Context, request *models.LineBookingUpdateRequest) (*models.LineBookingUpdateResponse, error) {
	request.StaffNote = strings.TrimSpace(request.StaffNote)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if target == nil || !visibleTo(target, request.UserRequestInfo) {
			return cerrors.NotFound()
		}

		target.StaffNote = request.StaffNote
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "staff_note", "update_date", "update_by"); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update line booking")
	}
	return &models.LineBookingUpdateResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}
