package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/bookingsource/models"
)

func (u *bookingSourceUseCase) CreateValidate(request *models.BookingSourceCreateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.ShortName, shortNameRules()...),
		validation.Field(&request.SortOrder, sortOrderRules()...),
	)
}

func (u *bookingSourceUseCase) Create(ctx context.Context, request *models.BookingSourceCreateRequest) (*models.BookingSourceCreateResponse, error) {
	normalize(&request.Name, &request.ShortName)
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var created *domain.BookingSource
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		actorID := request.UserRequestInfo.UserID
		now := time.Now()
		var err error
		created, err = u.repositoryRegistry.BookingSourceRepository.WithTx(appDB).Create(ctx, &domain.BookingSource{
			ID:         utils.NewID(),
			Name:       request.Name,
			ShortName:  request.ShortName,
			SortOrder:  request.SortOrder,
			Status:     true,
			CreateDate: now,
			UpdateDate: now,
			CreateBy:   actorID,
			UpdateBy:   actorID,
		})
		if err != nil {
			return uniqueConflict(err)
		}
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "create booking source")
	}
	return &models.BookingSourceCreateResponse{BookingSourceItem: models.NewBookingSourceItem(created)}, nil
}
