package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type TourProgramPeriodRepository interface {
	WithTx(tx query.AppIDB) TourProgramPeriodRepository
	Create(ctx context.Context, obj *domain.TourProgramPeriod) (*domain.TourProgramPeriod, error)
	Update(ctx context.Context, obj *domain.TourProgramPeriod, columns ...string) (*domain.TourProgramPeriod, error)
	FindByID(ctx context.Context, id string) (*domain.TourProgramPeriod, error)
	// ListByTour returns every period of the tour (active and deleted) by departure date.
	ListByTour(ctx context.Context, tourProgramID string) ([]domain.TourProgramPeriod, error)
}

type tourProgramPeriodRepository struct {
	query query.DB[domain.TourProgramPeriod]
}

func NewTourProgramPeriodRepository(db query.AppIDB) TourProgramPeriodRepository {
	return &tourProgramPeriodRepository{query: query.New[domain.TourProgramPeriod](db)}
}

func (r *tourProgramPeriodRepository) WithTx(tx query.AppIDB) TourProgramPeriodRepository {
	return NewTourProgramPeriodRepository(tx)
}

func (r *tourProgramPeriodRepository) Create(ctx context.Context, obj *domain.TourProgramPeriod) (*domain.TourProgramPeriod, error) {
	return r.query.Insert(ctx, obj)
}

func (r *tourProgramPeriodRepository) Update(ctx context.Context, obj *domain.TourProgramPeriod, columns ...string) (*domain.TourProgramPeriod, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *tourProgramPeriodRepository) FindByID(ctx context.Context, id string) (*domain.TourProgramPeriod, error) {
	return r.query.FindByID(ctx, id)
}

func (r *tourProgramPeriodRepository) ListByTour(ctx context.Context, tourProgramID string) ([]domain.TourProgramPeriod, error) {
	var items []domain.TourProgramPeriod
	err := r.query.NewSelect().
		Where("tpp.tour_program_id = ?", tourProgramID).
		OrderExpr("tpp.start_date ASC, tpp.end_date ASC, tpp.id ASC").
		Scan(ctx, &items)
	return items, err
}
