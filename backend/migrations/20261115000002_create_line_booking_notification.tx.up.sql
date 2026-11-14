-- A staff user's notification about a LINE booking: a comment on a booking they own or handle as
-- finance, or an @mention. One row per recipient; read_at is set when they open the booking.
CREATE TABLE line_booking_notification (
    id              UUID        PRIMARY KEY,
    user_id         UUID        NOT NULL REFERENCES app_user (id),
    line_booking_id UUID        NOT NULL REFERENCES line_booking (id),
    event_id        UUID        NOT NULL REFERENCES line_booking_event (id),
    kind            TEXT        NOT NULL,
    read_at         TIMESTAMPTZ NULL,
    create_date     TIMESTAMPTZ NOT NULL
);

CREATE INDEX line_booking_notification_user_idx ON line_booking_notification (user_id, create_date DESC);
