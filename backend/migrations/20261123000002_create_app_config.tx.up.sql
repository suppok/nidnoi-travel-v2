-- Typed app configuration: one row per code. Every code is registered in Go
-- (pkg/constants/appconfigcode) with the struct its value must decode into.
CREATE TABLE app_config (
    id          UUID        PRIMARY KEY,
    -- UPPER_SNAKE key, e.g. APP_SETTINGS. Matches a code in pkg/constants/appconfigcode.
    code        TEXT        NOT NULL,
    -- An object, an array or a {"value": ...} wrapper.
    value       JSONB       NOT NULL,
    -- Label for the value shape (informational).
    model_type  TEXT        NOT NULL DEFAULT '',
    description TEXT        NOT NULL DEFAULT '',
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL
);

CREATE UNIQUE INDEX app_config_code_key ON app_config (code);
