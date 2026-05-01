-- eo-catalog schema. Tenant-scoped STAC-style catalog of collections,
-- items (scenes), assets, and quality results.
--
-- Spatial data is stored as four bounding-box columns plus an optional
-- GeoJSON string. PostGIS-backed spatial indexing is added in a later
-- migration via raw SQL outside the sqlc-generated layer.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS collections (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    slug            text NOT NULL,
    title           text NOT NULL,
    description     text NOT NULL DEFAULT '',
    license         text NOT NULL DEFAULT '',
    bbox_lon_min    double precision,
    bbox_lat_min    double precision,
    bbox_lon_max    double precision,
    bbox_lat_max    double precision,
    temporal_start  timestamptz,
    temporal_end    timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS collections_tenant_idx ON collections (tenant_id);

CREATE TABLE IF NOT EXISTS items (
    id                uuid PRIMARY KEY,
    tenant_id         uuid NOT NULL,
    collection_id     uuid NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    mission           text NOT NULL,
    platform          text NOT NULL DEFAULT '',
    instrument        text NOT NULL DEFAULT '',
    datetime          timestamptz NOT NULL,
    bbox_lon_min      double precision,
    bbox_lat_min      double precision,
    bbox_lon_max      double precision,
    bbox_lat_max      double precision,
    geometry_geojson  text NOT NULL DEFAULT '',
    cloud_cover       double precision NOT NULL DEFAULT 0,
    properties_json   jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    created_by        text NOT NULL DEFAULT 'system',
    updated_by        text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS items_tenant_idx       ON items (tenant_id);
CREATE INDEX IF NOT EXISTS items_collection_idx   ON items (collection_id);
CREATE INDEX IF NOT EXISTS items_datetime_idx     ON items (datetime DESC);
CREATE INDEX IF NOT EXISTS items_bbox_lonlat_idx
    ON items (bbox_lon_min, bbox_lat_min, bbox_lon_max, bbox_lat_max);

CREATE TABLE IF NOT EXISTS assets (
    id          uuid PRIMARY KEY,
    item_id     uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    key         text NOT NULL,
    href        text NOT NULL,
    media_type  text NOT NULL DEFAULT '',
    title       text NOT NULL DEFAULT '',
    roles       text[] NOT NULL DEFAULT '{}',
    UNIQUE (item_id, key)
);

CREATE INDEX IF NOT EXISTS assets_item_idx ON assets (item_id);

CREATE TABLE IF NOT EXISTS quality_results (
    id                    uuid PRIMARY KEY,
    item_id               uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    cloud_cover           double precision NOT NULL,
    radiometric_rmse      double precision NOT NULL DEFAULT 0,
    geometric_accuracy_m  double precision NOT NULL DEFAULT 0,
    notes                 text NOT NULL DEFAULT '',
    computed_at           timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS quality_results_item_idx ON quality_results (item_id, computed_at DESC);
