package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
	"backend/pkg/domain/repomodels"
)

type SiteBuildRepository interface {
	WithTx(tx query.AppIDB) SiteBuildRepository
	Create(ctx context.Context, obj *domain.SiteBuild) (*domain.SiteBuild, error)
	// List returns builds newest first, with the Creator for manual builds.
	List(ctx context.Context, criteria repomodels.SiteBuildListCriteria) ([]domain.SiteBuild, int, error)
}

type siteBuildRepository struct{ query query.DB[domain.SiteBuild] }

func NewSiteBuildRepository(db query.AppIDB) SiteBuildRepository {
	return &siteBuildRepository{query: query.New[domain.SiteBuild](db)}
}

func (r *siteBuildRepository) WithTx(tx query.AppIDB) SiteBuildRepository {
	return NewSiteBuildRepository(tx)
}

func (r *siteBuildRepository) Create(ctx context.Context, obj *domain.SiteBuild) (*domain.SiteBuild, error) {
	return r.query.Insert(ctx, obj)
}

func (r *siteBuildRepository) List(ctx context.Context, criteria repomodels.SiteBuildListCriteria) ([]domain.SiteBuild, int, error) {
	var items []domain.SiteBuild
	q := r.query.NewSelect().Relation("Creator")
	if criteria.Trigger != "" {
		q = q.Where("sb.trigger = ?", criteria.Trigger)
	}
	total, err := q.
		OrderExpr("sb.create_date DESC, sb.id ASC").
		Limit(int64(criteria.PageSize)).
		Offset(int64((criteria.Page-1)*criteria.PageSize)).
		ScanAndCount(ctx, &items)
	if err != nil {
		return nil, 0, err
	}
	return items, int(total), nil
}
