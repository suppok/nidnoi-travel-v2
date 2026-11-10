package repository

import (
	"context"
	"strings"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type LineContactRepository interface {
	WithTx(tx query.AppIDB) LineContactRepository
	// CreateIfAbsent inserts obj unless a row with the same line_user_id exists. Safe under concurrent webhooks.
	CreateIfAbsent(ctx context.Context, obj *domain.LineContact) error
	Update(ctx context.Context, obj *domain.LineContact, columns ...string) (*domain.LineContact, error)
	FindByID(ctx context.Context, id string) (*domain.LineContact, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (*domain.LineContact, error)
	// List returns contacts with the most recent chat first.
	List(ctx context.Context, criteria repomodels.LineContactListCriteria) ([]domain.LineContact, int, error)
}

type lineContactRepository struct{ query query.DB[domain.LineContact] }

func NewLineContactRepository(db query.AppIDB) LineContactRepository {
	return &lineContactRepository{query: query.New[domain.LineContact](db)}
}

func (r *lineContactRepository) WithTx(tx query.AppIDB) LineContactRepository {
	return NewLineContactRepository(tx)
}

func (r *lineContactRepository) CreateIfAbsent(ctx context.Context, obj *domain.LineContact) error {
	_, err := r.query.NewInsert().Model(obj).On("CONFLICT (line_user_id) DO NOTHING").Exec(ctx)
	return err
}

func (r *lineContactRepository) Update(ctx context.Context, obj *domain.LineContact, columns ...string) (*domain.LineContact, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *lineContactRepository) FindByID(ctx context.Context, id string) (*domain.LineContact, error) {
	return r.query.FindByID(ctx, id)
}

func (r *lineContactRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*domain.LineContact, error) {
	obj := new(domain.LineContact)
	err := r.query.NewSelect().Model(obj).Where("lc.line_user_id = ?", lineUserID).Limit(1).Scan(ctx)
	return query.NotFoundAsNil(obj, err)
}

func (r *lineContactRepository) List(ctx context.Context, criteria repomodels.LineContactListCriteria) ([]domain.LineContact, int, error) {
	var items []domain.LineContact
	q := r.query.NewSelect().Model(&items).Where("lc.status = TRUE")

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		q = q.Where("lower(lc.display_name) LIKE ?", "%"+escapeLike(strings.ToLower(kw))+"%")
	}
	if criteria.Following != nil {
		q = q.Where("lc.following = ?", *criteria.Following)
	}

	total, err := q.
		OrderExpr("lc.last_message_at DESC NULLS LAST, lc.display_name ASC, lc.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page - 1) * criteria.PageSize)).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
