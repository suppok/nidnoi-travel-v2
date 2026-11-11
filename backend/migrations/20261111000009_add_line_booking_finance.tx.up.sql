-- Finance users each sales user sends confirmed LINE bookings to (chosen by the sales user).
-- Rows are replaced as a whole when the sales user saves their settings, so there is no status.
CREATE TABLE sales_finance (
    id              UUID        PRIMARY KEY,
    sales_user_id   UUID        NOT NULL REFERENCES app_user (id),
    finance_user_id UUID        NOT NULL REFERENCES app_user (id),
    is_default      BOOLEAN     NOT NULL DEFAULT FALSE,
    create_date     TIMESTAMPTZ NOT NULL,
    update_date     TIMESTAMPTZ NOT NULL,
    create_by       UUID        NOT NULL,
    update_by       UUID        NOT NULL,
    CONSTRAINT sales_finance_pair_key UNIQUE (sales_user_id, finance_user_id),
    CONSTRAINT sales_finance_not_self CHECK (sales_user_id <> finance_user_id)
);
CREATE UNIQUE INDEX sales_finance_default_key ON sales_finance (sales_user_id) WHERE is_default;

-- The finance user who issues the receipt, and when the booking was confirmed.
ALTER TABLE line_booking
    ADD COLUMN finance_user_id UUID        NULL REFERENCES app_user (id),
    ADD COLUMN confirmed_at    TIMESTAMPTZ NULL;
CREATE INDEX line_booking_finance_user_idx ON line_booking (finance_user_id, confirmed_at DESC) WHERE finance_user_id IS NOT NULL;

-- States: 10 new, 20 deposit paid (was "confirmed"), 25 paid in full, 30 cancelled.
-- Confirmed bookings that were not a deposit become "paid in full".
UPDATE line_booking SET state = 25 WHERE state = 20 AND NOT is_deposit;
