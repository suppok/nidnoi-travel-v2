package repository

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type AppUserRepository interface {
	WithTx(tx query.AppIDB) AppUserRepository
	Create(ctx context.Context, obj *domain.AppUser) (*domain.AppUser, error)
	Update(ctx context.Context, obj *domain.AppUser, columns ...string) (*domain.AppUser, error)
	FindByID(ctx context.Context, id string) (*domain.AppUser, error)
	FindByEmail(ctx context.Context, email string) (*domain.AppUser, error)
	List(ctx context.Context, criteria repomodels.AppUserListCriteria) ([]domain.AppUser, int, error)
	// ExistsWithRole reports whether any user (active or not) holds role.
	ExistsWithRole(ctx context.Context, role string) (bool, error)
}

type appUserRepository struct{ query query.DB[domain.AppUser] }

func NewAppUserRepository(db query.AppIDB) AppUserRepository {
	return &appUserRepository{query: query.New[domain.AppUser](db)}
}

func (r *appUserRepository) WithTx(tx query.AppIDB) AppUserRepository {
	return NewAppUserRepository(tx)
}

func (r *appUserRepository) Create(ctx context.Context, obj *domain.AppUser) (*domain.AppUser, error) {
	return r.query.Insert(ctx, obj)
}

func (r *appUserRepository) Update(ctx context.Context, obj *domain.AppUser, columns ...string) (*domain.AppUser, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *appUserRepository) FindByID(ctx context.Context, id string) (*domain.AppUser, error) {
	return r.query.FindByID(ctx, id)
}

func (r *appUserRepository) FindByEmail(ctx context.Context, email string) (*domain.AppUser, error) {
	obj := new(domain.AppUser)
	err := r.query.NewSelect().Where("au.email = ?", email).Limit(1).Scan(ctx, obj)
	return query.NotFoundAsNil(obj, err)
}

func (r *appUserRepository) List(ctx context.Context, criteria repomodels.AppUserListCriteria) ([]domain.AppUser, int, error) {
	var items []domain.AppUser
	q := r.query.NewSelect()

	if kw := strings.TrimSpace(criteria.Keyword); kw != "" {
		pattern := "%" + escapeLike(strings.ToLower(kw)) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(au.name) LIKE ?", pattern).WhereOr("au.email LIKE ?", pattern)
		})
	}
	if criteria.Role != "" {
		q = q.Where("au.roles @> ?", pgdialect.Array([]string{criteria.Role}))
	}
	if criteria.Status != nil {
		q = q.Where("au.status = ?", *criteria.Status)
	}

	total, err := q.
		OrderExpr("au.name ASC, au.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page-1)*criteria.PageSize)).
		ScanAndCount(ctx, &items)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}

func (r *appUserRepository) ExistsWithRole(ctx context.Context, role string) (bool, error) {
	return r.query.NewSelect().
		Where("au.roles @> ?", pgdialect.Array([]string{role})).
		Exists(ctx)
}

// escapeLike escapes LIKE wildcards in user input (the default escape character is backslash).
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
