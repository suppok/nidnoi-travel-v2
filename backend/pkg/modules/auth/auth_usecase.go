package auth

import (
	"context"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/modules/auth/models"
)

type AuthUseCase interface {
	Login(ctx context.Context, request *models.AuthLoginRequest) (*models.AuthLoginResponse, error)
	Refresh(ctx context.Context, request *models.AuthRefreshRequest) (*models.AuthRefreshResponse, error)
	Logout(ctx context.Context, request *models.AuthLogoutRequest) (*models.AuthLogoutResponse, error)
	Me(ctx context.Context, request *models.AuthMeRequest) (*models.AuthMeResponse, error)
	ChangePassword(ctx context.Context, request *models.AuthChangePasswordRequest) (*models.AuthChangePasswordResponse, error)

	// ResolveSession implements middleware.SessionResolver: the current state of an authenticated
	// user, or nil when the user no longer exists or is deactivated.
	ResolveSession(ctx context.Context, userID string) (*coremodels.UserRequestInfo, error)
}
