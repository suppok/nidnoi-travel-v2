-- Customer-site builds the backend triggered through the Cloudflare deploy hook (manual button,
-- debounced content edits, daily schedule). Builds from git pushes are not recorded here.
-- Append-only log: rows are never updated or soft-deleted (status stays TRUE).
CREATE TABLE site_build (
    id          UUID        PRIMARY KEY,
    -- Keep in sync with pkg/constants/sitebuild: manual, auto, daily.
    trigger     TEXT        NOT NULL,
    -- triggered (Cloudflare accepted the hook), failed (after retries), skipped (no hook configured).
    outcome     TEXT        NOT NULL,
    -- Customer-site paths the batch covered (informational; Cloudflare rebuilds the whole site).
    paths       TEXT[]      NOT NULL DEFAULT '{}',
    attempts    INTEGER     NOT NULL DEFAULT 0,
    -- Last trigger error, '' on success. Never contains the hook URL.
    error       TEXT        NOT NULL DEFAULT '',
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    -- The user who pressed "rebuild" (manual); the nil UUID for system builds (auto, daily).
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL,
    CONSTRAINT site_build_trigger_known CHECK (trigger IN ('manual', 'auto', 'daily')),
    CONSTRAINT site_build_outcome_known CHECK (outcome IN ('triggered', 'failed', 'skipped')),
    CONSTRAINT site_build_attempts CHECK (attempts >= 0)
);

CREATE INDEX site_build_create_date_idx ON site_build (create_date DESC);
