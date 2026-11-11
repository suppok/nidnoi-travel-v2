package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type BookingSourceRepository interface {
	WithTx(tx query.AppIDB) BookingSourceRepository
	Create(ctx context.Context, obj *domain.BookingSource) (*domain.BookingSource, error)
	Update(ctx context.Context, obj *domain.BookingSource, columns ...string) (*domain.BookingSource, error)
	FindByID(ctx context.Context, id string) (*domain.BookingSource, error)
	List(ctx context.Context, criteria repomodels.BookingSourceListCriteria) ([]domain.BookingSource, int, error)
}

type bookingSourceRepository struct {
	query query.DB[domain.BookingSource]
}

func NewBookingSourceRepository(db query.AppIDB) BookingSourceRepository {
	return &bookingSourceRepository{query: query.New[domain.BookingSource](db)}
}

func (r *bookingSourceRepository) WithTx(tx query.AppIDB) BookingSourceRepository {
	return NewBookingSourceRepository(tx)
}

func (r *bookingSourceRepository) Create(ctx context.Context, obj *domain.BookingSource) (*domain.BookingSource, error) {
	return r.query.Insert(ctx, obj)
}

func (r *bookingSourceRepository) Update(ctx context.Context, obj *domain.BookingSource, columns ...string) (*domain.BookingSource, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *bookingSourceRepository) FindByID(ctx context.Context, id string) (*domain.BookingSource, error) {
	obj := new(domain.BookingSource)
	err := r.query.NewSelect().Model(obj).Where("bs.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *bookingSourceRepository) List(ctx context.Context, criteria repomodels.BookingSourceListCriteria) ([]domain.BookingSource, int, error) {
	var items []domain.BookingSource
	q := r.query.NewSelect().Model(&items)

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(bs.name) LIKE ?", pattern).WhereOr("lower(bs.short_name) LIKE ?", pattern)
		})
	}
	if criteria.Status != nil {
		q = q.Where("bs.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("bs.sort_order ASC, bs.name ASC, bs.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
