package usecase

import (
	"context"
	"time"

	"backend/pkg/constants/wholesalestate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/wholesale/models"
)

func (u *wholesaleUseCase) CreateValidate(request *models.WholesaleCreateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.Code, codeRules()...),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.State, stateRules()...),
		validation.Field(&request.Contacts, contactsRules()...),
	)
}

func (u *wholesaleUseCase) Create(ctx context.Context, request *models.WholesaleCreateRequest) (*models.WholesaleCreateResponse, error) {
	normalize(&request.Code, &request.Name, request.Contacts)
	if request.State == 0 {
		request.State = wholesalestate.OnRequest
	}
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var created *domain.Wholesale
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		actorID := request.UserRequestInfo.UserID
		now := time.Now()
		var err error
		created, err = u.repositoryRegistry.WholesaleRepository.WithTx(appDB).Create(ctx, &domain.Wholesale{
			ID:         utils.NewID(),
			RefID:      request.RefID,
			Code:       request.Code,
			Name:       request.Name,
			State:      request.State,
			Status:     true,
			CreateDate: now,
			UpdateDate: now,
			CreateBy:   actorID,
			UpdateBy:   actorID,
		})
		if err != nil {
			return uniqueConflict(err)
		}
		created.Contacts, err = u.replaceContacts(ctx, appDB, created.ID, actorID, request.Contacts)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "create wholesale")
	}
	return &models.WholesaleCreateResponse{WholesaleItem: models.NewWholesaleItem(created)}, nil
}
