package usecase

import (
	"context"
	"time"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/user/models"
)

func (u *userUseCase) UpdateStatusValidate(request *models.UserUpdateStatusRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Status, validation.NotNil),
	)
}

// UpdateStatus deactivates (soft delete) or restores a user. Deactivation revokes every session.
func (u *userUseCase) UpdateStatus(ctx context.Context, request *models.UserUpdateStatusRequest) (*models.UserUpdateStatusResponse, error) {
	if err := u.UpdateStatusValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	actorID := request.UserRequestInfo.UserID
	if request.ID == actorID {
		return nil, cerrors.Conflict(errorcode.CannotModifySelf)
	}

	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if err := ensureCanManage(request.UserRequestInfo, target); err != nil {
			return err
		}
		if target.Status == *request.Status {
			return nil
		}

		now := time.Now()
		target.Status = *request.Status
		target.UpdateDate = now
		target.UpdateBy = actorID
		if _, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).Update(ctx, target, "status", "update_date", "update_by"); err != nil {
			return err
		}
		if !target.Status {
			return u.repositoryRegistry.RefreshTokenRepository.WithTx(appDB).RevokeAllForUser(ctx, target.ID, "", actorID, now)
		}
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update user status")
	}
	return nil, nil
}
