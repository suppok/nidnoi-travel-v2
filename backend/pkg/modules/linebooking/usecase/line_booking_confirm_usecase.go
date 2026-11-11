package usecase

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/linebookingstate"
	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
	"backend/pkg/repository"
)

func (u *lineBookingUseCase) ConfirmValidate(request *models.LineBookingConfirmRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.FinanceUserID, validation.Required, validation.UUID),
	)
}

// Confirm is the sales owner's "ยืนยัน": the payment is checked, so the booking becomes deposit paid
// or paid in full (by its payment plan) and goes to a finance user to issue the receipt.
func (u *lineBookingUseCase) Confirm(ctx context.Context, request *models.LineBookingConfirmRequest) (*models.LineBookingConfirmResponse, error) {
	if err := u.ConfirmValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		target, err := u.findNewForOwner(ctx, repo, request.ID, request.UserRequestInfo)
		if err != nil {
			return err
		}

		options, err := u.financeOptions(ctx, appDB, target)
		if err != nil {
			return err
		}
		if !slices.ContainsFunc(options, func(o models.LineBookingUser) bool { return o.ID == request.FinanceUserID }) {
			return cerrors.FieldError("financeUserId", "must be one of the sales owner's finance users")
		}

		now := time.Now()
		target.State = linebookingstate.PaidInFull
		if target.IsDeposit {
			target.State = linebookingstate.DepositPaid
		}
		target.FinanceUserID = null.StringFrom(request.FinanceUserID)
		target.ConfirmedAt = null.TimeFrom(now)
		target.UpdateDate = now
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "state", "finance_user_id", "confirmed_at", "update_date", "update_by"); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "confirm line booking")
	}
	return &models.LineBookingConfirmResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}

func (u *lineBookingUseCase) CancelValidate(request *models.LineBookingCancelRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
	)
}

// Cancel is the sales owner's "ยกเลิก" on a new booking.
func (u *lineBookingUseCase) Cancel(ctx context.Context, request *models.LineBookingCancelRequest) (*models.LineBookingCancelResponse, error) {
	if err := u.CancelValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		target, err := u.findNewForOwner(ctx, repo, request.ID, request.UserRequestInfo)
		if err != nil {
			return err
		}
		target.State = linebookingstate.Cancelled
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		if _, err = repo.Update(ctx, target, "state", "update_date", "update_by"); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "cancel line booking")
	}
	return &models.LineBookingCancelResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}

// findNewForOwner loads an active, new booking the caller may confirm or cancel: NOT_FOUND when
// they can't see it or it is deleted, PERMISSION_DENIED when they see it only as finance,
// LINE_BOOKING_NOT_NEW when it was already confirmed or cancelled.
func (u *lineBookingUseCase) findNewForOwner(ctx context.Context, repo repository.LineBookingRepository, id string, info *coremodels.UserRequestInfo) (*domain.LineBooking, error) {
	target, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if target == nil || !target.Status || !visibleTo(target, info) {
		return nil, cerrors.NotFound()
	}
	if !ownsAsSales(target, info) {
		return nil, cerrors.PermissionDenied()
	}
	if target.State != linebookingstate.New {
		return nil, cerrors.New(http.StatusConflict, errorcode.LineBookingNotNew)
	}
	return target, nil
}

// financeOptions are the finance users a booking can be sent to: the sales owner's active finance
// users, or every active finance user when the booking has no owner.
func (u *lineBookingUseCase) financeOptions(ctx context.Context, appDB query.AppIDB, b *domain.LineBooking) ([]models.LineBookingUser, error) {
	if !b.SalesUserID.Valid {
		users, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).ListActiveByRole(ctx, role.Finance)
		if err != nil {
			return nil, err
		}
		items := make([]models.LineBookingUser, len(users))
		for i, user := range users {
			items[i] = models.LineBookingUser{ID: user.ID, Name: user.Name}
		}
		return items, nil
	}
	links, err := u.repositoryRegistry.SalesFinanceRepository.WithTx(appDB).ListBySales(ctx, b.SalesUserID.String)
	if err != nil {
		return nil, err
	}
	items := []models.LineBookingUser{}
	for _, l := range links {
		if usableFinance(l.FinanceUser) {
			items = append(items, models.LineBookingUser{ID: l.FinanceUser.ID, Name: l.FinanceUser.Name})
		}
	}
	return items, nil
}

// usableFinance: an active user who still holds the finance role.
func usableFinance(user *domain.AppUser) bool {
	return user != nil && user.Status && slices.Contains(user.Roles, role.Finance)
}
