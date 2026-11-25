-- โปรไฟไหม้: staff pick departures to promote on the customer home page. fire_sale_order sorts them
-- (lowest first); the site shows the first few that are still upcoming and on a public, non-closed tour.
ALTER TABLE tour_program_period
    ADD COLUMN fire_sale       BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN fire_sale_order INTEGER NOT NULL DEFAULT 0,
    ADD CONSTRAINT tour_program_period_fire_sale_order_check CHECK (fire_sale_order >= 0),
    -- A promo always shows a struck-through price.
    ADD CONSTRAINT tour_program_period_fire_sale_discount_check
        CHECK (NOT fire_sale OR (price_before_discount IS NOT NULL AND price_before_discount > price));

CREATE INDEX tour_program_period_fire_sale_idx
    ON tour_program_period (fire_sale_order, start_date)
    WHERE fire_sale AND status;
