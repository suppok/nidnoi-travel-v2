package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/airline/models"
)

func (u *airlineUseCase) CreateValidate(request *models.AirlineCreateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.Code, codeRules()...),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.Logo, u.logoRules()...),
	)
}

func (u *airlineUseCase) Create(ctx context.Context, request *models.AirlineCreateRequest) (*models.AirlineCreateResponse, error) {
	normalize(&request.Code, &request.Name, &request.Logo)
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var created *domain.Airline
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		actorID := request.UserRequestInfo.UserID
		now := time.Now()
		var err error
		created, err = u.repositoryRegistry.AirlineRepository.WithTx(appDB).Create(ctx, &domain.Airline{
			ID:         utils.NewID(),
			RefID:      request.RefID,
			Code:       request.Code,
			Name:       request.Name,
			Logo:       request.Logo,
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
		return nil, cerrors.Wrap(err, "create airline")
	}
	return &models.AirlineCreateResponse{AirlineItem: models.NewAirlineItem(created)}, nil
}
