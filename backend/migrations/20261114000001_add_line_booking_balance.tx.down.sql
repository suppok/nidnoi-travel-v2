ALTER TABLE line_booking
    DROP COLUMN IF EXISTS balance_confirmed_at,
    DROP COLUMN IF EXISTS balance_submitted_by,
    DROP COLUMN IF EXISTS balance_submitted_at,
    DROP COLUMN IF EXISTS balance_slip_image;
