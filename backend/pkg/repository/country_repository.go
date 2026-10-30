package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type CountryRepository interface {
	WithTx(tx query.AppIDB) CountryRepository
	Create(ctx context.Context, obj *domain.Country) (*domain.Country, error)
	Update(ctx context.Context, obj *domain.Country, columns ...string) (*domain.Country, error)
	FindByID(ctx context.Context, id string) (*domain.Country, error)
	List(ctx context.Context, criteria repomodels.CountryListCriteria) ([]domain.Country, int, error)
}

type countryRepository struct{ query query.DB[domain.Country] }

func NewCountryRepository(db query.AppIDB) CountryRepository {
	return &countryRepository{query: query.New[domain.Country](db)}
}

func (r *countryRepository) WithTx(tx query.AppIDB) CountryRepository {
	return NewCountryRepository(tx)
}

func (r *countryRepository) Create(ctx context.Context, obj *domain.Country) (*domain.Country, error) {
	return r.query.Insert(ctx, obj)
}

func (r *countryRepository) Update(ctx context.Context, obj *domain.Country, columns ...string) (*domain.Country, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *countryRepository) FindByID(ctx context.Context, id string) (*domain.Country, error) {
	return r.query.FindByID(ctx, id)
}

func (r *countryRepository) List(ctx context.Context, criteria repomodels.CountryListCriteria) ([]domain.Country, int, error) {
	var items []domain.Country
	q := r.query.NewSelect()

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("co.name_th LIKE ?", pattern).WhereOr("lower(co.name_en) LIKE ?", pattern).WhereOr("co.slug LIKE ?", pattern)
		})
	}
	if criteria.Status != nil {
		q = q.Where("co.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("co.sort_order ASC, co.name_th ASC, co.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page-1)*criteria.PageSize)).
		ScanAndCount(ctx, &items)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
