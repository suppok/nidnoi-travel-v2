package repository

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type RefreshTokenRepository interface {
	WithTx(tx query.AppIDB) RefreshTokenRepository
	Create(ctx context.Context, obj *domain.RefreshToken) (*domain.RefreshToken, error)
	Update(ctx context.Context, obj *domain.RefreshToken, columns ...string) (*domain.RefreshToken, error)
	// FindByHashForUpdate locks the row so concurrent refreshes of the same token serialize.
	FindByHashForUpdate(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	// RevokeAllForUser revokes every unrevoked token of the user except exceptID (may be empty).
	RevokeAllForUser(ctx context.Context, userID, exceptID, updateBy string, at time.Time) error
}

type refreshTokenRepository struct{ query query.DB[domain.RefreshToken] }

func NewRefreshTokenRepository(db query.AppIDB) RefreshTokenRepository {
	return &refreshTokenRepository{query: query.New[domain.RefreshToken](db)}
}

func (r *refreshTokenRepository) WithTx(tx query.AppIDB) RefreshTokenRepository {
	return NewRefreshTokenRepository(tx)
}

func (r *refreshTokenRepository) Create(ctx context.Context, obj *domain.RefreshToken) (*domain.RefreshToken, error) {
	return r.query.Insert(ctx, obj)
}

func (r *refreshTokenRepository) Update(ctx context.Context, obj *domain.RefreshToken, columns ...string) (*domain.RefreshToken, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *refreshTokenRepository) FindByHashForUpdate(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	obj := new(domain.RefreshToken)
	err := r.query.NewSelect().Where("rt.token_hash = ?", tokenHash).For("UPDATE").Limit(1).Scan(ctx, obj)
	return query.NotFoundAsNil(obj, err)
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID, exceptID, updateBy string, at time.Time) error {
	q := r.query.NewUpdate().
		Set("revoked_at = ?", at).
		Set("update_date = ?", at).
		Set("update_by = ?", updateBy).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL")
	if exceptID != "" {
		q = q.Where("id <> ?", exceptID)
	}
	_, err := q.Exec(ctx)
	return err
}
