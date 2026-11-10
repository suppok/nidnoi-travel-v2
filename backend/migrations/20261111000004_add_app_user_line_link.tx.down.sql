ALTER TABLE app_user
    DROP COLUMN IF EXISTS line_user_id,
    DROP COLUMN IF EXISTS line_display_name,
    DROP COLUMN IF EXISTS line_linked_at,
    DROP COLUMN IF EXISTS line_link_code_hash,
    DROP COLUMN IF EXISTS line_link_code_expires_at;
