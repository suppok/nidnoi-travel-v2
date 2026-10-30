package usecase

import (
	"context"
	"strings"
	"time"

	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/city/models"
)

func (u *cityUseCase) CreateValidate(request *models.CityCreateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.CountryID, validation.Required, validation.UUID),
		validation.Field(&request.Slug, slugRules()...),
		validation.Field(&request.NameTh, nameRules()...),
		validation.Field(&request.NameEn, nameRules()...),
		validation.Field(&request.Description, descriptionRules()...),
		validation.Field(&request.HeroImage, heroImageRules()...),
		validation.Field(&request.SortOrder, sortOrderRules()...),
	)
}

func (u *cityUseCase) Create(ctx context.Context, request *models.CityCreateRequest) (*models.CityCreateResponse, error) {
	request.Slug = strings.TrimSpace(request.Slug)
	normalizeText(&request.NameTh, &request.NameEn, &request.Description, &request.HeroImage)
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var created *domain.City
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		country, err := u.activeCountry(ctx, appDB, request.CountryID)
		if err != nil {
			return err
		}

		now := time.Now()
		created, err = u.repositoryRegistry.CityRepository.WithTx(appDB).Create(ctx, &domain.City{
			ID:          utils.NewID(),
			RefID:       request.RefID,
			CountryID:   country.ID,
			Slug:        request.Slug,
			NameTh:      request.NameTh,
			NameEn:      request.NameEn,
			Description: request.Description,
			HeroImage:   request.HeroImage,
			SortOrder:   request.SortOrder,
			Status:      true,
			CreateDate:  now,
			UpdateDate:  now,
			CreateBy:    request.UserRequestInfo.UserID,
			UpdateBy:    request.UserRequestInfo.UserID,
		})
		if err != nil {
			return uniqueConflict(err)
		}
		created.Country = country
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "create city")
	}
	return &models.CityCreateResponse{CityItem: models.NewCityItem(created)}, nil
}
