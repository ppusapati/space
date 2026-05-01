-- sat-telemetry schema. Channels + frames + append-only samples.
--
-- For production deployments, telemetry_samples is intended to be a
-- TimescaleDB hypertable on (sample_time). The Atlas migration tooling
-- drives that upgrade in env-specific overlays; the base schema below
-- is portable plain Postgres and exercises the same indexes.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS channels (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    satellite_id    uuid NOT NULL,
    subsystem       text NOT NULL,
    name            text NOT NULL,
    units           text NOT NULL DEFAULT '',
    value_type      int  NOT NULL,
    min_value       double precision NOT NULL DEFAULT 0,
    max_value       double precision NOT NULL DEFAULT 0,
    sample_rate_hz  double precision NOT NULL DEFAULT 0 CHECK (sample_rate_hz >= 0),
    active          boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, satellite_id, subsystem, name)
);

CREATE INDEX IF NOT EXISTS channels_tenant_idx     ON channels (tenant_id);
CREATE INDEX IF NOT EXISTS channels_satellite_idx  ON channels (satellite_id);
CREATE INDEX IF NOT EXISTS channels_subsystem_idx  ON channels (subsystem);
CREATE INDEX IF NOT EXISTS channels_created_at_idx ON channels (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS telemetry_frames (
    id                  uuid PRIMARY KEY,
    tenant_id           uuid NOT NULL,
    satellite_id        uuid NOT NULL,
    apid                int  NOT NULL CHECK (apid >= 0),
    virtual_channel     int  NOT NULL CHECK (virtual_channel >= 0),
    sequence_count      bigint NOT NULL CHECK (sequence_count >= 0),
    sat_time            timestamptz NOT NULL,
    ground_time         timestamptz NOT NULL DEFAULT now(),
    payload_size_bytes  bigint NOT NULL CHECK (payload_size_bytes >= 0),
    payload_sha256      text NOT NULL,
    frame_type          text NOT NULL,
    created_by          text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS telemetry_frames_tenant_idx       ON telemetry_frames (tenant_id);
CREATE INDEX IF NOT EXISTS telemetry_frames_satellite_idx    ON telemetry_frames (satellite_id);
CREATE INDEX IF NOT EXISTS telemetry_frames_ground_time_idx  ON telemetry_frames (ground_time DESC, id DESC);
CREATE INDEX IF NOT EXISTS telemetry_frames_frame_type_idx   ON telemetry_frames (frame_type);

CREATE TABLE IF NOT EXISTS telemetry_samples (
    id              uuid NOT NULL,
    tenant_id       uuid NOT NULL,
    satellite_id    uuid NOT NULL,
    frame_id        uuid,
    channel_id      uuid NOT NULL,
    sample_time     timestamptz NOT NULL,
    value_double    double precision NOT NULL DEFAULT 0,
    value_int       bigint NOT NULL DEFAULT 0,
    value_bool      boolean NOT NULL DEFAULT false,
    value_text      text NOT NULL DEFAULT '',
    ingested_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (channel_id, sample_time, id)
);

CREATE INDEX IF NOT EXISTS telemetry_samples_channel_time_idx
    ON telemetry_samples (channel_id, sample_time DESC);
CREATE INDEX IF NOT EXISTS telemetry_samples_frame_idx
    ON telemetry_samples (frame_id);
CREATE INDEX IF NOT EXISTS telemetry_samples_satellite_time_idx
    ON telemetry_samples (satellite_id, sample_time DESC);
