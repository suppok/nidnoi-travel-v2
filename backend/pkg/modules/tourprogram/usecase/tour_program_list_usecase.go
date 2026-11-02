package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/tourprogram/models"
)

func (u *tourProgramUseCase) ListValidate(request *models.TourProgramListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.CountryID, validation.UUID),
		validation.Field(&request.CityID, validation.UUID),
		validation.Field(&request.WholesaleID, validation.UUID),
		validation.Field(&request.State, validation.In(knownStates()...)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *tourProgramUseCase) List(ctx context.Context, request *models.TourProgramListRequest) (*models.TourProgramListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	tours, total, err := u.repositoryRegistry.TourProgramRepository.List(ctx, repomodels.TourProgramListCriteria{
		Keyword:     request.Keyword,
		CountryID:   request.CountryID,
		CityID:      request.CityID,
		WholesaleID: request.WholesaleID,
		State:       request.State,
		Status:      request.Status,
		Page:        request.Page,
		PageSize:    request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list tour programs")
	}

	items := make([]models.TourProgramListItem, len(tours))
	for i := range tours {
		items[i] = models.NewTourProgramListItem(&tours[i])
	}
	return &models.TourProgramListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
