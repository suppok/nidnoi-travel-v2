package usecase

import (
	"context"
	"slices"

	"backend/pkg/constants/role"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/linebooking/models"
)

func (u *lineBookingUseCase) SalesUserList(ctx context.Context, request *models.LineBookingSalesUserListRequest) (*models.LineBookingSalesUserListResponse, error) {
	users, err := u.repositoryRegistry.AppUserRepository.ListActiveByRole(ctx, role.Sales)
	if err != nil {
		return nil, cerrors.Wrap(err, "list sales users")
	}
	res := &models.LineBookingSalesUserListResponse{Items: make([]models.LineBookingUser, len(users))}
	for i, user := range users {
		res.Items[i] = models.LineBookingUser{ID: user.ID, Name: user.Name}
		if request.UserRequestInfo != nil && user.ID == request.UserRequestInfo.UserID {
			res.DefaultID = user.ID
		}
	}
	return res, nil
}

// checkSalesUser reports whether id is an active user with the sales role.
func (u *lineBookingUseCase) checkSalesUser(ctx context.Context, appDB query.AppIDB, id string) error {
	user, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil || !user.Status || !slices.Contains(user.Roles, role.Sales) {
		return cerrors.FieldError("salesUserId", "must be an active sales user")
	}
	return nil
}
