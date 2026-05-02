-- 0006_saml_idps.sql — TASK-P1-IAM-006
--
-- SAML 2.0 IdP registry.
--
-- One row per registered IdP. attribute_mapping is JSONB so
-- adding a new IdP-specific knob (e.g. sub-organisation routing)
-- doesn't require a schema migration.
--
-- x509_cert is the IdP's signing certificate, PEM-encoded. The
-- chetana SP uses it to verify the assertion signature
-- (REQ-FUNC-PLT-IAM-007 acceptance #2: unsigned/invalidly-signed
-- assertions are rejected).
--
-- slo_url is optional — Single Logout is not implemented in v1.

CREATE TABLE IF NOT EXISTS saml_idps (
    id                bigserial PRIMARY KEY,
    name              text NOT NULL,
    entity_id         text NOT NULL UNIQUE,
    sso_url           text NOT NULL,
    slo_url           text,
    x509_cert         bytea NOT NULL,
    attribute_mapping jsonb NOT NULL,
    disabled          boolean NOT NULL DEFAULT false,
    created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS saml_idps_disabled_idx
    ON saml_idps (disabled) WHERE disabled = true;
