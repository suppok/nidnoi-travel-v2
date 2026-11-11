package usecase

import (
	"context"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain/repomodels"
	"backend/pkg/modules/linebooking/models"
)

const maxBookingSources = 500

// SourceList returns the active booking sources for the LIFF booking form.
func (u *lineBookingUseCase) SourceList(ctx context.Context, _ *models.LineBookingSourceListRequest) (*models.LineBookingSourceListResponse, error) {
	active := true
	sources, _, err := u.repositoryRegistry.BookingSourceRepository.List(ctx, repomodels.BookingSourceListCriteria{
		Status: &active, Page: 1, PageSize: maxBookingSources,
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "list booking sources")
	}
	res := &models.LineBookingSourceListResponse{Items: make([]models.LineBookingSource, len(sources))}
	for i, s := range sources {
		res.Items[i] = models.LineBookingSource{ID: s.ID, Name: s.Name, ShortName: s.ShortName}
	}
	return res, nil
}
