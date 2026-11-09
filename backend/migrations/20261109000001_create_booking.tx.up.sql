-- Booking numbers are <BK><nextval as 6 digits>, e.g. BK000001: the reference a customer quotes to staff.
CREATE SEQUENCE booking_no_seq START WITH 1 MAXVALUE 999999;

-- Booking requests from the customer site (/tour/<code>/booking/). A booking is a sales lead:
-- it reserves no seats and takes no payment; sales contacts the customer to confirm.
-- Contains customer personal data (PDPA): never copy prd rows into dev.
CREATE TABLE booking (
    id                     UUID        PRIMARY KEY,
    booking_no             TEXT        NOT NULL,
    tour_program_id        UUID        NOT NULL REFERENCES tour_program (id),
    tour_program_period_id UUID        NOT NULL REFERENCES tour_program_period (id),
    -- Snapshot at booking time, so the lead still reads correctly after the tour or period is edited.
    tour_code              TEXT        NOT NULL,
    tour_name              TEXT        NOT NULL,
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
    first_name             TEXT        NOT NULL,
    last_name              TEXT        NOT NULL DEFAULT '',
    email                  TEXT        NOT NULL,
    phone                  TEXT        NOT NULL,
    -- Customer's special requests ("ความต้องการพิเศษ").
    customer_note          TEXT        NOT NULL DEFAULT '',
    -- Sales workflow, keep in sync with pkg/constants/bookingstate: 10 new, 20 contacted, 30 confirmed, 40 cancelled.
    state                  INTEGER     NOT NULL DEFAULT 10,
    -- The staff user who first took the booking out of "new"; NULL while nobody has.
    assignee_id            UUID        NULL REFERENCES app_user (id),
    staff_note             TEXT        NOT NULL DEFAULT '',
    status                 BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date            TIMESTAMPTZ NOT NULL,
    update_date            TIMESTAMPTZ NOT NULL,
    -- The nil UUID for rows created by the customer site.
    create_by              UUID        NOT NULL,
    update_by              UUID        NOT NULL,
    CONSTRAINT booking_state_positive CHECK (state > 0),
    CONSTRAINT booking_counts CHECK (
        adult_twin_count >= 0 AND adult_single_count >= 0 AND child_with_bed_count >= 0 AND child_no_bed_count >= 0
        AND adult_twin_count + adult_single_count >= 1
        AND (adult_single_price IS NOT NULL OR adult_single_count = 0)
        AND (child_with_bed_price IS NOT NULL OR child_with_bed_count = 0)
        AND (child_no_bed_price IS NOT NULL OR child_no_bed_count = 0)
    ),
    CONSTRAINT booking_total CHECK (total_price >= 0)
);

CREATE UNIQUE INDEX booking_booking_no_key ON booking (booking_no);
CREATE INDEX booking_create_date_idx ON booking (create_date DESC);
CREATE INDEX booking_state_idx ON booking (state) WHERE status;
CREATE INDEX booking_tour_program_id_idx ON booking (tour_program_id);
