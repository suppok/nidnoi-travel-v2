package repository

import (
	"context"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type LineBookingEventRepository interface {
	WithTx(tx query.AppIDB) LineBookingEventRepository
	Create(ctx context.Context, obj *domain.LineBookingEvent) (*domain.LineBookingEvent, error)
	// ListByBooking returns a booking's history oldest first, with CreateUser.
	ListByBooking(ctx context.Context, lineBookingID string) ([]domain.LineBookingEvent, error)
}

type lineBookingEventRepository struct {
	query query.DB[domain.LineBookingEvent]
}

func NewLineBookingEventRepository(db query.AppIDB) LineBookingEventRepository {
	return &lineBookingEventRepository{query: query.New[domain.LineBookingEvent](db)}
}

func (r *lineBookingEventRepository) WithTx(tx query.AppIDB) LineBookingEventRepository {
	return NewLineBookingEventRepository(tx)
}

func (r *lineBookingEventRepository) Create(ctx context.Context, obj *domain.LineBookingEvent) (*domain.LineBookingEvent, error) {
	return r.query.Insert(ctx, obj)
}

func (r *lineBookingEventRepository) ListByBooking(ctx context.Context, lineBookingID string) ([]domain.LineBookingEvent, error) {
	var items []domain.LineBookingEvent
	err := r.query.NewSelect().Model(&items).Relation("CreateUser").
		Where("lbe.line_booking_id = ?", lineBookingID).
		OrderExpr("lbe.create_date ASC, lbe.id ASC").
		Scan(ctx)
	return items, err
}
