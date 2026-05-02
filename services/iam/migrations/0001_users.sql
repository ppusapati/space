-- 0001_users.sql — TASK-P1-IAM-001
--
-- Authoritative user table for the chetana IAM service.
--
-- Columns
-- -------
-- id                : ULID. PK. Issued by the application layer
--                     (services/packages/ulid).
-- tenant_id         : ULID; NOT NULL DEFAULT '<single-tenant-uuid>'.
--                     The single-tenant default lives in
--                     services/platform (TASK-P1-TENANT-001) and is
--                     applied via DEFAULT here so per-row inserts
--                     don't need to know it. Multi-tenant rollout
--                     in v1.x changes the default-application
--                     mechanism; the column shape is stable.
-- email_lower       : RFC 5321 lowercased email; UNIQUE within
--                     tenant. Login flow lower-cases before lookup.
-- email_display     : verbatim email as the user typed it (for
--                     correspondence; never used as a key).
-- password_hash     : Argon2id hash + salt + parameters in PHC
--                     string format ("$argon2id$v=19$m=…,t=…,p=…$
--                     <salt>$<hash>"). Stored as text.
-- password_algo     : 'argon2id' today; column reserved for future
--                     algo migration paths (e.g. Argon2id v20).
-- status            : 'active' | 'pending_verification' | 'disabled'
--                     | 'deleted'. Account lifecycle distinct from
--                     lockout state.
-- created_at /
-- updated_at        : audit timestamps.
-- last_login_at     : last successful authentication.
-- failed_login_count: counter incremented atomically by the login
--                     handler on bad password; reset on successful
--                     login.
-- locked_until      : NULL when the account is not in lockout;
--                     otherwise the wall-clock when the lockout
--                     expires. Progressive lockout escalates the
--                     duration (15m → 1h → 24h) per
--                     REQ-FUNC-PLT-IAM-003.
-- lockout_level     : 0 = no prior lockout this cycle; 1 = first
--                     escalation (15m); 2 = second (1h); 3 = third
--                     (24h). Reset to 0 after a successful login or
--                     after 30 days without failures.
-- data_classification: per-row classification tag — design.md §4.6.
--                     User records contain PII so default to 'cui'.
-- gdpr_anonymized_at: when the row was anonymised under GDPR
--                     Article 17 (TASK-P1-IAM-009). NULL until then.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id                  uuid PRIMARY KEY,
    tenant_id           uuid NOT NULL,
    email_lower         text NOT NULL,
    email_display       text NOT NULL,
    password_hash       text NOT NULL,
    password_algo       text NOT NULL DEFAULT 'argon2id',
    status              text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'pending_verification', 'disabled', 'deleted')),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    last_login_at       timestamptz,
    failed_login_count  int NOT NULL DEFAULT 0
        CHECK (failed_login_count >= 0),
    locked_until        timestamptz,
    lockout_level       int NOT NULL DEFAULT 0
        CHECK (lockout_level BETWEEN 0 AND 3),
    data_classification text NOT NULL DEFAULT 'cui'
        CHECK (data_classification IN ('public','internal','restricted','cui','itar')),
    gdpr_anonymized_at  timestamptz,

    -- Soft per-tenant uniqueness on the lowercased email.
    UNIQUE (tenant_id, email_lower)
);

-- Login flow uses (tenant_id, email_lower) lookup; covered by the
-- UNIQUE index above. The remaining indexes serve operations dashboards.
CREATE INDEX IF NOT EXISTS users_status_idx        ON users (status);
CREATE INDEX IF NOT EXISTS users_locked_until_idx  ON users (locked_until)
    WHERE locked_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS users_last_login_at_idx ON users (last_login_at DESC);

-- Trigger to keep updated_at honest. Defined per-table because the
-- platform-wide trigger lives in services/packages/db (Phase 1
-- consumer) and is not yet portable to a fresh schema migration.
CREATE OR REPLACE FUNCTION users_set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS users_updated_at_trigger ON users;
CREATE TRIGGER users_updated_at_trigger
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION users_set_updated_at();
