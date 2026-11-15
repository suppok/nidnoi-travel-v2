package usecase

import (
	"context"
	"slices"

	"backend/pkg/constants/role"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
)

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
