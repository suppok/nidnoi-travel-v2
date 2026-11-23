-- Transactional outbox for outgoing email. Use cases insert a row in the same transaction as the
-- change that triggers it; the in-process sender (services/emailoutbox) delivers it through the
-- mailer. Addresses, subject and bodies are resolved when the row is queued, so a later config
-- change never alters a queued email.
CREATE TABLE email_outbox (
    id                  UUID        PRIMARY KEY,
    line_booking_id     UUID        NOT NULL REFERENCES line_booking (id),
    -- Keep in sync with pkg/constants/emailoutbox (Type*).
    type                TEXT        NOT NULL,
    from_email          TEXT        NOT NULL,
    to_emails           TEXT[]      NOT NULL,
    cc_emails           TEXT[]      NOT NULL DEFAULT '{}',
    reply_to            TEXT        NOT NULL DEFAULT '',
    subject             TEXT        NOT NULL,
    body_html           TEXT        NOT NULL,
    body_text           TEXT        NOT NULL,
    -- pkg/constants/emailoutbox: 10 pending, 20 sent, 30 failed.
    state               INTEGER     NOT NULL DEFAULT 10,
    -- Send attempts in the current round; a manual resend starts a new round (send_round + 1,
    -- attempts 0). The provider idempotency key is "<id>-<send_round>".
    attempts            INTEGER     NOT NULL DEFAULT 0,
    send_round          INTEGER     NOT NULL DEFAULT 1,
    next_attempt_at     TIMESTAMPTZ NOT NULL,
    last_error          TEXT        NOT NULL DEFAULT '',
    provider_message_id TEXT        NOT NULL DEFAULT '',
    sent_at             TIMESTAMPTZ,
    status              BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date         TIMESTAMPTZ NOT NULL,
    update_date         TIMESTAMPTZ NOT NULL,
    create_by           UUID        NOT NULL,
    update_by           UUID        NOT NULL,
    CONSTRAINT email_outbox_state_known CHECK (state IN (10, 20, 30)),
    CONSTRAINT email_outbox_to_not_empty CHECK (cardinality(to_emails) > 0),
    CONSTRAINT email_outbox_attempts CHECK (attempts >= 0 AND send_round >= 1)
);

-- One email of each type per booking: a duplicate enqueue is a no-op.
CREATE UNIQUE INDEX email_outbox_booking_type_key ON email_outbox (line_booking_id, type);
CREATE INDEX email_outbox_due_idx ON email_outbox (state, next_attempt_at);
