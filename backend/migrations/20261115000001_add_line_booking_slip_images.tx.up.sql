-- A payment can come with several slips (split transfers). slip_image / balance_slip_image stay as
-- the first slip, so code that reads only them keeps working.
ALTER TABLE line_booking
    ADD COLUMN slip_images         TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN balance_slip_images TEXT[] NOT NULL DEFAULT '{}';

UPDATE line_booking SET slip_images = ARRAY[slip_image] WHERE slip_image <> '';
UPDATE line_booking SET balance_slip_images = ARRAY[balance_slip_image] WHERE balance_slip_image <> '';
