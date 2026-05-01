-- eo-pipeline schema. Tracks processing-job state per scene.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS jobs (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    item_id         uuid NOT NULL,
    stage           int NOT NULL,
    status          int NOT NULL,
    parameters_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    output_uri      text NOT NULL DEFAULT '',
    error_message   text NOT NULL DEFAULT '',
    started_at      timestamptz,
    finished_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'system',
    updated_by      text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS jobs_tenant_idx        ON jobs (tenant_id);
CREATE INDEX IF NOT EXISTS jobs_item_idx          ON jobs (item_id);
CREATE INDEX IF NOT EXISTS jobs_status_idx        ON jobs (status);
CREATE INDEX IF NOT EXISTS jobs_stage_idx         ON jobs (stage);
CREATE INDEX IF NOT EXISTS jobs_created_at_idx    ON jobs (created_at DESC, id DESC);
