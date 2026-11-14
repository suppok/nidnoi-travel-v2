package usecase

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/auth/models"
)

func (u *authUseCase) LoginValidate(request *models.AuthLoginRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Email, validation.Required, validation.Email),
		validation.Field(&request.Password, validation.Required),
	)
}

func (u *authUseCase) Login(ctx context.Context, request *models.AuthLoginRequest) (*models.AuthLoginResponse, error) {
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if err := u.LoginValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	user, err := u.repositoryRegistry.AppUserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		return nil, cerrors.Wrap(err, "find user")
	}

	// One error for unknown email, deactivated user and wrong password; bcrypt runs in every case.
	passwords := u.shareRegistry.ServiceRegistry.Password
	if user == nil || !user.Status {
		passwords.Matches("", request.Password)
		return nil, cerrors.New(http.StatusUnauthorized, errorcode.InvalidCredentials)
	}
	if !passwords.Matches(user.PasswordHash, request.Password) {
		return nil, cerrors.New(http.StatusUnauthorized, errorcode.InvalidCredentials)
	}

	return u.startSession(ctx, user, request.RememberMe, request.ClientIP, request.UserAgent)
}

// startSession signs user in (password or LINE login): records the login, creates the refresh
// token and issues the access token.
func (u *authUseCase) startSession(ctx context.Context, user *domain.AppUser, rememberMe bool, clientIP, userAgent string) (*models.AuthLoginResponse, error) {
	var refresh *domain.RefreshToken
	var refreshToken string
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		now := time.Now()

		user.LastLoginAt = null.TimeFrom(now)
		if _, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).Update(ctx, user, "last_login_at"); err != nil {
			return err
		}

		var err error
		refresh, refreshToken, err = u.createRefreshToken(ctx, appDB, user.ID, rememberMe, clientIP, userAgent, now)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "login")
	}

	s, err := u.issueSession(user.ID, refresh, refreshToken)
	if err != nil {
		return nil, cerrors.Wrap(err, "issue session")
	}
	return &models.AuthLoginResponse{
		AccessToken:      s.accessToken,
		ExpiresIn:        s.expiresIn,
		RefreshToken:     s.refreshToken,
		RefreshExpiresAt: s.refreshExpiresAt,
		Persistent:       s.persistent,
	}, nil
}
