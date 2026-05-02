-- 0003_mfa.sql — TASK-P1-IAM-003
--
-- MFA enrollment storage:
--
-- mfa_totp_secrets
-- ----------------
-- One row per user (PK on user_id). secret holds the raw 160-bit
-- shared secret; verified_at is NULL during the enrollment grace
-- window and set once the user has proven possession of the secret
-- by submitting a code that the verifier accepts.
--
-- mfa_backup_codes
-- ----------------
-- One row per code in the user's recovery book (typically 10 rows).
-- code_hash is bcrypt(plain) at cost 12. prefix is the leading
-- 4 characters of the plaintext, indexed so the verifier can locate
-- candidate rows in O(log n) and bcrypt-compare only those.
-- consumed_at is NULL until the code is used; non-NULL ⇒ already
-- spent (a subsequent presentation is a reuse and rejected).
--
-- A periodic sweep removes codes whose enrollment is replaced
-- (the application deletes the old set in the same transaction it
-- inserts the new one — see services/iam/internal/mfa/store.go).

CREATE TABLE IF NOT EXISTS mfa_totp_secrets (
    user_id     uuid PRIMARY KEY,
    secret      bytea NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    verified_at timestamptz
);

CREATE INDEX IF NOT EXISTS mfa_totp_secrets_verified_idx
    ON mfa_totp_secrets (verified_at) WHERE verified_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS mfa_backup_codes (
    id          bigserial PRIMARY KEY,
    user_id     uuid NOT NULL,
    prefix      text NOT NULL,
    code_hash   bytea NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    consumed_at timestamptz
);

CREATE INDEX IF NOT EXISTS mfa_backup_codes_user_idx
    ON mfa_backup_codes (user_id);
CREATE INDEX IF NOT EXISTS mfa_backup_codes_lookup_idx
    ON mfa_backup_codes (user_id, prefix) WHERE consumed_at IS NULL;
CREATE INDEX IF NOT EXISTS mfa_backup_codes_consumed_idx
    ON mfa_backup_codes (consumed_at) WHERE consumed_at IS NOT NULL;
