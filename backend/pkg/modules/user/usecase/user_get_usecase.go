package usecase

import (
	"context"
	"slices"

	"backend/pkg/constants/role"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/user/models"
)

func (u *userUseCase) GetValidate(request *models.UserGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *userUseCase) Get(ctx context.Context, request *models.UserGetRequest) (*models.UserGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.AppUserRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get user")
	}
	// The super admin is visible only to themselves.
	if found == nil || (slices.Contains(found.Roles, role.SuperAdmin) && !isSuperAdmin(request.UserRequestInfo)) {
		return nil, cerrors.NotFound()
	}
	return &models.UserGetResponse{UserItem: models.NewUserItem(found)}, nil
}
