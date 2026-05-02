-- 0004_webauthn.sql — TASK-P1-IAM-004
--
-- WebAuthn credential storage.
--
-- One row per registered credential. Multiple credentials per user
-- are normal (a user has both their phone's platform authenticator
-- and a hardware security key, say). Disabled rows stay in the
-- table for forensics; the application's User adapter filters
-- them out so they cannot satisfy an assertion.
--
-- Sign-count is the W3C §6.1.1 monotonic counter the authenticator
-- includes in every assertion. The application strictly increases
-- it on every successful assertion; if the authenticator returns
-- an equal-or-smaller value the credential is disabled with
-- disabled_reason = 'clone_detected' (REQ-FUNC-PLT-IAM-005
-- acceptance #2).

CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id                 bigserial PRIMARY KEY,
    user_id            uuid NOT NULL,
    credential_id      bytea NOT NULL UNIQUE,
    public_key         bytea NOT NULL,
    sign_count         bigint NOT NULL DEFAULT 0,
    transports         text NOT NULL DEFAULT '',
    attestation_type   text NOT NULL DEFAULT '',
    attestation_format text NOT NULL DEFAULT '',
    flags_uv           boolean NOT NULL DEFAULT false,
    flags_bs           boolean NOT NULL DEFAULT false,
    flags_be           boolean NOT NULL DEFAULT false,
    flags_up           boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL DEFAULT now(),
    last_used_at       timestamptz NOT NULL DEFAULT now(),
    disabled_at        timestamptz,
    disabled_reason    text NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS webauthn_credentials_user_idx
    ON webauthn_credentials (user_id) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS webauthn_credentials_disabled_idx
    ON webauthn_credentials (disabled_at) WHERE disabled_at IS NOT NULL;
