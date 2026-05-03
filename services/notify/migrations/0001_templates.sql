-- 0001_templates.sql — TASK-P1-NOTIFY-001
--
-- notification_templates
--   One row per (id, version, channel). Multiple versions of the
--   same template ID can coexist — the renderer always picks the
--   highest active version. variables_schema is JSONB so adding
--   a required variable to a future template version doesn't
--   require a schema migration.
--
-- notification_preferences
--   Per-user opt-out. Absence of a row = opted in by default.
--   Mandatory templates (security flows: login, MFA change,
--   password reset) IGNORE this table — the preferences.IsAllowed
--   helper short-circuits to true when the template's mandatory
--   bit is set.

CREATE TABLE IF NOT EXISTS notification_templates (
    id               text NOT NULL,
    version          int  NOT NULL,
    channel          text NOT NULL CHECK (channel IN ('email', 'sms', 'inapp')),
    body             text NOT NULL,
    variables_schema jsonb NOT NULL DEFAULT '{"required":[]}'::jsonb,
    mandatory        boolean NOT NULL DEFAULT false,
    active           boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id, version, channel)
);

CREATE INDEX IF NOT EXISTS notification_templates_active_idx
    ON notification_templates (id, channel, version DESC) WHERE active = true;

CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id     uuid NOT NULL,
    template_id text NOT NULL,
    opted_out   boolean NOT NULL DEFAULT false,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, template_id)
);

CREATE INDEX IF NOT EXISTS notification_preferences_user_idx
    ON notification_preferences (user_id);

-- Seed the mandatory security templates so the IAM service has
-- something to render on day one.
INSERT INTO notification_templates
  (id, version, channel, body, variables_schema, mandatory)
VALUES
  ('security.login.detected', 1, 'email',
   'Hi {{display_name}},\n\nWe detected a sign-in to your Chetana account from {{client_ip}} ({{user_agent}}) at {{occurred_at}}. If this was not you, reset your password immediately.',
   '{"required":["display_name","client_ip","user_agent","occurred_at"]}'::jsonb,
   true),
  ('security.password.reset', 1, 'email',
   'Hi {{display_name}},\n\nUse the link below to reset your Chetana password (expires {{expires_at}}):\n\n{{reset_link}}',
   '{"required":["display_name","reset_link","expires_at"]}'::jsonb,
   true),
  ('security.mfa.changed', 1, 'email',
   'Hi {{display_name}},\n\nYour Chetana MFA settings were changed at {{occurred_at}} ({{change_type}}). If this was not you, contact support immediately.',
   '{"required":["display_name","occurred_at","change_type"]}'::jsonb,
   true)
ON CONFLICT (id, version, channel) DO NOTHING;
