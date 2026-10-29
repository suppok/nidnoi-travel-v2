package usecase

import (
	"context"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain"
	"backend/pkg/modules/auth/commons"
	"backend/pkg/modules/auth/models"
)

// Refresh exchanges a refresh token for a new access token and rotates the refresh token.
// A rotated token presented again after the grace window means it was copied: every session of
// that user is revoked.
func (u *authUseCase) Refresh(ctx context.Context, request *models.AuthRefreshRequest) (*models.AuthRefreshResponse, error) {
	if request.RefreshToken == "" {
		return nil, cerrors.Unauthenticated()
	}

	var (
		userID      string
		next        *domain.RefreshToken
		nextToken   string
		rejected    bool
		reuseUserID string
	)
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		tokens := u.repositoryRegistry.RefreshTokenRepository.WithTx(appDB)
		now := time.Now()

		current, err := tokens.FindByHashForUpdate(ctx, commons.HashRefreshToken(request.RefreshToken))
		if err != nil {
			return err
		}
		if current == nil {
			rejected = true
			return nil
		}

		if current.RevokedAt.Valid {
			rejected = true
			if now.Sub(current.RevokedAt.Time) > u.coreRegistry.Config.Auth.RefreshReuseGrace {
				reuseUserID = current.UserID
				return tokens.RevokeAllForUser(ctx, current.UserID, "", current.UserID, now)
			}
			return nil
		}
		if !current.ExpiresAt.After(now) {
			rejected = true
			return nil
		}

		user, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).FindByID(ctx, current.UserID)
		if err != nil {
			return err
		}
		if user == nil || !user.Status {
			rejected = true
			return tokens.RevokeAllForUser(ctx, current.UserID, "", current.UserID, now)
		}

		next, nextToken, err = u.createRefreshToken(ctx, appDB, user.ID, current.Persistent, request.ClientIP, request.UserAgent, now)
		if err != nil {
			return err
		}

		current.RevokedAt = null.TimeFrom(now)
		current.ReplacedBy = null.StringFrom(next.ID)
		current.UpdateDate = now
		current.UpdateBy = user.ID
		if _, err := tokens.Update(ctx, current, "revoked_at", "replaced_by", "update_date", "update_by"); err != nil {
			return err
		}
		userID = user.ID
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "refresh")
	}
	if reuseUserID != "" {
		u.coreRegistry.Logger.Warn("refresh token reuse detected; revoked all sessions", zap.String("userId", reuseUserID))
	}
	if rejected {
		return nil, cerrors.Unauthenticated()
	}

	s, err := u.issueSession(userID, next, nextToken)
	if err != nil {
		return nil, cerrors.Wrap(err, "issue session")
	}
	return &models.AuthRefreshResponse{
		AccessToken:      s.accessToken,
		ExpiresIn:        s.expiresIn,
		RefreshToken:     s.refreshToken,
		RefreshExpiresAt: s.refreshExpiresAt,
		Persistent:       s.persistent,
	}, nil
}
