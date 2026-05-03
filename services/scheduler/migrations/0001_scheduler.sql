-- 0001_scheduler.sql — TASK-P1-PLT-SCHED-001
--
-- Distributed scheduler.
--
-- jobs        : the schedule + per-job policy.
-- job_runs    : the execution history (one row per attempt).
--
-- Multiple scheduler replicas read the same `jobs` table; the
-- per-job Redis lock ensures exactly-one runner per scheduled
-- tick (acceptance #1).

CREATE TABLE IF NOT EXISTS jobs (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     uuid NOT NULL,
    name          text NOT NULL,
    description   text NOT NULL DEFAULT '',
    schedule      text NOT NULL,        -- cron expression; "" = manual-only
    timezone      text NOT NULL DEFAULT 'UTC',
    enabled       boolean NOT NULL DEFAULT true,
    timeout_s     int NOT NULL DEFAULT 60,
    retry_policy  jsonb NOT NULL DEFAULT '{"max_attempts":1,"backoff_s":0}'::jsonb,
    payload       jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_run_at   timestamptz,
    next_run_at   timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS jobs_enabled_next_idx
    ON jobs (next_run_at) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS jobs_tenant_idx
    ON jobs (tenant_id);

CREATE TABLE IF NOT EXISTS job_runs (
    id           bigserial PRIMARY KEY,
    job_id       uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    tenant_id    uuid NOT NULL,
    runner_id    text NOT NULL,
    started_at   timestamptz NOT NULL DEFAULT now(),
    finished_at  timestamptz,
    status       text NOT NULL DEFAULT 'running'
        CHECK (status IN ('running','succeeded','failed','timeout','skipped')),
    exit_code    int  NOT NULL DEFAULT 0,
    output       text NOT NULL DEFAULT '',
    error_excerpt text NOT NULL DEFAULT '',
    attempt      int  NOT NULL DEFAULT 1,
    trigger      text NOT NULL DEFAULT 'cron'  CHECK (trigger IN ('cron','manual'))
);

CREATE INDEX IF NOT EXISTS job_runs_job_idx
    ON job_runs (job_id, started_at DESC);
CREATE INDEX IF NOT EXISTS job_runs_status_idx
    ON job_runs (status) WHERE status = 'running';
