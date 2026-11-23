// Package mailer sends transactional email. Use cases never call it: they queue rows in
// email_outbox and the outbox sender (services/emailoutbox) delivers them.
package mailer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"backend/pkg/core/config"
)

// Message is one email. Addresses are plain ("a@b.com") except From, which may carry a display
// name ("nidnoi travel <booking@example.com>").
type Message struct {
	From    string
	To      []string
	Cc      []string
	ReplyTo string
	Subject string
	HTML    string
	Text    string
	// IdempotencyKey makes a retried send of the same email a no-op at the provider.
	IdempotencyKey string
	// Inline images, referenced from HTML as "cid:<ContentID>".
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	ContentID   string
	Content     []byte
}

type Mailer interface {
	// Send delivers msg and returns the provider's message id. Failures are *SendError.
	Send(ctx context.Context, msg Message) (providerMessageID string, err error)
}

// SendError is a failed send. Retryable: rate limits, provider 5xx and network errors; any other
// rejection (bad address, invalid request) is permanent. Message never carries the API key.
type SendError struct {
	Retryable bool
	Status    int // HTTP status, 0 for network errors
	Message   string
}

func (e *SendError) Error() string {
	if e.Status == 0 {
		return "mail: " + e.Message
	}
	return fmt.Sprintf("mail: status %d: %s", e.Status, e.Message)
}

// IsRetryable reports whether err is worth retrying later. Unknown errors are retried.
func IsRetryable(err error) bool {
	var se *SendError
	if errors.As(err, &se) {
		return se.Retryable
	}
	return true
}

// New returns the dry-run mailer when cfg.DryRun is set, else the Resend client.
func New(cfg config.MailConfig, log *zap.Logger) Mailer {
	if cfg.DryRun {
		return NewDryRun(log)
	}
	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return NewResend(cfg.APIKey, resendEndpoint, timeout)
}
