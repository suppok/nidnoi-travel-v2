-- Staff link their own LINE account (from the backoffice QR code) so the LIFF app knows who they are.
-- The link code is a short-lived random token; only its SHA-256 is stored.
ALTER TABLE app_user
    ADD COLUMN line_user_id              TEXT        NULL,
    ADD COLUMN line_display_name         TEXT        NULL,
    ADD COLUMN line_linked_at            TIMESTAMPTZ NULL,
    ADD COLUMN line_link_code_hash       TEXT        NULL,
    ADD COLUMN line_link_code_expires_at TIMESTAMPTZ NULL;

-- One staff user per LINE account.
CREATE UNIQUE INDEX app_user_line_user_id_key ON app_user (line_user_id) WHERE line_user_id IS NOT NULL;
CREATE UNIQUE INDEX app_user_line_link_code_hash_key ON app_user (line_link_code_hash) WHERE line_link_code_hash IS NOT NULL;
