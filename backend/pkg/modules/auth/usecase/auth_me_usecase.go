package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/auth/models"
)

// Me returns the user loaded by the auth middleware for this request, plus their permission codes.
func (u *authUseCase) Me(_ context.Context, request *models.AuthMeRequest) (*models.AuthMeResponse, error) {
	info := request.UserRequestInfo
	if info == nil {
		return nil, cerrors.Unauthenticated()
	}
	return &models.AuthMeResponse{
		ID:                 info.UserID,
		Email:              info.Email,
		Name:               info.Name,
		Roles:              info.Roles,
		Permissions:        u.coreRegistry.Permissions.Codes(info.Roles),
		MustChangePassword: info.MustChangePassword,
	}, nil
}
