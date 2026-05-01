/**
 * Notifications domain module.
 *
 * Surfaces every standalone List RPC across the 2 notifications services
 * (notification / template). Both services expose a `ListTemplates` RPC
 * but they are intentionally distinct surfaces:
 *
 *   - `notification.NotificationService.ListTemplates` reads from the
 *     notification module's local `notifications.notification_templates`
 *     table — used by the send-side (NotificationService.SendNotification
 *     looks up a template_id from this list before rendering).
 *   - `notifications.template.TemplateService.ListTemplates` reads from
 *     the dedicated template-management service's `notifications.templates`
 *     table — used by template authoring (versioning, translations,
 *     rendering, preview).
 *
 * The two are connected through `template_id` references but live at
 * separate proto packages and storage. Both are wired here so the
 * notification authors see the full surface their workflows touch.
 *
 * Currently NOT wired (sub-resource requiring parent_id):
 *   - TemplateService/ListTemplateVersions — needs template_id; returns
 *     400 + `code: invalid_argument` under empty context (handler
 *     correctly routes through `errors.ToConnectError`, no defect).
 *
 * Wire-shape note for notification responses:
 *   notification proto declares `int64 total_count` (not int32). JSON
 *   wire encoding is the same camelCase string key — captured here for
 *   any future native-decoder change. The response also carries `page`,
 *   `page_size`, and `unread_count` fields that are server-side
 *   computed; ListPage's loader doesn't read them today (the loader
 *   reads totalCount + rows). If a future use-case needs the unread
 *   counter on the menu, extend the loader to return it as a side
 *   channel — do NOT shoehorn it into the totalCount lookup.
 */
import type { DomainModule } from '../index.js';

export const notifications: DomainModule = {
  id: 'notifications',
  label: 'Notifications',
  entities: [
    // ---- notification (send-side) ----
    {
      slug: 'notifications',
      label: 'Notifications',
      formId: 'alert_configuration',
      listEndpoint: '/notifications.notification.api.v1.NotificationService/ListNotifications',
      responseRowsKey: 'notifications',
      responseTotalKey: 'totalCount',
      columns: ['type', 'channel', 'priority', 'recipientId', 'subject', 'status'],
    },
    {
      slug: 'notification-templates',
      label: 'Notification Templates',
      formId: 'notification_template',
      listEndpoint: '/notifications.notification.api.v1.NotificationService/ListTemplates',
      responseRowsKey: 'templates',
      responseTotalKey: 'totalCount',
      columns: ['name', 'type', 'channel', 'language', 'active'],
    },
    // ---- template (authoring/management) ----
    {
      slug: 'templates',
      label: 'Templates',
      formId: 'notification_template_management',
      listEndpoint: '/notifications.template.api.v1.TemplateService/ListTemplates',
      responseRowsKey: 'templates',
      responseTotalKey: 'totalCount',
      columns: ['code', 'name', 'channel', 'contentType', 'subject', 'description'],
    },
  ],
};
