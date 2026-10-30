package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type WholesaleContactRepository interface {
	WithTx(tx query.AppIDB) WholesaleContactRepository
	// ReplaceForWholesale deletes the wholesale's contacts and inserts contacts (may be empty).
	ReplaceForWholesale(ctx context.Context, wholesaleID string, contacts []domain.WholesaleContact) error
}

type wholesaleContactRepository struct{ db query.AppIDB }

func NewWholesaleContactRepository(db query.AppIDB) WholesaleContactRepository {
	return &wholesaleContactRepository{db: db}
}

func (r *wholesaleContactRepository) WithTx(tx query.AppIDB) WholesaleContactRepository {
	return NewWholesaleContactRepository(tx)
}

func (r *wholesaleContactRepository) ReplaceForWholesale(ctx context.Context, wholesaleID string, contacts []domain.WholesaleContact) error {
	if _, err := r.db.NewDelete().Model((*domain.WholesaleContact)(nil)).Where("wholesale_id = ?", wholesaleID).Exec(ctx); err != nil {
		return err
	}
	if len(contacts) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&contacts).Exec(ctx)
	return err
}
