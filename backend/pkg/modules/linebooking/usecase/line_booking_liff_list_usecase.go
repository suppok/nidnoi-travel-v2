package usecase

import (
	"context"
	"strings"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/linebooking/models"
)

// liffListLimit caps the bookings listed for one customer chat; older ones still open by booking no.
const liffListLimit = 50

func (u *lineBookingUseCase) LIFFListValidate(request *models.LineBookingLIFFListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.LineContactID, validation.Required, validation.UUID),
	)
}

// LIFFList lists the bookings of one customer chat that LIFF may reopen (active and paid, the same
// rule as findEditable), so sales can pick one instead of typing its booking no. Like reopening,
// it isn't limited to the caller's own bookings.
func (u *lineBookingUseCase) LIFFList(ctx context.Context, request *models.LineBookingLIFFListRequest) (*models.LineBookingLIFFListResponse, error) {
	request.LineContactID = strings.TrimSpace(request.LineContactID)
	if err := u.LIFFListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	active := true
	items, _, err := u.repositoryRegistry.LineBookingRepository.List(ctx, repomodels.LineBookingListCriteria{
		Paid:          true,
		Status:        &active,
		LineContactID: request.LineContactID,
		Page:          1,
		PageSize:      liffListLimit,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list line bookings for LIFF")
	}
	result := make([]models.LineBookingItem, 0, len(items))
	for i := range items {
		if items[i].EditableFromLIFF() {
			result = append(result, models.NewLineBookingItem(&items[i]))
		}
	}
	return &models.LineBookingLIFFListResponse{Items: result}, nil
}
