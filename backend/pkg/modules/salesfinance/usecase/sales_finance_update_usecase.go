package usecase

import (
	"context"
	"slices"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/salesfinance/models"
)

func (u *salesFinanceUseCase) UpdateValidate(request *models.SalesFinanceUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.FinanceUserIDs, validation.Length(0, maxFinanceUsers), validation.Each(validation.Required, validation.UUID)),
		// In skips "" (no default).
		validation.Field(&request.DefaultFinanceUserID, validation.In(toAny(request.FinanceUserIDs)...).Error("must be one of financeUserIds")),
	)
}

func toAny(ids []string) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func (u *salesFinanceUseCase) Update(ctx context.Context, request *models.SalesFinanceUpdateRequest) (*models.SalesFinanceUpdateResponse, error) {
	if request.FinanceUserIDs == nil {
		request.FinanceUserIDs = []string{}
	}
	slices.Sort(request.FinanceUserIDs)
	request.FinanceUserIDs = slices.Compact(request.FinanceUserIDs)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.DefaultFinanceUserID == "" && len(request.FinanceUserIDs) == 1 {
		request.DefaultFinanceUserID = request.FinanceUserIDs[0]
	}
	me := request.UserRequestInfo.UserID

	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		users := u.repositoryRegistry.AppUserRepository.WithTx(appDB)
		now := time.Now()
		rows := make([]domain.SalesFinance, 0, len(request.FinanceUserIDs))
		for _, id := range request.FinanceUserIDs {
			user, err := users.FindByID(ctx, id)
			if err != nil {
				return err
			}
			if id == me || !usable(user) {
				return cerrors.FieldError("financeUserIds", "must be active finance users other than yourself")
			}
			rows = append(rows, domain.SalesFinance{
				ID: utils.NewID(), SalesUserID: me, FinanceUserID: id, IsDefault: id == request.DefaultFinanceUserID,
				CreateDate: now, UpdateDate: now, CreateBy: me, UpdateBy: me,
			})
		}
		return u.repositoryRegistry.SalesFinanceRepository.WithTx(appDB).ReplaceForSales(ctx, me, rows)
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update sales finance")
	}
	s, err := u.settings(ctx, me)
	if err != nil {
		return nil, cerrors.Wrap(err, "get sales finance")
	}
	return &models.SalesFinanceUpdateResponse{SalesFinanceSettings: *s}, nil
}
