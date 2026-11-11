ALTER TABLE line_booking
    DROP COLUMN IF EXISTS fix_requested_at,
    DROP COLUMN IF EXISTS fix_note,
    DROP COLUMN IF EXISTS customer_sent_contact_id,
    DROP COLUMN IF EXISTS customer_sent_at,
    DROP COLUMN IF EXISTS documents_approved_by,
    DROP COLUMN IF EXISTS documents_approved_at;
