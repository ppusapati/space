-- gi-analytics schema. GEOINT analysis jobs.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS analysis_jobs (
    id                   uuid PRIMARY KEY,
    tenant_id            uuid NOT NULL,
    type                 int  NOT NULL,
    status               int  NOT NULL,
    input_uris           text[] NOT NULL DEFAULT '{}',
    parameters_json      jsonb NOT NULL DEFAULT '{}'::jsonb,
    output_uri           text NOT NULL DEFAULT '',
    results_summary_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    error_message        text NOT NULL DEFAULT '',
    started_at           timestamptz,
    finished_at          timestamptz,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    created_by           text NOT NULL DEFAULT 'system',
    updated_by           text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS analysis_jobs_tenant_idx     ON analysis_jobs (tenant_id);
CREATE INDEX IF NOT EXISTS analysis_jobs_status_idx     ON analysis_jobs (status);
CREATE INDEX IF NOT EXISTS analysis_jobs_type_idx       ON analysis_jobs (type);
CREATE INDEX IF NOT EXISTS analysis_jobs_created_at_idx ON analysis_jobs (created_at DESC, id DESC);
