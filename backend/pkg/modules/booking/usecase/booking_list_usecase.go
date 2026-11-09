package usecase

import (
	"context"

	"backend/pkg/constants/bookingstate"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/booking/models"
)

func stateValues() []any {
	values := make([]any, len(bookingstate.All))
	for i, s := range bookingstate.All {
		values[i] = s
	}
	return values
}

func (u *bookingUseCase) ListValidate(request *models.BookingListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.State, validation.In(stateValues()...)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

// List returns bookings newest first.
func (u *bookingUseCase) List(ctx context.Context, request *models.BookingListRequest) (*models.BookingListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}
	bookings, total, err := u.repositoryRegistry.BookingRepository.List(ctx, repomodels.BookingListCriteria{
		Keyword:  request.Keyword,
		State:    request.State,
		Status:   request.Status,
		Page:     request.Page,
		PageSize: request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list bookings")
	}
	items := make([]models.BookingItem, len(bookings))
	for i := range bookings {
		items[i] = models.NewBookingItem(&bookings[i])
	}
	return &models.BookingListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
