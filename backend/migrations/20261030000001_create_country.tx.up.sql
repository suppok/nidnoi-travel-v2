-- Tour taxonomy: countries (also continents and multi-country regions such as "Scandinavia").
CREATE TABLE country (
    id          UUID        PRIMARY KEY,
    -- Legacy destination id from the old nidnoitravel.com site; NULL for rows added later.
    ref_id      INTEGER     NULL,
    -- Used in customer URLs (/tours/<slug>/); locked after create.
    slug        TEXT        NOT NULL,
    name_th     TEXT        NOT NULL,
    name_en     TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    hero_image  TEXT        NULL,
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    status      BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL,
    create_by   UUID        NOT NULL,
    update_by   UUID        NOT NULL,
    CONSTRAINT country_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug) <= 60)
);

CREATE UNIQUE INDEX country_slug_key ON country (slug);
CREATE UNIQUE INDEX country_ref_id_key ON country (ref_id);
