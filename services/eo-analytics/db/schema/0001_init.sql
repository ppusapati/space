-- eo-analytics schema. Model registry + inference-job state.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS models (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL,
    name          text NOT NULL,
    version       text NOT NULL,
    task          int  NOT NULL,
    framework     text NOT NULL,
    artefact_uri  text NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    active        boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    created_by    text NOT NULL DEFAULT 'system',
    updated_by    text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, name, version)
);

CREATE INDEX IF NOT EXISTS models_tenant_idx ON models (tenant_id);
CREATE INDEX IF NOT EXISTS models_task_idx   ON models (task);

CREATE TABLE IF NOT EXISTS inference_jobs (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL,
    model_id      uuid NOT NULL REFERENCES models(id),
    item_id       uuid NOT NULL,
    status        int  NOT NULL,
    output_uri    text NOT NULL DEFAULT '',
    error_message text NOT NULL DEFAULT '',
    started_at    timestamptz,
    finished_at   timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    created_by    text NOT NULL DEFAULT 'system',
    updated_by    text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS inference_jobs_tenant_idx ON inference_jobs (tenant_id);
CREATE INDEX IF NOT EXISTS inference_jobs_model_idx  ON inference_jobs (model_id);
CREATE INDEX IF NOT EXISTS inference_jobs_status_idx ON inference_jobs (status);
