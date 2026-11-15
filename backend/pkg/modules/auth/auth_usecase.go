package auth

import (
	"context"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/auth/models"
)

type AuthUseCase interface {
	Login(ctx context.Context, request *models.AuthLoginRequest) (*models.AuthLoginResponse, error)
	// LineLoginConfig says whether LINE Login is available on the login page.
	LineLoginConfig(ctx context.Context) (*models.AuthLineLoginConfigResponse, error)
	// LineLogin signs in the user who linked the LINE account LINE vouched for.
	LineLogin(ctx context.Context, request *models.AuthLineLoginRequest) (*models.AuthLoginResponse, error)
	Refresh(ctx context.Context, request *models.AuthRefreshRequest) (*models.AuthRefreshResponse, error)
	Logout(ctx context.Context, request *models.AuthLogoutRequest) (*models.AuthLogoutResponse, error)
	Me(ctx context.Context, request *models.AuthMeRequest) (*models.AuthMeResponse, error)
	ChangePassword(ctx context.Context, request *models.AuthChangePasswordRequest) (*models.AuthChangePasswordResponse, error)

	// LineLink links the caller's LINE account with a LINE Login code from the backoffice callback.
	LineLink(ctx context.Context, request *models.AuthLineLinkRequest) (*models.AuthLineLinkResponse, error)
	// LineUnlink removes the caller's LINE account link.
	LineUnlink(ctx context.Context, request *models.AuthLineUnlinkRequest) (*models.AuthLineUnlinkResponse, error)

	// ResolveSession implements middleware.SessionResolver: the current state of an authenticated
	// user, or nil when the user no longer exists or is deactivated.
	ResolveSession(ctx context.Context, userID string) (*coremodels.UserRequestInfo, error)
}
