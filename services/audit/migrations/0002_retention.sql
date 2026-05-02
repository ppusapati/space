-- 0002_retention.sql — TASK-P1-AUDIT-002
--
-- Retention tiers per REQ-FUNC-PLT-AUDIT-005:
--
--   • Online (Postgres):  5 years
--   • Cold (S3 Glacier):  7 additional years (12 years total)
--
-- Implementation:
--
--   1. Convert audit_events to a Timescale hypertable so the
--      retention drop is cheap (drop-chunk vs row-DELETE).
--      Skipped gracefully when the timescaledb extension is not
--      installed — dev environments stay on stock Postgres and
--      the retention sweep falls back to a row-DELETE.
--
--   2. Register the 5-year retention policy via
--      add_retention_policy() when Timescale is available.
--
--   3. Create the audit_archives table the Glacier-archival job
--      writes to: one row per archived chunk, carrying the
--      S3 URI + the chain attestation envelope of the dropped
--      rows so the cold copy is independently re-verifiable.

-- 1. Hypertable conversion (best-effort).
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        PERFORM create_hypertable(
            'audit_events',
            by_range('event_time', INTERVAL '1 month'),
            if_not_exists => true
        );
        PERFORM add_retention_policy(
            'audit_events',
            drop_after => INTERVAL '5 years',
            if_not_exists => true
        );
    END IF;
END
$$;

-- 2. audit_archives — one row per archived chunk.
CREATE TABLE IF NOT EXISTS audit_archives (
    id                bigserial PRIMARY KEY,
    tenant_id         uuid NOT NULL,
    archived_at       timestamptz NOT NULL DEFAULT now(),

    -- Range the archive covers.
    range_start       timestamptz NOT NULL,
    range_end         timestamptz NOT NULL,
    first_chain_seq   bigint NOT NULL,
    last_chain_seq    bigint NOT NULL,
    row_count         bigint NOT NULL,

    -- Glacier pointer.
    s3_bucket         text NOT NULL,
    s3_key            text NOT NULL,
    s3_storage_class  text NOT NULL DEFAULT 'GLACIER',
    s3_etag           text NOT NULL DEFAULT '',
    bytes_compressed  bigint NOT NULL DEFAULT 0,

    -- Chain attestation envelope (the same shape internal/export
    -- emits) so a future restore can re-verify the cold copy.
    envelope          jsonb NOT NULL,

    UNIQUE (s3_bucket, s3_key)
);

CREATE INDEX IF NOT EXISTS audit_archives_tenant_idx
    ON audit_archives (tenant_id);
CREATE INDEX IF NOT EXISTS audit_archives_range_idx
    ON audit_archives (tenant_id, range_start, range_end);

-- Grants: the audit service writes; auditors read.
GRANT SELECT, INSERT, UPDATE ON audit_archives TO audit_writer;
GRANT USAGE, SELECT ON SEQUENCE audit_archives_id_seq TO audit_writer;
GRANT SELECT ON audit_archives TO audit_reader;
