-- 0001_extensions.sql — TASK-P0-DB-001
--
-- Enables the Postgres extensions the platform requires cluster-wide:
--
--   • timescaledb — hypertables for telemetry_samples, audit_events,
--     processing_job_events. Required by REQ-FUNC-GS-TM-002,
--     REQ-FUNC-PLT-AUDIT-003.
--   • postgis     — geometry columns for AOIs, STAC item footprints,
--     ground stations. Required by REQ-FUNC-GI-AOI-001,
--     REQ-FUNC-EO-CAT-003.
--   • pg_trgm     — trigram indexes for STAC FTS, audit search.
--     Required by REQ-FUNC-PLT-AUDIT-004 (free-text on JSONB).
--
-- All three are idempotent — re-applying this migration is a no-op.

-- atlas:txmode none
--
-- TimescaleDB requires CREATE EXTENSION outside a transaction on some
-- versions (see https://docs.timescale.com/install/latest/installation-source/).
-- The directive above tells Atlas to run this file with autocommit.

CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
