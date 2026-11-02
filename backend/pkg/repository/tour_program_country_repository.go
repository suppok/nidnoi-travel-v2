package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type TourProgramCountryRepository interface {
	WithTx(tx query.AppIDB) TourProgramCountryRepository
	// ReplaceForTour deletes the tour's country links and inserts links.
	ReplaceForTour(ctx context.Context, tourProgramID string, links []domain.TourProgramCountry) error
}

type tourProgramCountryRepository struct{ db query.AppIDB }

func NewTourProgramCountryRepository(db query.AppIDB) TourProgramCountryRepository {
	return &tourProgramCountryRepository{db: db}
}

func (r *tourProgramCountryRepository) WithTx(tx query.AppIDB) TourProgramCountryRepository {
	return NewTourProgramCountryRepository(tx)
}

func (r *tourProgramCountryRepository) ReplaceForTour(ctx context.Context, tourProgramID string, links []domain.TourProgramCountry) error {
	if _, err := r.db.NewDelete().Model((*domain.TourProgramCountry)(nil)).Where("tour_program_id = ?", tourProgramID).Exec(ctx); err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&links).Exec(ctx)
	return err
}
