package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

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
