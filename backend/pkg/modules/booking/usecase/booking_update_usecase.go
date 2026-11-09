package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/bookingstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/booking/models"
)

const maxStaffNoteRunes = 2000

func (u *bookingUseCase) UpdateValidate(request *models.BookingUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.State, validation.Required, validation.In(stateValues()...)),
		validation.Field(&request.StaffNote, validation.RuneLength(0, maxStaffNoteRunes)),
	)
}

// Update moves a booking through the sales workflow and saves the staff note. The first user to
// take it out of "new" becomes its assignee.
func (u *bookingUseCase) Update(ctx context.Context, request *models.BookingUpdateRequest) (*models.BookingUpdateResponse, error) {
	request.StaffNote = strings.TrimSpace(request.StaffNote)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.Booking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.BookingRepository.WithTx(appDB)
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if target == nil {
			return cerrors.NotFound()
		}

		if !target.AssigneeID.Valid && request.State != bookingstate.New {
			target.AssigneeID = null.StringFrom(request.UserRequestInfo.UserID)
		}
		target.State = request.State
		target.StaffNote = request.StaffNote
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "state", "staff_note", "assignee_id", "update_date", "update_by"); err != nil {
			return err
		}
		// Reload for the assignee's name.
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update booking")
	}
	return &models.BookingUpdateResponse{BookingItem: models.NewBookingItem(updated)}, nil
}
