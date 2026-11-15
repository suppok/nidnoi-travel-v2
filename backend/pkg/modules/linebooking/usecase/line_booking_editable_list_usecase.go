package usecase

import (
	"context"
	"strings"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/linebooking/models"
)

// editableListLimit caps the bookings listed for one customer chat; older ones still open by booking no.
const editableListLimit = 50

func (u *lineBookingUseCase) EditableListValidate(request *models.LineBookingEditableListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.LineContactID, validation.Required, validation.UUID),
	)
}

// EditableList lists the bookings of one customer chat that can be reopened (active and paid, the same
// rule as findEditable), so sales can pick one instead of typing its booking no. Like reopening,
// it isn't limited to the caller's own bookings.
func (u *lineBookingUseCase) EditableList(ctx context.Context, request *models.LineBookingEditableListRequest) (*models.LineBookingEditableListResponse, error) {
	request.LineContactID = strings.TrimSpace(request.LineContactID)
	if err := u.EditableListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	active := true
	items, _, err := u.repositoryRegistry.LineBookingRepository.List(ctx, repomodels.LineBookingListCriteria{
		Paid:          true,
		Status:        &active,
		LineContactID: request.LineContactID,
		Page:          1,
		PageSize:      editableListLimit,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list editable line bookings")
	}
	result := make([]models.LineBookingItem, 0, len(items))
	for i := range items {
		if items[i].Reopenable() {
			result = append(result, models.NewLineBookingItem(&items[i]))
		}
	}
	return &models.LineBookingEditableListResponse{Items: result}, nil
}
