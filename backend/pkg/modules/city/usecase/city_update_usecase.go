package usecase

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/city/models"
	"backend/pkg/modules/shared/services/revalidation"
)

func (u *cityUseCase) UpdateValidate(request *models.CityUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.CountryID, validation.Required, validation.UUID),
		validation.Field(&request.NameTh, nameRules()...),
		validation.Field(&request.NameEn, nameRules()...),
		validation.Field(&request.Description, descriptionRules()...),
		validation.Field(&request.HeroImage, heroImageRules()...),
		validation.Field(&request.SortOrder, sortOrderRules()...),
	)
}

// Update changes the country, display fields and ref ID. The slug is never written (locked after create).
func (u *cityUseCase) Update(ctx context.Context, request *models.CityUpdateRequest) (*models.CityUpdateResponse, error) {
	normalizeText(&request.NameTh, &request.NameEn, &request.Description, &request.HeroImage)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.City
	var pathsBefore []string // a country move changes the landing URL: rebuild both
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}
		country := target.Country
		if country != nil {
			pathsBefore = revalidation.PathsForCity(country.Slug, target.Slug)
		}
		if request.CountryID != target.CountryID {
			if country, err = u.activeCountry(ctx, appDB, request.CountryID); err != nil {
				return err
			}
		}

		target.RefID = request.RefID
		target.CountryID = country.ID
		target.NameTh = request.NameTh
		target.NameEn = request.NameEn
		target.Description = request.Description
		target.HeroImage = request.HeroImage
		target.SortOrder = request.SortOrder
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		target.Country = nil // not a column; reattached below
		updated, err = u.repositoryRegistry.CityRepository.WithTx(appDB).Update(ctx, target,
			"ref_id", "country_id", "name_th", "name_en", "description", "hero_image", "sort_order", "update_date", "update_by")
		if err != nil {
			return uniqueConflict(err)
		}
		updated.Country = country
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update city")
	}
	u.enqueueIfShown(ctx, updated.ID, append(pathsBefore, revalidation.PathsForCity(updated.Country.Slug, updated.Slug)...))
	return &models.CityUpdateResponse{CityItem: models.NewCityItem(updated)}, nil
}
