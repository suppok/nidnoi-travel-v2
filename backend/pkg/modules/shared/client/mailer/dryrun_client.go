package mailer

import (
	"context"

	"go.uber.org/zap"
)

type dryRunClient struct{ log *zap.Logger }

// NewDryRun returns a mailer that only logs (MAIL_DRY_RUN=true). The body is never logged.
func NewDryRun(log *zap.Logger) Mailer {
	return &dryRunClient{log: log}
}

func (c *dryRunClient) Send(_ context.Context, msg Message) (string, error) {
	c.log.Info("mail dry run: not sent",
		zap.String("from", msg.From),
		zap.Strings("to", msg.To),
		zap.Strings("cc", msg.Cc),
		zap.String("replyTo", msg.ReplyTo),
		zap.String("subject", msg.Subject),
		zap.String("idempotencyKey", msg.IdempotencyKey))
	return "dryrun-" + msg.IdempotencyKey, nil
}
