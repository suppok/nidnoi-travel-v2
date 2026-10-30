package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/wholesale/models"
)

func (u *wholesaleUseCase) GetValidate(request *models.WholesaleGetRequest) error {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return validation.Errors{"id": err}
	}
	return nil
}

func (u *wholesaleUseCase) Get(ctx context.Context, request *models.WholesaleGetRequest) (*models.WholesaleGetResponse, error) {
	if err := u.GetValidate(request); err != nil {
		return nil, cerrors.NotFound()
	}

	found, err := u.repositoryRegistry.WholesaleRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get wholesale")
	}
	if found == nil {
		return nil, cerrors.NotFound()
	}
	return &models.WholesaleGetResponse{WholesaleItem: models.NewWholesaleItem(found)}, nil
}
