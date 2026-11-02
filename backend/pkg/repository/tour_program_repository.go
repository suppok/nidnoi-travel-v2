package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type TourProgramRepository interface {
	WithTx(tx query.AppIDB) TourProgramRepository
	// NextCodeNumber returns the next value of tour_code_seq. The use case formats the code.
	NextCodeNumber(ctx context.Context) (int64, error)
	Create(ctx context.Context, obj *domain.TourProgram) (*domain.TourProgram, error)
	Update(ctx context.Context, obj *domain.TourProgram, columns ...string) (*domain.TourProgram, error)
	// FindByID loads the tour with its Wholesale, countries, cities (with their Country) and images,
	// each in display order.
	FindByID(ctx context.Context, id string) (*domain.TourProgram, error)
	// List returns tours with their Wholesale and countries, most recently updated first.
	List(ctx context.Context, criteria repomodels.TourProgramListCriteria) ([]domain.TourProgram, int, error)
}

type tourProgramRepository struct {
	db    query.AppIDB
	query query.DB[domain.TourProgram]
}

func NewTourProgramRepository(db query.AppIDB) TourProgramRepository {
	return &tourProgramRepository{db: db, query: query.New[domain.TourProgram](db)}
}

func (r *tourProgramRepository) WithTx(tx query.AppIDB) TourProgramRepository {
	return NewTourProgramRepository(tx)
}

func (r *tourProgramRepository) NextCodeNumber(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.NewSelect().ColumnExpr("nextval('tour_code_seq')").Scan(ctx, &n)
	return n, err
}

func (r *tourProgramRepository) Create(ctx context.Context, obj *domain.TourProgram) (*domain.TourProgram, error) {
	return r.query.Insert(ctx, obj)
}

func (r *tourProgramRepository) Update(ctx context.Context, obj *domain.TourProgram, columns ...string) (*domain.TourProgram, error) {
	return r.query.Update(ctx, obj, columns...)
}

func orderedCountryLinks(q *bun.SelectQuery) *bun.SelectQuery {
	return q.OrderExpr("tpco.sort_order ASC")
}

func orderedCityLinks(q *bun.SelectQuery) *bun.SelectQuery {
	return q.OrderExpr("tpci.sort_order ASC")
}

func orderedImages(q *bun.SelectQuery) *bun.SelectQuery {
	return q.OrderExpr("tpi.sort_order ASC, tpi.id ASC")
}

func (r *tourProgramRepository) FindByID(ctx context.Context, id string) (*domain.TourProgram, error) {
	obj := new(domain.TourProgram)
	// Has-many relations need the destination as Model, not a Scan argument.
	err := r.query.NewSelect().Model(obj).
		Relation("Wholesale").
		Relation("CountryLinks", orderedCountryLinks).
		Relation("CountryLinks.Country").
		Relation("CityLinks", orderedCityLinks).
		Relation("CityLinks.City").
		Relation("Images", orderedImages).
		Where("tp.id = ?", id).
		Limit(1).
		Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *tourProgramRepository) List(ctx context.Context, criteria repomodels.TourProgramListCriteria) ([]domain.TourProgram, int, error) {
	var items []domain.TourProgram
	q := r.query.NewSelect().Model(&items).
		Relation("Wholesale").
		Relation("CountryLinks", orderedCountryLinks).
		Relation("CountryLinks.Country")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(tp.name) LIKE ?", pattern).
				WhereOr("lower(tp.code) LIKE ?", pattern).
				WhereOr("lower(tp.wholesale_tour_code) LIKE ?", pattern)
		})
	}
	if criteria.CountryID != "" {
		q = q.Where("EXISTS (SELECT 1 FROM tour_program_country x WHERE x.tour_program_id = tp.id AND x.country_id = ?)", criteria.CountryID)
	}
	if criteria.CityID != "" {
		q = q.Where("EXISTS (SELECT 1 FROM tour_program_city x WHERE x.tour_program_id = tp.id AND x.city_id = ?)", criteria.CityID)
	}
	if criteria.WholesaleID != "" {
		q = q.Where("tp.wholesale_id = ?", criteria.WholesaleID)
	}
	if criteria.State != 0 {
		q = q.Where("tp.state = ?", criteria.State)
	}
	if criteria.Status != nil {
		q = q.Where("tp.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("tp.update_date DESC, tp.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
