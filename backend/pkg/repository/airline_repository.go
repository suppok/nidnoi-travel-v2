package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type AirlineRepository interface {
	WithTx(tx query.AppIDB) AirlineRepository
	Create(ctx context.Context, obj *domain.Airline) (*domain.Airline, error)
	Update(ctx context.Context, obj *domain.Airline, columns ...string) (*domain.Airline, error)
	FindByID(ctx context.Context, id string) (*domain.Airline, error)
	List(ctx context.Context, criteria repomodels.AirlineListCriteria) ([]domain.Airline, int, error)
}

type airlineRepository struct{ query query.DB[domain.Airline] }

func NewAirlineRepository(db query.AppIDB) AirlineRepository {
	return &airlineRepository{query: query.New[domain.Airline](db)}
}

func (r *airlineRepository) WithTx(tx query.AppIDB) AirlineRepository {
	return NewAirlineRepository(tx)
}

func (r *airlineRepository) Create(ctx context.Context, obj *domain.Airline) (*domain.Airline, error) {
	return r.query.Insert(ctx, obj)
}

func (r *airlineRepository) Update(ctx context.Context, obj *domain.Airline, columns ...string) (*domain.Airline, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *airlineRepository) FindByID(ctx context.Context, id string) (*domain.Airline, error) {
	obj := new(domain.Airline)
	err := r.query.NewSelect().Model(obj).Where("al.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *airlineRepository) List(ctx context.Context, criteria repomodels.AirlineListCriteria) ([]domain.Airline, int, error) {
	var items []domain.Airline
	q := r.query.NewSelect().Model(&items)

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(al.name) LIKE ?", pattern).WhereOr("lower(al.code) LIKE ?", pattern)
		})
	}
	if criteria.Status != nil {
		q = q.Where("al.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("al.name ASC, al.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
