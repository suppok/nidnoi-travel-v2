package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type AppConfigRepository interface {
	WithTx(tx query.AppIDB) AppConfigRepository
	Create(ctx context.Context, obj *domain.AppConfig) (*domain.AppConfig, error)
	Update(ctx context.Context, obj *domain.AppConfig, columns ...string) (*domain.AppConfig, error)
	FindByID(ctx context.Context, id string) (*domain.AppConfig, error)
	// FindByCode returns the row with this code, active or not; nil when there is none.
	FindByCode(ctx context.Context, code string) (*domain.AppConfig, error)
	// List returns every row (active or not), ordered by code.
	List(ctx context.Context) ([]domain.AppConfig, error)
}

type appConfigRepository struct{ query query.DB[domain.AppConfig] }

func NewAppConfigRepository(db query.AppIDB) AppConfigRepository {
	return &appConfigRepository{query: query.New[domain.AppConfig](db)}
}

func (r *appConfigRepository) WithTx(tx query.AppIDB) AppConfigRepository {
	return NewAppConfigRepository(tx)
}

func (r *appConfigRepository) Create(ctx context.Context, obj *domain.AppConfig) (*domain.AppConfig, error) {
	return r.query.Insert(ctx, obj)
}

func (r *appConfigRepository) Update(ctx context.Context, obj *domain.AppConfig, columns ...string) (*domain.AppConfig, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *appConfigRepository) FindByID(ctx context.Context, id string) (*domain.AppConfig, error) {
	return r.query.FindByID(ctx, id)
}

func (r *appConfigRepository) FindByCode(ctx context.Context, code string) (*domain.AppConfig, error) {
	obj := new(domain.AppConfig)
	err := r.query.NewSelect().Where("ac.code = ?", code).Limit(1).Scan(ctx, obj)
	return query.NotFoundAsNil(obj, err)
}

func (r *appConfigRepository) List(ctx context.Context) ([]domain.AppConfig, error) {
	var items []domain.AppConfig
	if err := r.query.NewSelect().OrderExpr("ac.code ASC").Scan(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
