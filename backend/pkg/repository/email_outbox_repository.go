package repository

import (
	"context"
	"time"

	"backend/pkg/constants/emailoutboxstate"
	"backend/pkg/core/database/query"
	"backend/pkg/domain"
)

type EmailOutboxRepository interface {
	WithTx(tx query.AppIDB) EmailOutboxRepository
	// CreateIfAbsent inserts obj unless the booking already has an email of its type
	// (ON CONFLICT DO NOTHING); created is false for a duplicate.
	CreateIfAbsent(ctx context.Context, obj *domain.EmailOutbox) (created bool, err error)
	Update(ctx context.Context, obj *domain.EmailOutbox, columns ...string) (*domain.EmailOutbox, error)
	FindByBooking(ctx context.Context, lineBookingID, emailType string) (*domain.EmailOutbox, error)
	// ClaimDue takes up to limit active pending rows due at now, oldest first, in one statement:
	// it counts the attempt and pushes next_attempt_at to leaseUntil, so the rows are not picked again
	// while they are being sent (FOR UPDATE SKIP LOCKED). A row whose send never finishes (crash) is
	// picked again after the lease.
	ClaimDue(ctx context.Context, now, leaseUntil time.Time, limit int) ([]domain.EmailOutbox, error)
	// Finish stores the outcome of a send of round sendRound. It changes nothing when the row was
	// resent (new round) or is no longer pending meanwhile.
	Finish(ctx context.Context, id string, sendRound int, outcome EmailOutboxOutcome) error
}

// EmailOutboxOutcome is the result of one send.
type EmailOutboxOutcome struct {
	State             int // emailoutbox.State*
	NextAttemptAt     time.Time
	LastError         string
	ProviderMessageID string
	SentAt            *time.Time
	Now               time.Time
}

type emailOutboxRepository struct{ query query.DB[domain.EmailOutbox] }

func NewEmailOutboxRepository(db query.AppIDB) EmailOutboxRepository {
	return &emailOutboxRepository{query: query.New[domain.EmailOutbox](db)}
}

func (r *emailOutboxRepository) WithTx(tx query.AppIDB) EmailOutboxRepository {
	return NewEmailOutboxRepository(tx)
}

func (r *emailOutboxRepository) CreateIfAbsent(ctx context.Context, obj *domain.EmailOutbox) (bool, error) {
	res, err := r.query.NewInsert().Model(obj).On("CONFLICT (line_booking_id, type) DO NOTHING").Exec(ctx)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *emailOutboxRepository) Update(ctx context.Context, obj *domain.EmailOutbox, columns ...string) (*domain.EmailOutbox, error) {
	return r.query.Update(ctx, obj, columns...)
}

func (r *emailOutboxRepository) FindByBooking(ctx context.Context, lineBookingID, emailType string) (*domain.EmailOutbox, error) {
	obj := new(domain.EmailOutbox)
	err := r.query.NewSelect().
		Where("eo.line_booking_id = ?", lineBookingID).
		Where("eo.type = ?", emailType).
		Limit(1).
		Scan(ctx, obj)
	return query.NotFoundAsNil(obj, err)
}

func (r *emailOutboxRepository) ClaimDue(ctx context.Context, now, leaseUntil time.Time, limit int) ([]domain.EmailOutbox, error) {
	due := r.query.NewSelect().
		Column("eo.id").
		Where("eo.state = ?", emailoutboxstate.Pending).
		Where("eo.status = TRUE").
		Where("eo.next_attempt_at <= ?", now).
		OrderExpr("eo.next_attempt_at ASC").
		Limit(int64(limit)).
		For("UPDATE SKIP LOCKED")
	var rows []domain.EmailOutbox
	err := r.query.NewUpdate().
		Set("attempts = eo.attempts + 1").
		Set("next_attempt_at = ?", leaseUntil).
		Set("update_date = ?", now).
		Where("eo.id IN (?)", due).
		Returning("*").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *emailOutboxRepository) Finish(ctx context.Context, id string, sendRound int, o EmailOutboxOutcome) error {
	q := r.query.NewUpdate().
		Set("state = ?", o.State).
		Set("next_attempt_at = ?", o.NextAttemptAt).
		Set("last_error = ?", o.LastError).
		Set("update_date = ?", o.Now).
		Where("eo.id = ?", id).
		Where("eo.send_round = ?", sendRound).
		Where("eo.state = ?", emailoutboxstate.Pending)
	if o.ProviderMessageID != "" {
		q = q.Set("provider_message_id = ?", o.ProviderMessageID)
	}
	if o.SentAt != nil {
		q = q.Set("sent_at = ?", *o.SentAt)
	}
	_, err := q.Exec(ctx)
	return err
}
