package usecase

import (
	"context"
	"net/http"
	"time"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/auth/commons"
	"backend/pkg/modules/auth/models"
)

func (u *authUseCase) ChangePasswordValidate(request *models.AuthChangePasswordRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.CurrentPassword, validation.Required),
		validation.Field(&request.NewPassword, validation.Required, u.shareRegistry.ServiceRegistry.Password.Rule()),
	)
}

// ChangePassword replaces the caller's password, clears mustChangePassword and signs out every
// other session (the current one, identified by its refresh cookie, stays).
func (u *authUseCase) ChangePassword(ctx context.Context, request *models.AuthChangePasswordRequest) (*models.AuthChangePasswordResponse, error) {
	if request.UserRequestInfo == nil {
		return nil, cerrors.Unauthenticated()
	}
	if err := u.ChangePasswordValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.NewPassword == request.CurrentPassword {
		return nil, cerrors.FieldError("newPassword", "must differ from the current password")
	}

	passwords := u.shareRegistry.ServiceRegistry.Password
	userID := request.UserRequestInfo.UserID

	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		users := u.repositoryRegistry.AppUserRepository.WithTx(appDB)
		tokens := u.repositoryRegistry.RefreshTokenRepository.WithTx(appDB)

		user, err := users.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil || !user.Status {
			return cerrors.Unauthenticated()
		}
		if !passwords.Matches(user.PasswordHash, request.CurrentPassword) {
			return cerrors.New(http.StatusBadRequest, errorcode.InvalidPassword)
		}

		hash, err := passwords.Hash(request.NewPassword)
		if err != nil {
			return err
		}
		now := time.Now()
		user.PasswordHash = hash
		user.MustChangePassword = false
		user.UpdateDate = now
		user.UpdateBy = user.ID
		if _, err := users.Update(ctx, user, "password_hash", "must_change_password", "update_date", "update_by"); err != nil {
			return err
		}

		keepID := ""
		if request.RefreshToken != "" {
			current, err := tokens.FindByHashForUpdate(ctx, commons.HashRefreshToken(request.RefreshToken))
			if err != nil {
				return err
			}
			if current != nil && current.UserID == user.ID {
				keepID = current.ID
			}
		}
		return tokens.RevokeAllForUser(ctx, user.ID, keepID, user.ID, now)
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "change password")
	}
	return nil, nil
}
