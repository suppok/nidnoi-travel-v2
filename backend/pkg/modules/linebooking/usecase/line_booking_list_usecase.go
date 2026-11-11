package usecase

import (
	"context"

	"backend/pkg/constants/linebookingstate"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/linebooking/models"
)

func stateValues() []any {
	values := make([]any, len(linebookingstate.All))
	for i, s := range linebookingstate.All {
		values[i] = s
	}
	return values
}

// assignedTo is the caller's id when they asked for the bookings sent to them (finance alerts).
func assignedTo(request *models.LineBookingListRequest) string {
	if request.AssignedToMe && request.UserRequestInfo != nil {
		return request.UserRequestInfo.UserID
	}
	return ""
}

// ownedBy is the caller's id when they asked for the bookings they own (sales alerts).
func ownedBy(request *models.LineBookingListRequest) string {
	if request.OwnedByMe && request.UserRequestInfo != nil {
		return request.UserRequestInfo.UserID
	}
	return ""
}

func (u *lineBookingUseCase) ListValidate(request *models.LineBookingListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Keyword, validation.RuneLength(0, 100)),
		validation.Field(&request.State, validation.In(stateValues()...)),
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxPageSize)),
	)
}

func (u *lineBookingUseCase) List(ctx context.Context, request *models.LineBookingListRequest) (*models.LineBookingListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}
	items, total, err := u.repositoryRegistry.LineBookingRepository.List(ctx, repomodels.LineBookingListCriteria{
		Keyword: request.Keyword,
		State:   request.State,
		Paid:    request.Paid,
		Status:  request.Status,
		// Sales see the bookings they own, finance the bookings sent to them.
		Scope:             ownerScope(request.UserRequestInfo),
		FinanceUserID:     assignedTo(request),
		SalesUserID:       ownedBy(request),
		DocumentsDone:     request.DocumentsDone,
		DocumentsApproved: request.DocumentsApproved,
		Page:              request.Page,
		PageSize:          request.PageSize,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list line bookings")
	}
	result := make([]models.LineBookingItem, len(items))
	for i := range items {
		result[i] = models.NewLineBookingItem(&items[i])
	}
	return &models.LineBookingListResponse{Items: result, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}
