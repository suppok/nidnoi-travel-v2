package usecase

import (
	"context"

	"backend/pkg/constants/role"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/user/models"
)

func (u *userUseCase) SalesList(ctx context.Context, request *models.UserSalesListRequest) (*models.UserSalesListResponse, error) {
	users, err := u.repositoryRegistry.AppUserRepository.ListActiveByRole(ctx, role.Sales)
	if err != nil {
		return nil, cerrors.Wrap(err, "list sales users")
	}
	res := &models.UserSalesListResponse{Items: make([]models.UserSalesItem, len(users))}
	for i, user := range users {
		res.Items[i] = models.UserSalesItem{ID: user.ID, Name: user.Name, Phone: user.Phone}
		if request.UserRequestInfo != nil && user.ID == request.UserRequestInfo.UserID {
			res.DefaultID = user.ID
		}
	}
	return res, nil
}
