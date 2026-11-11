package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type SalesFinanceRepository interface {
	WithTx(tx query.AppIDB) SalesFinanceRepository
	// ListBySales returns the sales user's finance links with their FinanceUser, by finance name.
	// Links to deactivated users are included; callers filter on FinanceUser.Status.
	ListBySales(ctx context.Context, salesUserID string) ([]domain.SalesFinance, error)
	// ReplaceForSales deletes the sales user's links and inserts rows.
	ReplaceForSales(ctx context.Context, salesUserID string, rows []domain.SalesFinance) error
}

type salesFinanceRepository struct{ db query.AppIDB }

func NewSalesFinanceRepository(db query.AppIDB) SalesFinanceRepository {
	return &salesFinanceRepository{db: db}
}

func (r *salesFinanceRepository) WithTx(tx query.AppIDB) SalesFinanceRepository {
	return NewSalesFinanceRepository(tx)
}

func (r *salesFinanceRepository) ListBySales(ctx context.Context, salesUserID string) ([]domain.SalesFinance, error) {
	var items []domain.SalesFinance
	err := r.db.NewSelect().Model(&items).Relation("FinanceUser").
		Where("sf.sales_user_id = ?", salesUserID).
		OrderExpr("finance_user.name ASC, sf.id ASC").
		Scan(ctx)
	return items, err
}

func (r *salesFinanceRepository) ReplaceForSales(ctx context.Context, salesUserID string, rows []domain.SalesFinance) error {
	if _, err := r.db.NewDelete().Model((*domain.SalesFinance)(nil)).Where("sales_user_id = ?", salesUserID).Exec(ctx); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&rows).Exec(ctx)
	return err
}
