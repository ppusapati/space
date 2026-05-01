-- sat-simulation schema. Reusable scenario specs + per-run records.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS scenarios (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL,
    slug         text NOT NULL,
    title        text NOT NULL,
    description  text NOT NULL DEFAULT '',
    spec_json    jsonb NOT NULL DEFAULT '{}'::jsonb,
    active       boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    created_by   text NOT NULL DEFAULT 'system',
    updated_by   text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS scenarios_tenant_idx     ON scenarios (tenant_id);
CREATE INDEX IF NOT EXISTS scenarios_created_at_idx ON scenarios (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS simulation_runs (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    satellite_id    uuid NOT NULL,
    scenario_id     uuid NOT NULL REFERENCES scenarios (id),
    mode            int  NOT NULL,
    status          int  NOT NULL,
    parameters_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    log_uri         text NOT NULL DEFAULT '',
    telemetry_uri   text NOT NULL DEFAULT '',
    results_json    jsonb NOT NULL DEFAULT '{}'::jsonb,
    score           double precision NOT NULL DEFAULT 0,
    started_at      timestamptz,
    finished_at     timestamptz,
    error_message   text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS simulation_runs_tenant_idx     ON simulation_runs (tenant_id);
CREATE INDEX IF NOT EXISTS simulation_runs_satellite_idx  ON simulation_runs (satellite_id);
CREATE INDEX IF NOT EXISTS simulation_runs_scenario_idx   ON simulation_runs (scenario_id);
CREATE INDEX IF NOT EXISTS simulation_runs_status_idx     ON simulation_runs (status);
CREATE INDEX IF NOT EXISTS simulation_runs_mode_idx       ON simulation_runs (mode);
CREATE INDEX IF NOT EXISTS simulation_runs_created_at_idx ON simulation_runs (created_at DESC, id DESC);
