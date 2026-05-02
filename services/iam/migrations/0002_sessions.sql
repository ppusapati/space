-- 0002_sessions.sql — TASK-P1-IAM-002
--
-- IAM session + refresh-token bookkeeping.
--
-- sessions
-- --------
-- One row per logical login session. The JWT carries session_id;
-- middleware on every protected RPC checks the session row's
-- revoked_at field for immediate invalidation (logout / admin
-- revoke).
--
-- refresh_tokens
-- --------------
-- One row per refresh-token credential. consumed_at NULL ⇒ valid;
-- non-NULL ⇒ already used (a subsequent presentation of the same
-- token is a reuse attempt and triggers family invalidation).
-- See services/iam/internal/token/refresh.go for the full design.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS sessions (
    id                  text PRIMARY KEY,
    user_id             uuid NOT NULL,
    tenant_id           uuid NOT NULL,
    issued_at           timestamptz NOT NULL DEFAULT now(),
    last_seen_at        timestamptz NOT NULL DEFAULT now(),
    absolute_expires_at timestamptz NOT NULL,
    idle_expires_at     timestamptz NOT NULL,
    revoked_at          timestamptz,
    revoked_by          text,
    client_ip           text NOT NULL DEFAULT '',
    user_agent          text NOT NULL DEFAULT '',
    amr                 text[] NOT NULL DEFAULT ARRAY[]::text[],
    data_classification text NOT NULL DEFAULT 'cui'
        CHECK (data_classification IN ('public','internal','restricted','cui','itar'))
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_tenant_id_idx ON sessions (tenant_id);
CREATE INDEX IF NOT EXISTS sessions_revoked_idx ON sessions (revoked_at) WHERE revoked_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS sessions_absolute_expires_idx ON sessions (absolute_expires_at);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          text PRIMARY KEY,
    token_hash  text NOT NULL,
    family_id   text NOT NULL,
    parent_id   text REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    user_id     uuid NOT NULL,
    tenant_id   uuid NOT NULL,
    session_id  text NOT NULL,
    issued_at   timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked     boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS refresh_tokens_family_idx     ON refresh_tokens (family_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx       ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_session_idx    ON refresh_tokens (session_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_idx    ON refresh_tokens (expires_at);
CREATE INDEX IF NOT EXISTS refresh_tokens_consumed_idx   ON refresh_tokens (consumed_at) WHERE consumed_at IS NOT NULL;

-- Garbage-collection: refresh_tokens older than 14 days (well past
-- the 7-day TTL) are dropped by a periodic job. The job lives in
-- the platform scheduler service (TASK-P1-PLT-SCHED-001); this
-- migration only declares the index that makes the sweep cheap.
CREATE INDEX IF NOT EXISTS refresh_tokens_gc_idx
    ON refresh_tokens (expires_at)
    WHERE consumed_at IS NULL OR consumed_at < (now() - INTERVAL '14 days');
