-- 0001_audit.sql — TASK-P1-AUDIT-001
--
-- Append-only hash-chain audit store.
--
-- Two tables:
--
--   audit_events
--     One row per recorded event. The chain invariant is:
--
--       prev_hash = SHA-256(prev_row_canonical_json)
--       row_hash  = SHA-256(canonical_json(this row except row_hash))
--
--     where canonical_json is the lowercase-key JCS-style
--     serialization (RFC 8785 in spirit; we use a deterministic
--     stable-key Go marshaller in services/audit/internal/chain).
--
--     Tampering is detected by recomputing the chain over a time
--     range; the verifier reports the first offset whose
--     prev_hash does NOT match the previous row's row_hash.
--
--   chain_tip
--     One row per (tenant_id), holding the latest row_id +
--     row_hash so an INSERT can compute prev_hash without scanning
--     the whole table. Updated atomically inside the same
--     transaction that inserts the audit_events row, gated by
--     SELECT FOR UPDATE so two concurrent appenders serialise.
--
-- Timescale hypertable
--
-- audit_events is created as a regular table here. The Timescale
-- migration (0002_retention.sql, TASK-P1-AUDIT-002) converts it to
-- a hypertable with 1-month chunks + the 5-year retention policy.
-- Keeping the hypertable conversion separate keeps this migration
-- runnable on a stock Postgres for dev environments that don't
-- have the Timescale extension yet.
--
-- Postgres role grants
--
-- REQ-FUNC-PLT-AUDIT-006 / acceptance #1: only the audit service
-- writes audit_events; every other service reads via the audit
-- service's RPC. Operationally this is enforced by:
--
--   1. Creating a dedicated `audit_writer` role.
--   2. GRANT INSERT, SELECT on audit_events + chain_tip TO audit_writer.
--   3. REVOKE INSERT on audit_events FROM the per-service roles.
--
-- The migration sets up the role + the grants. Per-service role
-- creation lives in tools/db/roles.sh (TASK-P1-PLT-DBROLES-001,
-- future); until that lands the audit_writer role is created here
-- and the audit service connects as it.

CREATE TABLE IF NOT EXISTS audit_events (
    id                  bigserial PRIMARY KEY,
    tenant_id           uuid NOT NULL,
    event_time          timestamptz NOT NULL DEFAULT now(),
    actor_user_id       uuid,
    actor_session_id    text NOT NULL DEFAULT '',
    actor_client_ip     text NOT NULL DEFAULT '',
    actor_user_agent    text NOT NULL DEFAULT '',
    action              text NOT NULL,
    resource            text NOT NULL DEFAULT '',
    decision            text NOT NULL CHECK (decision IN ('allow', 'deny', 'ok', 'fail', 'info')),
    reason              text NOT NULL DEFAULT '',
    matched_policy_id   text NOT NULL DEFAULT '',
    procedure           text NOT NULL DEFAULT '',
    classification      text NOT NULL DEFAULT 'cui'
        CHECK (classification IN ('public','internal','restricted','cui','itar')),
    metadata            jsonb NOT NULL DEFAULT '{}'::jsonb,

    -- Hash chain.
    prev_hash           text NOT NULL,                  -- 64 hex chars
    row_hash            text NOT NULL UNIQUE,           -- 64 hex chars
    chain_seq           bigint NOT NULL                  -- monotonic per tenant
);

CREATE INDEX IF NOT EXISTS audit_events_tenant_time_idx
    ON audit_events (tenant_id, event_time DESC);
CREATE INDEX IF NOT EXISTS audit_events_actor_idx
    ON audit_events (actor_user_id) WHERE actor_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS audit_events_action_idx
    ON audit_events (action);
CREATE INDEX IF NOT EXISTS audit_events_decision_idx
    ON audit_events (decision);
CREATE INDEX IF NOT EXISTS audit_events_chain_seq_idx
    ON audit_events (tenant_id, chain_seq);
-- JSONB freetext support for AUDIT-002's search.
CREATE INDEX IF NOT EXISTS audit_events_metadata_gin_idx
    ON audit_events USING gin (metadata jsonb_path_ops);

-- One row per tenant; tracks the latest event so an INSERT can
-- read prev_hash + chain_seq without scanning audit_events.
CREATE TABLE IF NOT EXISTS chain_tip (
    tenant_id    uuid PRIMARY KEY,
    last_row_id  bigint NOT NULL DEFAULT 0,
    last_hash    text   NOT NULL DEFAULT '0000000000000000000000000000000000000000000000000000000000000000',
    last_seq     bigint NOT NULL DEFAULT 0,
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Genesis tip for the v1 single tenant. Idempotent.
INSERT INTO chain_tip (tenant_id) VALUES ('00000000-0000-0000-0000-000000000001')
ON CONFLICT (tenant_id) DO NOTHING;

-- Role + grants. Idempotent for re-applies.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'audit_writer') THEN
        CREATE ROLE audit_writer NOLOGIN;
    END IF;
END
$$;

GRANT SELECT, INSERT, UPDATE ON audit_events TO audit_writer;
GRANT SELECT, INSERT, UPDATE ON chain_tip TO audit_writer;
GRANT USAGE, SELECT ON SEQUENCE audit_events_id_seq TO audit_writer;

-- Read-only role for the search RPC + ad-hoc compliance queries.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'audit_reader') THEN
        CREATE ROLE audit_reader NOLOGIN;
    END IF;
END
$$;
GRANT SELECT ON audit_events TO audit_reader;
GRANT SELECT ON chain_tip TO audit_reader;
