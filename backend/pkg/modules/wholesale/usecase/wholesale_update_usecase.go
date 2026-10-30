package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/wholesale/models"
)

func (u *wholesaleUseCase) UpdateValidate(request *models.WholesaleUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.Code, codeRules()...),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.State, stateRules()...),
		validation.Field(&request.Contacts, contactsRules()...),
	)
}

// Update replaces every field of the wholesale, including its whole contact list.
func (u *wholesaleUseCase) Update(ctx context.Context, request *models.WholesaleUpdateRequest) (*models.WholesaleUpdateResponse, error) {
	normalize(&request.Code, &request.Name, request.Contacts)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.Wholesale
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}

		actorID := request.UserRequestInfo.UserID
		target.RefID = request.RefID
		target.Code = request.Code
		target.Name = request.Name
		target.State = request.State
		target.UpdateDate = time.Now()
		target.UpdateBy = actorID
		target.Contacts = nil // not a column; replaced below
		updated, err = u.repositoryRegistry.WholesaleRepository.WithTx(appDB).Update(ctx, target,
			"ref_id", "code", "name", "state", "update_date", "update_by")
		if err != nil {
			return uniqueConflict(err)
		}
		updated.Contacts, err = u.replaceContacts(ctx, appDB, updated.ID, actorID, request.Contacts)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update wholesale")
	}
	return &models.WholesaleUpdateResponse{WholesaleItem: models.NewWholesaleItem(updated)}, nil
}
