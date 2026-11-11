package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/linebooking/models"
)

// FinanceUserList returns the finance users a booking can be sent to on confirm, with the sales
// owner's default, for the confirm dialog.
func (u *lineBookingUseCase) FinanceUserList(ctx context.Context, request *models.LineBookingFinanceUserListRequest) (*models.LineBookingFinanceUserListResponse, error) {
	if err := validation.ValidateStruct(request, validation.Field(&request.ID, validation.Required, validation.UUID)); err != nil {
		return nil, cerrors.Validation(err)
	}
	b, err := u.repositoryRegistry.LineBookingRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking")
	}
	if b == nil || !visibleTo(b, request.UserRequestInfo) {
		return nil, cerrors.NotFound()
	}
	items, err := u.financeOptions(ctx, u.coreRegistry.AppDB, b)
	if err != nil {
		return nil, cerrors.Wrap(err, "list finance users")
	}
	res := &models.LineBookingFinanceUserListResponse{Items: items}
	if b.SalesUserID.Valid {
		links, err := u.repositoryRegistry.SalesFinanceRepository.ListBySales(ctx, b.SalesUserID.String)
		if err != nil {
			return nil, cerrors.Wrap(err, "list finance users")
		}
		for _, l := range links {
			if l.IsDefault && usableFinance(l.FinanceUser) {
				res.DefaultID = l.FinanceUserID
			}
		}
	}
	return res, nil
}
