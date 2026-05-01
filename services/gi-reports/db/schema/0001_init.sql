-- gi-reports schema. Templates + generated reports.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS report_templates (
    id                 uuid PRIMARY KEY,
    tenant_id          uuid NOT NULL,
    slug               text NOT NULL,
    name               text NOT NULL,
    description        text NOT NULL DEFAULT '',
    template_uri       text NOT NULL,
    format             int  NOT NULL,
    parameters_schema  text NOT NULL DEFAULT '',
    active             boolean NOT NULL DEFAULT true,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    created_by         text NOT NULL DEFAULT 'system',
    updated_by         text NOT NULL DEFAULT 'system',
    UNIQUE (tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS report_templates_tenant_idx     ON report_templates (tenant_id);
CREATE INDEX IF NOT EXISTS report_templates_format_idx     ON report_templates (format);
CREATE INDEX IF NOT EXISTS report_templates_created_at_idx ON report_templates (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS reports (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    template_id     uuid NOT NULL REFERENCES report_templates (id),
    status          int  NOT NULL,
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

CREATE INDEX IF NOT EXISTS reports_tenant_idx     ON reports (tenant_id);
CREATE INDEX IF NOT EXISTS reports_template_idx   ON reports (template_id);
CREATE INDEX IF NOT EXISTS reports_status_idx     ON reports (status);
CREATE INDEX IF NOT EXISTS reports_created_at_idx ON reports (created_at DESC, id DESC);
