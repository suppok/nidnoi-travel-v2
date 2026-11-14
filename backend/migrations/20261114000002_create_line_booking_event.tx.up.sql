-- History of a LINE booking: one append-only row per action (created, confirmed, documents, LIFF
-- edits with their field changes, balance, ...). Rows are never updated or deleted, so the table has
-- no status / update columns.
CREATE TABLE line_booking_event (
    id              UUID        PRIMARY KEY,
    line_booking_id UUID        NOT NULL REFERENCES line_booking (id),
    action          TEXT        NOT NULL,
    detail          JSONB       NOT NULL DEFAULT '{}',
    create_date     TIMESTAMPTZ NOT NULL,
    create_by       UUID        NOT NULL REFERENCES app_user (id)
);

CREATE INDEX line_booking_event_booking_idx ON line_booking_event (line_booking_id, create_date);
