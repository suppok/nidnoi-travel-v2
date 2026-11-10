-- LINE bookings submitted by sales staff from the LIFF app inside LINE OA.
-- Separate from `booking` (customer self-service leads): these are sales-confirmed with payment slips.
-- Contains customer personal data (PDPA): never copy prd rows into dev.
CREATE TABLE line_booking (
    id                     UUID        PRIMARY KEY,
    booking_no             TEXT        NOT NULL,
    -- Sales who submitted via LIFF (LINE profile, not a backoffice user).
    line_user_id           TEXT        NOT NULL,
    line_display_name      TEXT        NOT NULL,
    -- Tour snapshot at booking time.
    tour_program_id        UUID        NOT NULL REFERENCES tour_program (id),
    tour_code              TEXT        NOT NULL,
    tour_name              TEXT        NOT NULL,
    tour_program_period_id UUID        NOT NULL REFERENCES tour_program_period (id),
    start_date             DATE        NOT NULL,
    end_date               DATE        NOT NULL,
    -- Travellers per price row and the unit price quoted (whole baht). A NULL price means the period
    -- had no such row, and its count is then 0.
    adult_twin_count       INTEGER     NOT NULL DEFAULT 0,
    adult_twin_price       INTEGER     NOT NULL,
    adult_single_count     INTEGER     NOT NULL DEFAULT 0,
    adult_single_price     INTEGER     NULL,
    child_with_bed_count   INTEGER     NOT NULL DEFAULT 0,
    child_with_bed_price   INTEGER     NULL,
    child_no_bed_count     INTEGER     NOT NULL DEFAULT 0,
    child_no_bed_price     INTEGER     NULL,
    total_price            INTEGER     NOT NULL,
    -- Customer contact.
    customer_name          TEXT        NOT NULL,
    customer_phone         TEXT        NOT NULL,
    -- Payment slip image URL (R2).
    slip_image             TEXT        NOT NULL,
    note                   TEXT        NOT NULL DEFAULT '',
    -- Sales workflow: 10 new, 20 confirmed, 30 cancelled. Keep in sync with pkg/constants/linebookingstate.
    state                  INTEGER     NOT NULL DEFAULT 10,
    staff_note             TEXT        NOT NULL DEFAULT '',
    status                 BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date            TIMESTAMPTZ NOT NULL,
    update_date            TIMESTAMPTZ NOT NULL,
    -- The nil UUID for rows created via LIFF (no backoffice user).
    create_by              UUID        NOT NULL,
    update_by              UUID        NOT NULL,
    CONSTRAINT line_booking_state_positive CHECK (state > 0),
    CONSTRAINT line_booking_counts CHECK (
        adult_twin_count >= 0 AND adult_single_count >= 0 AND child_with_bed_count >= 0 AND child_no_bed_count >= 0
        AND adult_twin_count + adult_single_count >= 1
        AND (adult_single_price IS NOT NULL OR adult_single_count = 0)
        AND (child_with_bed_price IS NOT NULL OR child_with_bed_count = 0)
        AND (child_no_bed_price IS NOT NULL OR child_no_bed_count = 0)
    ),
    CONSTRAINT line_booking_total CHECK (total_price >= 0)
);

CREATE UNIQUE INDEX line_booking_booking_no_key ON line_booking (booking_no);
CREATE INDEX line_booking_create_date_idx ON line_booking (create_date DESC);
CREATE INDEX line_booking_state_idx ON line_booking (state) WHERE status;
CREATE INDEX line_booking_line_user_id_idx ON line_booking (line_user_id);
CREATE INDEX line_booking_tour_program_id_idx ON line_booking (tour_program_id);
