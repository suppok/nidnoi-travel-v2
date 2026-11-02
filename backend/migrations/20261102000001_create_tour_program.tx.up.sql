-- Public tour codes are <tourCode.prefix><nextval as 6 digits>, e.g. NIDN000001. Gaps after a rollback
-- are fine. MAXVALUE keeps codes at 6 digits (nextval fails instead of producing NIDN1000000).
-- The legacy import moves this past the highest imported code with setval.
CREATE SEQUENCE tour_code_seq START WITH 1 MAXVALUE 999999;

-- Tour programs. code is the public, stable identifier used in customer URLs (/tour/<code>/).
CREATE TABLE tour_program (
    id                  UUID        PRIMARY KEY,
    -- Legacy id from the old nidnoitravel.com site; NULL for rows added later.
    ref_id              INTEGER     NULL,
    -- Generated from tour_code_seq at create; never changes.
    code                TEXT        NOT NULL,
    name                TEXT        NOT NULL,
    wholesale_id        UUID        NOT NULL REFERENCES wholesale (id),
    -- The wholesale's own code for this program (staff only).
    wholesale_tour_code TEXT        NOT NULL DEFAULT '',
    -- Workflow state, keep in sync with pkg/constants/tourprogramstate: 10 draft, 20 selling, 30 closed, 40 cancelled.
    state               INTEGER     NOT NULL DEFAULT 10,
    days                INTEGER     NOT NULL,
    nights              INTEGER     NOT NULL,
    airline             TEXT        NOT NULL DEFAULT '',
    -- Short summary for tour cards and the meta description.
    highlight           TEXT        NOT NULL DEFAULT '',
    description         TEXT        NOT NULL DEFAULT '',
    -- Media URLs on the R2 public domain.
    banner_image        TEXT        NULL,
    pdf_file            TEXT        NULL, -- public itinerary download
    docx_file           TEXT        NULL, -- staff only (editable source)
    internal_note       TEXT        NOT NULL DEFAULT '',
    status              BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date         TIMESTAMPTZ NOT NULL,
    update_date         TIMESTAMPTZ NOT NULL,
    create_by           UUID        NOT NULL,
    update_by           UUID        NOT NULL,
    CONSTRAINT tour_program_state_positive CHECK (state > 0),
    CONSTRAINT tour_program_duration CHECK (days >= 1 AND nights >= 0 AND nights <= days)
);

CREATE UNIQUE INDEX tour_program_code_key ON tour_program (code);
CREATE UNIQUE INDEX tour_program_ref_id_key ON tour_program (ref_id);
CREATE INDEX tour_program_wholesale_id_idx ON tour_program (wholesale_id);

-- Countries of a tour (one or more), in display order. Replaced as a whole when the tour is saved.
CREATE TABLE tour_program_country (
    tour_program_id UUID    NOT NULL REFERENCES tour_program (id) ON DELETE CASCADE,
    country_id      UUID    NOT NULL REFERENCES country (id),
    sort_order      INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (tour_program_id, country_id)
);

CREATE INDEX tour_program_country_country_id_idx ON tour_program_country (country_id);

-- Cities of a tour (zero or more), in display order. Replaced as a whole when the tour is saved.
CREATE TABLE tour_program_city (
    tour_program_id UUID    NOT NULL REFERENCES tour_program (id) ON DELETE CASCADE,
    city_id         UUID    NOT NULL REFERENCES city (id),
    sort_order      INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (tour_program_id, city_id)
);

CREATE INDEX tour_program_city_city_id_idx ON tour_program_city (city_id);

-- Gallery images of a tour, in display order. Replaced as a whole when the tour is saved, so there
-- is no soft-delete status.
CREATE TABLE tour_program_image (
    id              UUID        PRIMARY KEY,
    tour_program_id UUID        NOT NULL REFERENCES tour_program (id) ON DELETE CASCADE,
    url             TEXT        NOT NULL,
    sort_order      INTEGER     NOT NULL DEFAULT 0,
    create_date     TIMESTAMPTZ NOT NULL,
    update_date     TIMESTAMPTZ NOT NULL,
    create_by       UUID        NOT NULL,
    update_by       UUID        NOT NULL
);

CREATE INDEX tour_program_image_tour_program_id_idx ON tour_program_image (tour_program_id, sort_order);
