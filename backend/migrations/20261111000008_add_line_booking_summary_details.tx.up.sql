-- Everything the backoffice "สรุปยอดโอน" summary holds, so a LIFF booking keeps the deal as told to
-- the customer. All columns have defaults, so older code keeps inserting rows.
-- Unit prices (adult_twin_price, ...) are now entered by sales, defaulting to the period's prices;
-- total_price is the summary's grand total: price rows - discount + charges + card fee.
ALTER TABLE line_booking
    ADD COLUMN booking_source_id           UUID    NULL REFERENCES booking_source (id),
    -- The customer's name and booking number on that source (e.g. their LINE name).
    ADD COLUMN ref_name                    TEXT    NOT NULL DEFAULT '',
    ADD COLUMN ref_booking_no              TEXT    NOT NULL DEFAULT '',
    ADD COLUMN customer_address            TEXT    NOT NULL DEFAULT '',
    ADD COLUMN infant_count                INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN infant_price                INTEGER NULL,
    -- Shown to the customer as "(จาก …)" next to the adult price.
    ADD COLUMN adult_price_before_discount INTEGER NULL,
    -- Off every traveller's price (adults, children, infants).
    ADD COLUMN discount_per_person         INTEGER NOT NULL DEFAULT 0,
    -- Extra charges: [{"kind":"visa|foreignPassport|joinLand|fuelTax|airportTax|option","name":"","count":2,"price":1500}]
    ADD COLUMN charges                     JSONB   NOT NULL DEFAULT '[]',
    ADD COLUMN payment_method              TEXT    NOT NULL DEFAULT 'transfer',
    -- Card fee rate in basis points (300 = 3%) and the fee charged, on the deposit when there is one.
    ADD COLUMN credit_fee_bps              INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN card_fee                    INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN is_deposit                  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN deposit_per_person          INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN deposit_due_date            DATE    NULL,
    ADD COLUMN deposit_due_time            TEXT    NOT NULL DEFAULT '', -- HH:mm, '' = none
    ADD COLUMN remaining_due_date          DATE    NULL,
    ADD COLUMN full_due_date               DATE    NULL,
    ADD COLUMN full_due_time               TEXT    NOT NULL DEFAULT '',
    ADD CONSTRAINT line_booking_payment_method CHECK (payment_method IN ('transfer', 'credit')),
    ADD CONSTRAINT line_booking_infants CHECK (infant_count >= 0 AND (infant_price IS NOT NULL OR infant_count = 0)),
    ADD CONSTRAINT line_booking_amounts CHECK (
        discount_per_person >= 0 AND credit_fee_bps >= 0 AND card_fee >= 0 AND deposit_per_person >= 0
    );

CREATE INDEX line_booking_booking_source_id_idx ON line_booking (booking_source_id) WHERE booking_source_id IS NOT NULL;
