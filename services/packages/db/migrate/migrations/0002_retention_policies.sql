-- 0002_retention_policies.sql — TASK-P0-DB-001
--
-- Defines TimescaleDB retention policies for the platform-wide hypertables
-- the spec expects:
--
--   telemetry_samples       (sat-telemetry, Phase 2)
--     • raw rows                 retained 7 days   (hot hypertable)
--     • 1-min continuous agg     retained 90 days
--     • 1-h continuous agg       retained 5 years
--   audit_events            (audit, Phase 1)
--     • online retention         5 years
--     • cold archive             pointer table (no retention here; archive
--                                worker writes to S3 Glacier per
--                                REQ-FUNC-PLT-AUDIT-003)
--   processing_job_events   (eo-pipeline, Phase 3)
--     • 1 year
--
-- These tables DO NOT YET EXIST in Phase 0. The DO blocks below check for
-- their presence and are no-ops until the owning service's per-service
-- migration creates the hypertable. Re-applying this migration after the
-- tables exist is idempotent — add_retention_policy(if_not_exists => true).
--
-- NOTE on database scope:
-- Per design.md §5.1 each service owns its own logical Postgres database.
-- This Atlas project applies its migrations to the cluster-wide
-- `postgres` database (initial connection) AND to each service database
-- via the runner's per-database loop. Retention policies attach to the
-- table's owning database. The DO blocks below therefore check for table
-- existence in the CURRENT database — they activate only when applied to
-- the database that hosts the hypertable.

-- atlas:txmode none

-- ---------------------------------------------------------------------
-- telemetry_samples — sat-telemetry database
-- ---------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'telemetry_samples'
    ) THEN
        -- Promote to hypertable if not already. Chunk interval = 1 day per
        -- design.md §5.1.
        PERFORM create_hypertable(
            'telemetry_samples',
            'sample_time',
            chunk_time_interval => INTERVAL '1 day',
            if_not_exists       => true,
            migrate_data        => true
        );

        -- Raw retention: 7 days.
        PERFORM add_retention_policy(
            'telemetry_samples',
            INTERVAL '7 days',
            if_not_exists => true
        );
    END IF;
END$$;

-- 1-minute continuous aggregate retention: 90 days.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM timescaledb_information.continuous_aggregates
        WHERE view_name = 'telemetry_samples_1min'
    ) THEN
        PERFORM add_retention_policy(
            'telemetry_samples_1min',
            INTERVAL '90 days',
            if_not_exists => true
        );
    END IF;
END$$;

-- 1-hour continuous aggregate retention: 5 years.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM timescaledb_information.continuous_aggregates
        WHERE view_name = 'telemetry_samples_1h'
    ) THEN
        PERFORM add_retention_policy(
            'telemetry_samples_1h',
            INTERVAL '5 years',
            if_not_exists => true
        );
    END IF;
END$$;

-- ---------------------------------------------------------------------
-- audit_events — audit database
-- ---------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'audit_events'
    ) THEN
        -- Promote to hypertable. Chunk interval = 1 month per design.md §5.1.
        PERFORM create_hypertable(
            'audit_events',
            'event_time',
            chunk_time_interval => INTERVAL '1 month',
            if_not_exists       => true,
            migrate_data        => true
        );

        -- Online retention: 5 years (REQ-FUNC-PLT-AUDIT-003).
        -- Cold archival to S3 Glacier handled by the audit service's
        -- archive worker (TASK-P1-AUDIT-002), not by Timescale retention.
        PERFORM add_retention_policy(
            'audit_events',
            INTERVAL '5 years',
            if_not_exists => true
        );
    END IF;
END$$;

-- ---------------------------------------------------------------------
-- processing_job_events — eo-pipeline database
-- ---------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'processing_job_events'
    ) THEN
        PERFORM create_hypertable(
            'processing_job_events',
            'event_time',
            chunk_time_interval => INTERVAL '1 day',
            if_not_exists       => true,
            migrate_data        => true
        );

        PERFORM add_retention_policy(
            'processing_job_events',
            INTERVAL '1 year',
            if_not_exists => true
        );
    END IF;
END$$;
