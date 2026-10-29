package usecase

import (
	"context"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/auth/commons"
	"backend/pkg/modules/auth/models"
)

// Logout revokes the presented refresh token. It succeeds even without a valid token, so the
// client can always clear its state.
func (u *authUseCase) Logout(ctx context.Context, request *models.AuthLogoutRequest) (*models.AuthLogoutResponse, error) {
	if request.RefreshToken == "" {
		return nil, nil
	}

	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		tokens := u.repositoryRegistry.RefreshTokenRepository.WithTx(appDB)
		current, err := tokens.FindByHashForUpdate(ctx, commons.HashRefreshToken(request.RefreshToken))
		if err != nil || current == nil || current.RevokedAt.Valid {
			return err
		}
		now := time.Now()
		current.RevokedAt = null.TimeFrom(now)
		current.UpdateDate = now
		current.UpdateBy = current.UserID
		_, err = tokens.Update(ctx, current, "revoked_at", "update_date", "update_by")
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "logout")
	}
	return nil, nil
}
