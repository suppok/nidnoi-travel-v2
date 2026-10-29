package user

import (
	"context"

	"backend/pkg/modules/user/models"
)

type UserUseCase interface {
	List(ctx context.Context, request *models.UserListRequest) (*models.UserListResponse, error)
	Get(ctx context.Context, request *models.UserGetRequest) (*models.UserGetResponse, error)
	Create(ctx context.Context, request *models.UserCreateRequest) (*models.UserCreateResponse, error)
	Update(ctx context.Context, request *models.UserUpdateRequest) (*models.UserUpdateResponse, error)
	UpdateStatus(ctx context.Context, request *models.UserUpdateStatusRequest) (*models.UserUpdateStatusResponse, error)
	ResetPassword(ctx context.Context, request *models.UserResetPasswordRequest) (*models.UserResetPasswordResponse, error)
}
