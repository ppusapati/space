-- 0007_password_resets.sql — TASK-P1-IAM-008
--
-- Password reset tokens.
--
-- One row per outstanding (or recently consumed) reset token.
-- token_hash is SHA-256 of the bearer secret so a database read
-- does not enable forgery. Single-use is enforced by the
-- consumed_at column; subsequent presentation of the same token
-- returns ErrTokenAlreadyUsed.
--
-- Lifetimes are tight (1h TTL per REQ-FUNC-PLT-IAM-010) and the
-- (user_id, issued_at) index supports the per-user 3-per-hour
-- rate-limit count.

CREATE TABLE IF NOT EXISTS password_resets (
    id          text PRIMARY KEY,
    token_hash  text NOT NULL,
    user_id     uuid NOT NULL,
    issued_at   timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    consumed_at timestamptz
);

CREATE INDEX IF NOT EXISTS password_resets_user_idx
    ON password_resets (user_id);
CREATE INDEX IF NOT EXISTS password_resets_user_issued_idx
    ON password_resets (user_id, issued_at);
CREATE INDEX IF NOT EXISTS password_resets_expires_idx
    ON password_resets (expires_at);
CREATE INDEX IF NOT EXISTS password_resets_consumed_idx
    ON password_resets (consumed_at) WHERE consumed_at IS NOT NULL;
