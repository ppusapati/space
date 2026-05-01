-- gs-mc schema. Ground stations + antennas.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS ground_stations (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    slug            text NOT NULL,
    name            text NOT NULL,
    country_code    text NOT NULL,
    latitude_deg    double precision NOT NULL CHECK (latitude_deg BETWEEN -90 AND 90),
    longitude_deg   double precision NOT NULL CHECK (longitude_deg BETWEEN -180 AND 180),
    altitude_m      double precision NOT NULL DEFAULT 0,
    active          boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS ground_stations_tenant_idx     ON ground_stations (tenant_id);
CREATE INDEX IF NOT EXISTS ground_stations_created_at_idx ON ground_stations (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS antennas (
    id                   uuid PRIMARY KEY,
    tenant_id            uuid NOT NULL,
    station_id           uuid NOT NULL REFERENCES ground_stations (id),
    slug                 text NOT NULL,
    name                 text NOT NULL,
    band                 int  NOT NULL,
    min_freq_hz          bigint NOT NULL CHECK (min_freq_hz >= 0),
    max_freq_hz          bigint NOT NULL CHECK (max_freq_hz >= 0),
    polarization         int  NOT NULL,
    gain_dbi             double precision NOT NULL DEFAULT 0,
    slew_rate_deg_per_s  double precision NOT NULL DEFAULT 0,
    active               boolean NOT NULL DEFAULT true,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    created_by           text NOT NULL DEFAULT 'system',
    updated_by           text NOT NULL DEFAULT 'system',
    UNIQUE (station_id, slug)
);

CREATE INDEX IF NOT EXISTS antennas_tenant_idx     ON antennas (tenant_id);
CREATE INDEX IF NOT EXISTS antennas_station_idx    ON antennas (station_id);
CREATE INDEX IF NOT EXISTS antennas_band_idx       ON antennas (band);
CREATE INDEX IF NOT EXISTS antennas_created_at_idx ON antennas (created_at DESC, id DESC);
