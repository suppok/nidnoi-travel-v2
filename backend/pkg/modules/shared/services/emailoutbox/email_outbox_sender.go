package emailoutbox

import (
	"context"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/emailoutboxstate"
	"backend/pkg/core/config"
	"backend/pkg/domain"
	"backend/pkg/modules/shared/client/mailer"
	"backend/pkg/modules/shared/client/objectstorage"
	"backend/pkg/repository"
)

// lease is how long a claimed row is hidden from other ticks while it is being sent. It must be
// well above mail.httpTimeout; a row whose send never finishes (crash) is retried after it, and the
// provider's idempotency key stops a double send.
const lease = 5 * time.Minute

// maxBatchesPerTick bounds one tick when a burst is queued.
const maxBatchesPerTick = 10

// Sender delivers due outbox rows. It runs in-process; the API runs as exactly one instance, but the
// claim (FOR UPDATE SKIP LOCKED) stays safe if two ever overlap.
type Sender struct {
	repo   repository.EmailOutboxRepository
	mailer mailer.Mailer
	images *inlineImages
	log    *zap.Logger
	cfg    config.MailOutboxConfig
	now    func() time.Time
}

// NewSender: storage serves the uploaded inline images (the email contact card).
func NewSender(repo repository.EmailOutboxRepository, m mailer.Mailer, storage objectstorage.ObjectStorage, log *zap.Logger, cfg config.MailOutboxConfig) *Sender {
	return &Sender{repo: repo, mailer: m, images: newInlineImages(storage), log: log, cfg: cfg, now: time.Now}
}

// Run sends due emails every tick until ctx is cancelled. Unsent rows simply wait in the table.
func (s *Sender) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.TickInterval)
	defer ticker.Stop()
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Sender) tick(ctx context.Context) {
	for range maxBatchesPerTick {
		if ctx.Err() != nil {
			return
		}
		now := s.now()
		rows, err := s.repo.ClaimDue(ctx, now, now.Add(lease), s.cfg.BatchSize)
		if err != nil {
			if ctx.Err() == nil {
				s.log.Error("email outbox: claim due rows", zap.Error(err))
			}
			return
		}
		for i := range rows {
			s.send(ctx, &rows[i])
		}
		if len(rows) < s.cfg.BatchSize {
			return
		}
	}
}

func (s *Sender) send(ctx context.Context, row *domain.EmailOutbox) {
	// Inline images are attached at send time from the cid: references in the HTML, so the outbox row
	// stays small. A card that can't be loaded fails this attempt (retryable), like a provider error.
	var id string
	attachments, sendErr := s.images.attachments(ctx, row.BodyHTML)
	if sendErr == nil {
		id, sendErr = s.mailer.Send(ctx, mailer.Message{
			From:           row.FromEmail,
			To:             row.ToEmails,
			Cc:             row.CcEmails,
			ReplyTo:        row.ReplyTo,
			Subject:        row.Subject,
			HTML:           row.BodyHTML,
			Text:           row.BodyText,
			IdempotencyKey: row.IdempotencyKey(),
			Attachments:    attachments,
		})
	}
	now := s.now()
	d := Decide(sendErr, mailer.IsRetryable(sendErr), row.Attempts, s.cfg.MaxAttempts, now)
	outcome := repository.EmailOutboxOutcome{State: d.State, NextAttemptAt: d.NextAttemptAt, Now: now}
	fields := []zap.Field{
		zap.String("outboxId", row.ID),
		zap.String("lineBookingId", row.LineBookingID),
		zap.String("type", row.Type),
		zap.Int("attempt", row.Attempts),
		zap.Int("round", row.SendRound),
	}
	if sendErr == nil {
		outcome.ProviderMessageID = id
		outcome.SentAt = &now
		s.log.Info("email sent", append(fields, zap.String("providerMessageId", id))...)
	} else {
		outcome.LastError = sendErr.Error()
		fields = append(fields, zap.Error(sendErr))
		if d.State == emailoutboxstate.Failed {
			s.log.Error("email failed: giving up", fields...)
		} else {
			s.log.Error("email send failed: will retry", append(fields, zap.Time("nextAttemptAt", d.NextAttemptAt))...)
		}
	}
	// Record the outcome even when shutdown cancelled ctx mid-send.
	if err := s.repo.Finish(context.WithoutCancel(ctx), row.ID, row.SendRound, outcome); err != nil {
		s.log.Error("email outbox: record outcome", append(fields, zap.NamedError("recordError", err))...)
	}
}
