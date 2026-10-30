-- Tour taxonomy: cities (also provinces and sub-country regions such as "Hokkaido"). Each belongs to one country.
CREATE TABLE city (
    id          UUID        PRIMARY KEY,
    -- Legacy destination id from the old nidnoitravel.com site; NULL for rows added later.
    ref_id      INTEGER     NULL,
    country_id  UUID        NOT NULL REFERENCES country (id),
    -- Used in customer URLs (/tours/<country>/<slug>/); locked after create.
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
    CONSTRAINT city_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug) <= 60)
);

CREATE UNIQUE INDEX city_slug_key ON city (slug);
CREATE UNIQUE INDEX city_ref_id_key ON city (ref_id);
CREATE INDEX city_country_id_idx ON city (country_id);
