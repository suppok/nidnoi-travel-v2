package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/bookingsource/models"
)

func (u *bookingSourceUseCase) ListValidate(request *models.BookingSourceListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *bookingSourceUseCase) List(ctx context.Context, request *models.BookingSourceListRequest) (*models.BookingSourceListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}

	sources, total, err := u.repositoryRegistry.BookingSourceRepository.List(ctx, repomodels.BookingSourceListCriteria{
		Keyword:  request.Keyword,
		Status:   request.Status,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list booking sources")
	}

	items := make([]models.BookingSourceItem, len(sources))
	for i := range sources {
		items[i] = models.NewBookingSourceItem(&sources[i])
	}
	return &models.BookingSourceListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
