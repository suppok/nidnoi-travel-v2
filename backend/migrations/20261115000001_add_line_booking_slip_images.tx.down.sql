ALTER TABLE line_booking
    DROP COLUMN IF EXISTS balance_slip_images,
    DROP COLUMN IF EXISTS slip_images;
