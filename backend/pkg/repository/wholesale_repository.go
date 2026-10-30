package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type WholesaleRepository interface {
	WithTx(tx query.AppIDB) WholesaleRepository
	Create(ctx context.Context, obj *domain.Wholesale) (*domain.Wholesale, error)
	Update(ctx context.Context, obj *domain.Wholesale, columns ...string) (*domain.Wholesale, error)
	// FindByID loads the wholesale with its Contacts in display order.
	FindByID(ctx context.Context, id string) (*domain.Wholesale, error)
	// List returns wholesales with their Contacts in display order.
	List(ctx context.Context, criteria repomodels.WholesaleListCriteria) ([]domain.Wholesale, int, error)
}

type wholesaleRepository struct{ query query.DB[domain.Wholesale] }

func NewWholesaleRepository(db query.AppIDB) WholesaleRepository {
	return &wholesaleRepository{query: query.New[domain.Wholesale](db)}
}

func (r *wholesaleRepository) WithTx(tx query.AppIDB) WholesaleRepository {
	return NewWholesaleRepository(tx)
}

func (r *wholesaleRepository) Create(ctx context.Context, obj *domain.Wholesale) (*domain.Wholesale, error) {
	return r.query.Insert(ctx, obj)
}

func (r *wholesaleRepository) Update(ctx context.Context, obj *domain.Wholesale, columns ...string) (*domain.Wholesale, error) {
	return r.query.Update(ctx, obj, columns...)
}

func orderedContacts(q *bun.SelectQuery) *bun.SelectQuery {
	return q.OrderExpr("wc.sort_order ASC, wc.id ASC")
}

func (r *wholesaleRepository) FindByID(ctx context.Context, id string) (*domain.Wholesale, error) {
	obj := new(domain.Wholesale)
	// Has-many relations need the destination as Model, not a Scan argument.
	err := r.query.NewSelect().Model(obj).Relation("Contacts", orderedContacts).Where("ws.id = ?", id).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *wholesaleRepository) List(ctx context.Context, criteria repomodels.WholesaleListCriteria) ([]domain.Wholesale, int, error) {
	var items []domain.Wholesale
	q := r.query.NewSelect().Model(&items).Relation("Contacts", orderedContacts)

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(ws.name) LIKE ?", pattern).
				WhereOr("lower(ws.code) LIKE ?", pattern).
				WhereOr("EXISTS (SELECT 1 FROM wholesale_contact c WHERE c.wholesale_id = ws.id AND (lower(c.name) LIKE ? OR c.phone LIKE ?))", pattern, pattern)
		})
	}
	if criteria.State != 0 {
		q = q.Where("ws.state = ?", criteria.State)
	}
	if criteria.Status != nil {
		q = q.Where("ws.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("ws.name ASC, ws.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
