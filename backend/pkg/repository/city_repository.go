package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type CityRepository interface {
	WithTx(tx query.AppIDB) CityRepository
	Create(ctx context.Context, obj *domain.City) (*domain.City, error)
	Update(ctx context.Context, obj *domain.City, columns ...string) (*domain.City, error)
	// FindByID loads the city with its Country.
	FindByID(ctx context.Context, id string) (*domain.City, error)
	// List returns cities with their Country.
	List(ctx context.Context, criteria repomodels.CityListCriteria) ([]domain.City, int, error)
	// FindByIDs returns the rows among ids that exist, in no particular order.
	FindByIDs(ctx context.Context, ids []string) ([]domain.City, error)
}

type cityRepository struct{ query query.DB[domain.City] }

func NewCityRepository(db query.AppIDB) CityRepository {
	return &cityRepository{query: query.New[domain.City](db)}
}

func (r *cityRepository) WithTx(tx query.AppIDB) CityRepository {
	return NewCityRepository(tx)
}

func (r *cityRepository) Create(ctx context.Context, obj *domain.City) (*domain.City, error) {
	return r.query.Insert(ctx, obj)
}

func (r *cityRepository) Update(ctx context.Context, obj *domain.City, columns ...string) (*domain.City, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *cityRepository) FindByID(ctx context.Context, id string) (*domain.City, error) {
	obj := new(domain.City)
	err := r.query.NewSelect().Relation("Country").Where("ci.id = ?", id).Limit(1).Scan(ctx, obj)
	return query.NotFoundAsNil(obj, err)
}

func (r *cityRepository) List(ctx context.Context, criteria repomodels.CityListCriteria) ([]domain.City, int, error) {
	var items []domain.City
	q := r.query.NewSelect().Relation("Country")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("ci.name_th LIKE ?", pattern).WhereOr("lower(ci.name_en) LIKE ?", pattern).WhereOr("ci.slug LIKE ?", pattern)
		})
	}
	if criteria.CountryID != "" {
		q = q.Where("ci.country_id = ?", criteria.CountryID)
	}
	if criteria.Status != nil {
		q = q.Where("ci.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("ci.sort_order ASC, ci.name_th ASC, ci.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page-1)*criteria.PageSize)).
		ScanAndCount(ctx, &items)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}

func (r *cityRepository) FindByIDs(ctx context.Context, ids []string) ([]domain.City, error) {
	var items []domain.City
	if len(ids) == 0 {
		return items, nil
	}
	err := r.query.NewSelect().Where("ci.id IN (?)", bun.In(ids)).Scan(ctx, &items)
	return items, err
}
