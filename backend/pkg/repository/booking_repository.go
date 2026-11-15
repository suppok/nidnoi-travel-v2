package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/constants/bookingstate"
	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type BookingRepository interface {
	WithTx(tx query.AppIDB) BookingRepository
	Create(ctx context.Context, obj *domain.Booking) (*domain.Booking, error)
	Update(ctx context.Context, obj *domain.Booking, columns ...string) (*domain.Booking, error)
	// FindByID loads the booking with its Assignee.
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	// List returns bookings newest first, with their Assignee.
	List(ctx context.Context, criteria repomodels.BookingListCriteria) ([]domain.Booking, int, error)

	// --- Dashboard aggregates: active bookings created in the range. ---

	StatsByState(ctx context.Context, rng repomodels.DateRange) ([]repomodels.StateCount, error)
	// StatsBuckets counts bookings per date_trunc(unit) of the create date in Thailand time. Empty
	// buckets are absent; Revenue is 0 (requests are leads, not sales).
	StatsBuckets(ctx context.Context, rng repomodels.DateRange, unit string) ([]repomodels.BucketCount, error)
	// CountNew counts the active bookings still in state new (any create date).
	CountNew(ctx context.Context) (int, error)
}

type bookingRepository struct{ query query.DB[domain.Booking] }

func NewBookingRepository(db query.AppIDB) BookingRepository {
	return &bookingRepository{query: query.New[domain.Booking](db)}
}

func (r *bookingRepository) WithTx(tx query.AppIDB) BookingRepository {
	return NewBookingRepository(tx)
}

func (r *bookingRepository) Create(ctx context.Context, obj *domain.Booking) (*domain.Booking, error) {
	return r.query.Insert(ctx, obj)
}

func (r *bookingRepository) Update(ctx context.Context, obj *domain.Booking, columns ...string) (*domain.Booking, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *bookingRepository) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	obj := new(domain.Booking)
	err := r.query.NewSelect().Model(obj).Relation("Assignee").Where("bk.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *bookingRepository) List(ctx context.Context, criteria repomodels.BookingListCriteria) ([]domain.Booking, int, error) {
	var items []domain.Booking
	q := r.query.NewSelect().Model(&items).Relation("Assignee")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(bk.booking_no) LIKE ?", pattern).
				WhereOr("lower(bk.first_name || ' ' || bk.last_name) LIKE ?", pattern).
				WhereOr("lower(bk.email) LIKE ?", pattern).
				WhereOr("bk.phone LIKE ?", pattern).
				WhereOr("lower(bk.tour_code) LIKE ?", pattern)
		})
	}
	if criteria.State != 0 {
		q = q.Where("bk.state = ?", criteria.State)
	}
	if criteria.Status != nil {
		q = q.Where("bk.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("bk.create_date DESC, bk.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}

func (r *bookingRepository) statsQuery(rng repomodels.DateRange) *bun.SelectQuery {
	return r.query.NewSelect().
		Where("bk.status = TRUE").
		Where("bk.create_date >= ?", rng.From).
		Where("bk.create_date < ?", rng.To)
}

func (r *bookingRepository) StatsByState(ctx context.Context, rng repomodels.DateRange) ([]repomodels.StateCount, error) {
	var rows []repomodels.StateCount
	err := r.statsQuery(rng).
		ColumnExpr("bk.state AS state").
		ColumnExpr("count(*) AS count").
		GroupExpr("bk.state").
		OrderExpr("bk.state").
		Scan(ctx, &rows)
	return rows, err
}

func (r *bookingRepository) StatsBuckets(ctx context.Context, rng repomodels.DateRange, unit string) ([]repomodels.BucketCount, error) {
	var rows []repomodels.BucketCount
	err := r.statsQuery(rng).
		ColumnExpr("date_trunc(?, bk.create_date AT TIME ZONE 'Asia/Bangkok')::date AS bucket", unit).
		ColumnExpr("count(*) AS count").
		GroupExpr("1").
		OrderExpr("1").
		Scan(ctx, &rows)
	return rows, err
}

func (r *bookingRepository) CountNew(ctx context.Context) (int, error) {
	n, err := r.query.NewSelect().
		Where("bk.status = TRUE").
		Where("bk.state = ?", bookingstate.New).
		Count(ctx)
	return int(n), err
}
