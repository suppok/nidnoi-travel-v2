package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type TourProgramImageRepository interface {
	WithTx(tx query.AppIDB) TourProgramImageRepository
	// ReplaceForTour deletes the tour's gallery images and inserts images (may be empty).
	ReplaceForTour(ctx context.Context, tourProgramID string, images []domain.TourProgramImage) error
}

type tourProgramImageRepository struct{ db query.AppIDB }

func NewTourProgramImageRepository(db query.AppIDB) TourProgramImageRepository {
	return &tourProgramImageRepository{db: db}
}

func (r *tourProgramImageRepository) WithTx(tx query.AppIDB) TourProgramImageRepository {
	return NewTourProgramImageRepository(tx)
}

func (r *tourProgramImageRepository) ReplaceForTour(ctx context.Context, tourProgramID string, images []domain.TourProgramImage) error {
	if _, err := r.db.NewDelete().Model((*domain.TourProgramImage)(nil)).Where("tour_program_id = ?", tourProgramID).Exec(ctx); err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&images).Exec(ctx)
	return err
}
