-- gs-ingest schema. Ingest sessions + downlink frames.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS ingest_sessions (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    booking_id      uuid NOT NULL,
    pass_id         uuid NOT NULL,
    station_id      uuid NOT NULL,
    satellite_id    uuid NOT NULL,
    status          int  NOT NULL,
    started_at      timestamptz,
    completed_at    timestamptz,
    frames_received bigint NOT NULL DEFAULT 0 CHECK (frames_received >= 0),
    bytes_received  bigint NOT NULL DEFAULT 0 CHECK (bytes_received  >= 0),
    error_message   text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS ingest_sessions_tenant_idx     ON ingest_sessions (tenant_id);
CREATE INDEX IF NOT EXISTS ingest_sessions_station_idx    ON ingest_sessions (station_id);
CREATE INDEX IF NOT EXISTS ingest_sessions_satellite_idx  ON ingest_sessions (satellite_id);
CREATE INDEX IF NOT EXISTS ingest_sessions_status_idx     ON ingest_sessions (status);
CREATE INDEX IF NOT EXISTS ingest_sessions_created_at_idx ON ingest_sessions (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS downlink_frames (
    id                  uuid PRIMARY KEY,
    tenant_id           uuid NOT NULL,
    session_id          uuid NOT NULL REFERENCES ingest_sessions (id),
    apid                int  NOT NULL CHECK (apid >= 0),
    virtual_channel     int  NOT NULL CHECK (virtual_channel >= 0),
    sequence_count      bigint NOT NULL CHECK (sequence_count >= 0),
    ground_time         timestamptz NOT NULL DEFAULT now(),
    payload_size_bytes  bigint NOT NULL CHECK (payload_size_bytes >= 0),
    payload_sha256      text NOT NULL,
    payload_uri         text NOT NULL,
    frame_type          text NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    created_by          text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS downlink_frames_tenant_idx       ON downlink_frames (tenant_id);
CREATE INDEX IF NOT EXISTS downlink_frames_session_idx      ON downlink_frames (session_id);
CREATE INDEX IF NOT EXISTS downlink_frames_ground_time_idx  ON downlink_frames (ground_time DESC, id DESC);
CREATE INDEX IF NOT EXISTS downlink_frames_frame_type_idx   ON downlink_frames (frame_type);
