-- What the tour price includes / excludes (shown on the customer site), and whether the tour stops at
-- government shops (ลงร้านรัฐบาล): NULL = not specified.
ALTER TABLE tour_program
    ADD COLUMN government_shop_stops BOOLEAN NULL,
    ADD COLUMN included_items TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN excluded_items TEXT[] NOT NULL DEFAULT '{}';
