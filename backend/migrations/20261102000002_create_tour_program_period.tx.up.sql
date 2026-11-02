-- Departures of a tour with prices and seats: the data that changes most often.
-- Prices are whole Thai baht.
CREATE TABLE tour_program_period (
    id                    UUID        PRIMARY KEY,
    tour_program_id       UUID        NOT NULL REFERENCES tour_program (id),
    start_date            DATE        NOT NULL,
    end_date              DATE        NOT NULL,
    -- Adult price, twin/triple sharing. The lowest among active periods is the tour's starting price.
    price                 INTEGER     NOT NULL,
    -- Shown struck through next to price when set.
    price_before_discount INTEGER     NULL,
    child_with_bed_price  INTEGER     NULL,
    child_no_bed_price    INTEGER     NULL,
    single_supplement     INTEGER     NULL,
    deposit               INTEGER     NULL,
    seats_total           INTEGER     NOT NULL,
    seats_available       INTEGER     NOT NULL,
    -- Staff only.
    note                  TEXT        NOT NULL DEFAULT '',
    status                BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date           TIMESTAMPTZ NOT NULL,
    update_date           TIMESTAMPTZ NOT NULL,
    create_by             UUID        NOT NULL,
    update_by             UUID        NOT NULL,
    CONSTRAINT tour_program_period_dates CHECK (end_date >= start_date),
    CONSTRAINT tour_program_period_prices CHECK (
        price >= 0
        AND (price_before_discount IS NULL OR price_before_discount >= 0)
        AND (child_with_bed_price IS NULL OR child_with_bed_price >= 0)
        AND (child_no_bed_price IS NULL OR child_no_bed_price >= 0)
        AND (single_supplement IS NULL OR single_supplement >= 0)
        AND (deposit IS NULL OR deposit >= 0)
    ),
    CONSTRAINT tour_program_period_seats CHECK (seats_total >= 0 AND seats_available >= 0 AND seats_available <= seats_total)
);

CREATE INDEX tour_program_period_tour_program_id_idx ON tour_program_period (tour_program_id, start_date);
