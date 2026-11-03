-- Airlines that operate our tours' flights.
CREATE TABLE airline (
    id          UUID        PRIMARY KEY,
    -- Legacy id from the old nidnoitravel.com site; NULL for rows added later.
    ref_id      INTEGER     NULL,
    -- IATA designator (TG, FD), or the ICAO code (3 letters) when there is no IATA one. Uppercase, unique.
    code        TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    -- Logo URL on the R2 public domain (POST /media/upload, purpose airline-logo).
    logo        TEXT        NULL,
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL,
    CONSTRAINT airline_code_format CHECK (code ~ '^[A-Z0-9]{2,3}$')
);

CREATE UNIQUE INDEX airline_code_key ON airline (code);
CREATE UNIQUE INDEX airline_ref_id_key ON airline (ref_id);
