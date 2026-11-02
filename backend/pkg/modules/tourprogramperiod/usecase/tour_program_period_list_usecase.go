package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/tourprogramperiod/models"
)

func (u *tourProgramPeriodUseCase) ListValidate(request *models.TourProgramPeriodListRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.TourProgramID, validation.Required, validation.UUID),
	)
}

// List returns the tour's periods by departure date, optionally filtered by status.
func (u *tourProgramPeriodUseCase) List(ctx context.Context, request *models.TourProgramPeriodListRequest) (*models.TourProgramPeriodListResponse, error) {
	if err := u.ListValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	periods, err := u.repositoryRegistry.TourProgramPeriodRepository.ListByTour(ctx, request.TourProgramID)
	if err != nil {
		return nil, cerrors.Wrap(err, "list tour program periods")
	}

	items := make([]models.TourProgramPeriodItem, 0, len(periods))
	for i := range periods {
		if request.Status != nil && periods[i].Status != *request.Status {
			continue
		}
		items = append(items, models.NewTourProgramPeriodItem(&periods[i]))
	}
	return &models.TourProgramPeriodListResponse{Items: items}, nil
}
