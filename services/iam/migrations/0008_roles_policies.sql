-- 0008_roles_policies.sql — TASK-P1-AUTHZ-001
--
-- RBAC + ABAC policy storage.
--
-- Tables:
--
--   roles
--     One row per named role (e.g. "operator", "mission_lead",
--     "admin"). Roles are tenant-scoped — even in single-tenant
--     mode every row carries the tenant_id for forward compat
--     with TASK-P1-TENANT-001.
--
--   role_permissions
--     Many-to-many: role_id → permission ({module}.{resource}.{action}).
--     A user holds the union of permissions across every role
--     they're granted. Wildcards are stored verbatim.
--
--   user_roles
--     Many-to-many: user_id → role_id. The login flow's role
--     projection reads this table and stamps the resulting
--     []string into the JWT's `roles` claim.
--
--   policies
--     The chetana ABAC layer. One row per rule; the loader
--     (services/iam/internal/policy/loader.go) reads the active
--     subset and projects it into authzv1.Policy. Rules are
--     evaluated highest priority first; deny-wins on ties.
--
--     Columns:
--       id                  text PK; stable identifier the audit
--                           chain references for replayability.
--       description         human-readable summary.
--       effect              CHECK ('allow' | 'deny').
--       priority            int; higher wins.
--       permission          {module}.{resource}.{action} pattern;
--                           wildcards allowed at segment boundaries.
--       roles               text[]; any-of, empty = any-role.
--       min_clearance       enum-like CHECK to keep typos out.
--       require_us_person   ITAR gate.
--       tenant              empty / '*' = every tenant.
--       notes               surfaced in audit events.
--       disabled            boolean, default false.

CREATE TABLE IF NOT EXISTS roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id    uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission text NOT NULL,
    PRIMARY KEY (role_id, permission)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id    uuid NOT NULL,
    role_id    uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at timestamptz NOT NULL DEFAULT now(),
    granted_by text NOT NULL DEFAULT '',
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS user_roles_user_idx ON user_roles (user_id);

CREATE TABLE IF NOT EXISTS policies (
    id                text PRIMARY KEY,
    description       text NOT NULL DEFAULT '',
    effect            text NOT NULL CHECK (effect IN ('allow', 'deny')),
    priority          int NOT NULL DEFAULT 0,
    permission        text NOT NULL,
    roles             text[] NOT NULL DEFAULT ARRAY[]::text[],
    min_clearance     text CHECK (min_clearance IN
                                 ('public', 'internal', 'restricted', 'cui', 'itar')),
    require_us_person boolean NOT NULL DEFAULT false,
    tenant            text NOT NULL DEFAULT '',
    notes             text NOT NULL DEFAULT '',
    disabled          boolean NOT NULL DEFAULT false,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS policies_priority_idx
    ON policies (priority DESC) WHERE NOT disabled;
CREATE INDEX IF NOT EXISTS policies_permission_idx
    ON policies (permission) WHERE NOT disabled;

-- Seed: a single super-admin allow + a default ITAR deny that
-- catches misconfigured itar resources before any allow can fire.
-- These are idempotent (ON CONFLICT DO NOTHING) so re-running the
-- migration is safe.
INSERT INTO policies (id, description, effect, priority, permission, roles, notes)
VALUES
  ('seed-superadmin-allow',
   'Super-admin role: full wildcard access. Granted only to bootstrap operators.',
   'allow', 1000, '*', ARRAY['admin'],
   'Seeded by migration 0008. Disable + replace with tenant-specific rules in production.')
ON CONFLICT (id) DO NOTHING;

INSERT INTO policies (id, description, effect, priority, permission, min_clearance, require_us_person, notes)
VALUES
  ('seed-itar-default-deny',
   'Defence-in-depth ITAR gate: any itar-classified permission requires a US person AND itar clearance.',
   'deny', 900, '*', 'itar', true,
   'Seeded by migration 0008. Catches resources tagged itar that lack their own explicit gate.')
ON CONFLICT (id) DO NOTHING;
