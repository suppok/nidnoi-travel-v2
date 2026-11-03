package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"

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
	// ListPublicByTours returns the active upcoming periods (start_date > departAfter) of the tours,
	// by tour and departure date.
	ListPublicByTours(ctx context.Context, tourProgramIDs []string, departAfter time.Time) ([]domain.TourProgramPeriod, error)
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

func (r *tourProgramPeriodRepository) ListPublicByTours(ctx context.Context, tourProgramIDs []string, departAfter time.Time) ([]domain.TourProgramPeriod, error) {
	var items []domain.TourProgramPeriod
	if len(tourProgramIDs) == 0 {
		return items, nil
	}
	err := r.query.NewSelect().
		Where("tpp.tour_program_id IN (?)", bun.In(tourProgramIDs)).
		Where("tpp.status = TRUE").
		Where("tpp.start_date > ?", departAfter.Format(time.DateOnly)).
		OrderExpr("tpp.tour_program_id ASC, tpp.start_date ASC, tpp.end_date ASC, tpp.id ASC").
		Scan(ctx, &items)
	return items, err
}
