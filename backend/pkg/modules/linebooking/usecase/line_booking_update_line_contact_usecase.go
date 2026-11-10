package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

func (u *lineBookingUseCase) UpdateLineContactValidate(request *models.LineBookingUpdateLineContactRequest) error {
	fields := validation.Errors{}
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		fields["id"] = err
	}
	if request.LineContactID.Valid {
		if err := validation.Validate(request.LineContactID.String, validation.Required, validation.UUID); err != nil {
			fields["lineContactId"] = err
		}
	}
	return fields.Filter()
}

func (u *lineBookingUseCase) UpdateLineContact(ctx context.Context, request *models.LineBookingUpdateLineContactRequest) (*models.LineBookingUpdateLineContactResponse, error) {
	if request.LineContactID.Valid {
		request.LineContactID = null.StringFrom(strings.TrimSpace(request.LineContactID.String))
	}
	if err := u.UpdateLineContactValidate(request); err != nil {
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

		if request.LineContactID.Valid {
			contact, err := u.repositoryRegistry.LineContactRepository.WithTx(appDB).FindByID(ctx, request.LineContactID.String)
			if err != nil {
				return err
			}
			if contact == nil || !contact.Status {
				return cerrors.FieldError("lineContactId", "no such LINE contact")
			}
		}

		target.LineContactID = request.LineContactID
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "line_contact_id", "update_date", "update_by"); err != nil {
			return err
		}
		// Reload for the contact's name and picture.
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update line booking contact")
	}
	return &models.LineBookingUpdateLineContactResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}
