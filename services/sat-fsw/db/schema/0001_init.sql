-- sat-fsw schema. Tracks firmware builds and deployment manifests.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS firmware_builds (
    id                   uuid PRIMARY KEY,
    tenant_id            uuid NOT NULL,
    target_platform      text NOT NULL,
    subsystem            text NOT NULL,
    version              text NOT NULL,
    git_sha              text NOT NULL,
    artefact_uri         text NOT NULL,
    artefact_size_bytes  bigint NOT NULL CHECK (artefact_size_bytes > 0),
    artefact_sha256      text NOT NULL,
    status               int NOT NULL,
    notes                text NOT NULL DEFAULT '',
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    created_by           text NOT NULL DEFAULT 'system',
    updated_by           text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, subsystem, version, git_sha)
);

CREATE INDEX IF NOT EXISTS firmware_builds_tenant_idx     ON firmware_builds (tenant_id);
CREATE INDEX IF NOT EXISTS firmware_builds_subsystem_idx  ON firmware_builds (subsystem);
CREATE INDEX IF NOT EXISTS firmware_builds_status_idx     ON firmware_builds (status);
CREATE INDEX IF NOT EXISTS firmware_builds_created_at_idx ON firmware_builds (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS deployment_manifests (
    id                uuid PRIMARY KEY,
    tenant_id         uuid NOT NULL,
    satellite_id      uuid NOT NULL,
    manifest_version  text NOT NULL,
    status            int NOT NULL,
    assignments_json  jsonb NOT NULL DEFAULT '{}'::jsonb,
    notes             text NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    created_by        text NOT NULL DEFAULT 'system',
    updated_by        text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, satellite_id, manifest_version)
);

CREATE INDEX IF NOT EXISTS deployment_manifests_tenant_idx     ON deployment_manifests (tenant_id);
CREATE INDEX IF NOT EXISTS deployment_manifests_satellite_idx  ON deployment_manifests (satellite_id);
CREATE INDEX IF NOT EXISTS deployment_manifests_status_idx     ON deployment_manifests (status);
CREATE INDEX IF NOT EXISTS deployment_manifests_created_at_idx ON deployment_manifests (created_at DESC, id DESC);
