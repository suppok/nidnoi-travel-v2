package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/country/models"
)

func (u *countryUseCase) UpdateValidate(request *models.CountryUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.NameTh, nameRules()...),
		validation.Field(&request.NameEn, nameRules()...),
		validation.Field(&request.Description, descriptionRules()...),
		validation.Field(&request.HeroImage, heroImageRules()...),
		validation.Field(&request.SortOrder, sortOrderRules()...),
	)
}

// Update changes the display fields and ref ID. The slug is never written (locked after create).
func (u *countryUseCase) Update(ctx context.Context, request *models.CountryUpdateRequest) (*models.CountryUpdateResponse, error) {
	normalizeText(&request.NameTh, &request.NameEn, &request.Description, &request.HeroImage)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.Country
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}

		target.RefID = request.RefID
		target.NameTh = request.NameTh
		target.NameEn = request.NameEn
		target.Description = request.Description
		target.HeroImage = request.HeroImage
		target.SortOrder = request.SortOrder
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		updated, err = u.repositoryRegistry.CountryRepository.WithTx(appDB).Update(ctx, target,
			"ref_id", "name_th", "name_en", "description", "hero_image", "sort_order", "update_date", "update_by")
		return uniqueConflict(err)
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update country")
	}
	u.enqueueIfShown(ctx, updated)
	return &models.CountryUpdateResponse{CountryItem: models.NewCountryItem(updated)}, nil
}
