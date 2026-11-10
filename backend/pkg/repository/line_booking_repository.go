package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type LineBookingRepository interface {
	WithTx(tx query.AppIDB) LineBookingRepository
	Create(ctx context.Context, obj *domain.LineBooking) (*domain.LineBooking, error)
	Update(ctx context.Context, obj *domain.LineBooking, columns ...string) (*domain.LineBooking, error)
	// FindByID loads the booking with its LineContact.
	FindByID(ctx context.Context, id string) (*domain.LineBooking, error)
	// List returns LINE bookings newest first, with their LineContact.
	List(ctx context.Context, criteria repomodels.LineBookingListCriteria) ([]domain.LineBooking, int, error)
}

type lineBookingRepository struct{ query query.DB[domain.LineBooking] }

func NewLineBookingRepository(db query.AppIDB) LineBookingRepository {
	return &lineBookingRepository{query: query.New[domain.LineBooking](db)}
}

func (r *lineBookingRepository) WithTx(tx query.AppIDB) LineBookingRepository {
	return NewLineBookingRepository(tx)
}

func (r *lineBookingRepository) Create(ctx context.Context, obj *domain.LineBooking) (*domain.LineBooking, error) {
	return r.query.Insert(ctx, obj)
}

func (r *lineBookingRepository) Update(ctx context.Context, obj *domain.LineBooking, columns ...string) (*domain.LineBooking, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *lineBookingRepository) FindByID(ctx context.Context, id string) (*domain.LineBooking, error) {
	obj := new(domain.LineBooking)
	err := r.query.NewSelect().Model(obj).Relation("LineContact").Where("lb.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *lineBookingRepository) List(ctx context.Context, criteria repomodels.LineBookingListCriteria) ([]domain.LineBooking, int, error) {
	var items []domain.LineBooking
	q := r.query.NewSelect().Model(&items).Relation("LineContact")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(lb.booking_no) LIKE ?", pattern).
				WhereOr("lower(lb.customer_name) LIKE ?", pattern).
				WhereOr("lb.customer_phone LIKE ?", pattern).
				WhereOr("lower(lb.tour_code) LIKE ?", pattern).
				WhereOr("lower(lb.line_display_name) LIKE ?", pattern)
		})
	}
	if criteria.State != 0 {
		q = q.Where("lb.state = ?", criteria.State)
	}
	if criteria.Status != nil {
		q = q.Where("lb.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("lb.create_date DESC, lb.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
