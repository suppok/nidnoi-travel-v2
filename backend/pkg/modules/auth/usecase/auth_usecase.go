package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	"backend/pkg/modules/auth"
	"backend/pkg/modules/auth/commons"
	"backend/pkg/registry"
)

type authUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewAuthUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) auth.AuthUseCase {
	return &authUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// session is what login and refresh hand back to their responses.
type session struct {
	accessToken      string
	expiresIn        int
	refreshToken     string
	refreshExpiresAt time.Time
	persistent       bool
}

// createRefreshToken inserts a new refresh token row inside the caller's transaction.
func (u *authUseCase) createRefreshToken(
	ctx context.Context, appDB query.AppIDB, userID string, persistent bool, clientIP, userAgent string, now time.Time,
) (*domain.RefreshToken, string, error) {
	token, err := commons.NewRefreshToken()
	if err != nil {
		return nil, "", err
	}

	ttl := u.coreRegistry.Config.Auth.SessionRefreshTokenTTL
	if persistent {
		ttl = u.coreRegistry.Config.Auth.RefreshTokenTTL
	}

	row, err := u.repositoryRegistry.RefreshTokenRepository.WithTx(appDB).Create(ctx, &domain.RefreshToken{
		ID:         utils.NewID(),
		UserID:     userID,
		TokenHash:  commons.HashRefreshToken(token),
		Persistent: persistent,
		ExpiresAt:  now.Add(ttl),
		UserAgent:  truncate(userAgent, 512),
		ClientIP:   truncate(clientIP, 64),
		Status:     true,
		CreateDate: now,
		UpdateDate: now,
		CreateBy:   userID,
		UpdateBy:   userID,
	})
	if err != nil {
		return nil, "", err
	}
	return row, token, nil
}

func (u *authUseCase) issueSession(userID string, refresh *domain.RefreshToken, refreshToken string) (*session, error) {
	accessToken, ttl, err := u.coreRegistry.JWT.Issue(userID)
	if err != nil {
		return nil, err
	}
	return &session{
		accessToken:      accessToken,
		expiresIn:        int(ttl.Seconds()),
		refreshToken:     refreshToken,
		refreshExpiresAt: refresh.ExpiresAt,
		persistent:       refresh.Persistent,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
