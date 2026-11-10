package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/auth/models"
)

// Me returns the user loaded by the auth middleware for this request, plus their permission codes.
func (u *authUseCase) Me(ctx context.Context, request *models.AuthMeRequest) (*models.AuthMeResponse, error) {
	info := request.UserRequestInfo
	if info == nil {
		return nil, cerrors.Unauthenticated()
	}
	user, err := u.repositoryRegistry.AppUserRepository.FindByID(ctx, info.UserID)
	if err != nil {
		return nil, cerrors.Wrap(err, "me")
	}
	if user == nil {
		return nil, cerrors.Unauthenticated()
	}
	var line *models.AuthLineAccount
	if user.LineUserID.Valid {
		line = &models.AuthLineAccount{DisplayName: user.LineDisplayName.String, LinkedAt: user.LineLinkedAt.Time}
	}
	return &models.AuthMeResponse{
		ID:                 info.UserID,
		Email:              info.Email,
		Name:               info.Name,
		Roles:              info.Roles,
		Permissions:        u.coreRegistry.Permissions.Codes(info.Roles),
		MustChangePassword: info.MustChangePassword,
		Line:               line,
	}, nil
}
