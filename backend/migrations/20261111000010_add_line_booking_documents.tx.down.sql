DROP INDEX IF EXISTS line_booking_documents_at_idx;
ALTER TABLE line_booking
    DROP COLUMN IF EXISTS documents_by,
    DROP COLUMN IF EXISTS documents_at,
    DROP COLUMN IF EXISTS re_file,
    DROP COLUMN IF EXISTS iv_file,
    DROP COLUMN IF EXISTS ai_file;
