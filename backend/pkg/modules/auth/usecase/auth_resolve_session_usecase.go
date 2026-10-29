package usecase

import (
	"context"

	coremodels "backend/pkg/core/api/models"
	cerrors "backend/pkg/core/errors"
)

// ResolveSession loads the user on every authenticated request (one primary-key lookup), so
// deactivation, role changes and a completed password change apply immediately.
func (u *authUseCase) ResolveSession(ctx context.Context, userID string) (*coremodels.UserRequestInfo, error) {
	user, err := u.repositoryRegistry.AppUserRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, cerrors.Wrap(err, "resolve session")
	}
	if user == nil || !user.Status {
		return nil, nil
	}
	return &coremodels.UserRequestInfo{
		UserID:             user.ID,
		Email:              user.Email,
		Name:               user.Name,
		Roles:              user.Roles,
		MustChangePassword: user.MustChangePassword,
	}, nil
}
