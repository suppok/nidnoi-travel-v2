-- Backoffice refresh tokens. Only a SHA-256 hash of the token is stored; the token itself lives
-- in the HttpOnly cookie. Rotated on every refresh (revoked_at set, replaced_by points to the new row).
CREATE TABLE refresh_token (
    id          UUID        PRIMARY KEY,
    user_id     UUID        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL,
    persistent  BOOLEAN     NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ NULL,
    replaced_by UUID        NULL,
    user_agent  TEXT        NOT NULL DEFAULT '',
    client_ip   TEXT        NOT NULL DEFAULT '',
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL
);

CREATE UNIQUE INDEX refresh_token_token_hash_key ON refresh_token (token_hash);
CREATE INDEX refresh_token_user_active_idx ON refresh_token (user_id) WHERE revoked_at IS NULL;
