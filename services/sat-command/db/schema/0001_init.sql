-- sat-command schema. Command catalog + time-tagged uplink queue.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS command_defs (
    id                 uuid PRIMARY KEY,
    tenant_id          uuid NOT NULL,
    satellite_id       uuid,                 -- NULL = tenant-wide command
    subsystem          text NOT NULL,
    name               text NOT NULL,
    opcode             bigint NOT NULL CHECK (opcode >= 0),
    parameters_schema  text NOT NULL DEFAULT '',
    description        text NOT NULL DEFAULT '',
    active             boolean NOT NULL DEFAULT true,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    created_by         text NOT NULL DEFAULT 'system',
    updated_by         text NOT NULL DEFAULT 'system'
);

-- Uniqueness: per-satellite commands and tenant-wide commands have separate
-- partial unique indexes to permit NULL satellite_id without COALESCE shenanigans.
CREATE UNIQUE INDEX IF NOT EXISTS command_defs_unique_per_satellite
    ON command_defs (tenant_id, satellite_id, subsystem, name)
    WHERE satellite_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS command_defs_unique_tenant_wide
    ON command_defs (tenant_id, subsystem, name)
    WHERE satellite_id IS NULL;

CREATE INDEX IF NOT EXISTS command_defs_tenant_idx     ON command_defs (tenant_id);
CREATE INDEX IF NOT EXISTS command_defs_satellite_idx  ON command_defs (satellite_id);
CREATE INDEX IF NOT EXISTS command_defs_subsystem_idx  ON command_defs (subsystem);
CREATE INDEX IF NOT EXISTS command_defs_created_at_idx ON command_defs (created_at DESC, id DESC);

-- Per-satellite monotonic sequence counter for uplink ordering.
CREATE TABLE IF NOT EXISTS uplink_sequences (
    satellite_id  uuid PRIMARY KEY,
    next_value    bigint NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS uplink_requests (
    id                  uuid PRIMARY KEY,
    tenant_id           uuid NOT NULL,
    satellite_id        uuid NOT NULL,
    command_def_id      uuid NOT NULL REFERENCES command_defs (id),
    parameters_json     text NOT NULL DEFAULT '',
    scheduled_release   timestamptz NOT NULL,
    status              int  NOT NULL,
    sequence_number     bigint NOT NULL,
    gateway_id          text NOT NULL DEFAULT '',
    submitted_at        timestamptz NOT NULL DEFAULT now(),
    released_at         timestamptz,
    acked_at            timestamptz,
    completed_at        timestamptz,
    error_message       text NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    created_by          text NOT NULL DEFAULT 'system',
    updated_by          text NOT NULL DEFAULT 'system',
    UNIQUE (satellite_id, sequence_number)
);

CREATE INDEX IF NOT EXISTS uplink_requests_tenant_idx     ON uplink_requests (tenant_id);
CREATE INDEX IF NOT EXISTS uplink_requests_satellite_idx  ON uplink_requests (satellite_id);
CREATE INDEX IF NOT EXISTS uplink_requests_status_idx     ON uplink_requests (status);
CREATE INDEX IF NOT EXISTS uplink_requests_release_idx    ON uplink_requests (scheduled_release ASC, id ASC);
CREATE INDEX IF NOT EXISTS uplink_requests_created_at_idx ON uplink_requests (created_at DESC, id DESC);
