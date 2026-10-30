-- Wholesale tour operators we source tours from, with their contact people.
CREATE TABLE wholesale (
    id          UUID        PRIMARY KEY,
    -- Legacy id from the old nidnoitravel.com site; NULL for rows added later.
    ref_id      INTEGER     NULL,
    -- Short code staff use to refer to the wholesale (uppercase, unique).
    code        TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    -- Workflow state, keep in sync with pkg/constants/wholesalestate: 10 on request, 20 connected, 30 blocked, 40 declined.
    state       INTEGER     NOT NULL DEFAULT 10,
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL,
    CONSTRAINT wholesale_code_upper CHECK (code = upper(code)),
    CONSTRAINT wholesale_state_positive CHECK (state > 0)
);

CREATE UNIQUE INDEX wholesale_code_key ON wholesale (code);
CREATE UNIQUE INDEX wholesale_ref_id_key ON wholesale (ref_id);

-- Contact people of a wholesale. Replaced as a whole when the wholesale is saved, so there is no
-- soft-delete status.
CREATE TABLE wholesale_contact (
    id           UUID        PRIMARY KEY,
    wholesale_id UUID        NOT NULL REFERENCES wholesale (id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    phone        TEXT        NOT NULL,
    sort_order   INTEGER     NOT NULL DEFAULT 0,
    create_date  TIMESTAMPTZ NOT NULL,
    update_date  TIMESTAMPTZ NOT NULL,
    create_by    UUID        NOT NULL,
    update_by    UUID        NOT NULL
);

CREATE INDEX wholesale_contact_wholesale_id_idx ON wholesale_contact (wholesale_id, sort_order);
