-- 0001_tenants.sql — TASK-P1-TENANT-001
--
-- Platform tenants table.
--
-- chetana ships as single-tenant in v1; one row is seeded below.
-- Every domain table across every service carries
-- `tenant_id NOT NULL DEFAULT '<single-tenant-uuid>'` so the
-- schema is forward-compatible with the multi-tenant runtime
-- that lands in v1.x.
--
-- Design notes (REQ-FUNC-PLT-TENANT-003):
--
--   • PostgreSQL Row-Level Security is intentionally NOT enabled.
--     Reasoning:
--
--       1. RLS implicitly filters rows that the application-layer
--          audit chain never sees, weakening forensic replay.
--
--       2. The single-tenant v1 deployment has no rows for RLS to
--          filter — every query already runs in the one tenant's
--          scope.
--
--       3. The `tools/lint/tenantid` static analyser gives us most
--          of the safety RLS provides (asserts every domain table
--          carries `tenant_id`) at lower operational cost and
--          without bypassing the audit hook.
--
--   • security_policy and quotas are JSONB so adding a new knob
--     does not require a schema migration. New fields MUST
--     default to a zero-value that matches the legacy behaviour
--     so pre-existing rows keep working.

CREATE TABLE IF NOT EXISTS tenants (
    id                  uuid PRIMARY KEY,
    name                text NOT NULL,
    display_name        text NOT NULL,
    status              text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'archived')),
    data_classification text NOT NULL DEFAULT 'cui'
        CHECK (data_classification IN ('public','internal','restricted','cui','itar')),
    security_policy     jsonb NOT NULL DEFAULT '{}'::jsonb,
    quotas              jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS tenants_status_idx
    ON tenants (status) WHERE status <> 'active';

-- Seed: the single tenant for the v1 deployment. Idempotent —
-- re-running the migration is safe.
--
-- The default security_policy + quotas mirror the v1 IAM defaults
-- (see services/platform/internal/tenant/store.go's
-- DefaultSecurityPolicy + DefaultQuotas).
INSERT INTO tenants
  (id, name, display_name, status, data_classification, security_policy, quotas)
VALUES
  ('00000000-0000-0000-0000-000000000001',
   'chetana',
   'Chetana Single-Tenant Deployment',
   'active',
   'cui',
   '{
      "mfa_required": false,
      "session_idle_timeout": 3600000000000,
      "session_absolute_limit": 86400000000000,
      "max_concurrent_sessions": 5,
      "password_min_length": 12,
      "password_require_mixed": true
    }'::jsonb,
   '{
      "max_users": 1000,
      "max_roles_per_user": 32,
      "max_api_requests_hour": 1000000
    }'::jsonb)
ON CONFLICT (id) DO NOTHING;
