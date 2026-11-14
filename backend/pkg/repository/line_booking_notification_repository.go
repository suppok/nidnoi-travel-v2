package repository

import (
	"context"
	"time"

	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type LineBookingNotificationRepository interface {
	WithTx(tx query.AppIDB) LineBookingNotificationRepository
	Create(ctx context.Context, obj *domain.LineBookingNotification) (*domain.LineBookingNotification, error)
	// List returns a user's notifications newest first, with Event.CreateUser and LineBooking.
	List(ctx context.Context, userID string, unreadOnly bool, page, pageSize int) ([]domain.LineBookingNotification, int, error)
	// MarkRead marks the user's unread notifications on one booking read.
	MarkRead(ctx context.Context, userID, lineBookingID string, at time.Time) error
}

type lineBookingNotificationRepository struct {
	query query.DB[domain.LineBookingNotification]
}

func NewLineBookingNotificationRepository(db query.AppIDB) LineBookingNotificationRepository {
	return &lineBookingNotificationRepository{query: query.New[domain.LineBookingNotification](db)}
}

func (r *lineBookingNotificationRepository) WithTx(tx query.AppIDB) LineBookingNotificationRepository {
	return NewLineBookingNotificationRepository(tx)
}

func (r *lineBookingNotificationRepository) Create(ctx context.Context, obj *domain.LineBookingNotification) (*domain.LineBookingNotification, error) {
	return r.query.Insert(ctx, obj)
}

func (r *lineBookingNotificationRepository) List(ctx context.Context, userID string, unreadOnly bool, page, pageSize int) ([]domain.LineBookingNotification, int, error) {
	var items []domain.LineBookingNotification
	q := r.query.NewSelect().Model(&items).
		Relation("Event").Relation("Event.CreateUser").Relation("LineBooking").
		Where("lbn.user_id = ?", userID)
	if unreadOnly {
		q = q.Where("lbn.read_at IS NULL")
	}
	total, err := q.OrderExpr("lbn.create_date DESC, lbn.id ASC").
		Limit(int64(pageSize)).
		Offset(int64((page - 1) * pageSize)).
		ScanAndCount(ctx)
	return items, int(total), err
}

func (r *lineBookingNotificationRepository) MarkRead(ctx context.Context, userID, lineBookingID string, at time.Time) error {
	_, err := r.query.NewUpdate().Model((*domain.LineBookingNotification)(nil)).
		Set("read_at = ?", at).
		Where("user_id = ?", userID).
		Where("line_booking_id = ?", lineBookingID).
		Where("read_at IS NULL").
		Exec(ctx)
	return err
}
