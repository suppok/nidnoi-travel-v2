// Package emailoutbox queues and delivers outgoing email through the email_outbox table
// (transactional outbox). Use cases call NewRow and insert it inside their own transaction; the
// Sender worker, started from app.go, delivers due rows through the mailer. Nothing is sent inline.
package emailoutbox

import (
	"time"

	"backend/pkg/constants/emailoutboxstate"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	"backend/pkg/modules/shared/services/notification"
)

// Addresses of one email. From is MAIL_FROM; Reply-To and Cc come from the caller's app config
// (e.g. WHOLESALE_EMAIL). They are stored on the row, so a later config change never alters a queued email.
type Addresses struct {
	From    string
	To      []string
	Cc      []string
	ReplyTo string
}

// NewRow builds a pending outbox row.
func NewRow(addr Addresses, lineBookingID, emailType string, email notification.Email, actorID string, now time.Time) *domain.EmailOutbox {
	return &domain.EmailOutbox{
		ID:            utils.NewID(),
		LineBookingID: lineBookingID,
		Type:          emailType,
		FromEmail:     addr.From,
		ToEmails:      append([]string{}, addr.To...),
		CcEmails:      append([]string{}, addr.Cc...),
		ReplyTo:       addr.ReplyTo,
		Subject:       email.Subject,
		BodyHTML:      email.HTML,
		BodyText:      email.Text,
		State:         emailoutboxstate.Pending,
		SendRound:     1,
		NextAttemptAt: now,
		Status:        true,
		CreateDate:    now,
		UpdateDate:    now,
		CreateBy:      actorID,
		UpdateBy:      actorID,
	}
}

// retryDelays are the waits after the 1st, 2nd, ... failed attempt; the last one repeats.
var retryDelays = []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour}

// Decision is what to do with a row after a send attempt.
type Decision struct {
	State         int
	NextAttemptAt time.Time
}

// Decide maps the result of attempt number attempts (1-based, already counted) to the row's next
// state: sent on success; failed on a permanent error or when maxAttempts is reached; otherwise
// pending again after the backoff.
func Decide(sendErr error, retryable bool, attempts, maxAttempts int, now time.Time) Decision {
	if sendErr == nil {
		return Decision{State: emailoutboxstate.Sent, NextAttemptAt: now}
	}
	if !retryable || attempts >= maxAttempts {
		return Decision{State: emailoutboxstate.Failed, NextAttemptAt: now}
	}
	i := min(max(attempts-1, 0), len(retryDelays)-1)
	return Decision{State: emailoutboxstate.Pending, NextAttemptAt: now.Add(retryDelays[i])}
}
