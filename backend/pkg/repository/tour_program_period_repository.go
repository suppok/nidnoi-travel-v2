package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
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

	// --- Dashboard aggregates: active periods of active, public tours departing in [from, to] (dates). ---

	// SeatStats sums their seats; lowSeats is the most seats left that still counts as "low".
	SeatStats(ctx context.Context, from, to time.Time, lowSeats int) (*repomodels.PeriodSeatStats, error)
	// ListUpcoming returns the ones with seats left, fewest seats first, then by departure; at most limit rows.
	ListUpcoming(ctx context.Context, from, to time.Time, limit int) ([]repomodels.UpcomingPeriod, error)
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

func (r *tourProgramPeriodRepository) upcomingQuery(from, to time.Time) *bun.SelectQuery {
	return r.query.NewSelect().
		Join("JOIN tour_program AS tp ON tp.id = tpp.tour_program_id").
		Where("tpp.status = TRUE").
		Where("tp.status = TRUE").
		Where("tp.state IN (?)", bun.In(tourprogramstate.Public)).
		Where("tpp.start_date >= ?", from.Format(time.DateOnly)).
		Where("tpp.start_date <= ?", to.Format(time.DateOnly))
}

func (r *tourProgramPeriodRepository) SeatStats(ctx context.Context, from, to time.Time, lowSeats int) (*repomodels.PeriodSeatStats, error) {
	out := new(repomodels.PeriodSeatStats)
	err := r.upcomingQuery(from, to).
		ColumnExpr("count(*) AS periods").
		ColumnExpr("coalesce(sum(tpp.seats_total), 0) AS seats_total").
		ColumnExpr("coalesce(sum(tpp.seats_available), 0) AS seats_available").
		ColumnExpr("count(*) FILTER (WHERE tpp.seats_available <= 0) AS sold_out").
		ColumnExpr("count(*) FILTER (WHERE tpp.seats_available BETWEEN 1 AND ?) AS low_seats", lowSeats).
		Scan(ctx, out)
	return out, err
}

func (r *tourProgramPeriodRepository) ListUpcoming(ctx context.Context, from, to time.Time, limit int) ([]repomodels.UpcomingPeriod, error) {
	var rows []repomodels.UpcomingPeriod
	err := r.upcomingQuery(from, to).
		ColumnExpr("tpp.id AS period_id").
		ColumnExpr("tp.id AS tour_program_id").
		ColumnExpr("tp.code AS tour_code").
		ColumnExpr("tp.name AS tour_name").
		ColumnExpr("tpp.start_date, tpp.end_date, tpp.price, tpp.seats_total, tpp.seats_available").
		Where("tpp.seats_available > 0").
		OrderExpr("tpp.seats_available ASC, tpp.start_date ASC, tp.code ASC").
		Limit(int64(limit)).
		Scan(ctx, &rows)
	return rows, err
}
