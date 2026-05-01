-- gi-tiles schema. Tile-set catalog (raster + vector).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tile_sets (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL,
    slug         text NOT NULL,
    name         text NOT NULL,
    description  text NOT NULL DEFAULT '',
    format       int  NOT NULL,
    projection   text NOT NULL,
    min_zoom     int  NOT NULL CHECK (min_zoom BETWEEN 0 AND 30),
    max_zoom     int  NOT NULL CHECK (max_zoom BETWEEN 0 AND 30 AND max_zoom >= min_zoom),
    source_uri   text NOT NULL,
    attribution  text NOT NULL DEFAULT '',
    active       boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    created_by   text NOT NULL DEFAULT 'system',
    updated_by   text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS tile_sets_tenant_idx     ON tile_sets (tenant_id);
CREATE INDEX IF NOT EXISTS tile_sets_format_idx     ON tile_sets (format);
CREATE INDEX IF NOT EXISTS tile_sets_created_at_idx ON tile_sets (created_at DESC, id DESC);
