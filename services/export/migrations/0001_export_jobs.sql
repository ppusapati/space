-- 0001_export_jobs.sql — TASK-P1-EXPORT-001
--
-- Export job queue.
--
-- Design choices:
--
--   • Postgres-backed queue (not Kafka): jobs are typically
--     long-running (10s of seconds for a 1GB CSV) and need
--     observable state per job, not per message. The
--     `FOR UPDATE SKIP LOCKED` worker checkout gives us cheap
--     parallelism without needing a separate broker.
--
--   • Lease semantics: every checkout sets `leased_until = now() +
--     lease_ttl`. A second worker can re-claim the job once the
--     lease elapses — this is the crash-recovery path
--     (acceptance #2: crashed worker → another picks up within
--     `lease_ttl + jitter`).
--
--   • Soft retention: completed + failed jobs stay in the table
--     for `retention_after_done` so users can re-fetch the
--     presigned URL. The cleanup sweep
--     (services/export/internal/cleanup/cron.go, daily) deletes
--     expired rows AND the matching S3 objects in one txn.

CREATE TABLE IF NOT EXISTS export_jobs (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL,
    requested_by    uuid NOT NULL,
    kind            text NOT NULL,    -- "gdpr_sar" | "audit_csv" | "audit_json" | future
    payload         jsonb NOT NULL DEFAULT '{}'::jsonb,

    -- Lifecycle.
    status          text NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued','running','succeeded','failed','expired')),
    leased_by       text NOT NULL DEFAULT '',
    leased_until    timestamptz,

    attempts        int  NOT NULL DEFAULT 0,
    max_attempts    int  NOT NULL DEFAULT 5,
    last_error      text NOT NULL DEFAULT '',

    -- Output.
    s3_bucket       text NOT NULL DEFAULT '',
    s3_key          text NOT NULL DEFAULT '',
    presigned_url   text NOT NULL DEFAULT '',
    presigned_until timestamptz,
    bytes_total     bigint NOT NULL DEFAULT 0,

    -- Bookkeeping.
    enqueued_at     timestamptz NOT NULL DEFAULT now(),
    started_at      timestamptz,
    completed_at    timestamptz,
    expires_at      timestamptz NOT NULL DEFAULT now() + interval '7 days'
);

CREATE INDEX IF NOT EXISTS export_jobs_status_idx
    ON export_jobs (status) WHERE status IN ('queued', 'running');
CREATE INDEX IF NOT EXISTS export_jobs_lease_idx
    ON export_jobs (leased_until) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS export_jobs_tenant_idx
    ON export_jobs (tenant_id, enqueued_at DESC);
CREATE INDEX IF NOT EXISTS export_jobs_expiry_idx
    ON export_jobs (expires_at);
