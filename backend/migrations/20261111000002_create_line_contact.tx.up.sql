-- Customers who chatted with, followed or blocked the LINE OA, collected from Messaging API webhooks.
-- Only the LINE profile is stored, never message text. Personal data (PDPA): never copy prd rows into dev.
CREATE TABLE line_contact (
    id                 UUID        PRIMARY KEY,
    -- Messaging API user ID (U + 32 hex). Matches LIFF user IDs only when both channels share a provider.
    line_user_id       TEXT        NOT NULL,
    display_name       TEXT        NOT NULL DEFAULT '',
    picture_url        TEXT        NULL,
    -- True while the user is a friend of the OA (push messages only reach friends).
    following          BOOLEAN     NOT NULL DEFAULT FALSE,
    last_message_at    TIMESTAMPTZ NULL,
    profile_updated_at TIMESTAMPTZ NULL,
    status             BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date        TIMESTAMPTZ NOT NULL,
    update_date        TIMESTAMPTZ NOT NULL,
    -- The nil UUID: rows are created by webhooks, not by a backoffice user.
    create_by          UUID        NOT NULL,
    update_by          UUID        NOT NULL
);

CREATE UNIQUE INDEX line_contact_line_user_id_key ON line_contact (line_user_id);
CREATE INDEX line_contact_last_message_at_idx ON line_contact (last_message_at DESC NULLS LAST);
