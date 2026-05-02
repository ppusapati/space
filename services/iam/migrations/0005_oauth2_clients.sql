-- 0005_oauth2_clients.sql — TASK-P1-IAM-005
--
-- OAuth 2.1 + OIDC tables.
--
-- oauth2_clients
-- --------------
-- One row per registered OAuth client. client_secret_hash is
-- argon2id-encoded (PHC string) for confidential clients; NULL
-- for public clients. redirect_uris is an exact-match allow-list.
-- token_endpoint_auth_method pins the channel the client is
-- expected to authenticate over at /oauth2/token.
--
-- oauth2_auth_codes
-- -----------------
-- Short-lived (10-minute TTL) authorisation codes. code_hash is
-- SHA-256 of the bearer string so a DB read does not enable
-- redemption forgery. Single-use: consumed_at is set at /token
-- redemption; subsequent presentation returns ErrAuthCodeReused.
-- code_challenge + code_challenge_method bind PKCE at issue time.

CREATE TABLE IF NOT EXISTS oauth2_clients (
    client_id                  text PRIMARY KEY,
    client_secret_hash         text,
    redirect_uris              text[] NOT NULL,
    grant_types                text[] NOT NULL,
    scopes                     text[] NOT NULL DEFAULT ARRAY[]::text[],
    token_endpoint_auth_method text NOT NULL DEFAULT 'client_secret_basic'
        CHECK (token_endpoint_auth_method IN
               ('client_secret_basic', 'client_secret_post', 'none')),
    disabled                   boolean NOT NULL DEFAULT false,
    created_at                 timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS oauth2_clients_disabled_idx
    ON oauth2_clients (disabled) WHERE disabled = true;

CREATE TABLE IF NOT EXISTS oauth2_auth_codes (
    id                    text PRIMARY KEY,
    code_hash             text NOT NULL,
    client_id             text NOT NULL REFERENCES oauth2_clients (client_id) ON DELETE CASCADE,
    user_id               uuid NOT NULL,
    tenant_id             uuid NOT NULL,
    session_id            text NOT NULL,
    redirect_uri          text NOT NULL,
    scopes                text[] NOT NULL DEFAULT ARRAY[]::text[],
    code_challenge        text NOT NULL,
    code_challenge_method text NOT NULL CHECK (code_challenge_method IN ('S256')),
    nonce                 text NOT NULL DEFAULT '',
    issued_at             timestamptz NOT NULL DEFAULT now(),
    expires_at            timestamptz NOT NULL,
    consumed_at           timestamptz
);

CREATE INDEX IF NOT EXISTS oauth2_auth_codes_client_idx
    ON oauth2_auth_codes (client_id);
CREATE INDEX IF NOT EXISTS oauth2_auth_codes_user_idx
    ON oauth2_auth_codes (user_id);
CREATE INDEX IF NOT EXISTS oauth2_auth_codes_expires_idx
    ON oauth2_auth_codes (expires_at);
CREATE INDEX IF NOT EXISTS oauth2_auth_codes_consumed_idx
    ON oauth2_auth_codes (consumed_at) WHERE consumed_at IS NOT NULL;
