package usecase

import (
	"context"

	"backend/pkg/constants/role"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/user/models"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func (u *userUseCase) ListValidate(request *models.UserListRequest) error {
	known := make([]any, len(role.All))
	for i, r := range role.All {
		known[i] = r
	}
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.Role, validation.In(known...)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *userUseCase) List(ctx context.Context, request *models.UserListRequest) (*models.UserListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	users, total, err := u.repositoryRegistry.AppUserRepository.List(ctx, repomodels.AppUserListCriteria{
		Keyword:  request.Keyword,
		Role:     request.Role,
		Status:   request.Status,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list users")
	}

	items := make([]models.UserItem, len(users))
	for i := range users {
		items[i] = models.NewUserItem(&users[i])
	}
	return &models.UserListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
