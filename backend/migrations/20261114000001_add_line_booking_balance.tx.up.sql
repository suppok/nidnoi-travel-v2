-- The balance of a deposit booking: sales record it from LIFF with a slip (balance_submitted_at),
-- the sales owner confirms it in the backoffice (balance_confirmed_at, state 25) and finance then
-- issues the RE.
ALTER TABLE line_booking
    ADD COLUMN balance_slip_image   TEXT        NOT NULL DEFAULT '',
    ADD COLUMN balance_submitted_at TIMESTAMPTZ NULL,
    ADD COLUMN balance_submitted_by UUID        NULL REFERENCES app_user (id),
    ADD COLUMN balance_confirmed_at TIMESTAMPTZ NULL;
