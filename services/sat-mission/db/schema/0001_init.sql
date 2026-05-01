-- sat-mission schema. Satellite registry per tenant.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS satellites (
    id                       uuid PRIMARY KEY,
    tenant_id                uuid NOT NULL,
    name                     text NOT NULL,
    norad_id                 text NOT NULL DEFAULT '',
    international_designator text NOT NULL DEFAULT '',
    tle_line1                text NOT NULL DEFAULT '',
    tle_line2                text NOT NULL DEFAULT '',
    current_mode             int  NOT NULL DEFAULT 0,
    last_state_rx_km         double precision,
    last_state_ry_km         double precision,
    last_state_rz_km         double precision,
    last_state_vx_km_s       double precision,
    last_state_vy_km_s       double precision,
    last_state_vz_km_s       double precision,
    last_state_epoch         timestamptz,
    config_json              jsonb NOT NULL DEFAULT '{}'::jsonb,
    active                   boolean NOT NULL DEFAULT true,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now(),
    created_by               text NOT NULL DEFAULT 'system',
    updated_by               text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS satellites_tenant_idx ON satellites (tenant_id);
CREATE INDEX IF NOT EXISTS satellites_norad_idx  ON satellites (norad_id) WHERE norad_id <> '';
