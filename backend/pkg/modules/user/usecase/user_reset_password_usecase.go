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

func (u *userUseCase) ResetPasswordValidate(request *models.UserResetPasswordRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.TemporaryPassword, validation.Required, u.shareRegistry.ServiceRegistry.Password.Rule()),
	)
}

// ResetPassword sets a temporary password chosen by an admin, forces a change at next login and
// signs the user out everywhere. Staff change their own password through PUT /auth/password.
func (u *userUseCase) ResetPassword(ctx context.Context, request *models.UserResetPasswordRequest) (*models.UserResetPasswordResponse, error) {
	if err := u.ResetPasswordValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	actorID := request.UserRequestInfo.UserID
	if request.ID == actorID {
		return nil, cerrors.Conflict(errorcode.CannotModifySelf)
	}

	hash, err := u.shareRegistry.ServiceRegistry.Password.Hash(request.TemporaryPassword)
	if err != nil {
		return nil, cerrors.Wrap(err, "hash password")
	}

	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		if err := ensureCanManage(request.UserRequestInfo, target); err != nil {
			return err
		}

		now := time.Now()
		target.PasswordHash = hash
		target.MustChangePassword = true
		target.UpdateDate = now
		target.UpdateBy = actorID
		if _, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).Update(ctx, target, "password_hash", "must_change_password", "update_date", "update_by"); err != nil {
			return err
		}
		return u.repositoryRegistry.RefreshTokenRepository.WithTx(appDB).RevokeAllForUser(ctx, target.ID, "", actorID, now)
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "reset password")
	}
	return nil, nil
}
