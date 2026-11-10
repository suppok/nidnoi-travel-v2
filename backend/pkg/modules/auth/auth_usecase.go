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

	// CreateLineLink starts linking the caller's LINE account: a one-time LIFF URL for a QR code.
	CreateLineLink(ctx context.Context, request *models.AuthLineLinkCreateRequest) (*models.AuthLineLinkCreateResponse, error)
	// LineUnlink removes the caller's LINE account link.
	LineUnlink(ctx context.Context, request *models.AuthLineUnlinkRequest) (*models.AuthLineUnlinkResponse, error)
	// LIFFLineLink finishes linking from the LIFF app with the code and the LIFF access token.
	LIFFLineLink(ctx context.Context, request *models.LIFFLineLinkRequest) (*models.LIFFLineLinkResponse, error)
	// LIFFMe returns the staff user linked to the LIFF caller.
	LIFFMe(ctx context.Context, request *models.LIFFMeRequest) (*models.LIFFMeResponse, error)

	// ResolveLIFFSession implements middleware.LIFFSessionResolver.
	ResolveLIFFSession(ctx context.Context, accessToken string) (*coremodels.UserRequestInfo, error)

	// ResolveSession implements middleware.SessionResolver: the current state of an authenticated
	// user, or nil when the user no longer exists or is deactivated.
	ResolveSession(ctx context.Context, userID string) (*coremodels.UserRequestInfo, error)
}
