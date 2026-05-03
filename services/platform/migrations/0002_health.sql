-- 0002_health.sql — TASK-P1-PLT-HEALTH-001
--
-- Aggregate health state + alerting incidents.
--
-- service_health rolls up the latest /ready snapshot per
-- registered service (one row per service). The aggregator
-- UPSERTs on every poll tick.
--
-- incidents records each (service, state) transition that the
-- alerter promotes to a paging or warning event. Dedupe is on
-- (service, state) WHERE resolved_at IS NULL — a service can
-- have at most one open incident per state at a time.

CREATE TABLE IF NOT EXISTS service_health (
    service       text PRIMARY KEY,
    last_seen_at  timestamptz NOT NULL DEFAULT now(),
    last_status   text NOT NULL CHECK (last_status IN ('ok','degraded','down','unknown')),
    last_error    text NOT NULL DEFAULT '',
    error_count   bigint NOT NULL DEFAULT 0,
    success_count bigint NOT NULL DEFAULT 0,
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS service_health_status_idx
    ON service_health (last_status) WHERE last_status <> 'ok';

CREATE TABLE IF NOT EXISTS health_incidents (
    id            bigserial PRIMARY KEY,
    service       text NOT NULL,
    state         text NOT NULL CHECK (state IN ('flap','sustained_failure')),
    severity      text NOT NULL DEFAULT 'warn' CHECK (severity IN ('warn','page')),
    opened_at     timestamptz NOT NULL DEFAULT now(),
    resolved_at   timestamptz,
    transitions   int NOT NULL DEFAULT 0,
    note          text NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS health_incidents_open_idx
    ON health_incidents (service, state) WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS health_incidents_opened_idx
    ON health_incidents (opened_at DESC);

-- transitions log (rolling) — used by the flap detector. Old rows
-- are pruned by the aggregator's sweep.
CREATE TABLE IF NOT EXISTS service_transitions (
    id           bigserial PRIMARY KEY,
    service      text NOT NULL,
    from_status  text NOT NULL,
    to_status    text NOT NULL,
    transitioned_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS service_transitions_lookup_idx
    ON service_transitions (service, transitioned_at DESC);
