package repository

import (
	"context"
	"strings"

	"time"

	"github.com/uptrace/bun"

	"backend/pkg/constants/tourprogramstate"
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
	// FindByID loads the tour with its Wholesale, Airline, countries, cities (each with its Country) and images,
	// each in display order.
	FindByID(ctx context.Context, id string) (*domain.TourProgram, error)
	// List returns tours with their Wholesale and countries, most recently updated first.
	List(ctx context.Context, criteria repomodels.TourProgramListCriteria) ([]domain.TourProgram, int, error)

	// --- Public (customer site): active tours in a public state only. ---

	// ListPublic returns tours with their Airline, countries and cities (cities with their Country),
	// ordered by code.
	ListPublic(ctx context.Context, criteria repomodels.TourProgramPublicListCriteria) ([]domain.TourProgram, int, error)
	// FindPublicByCode loads the tour like ListPublic plus its gallery images; nil when the code is
	// unknown, deleted or not public.
	FindPublicByCode(ctx context.Context, code string) (*domain.TourProgram, error)
	// FindPublicStateByCode reads only id, code, name and state (periods endpoint, booking requests).
	FindPublicStateByCode(ctx context.Context, code string) (*domain.TourProgram, error)
	// PublicCountryStats / PublicCityStats count public tours per country/city and their lowest
	// upcoming period price (start_date > departAfter). Rows without public tours are absent.
	PublicCountryStats(ctx context.Context, departAfter time.Time) ([]repomodels.TaxonomyPublicStats, error)
	PublicCityStats(ctx context.Context, departAfter time.Time) ([]repomodels.TaxonomyPublicStats, error)
	// CountPublicByCountry / CountPublicByCity count public tours linked to the country/city.
	CountPublicByCountry(ctx context.Context, countryID string) (int, error)
	CountPublicByCity(ctx context.Context, cityID string) (int, error)
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
		Relation("Airline").
		Relation("CountryLinks", orderedCountryLinks).
		Relation("CountryLinks.Country").
		Relation("CityLinks", orderedCityLinks).
		Relation("CityLinks.City").
		Relation("CityLinks.City.Country").
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

// wherePublic hard-codes the customer-site visibility rule: active and in a public state.
func wherePublic(q *bun.SelectQuery) *bun.SelectQuery {
	return q.Where("tp.status = TRUE").Where("tp.state IN (?)", bun.In(tourprogramstate.Public))
}

// publicRelations loads what public DTOs render. Soft-deleted countries/cities are dropped by the use case.
func publicRelations(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		Relation("Airline").
		Relation("CountryLinks", orderedCountryLinks).
		Relation("CountryLinks.Country").
		Relation("CityLinks", orderedCityLinks).
		Relation("CityLinks.City").
		Relation("CityLinks.City.Country")
}

func (r *tourProgramRepository) ListPublic(ctx context.Context, criteria repomodels.TourProgramPublicListCriteria) ([]domain.TourProgram, int, error) {
	var items []domain.TourProgram
	q := wherePublic(publicRelations(r.query.NewSelect().Model(&items)))

	if criteria.CountrySlug != "" {
		q = q.Where(`EXISTS (SELECT 1 FROM tour_program_country x JOIN country c ON c.id = x.country_id
			WHERE x.tour_program_id = tp.id AND c.slug = ? AND c.status = TRUE)`, criteria.CountrySlug)
	}
	if criteria.CitySlug != "" {
		q = q.Where(`EXISTS (SELECT 1 FROM tour_program_city x JOIN city c ON c.id = x.city_id
			WHERE x.tour_program_id = tp.id AND c.slug = ? AND c.status = TRUE)`, criteria.CitySlug)
	}
	if criteria.DateFrom.Valid || criteria.DateTo.Valid {
		sub := r.db.NewSelect().TableExpr("tour_program_period AS p").ColumnExpr("1").
			Where("p.tour_program_id = tp.id").
			Where("p.status = TRUE").
			Where("p.start_date > ?", criteria.DepartAfter.Format(time.DateOnly))
		if criteria.DateFrom.Valid {
			sub = sub.Where("p.start_date >= ?", criteria.DateFrom.Time.Format(time.DateOnly))
		}
		if criteria.DateTo.Valid {
			sub = sub.Where("p.start_date <= ?", criteria.DateTo.Time.Format(time.DateOnly))
		}
		q = q.Where("EXISTS (?)", sub)
	}

	total, err := q.
		OrderExpr("tp.code ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}

func (r *tourProgramRepository) FindPublicByCode(ctx context.Context, code string) (*domain.TourProgram, error) {
	obj := new(domain.TourProgram)
	err := wherePublic(publicRelations(r.query.NewSelect().Model(obj))).
		Relation("Images", orderedImages).
		Where("tp.code = ?", code).
		Limit(1).
		Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *tourProgramRepository) FindPublicStateByCode(ctx context.Context, code string) (*domain.TourProgram, error) {
	obj := new(domain.TourProgram)
	err := wherePublic(r.query.NewSelect().Model(obj).Column("id", "code", "name", "state")).
		Where("tp.code = ?", code).
		Limit(1).
		Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

// publicStats aggregates public tours per taxonomy id through a link table
// (tour_program_country.country_id or tour_program_city.city_id).
func (r *tourProgramRepository) publicStats(ctx context.Context, linkTable, idColumn string, departAfter time.Time) ([]repomodels.TaxonomyPublicStats, error) {
	var rows []repomodels.TaxonomyPublicStats
	err := r.db.NewSelect().
		TableExpr("tour_program AS tp").
		Join("JOIN ? AS x ON x.tour_program_id = tp.id", bun.Ident(linkTable)).
		Join("LEFT JOIN tour_program_period AS p ON p.tour_program_id = tp.id AND p.status = TRUE AND p.start_date > ?",
			departAfter.Format(time.DateOnly)).
		ColumnExpr("x.? AS id", bun.Ident(idColumn)).
		ColumnExpr("count(DISTINCT tp.id) AS tour_count").
		ColumnExpr("min(p.price) AS starting_price").
		Where("tp.status = TRUE").
		Where("tp.state IN (?)", bun.In(tourprogramstate.Public)).
		GroupExpr("x.?", bun.Ident(idColumn)).
		Scan(ctx, &rows)
	return rows, err
}

func (r *tourProgramRepository) PublicCountryStats(ctx context.Context, departAfter time.Time) ([]repomodels.TaxonomyPublicStats, error) {
	return r.publicStats(ctx, "tour_program_country", "country_id", departAfter)
}

func (r *tourProgramRepository) PublicCityStats(ctx context.Context, departAfter time.Time) ([]repomodels.TaxonomyPublicStats, error) {
	return r.publicStats(ctx, "tour_program_city", "city_id", departAfter)
}

func (r *tourProgramRepository) CountPublicByCountry(ctx context.Context, countryID string) (int, error) {
	n, err := wherePublic(r.query.NewSelect()).
		Where("EXISTS (SELECT 1 FROM tour_program_country x WHERE x.tour_program_id = tp.id AND x.country_id = ?)", countryID).
		Count(ctx)
	return int(n), err
}

func (r *tourProgramRepository) CountPublicByCity(ctx context.Context, cityID string) (int, error) {
	n, err := wherePublic(r.query.NewSelect()).
		Where("EXISTS (SELECT 1 FROM tour_program_city x WHERE x.tour_program_id = tp.id AND x.city_id = ?)", cityID).
		Count(ctx)
	return int(n), err
}
