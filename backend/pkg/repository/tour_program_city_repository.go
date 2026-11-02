package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type TourProgramCityRepository interface {
	WithTx(tx query.AppIDB) TourProgramCityRepository
	// ReplaceForTour deletes the tour's city links and inserts links (may be empty).
	ReplaceForTour(ctx context.Context, tourProgramID string, links []domain.TourProgramCity) error
}

type tourProgramCityRepository struct{ db query.AppIDB }

func NewTourProgramCityRepository(db query.AppIDB) TourProgramCityRepository {
	return &tourProgramCityRepository{db: db}
}

func (r *tourProgramCityRepository) WithTx(tx query.AppIDB) TourProgramCityRepository {
	return NewTourProgramCityRepository(tx)
}

func (r *tourProgramCityRepository) ReplaceForTour(ctx context.Context, tourProgramID string, links []domain.TourProgramCity) error {
	if _, err := r.db.NewDelete().Model((*domain.TourProgramCity)(nil)).Where("tour_program_id = ?", tourProgramID).Exec(ctx); err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&links).Exec(ctx)
	return err
}
