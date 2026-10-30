package usecase

import (
	"context"
	"strings"
	"time"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/country/models"
)

func (u *countryUseCase) CreateValidate(request *models.CountryCreateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.RefID, refIDRules()...),
		validation.Field(&request.Slug, slugRules()...),
		validation.Field(&request.NameTh, nameRules()...),
		validation.Field(&request.NameEn, nameRules()...),
		validation.Field(&request.Description, descriptionRules()...),
		validation.Field(&request.HeroImage, heroImageRules()...),
		validation.Field(&request.SortOrder, sortOrderRules()...),
	)
}

func (u *countryUseCase) Create(ctx context.Context, request *models.CountryCreateRequest) (*models.CountryCreateResponse, error) {
	request.Slug = strings.TrimSpace(request.Slug)
	normalizeText(&request.NameTh, &request.NameEn, &request.Description, &request.HeroImage)
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	now := time.Now()
	created, err := u.repositoryRegistry.CountryRepository.Create(ctx, &domain.Country{
		ID:          utils.NewID(),
		RefID:       request.RefID,
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
		return nil, cerrors.Wrap(uniqueConflict(err), "create country")
	}
	return &models.CountryCreateResponse{CountryItem: models.NewCountryItem(created)}, nil
}
