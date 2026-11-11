-- Channels a booking comes through (LINE OA, Facebook, agents, ...), picked in the backoffice
-- sales tools (booking summary).
CREATE TABLE booking_source (
    id          UUID        PRIMARY KEY,
    name        TEXT        NOT NULL,
    -- Short label for the source's own reference name, shown as "ชื่อ <short_name>"; empty = plain "ชื่อ".
    short_name  TEXT        NOT NULL DEFAULT '',
    -- Display order in dropdowns, ascending.
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL
);

CREATE UNIQUE INDEX booking_source_name_key ON booking_source (lower(name));
