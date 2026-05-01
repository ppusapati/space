-- gs-scheduler schema. Predicted contact passes + tenant bookings.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS contact_passes (
    id                 uuid PRIMARY KEY,
    tenant_id          uuid NOT NULL,
    station_id         uuid NOT NULL,
    satellite_id       uuid NOT NULL,
    aos_time           timestamptz NOT NULL,
    tca_time           timestamptz NOT NULL,
    los_time           timestamptz NOT NULL,
    max_elevation_deg  double precision NOT NULL CHECK (max_elevation_deg BETWEEN 0 AND 90),
    aos_azimuth_deg    double precision NOT NULL CHECK (aos_azimuth_deg >= 0 AND aos_azimuth_deg < 360),
    los_azimuth_deg    double precision NOT NULL CHECK (los_azimuth_deg >= 0 AND los_azimuth_deg < 360),
    source             text NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    created_by         text NOT NULL DEFAULT 'system',
    updated_by         text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS contact_passes_tenant_idx     ON contact_passes (tenant_id);
CREATE INDEX IF NOT EXISTS contact_passes_station_idx    ON contact_passes (station_id);
CREATE INDEX IF NOT EXISTS contact_passes_satellite_idx  ON contact_passes (satellite_id);
CREATE INDEX IF NOT EXISTS contact_passes_aos_idx        ON contact_passes (aos_time ASC, id ASC);
CREATE INDEX IF NOT EXISTS contact_passes_created_at_idx ON contact_passes (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS bookings (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    pass_id         uuid NOT NULL REFERENCES contact_passes (id),
    priority        int  NOT NULL CHECK (priority BETWEEN 0 AND 100),
    status          int  NOT NULL,
    purpose         text NOT NULL,
    notes           text NOT NULL DEFAULT '',
    scheduled_at    timestamptz,
    completed_at    timestamptz,
    error_message   text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS bookings_tenant_idx     ON bookings (tenant_id);
CREATE INDEX IF NOT EXISTS bookings_pass_idx       ON bookings (pass_id);
CREATE INDEX IF NOT EXISTS bookings_status_idx     ON bookings (status);
CREATE INDEX IF NOT EXISTS bookings_created_at_idx ON bookings (created_at DESC, id DESC);
