/**
 * Platform domain module.
 *
 * Surfaces every standalone List RPC across the 12 platform services
 * (apigateway / barcodeqr / batch / filestorage / integration / print /
 * queue / scheduler / sla / systemsettings / webhook). Sub-resource
 * Lists requiring a parent_id (ListPartitions/ListErrors needing
 * execution_id, ListVersions/ListAccess needing file_id, ListConsumers
 * needing queue_id, ListHolidays needing calendar_id, ListPreferences
 * needing user_id, ListDeliveries needing webhook_id, ListGroupPrinters
 * needing group_id) are intentionally NOT wired — they error 400/500
 * under empty context and create broken UX.
 *
 * Every (formId, listEndpoint, responseRowsKey) triple is verified live
 * against the JWT monolith on 2026-04-29:
 *   - formId → FormService.GetFormSchema returns 200 with formDefinition
 *   - listEndpoint → returns 200 + rows array under empty-tenant context
 *   - responseRowsKey matches the proto field name of the rows array
 *
 * Form catalog observations:
 *   - The platform module form catalog has 16 forms, but only 5 have
 *     direct semantic matches to platform entities — `api_management`,
 *     `scheduled_job_management`, `backup_recovery`,
 *     `system_configuration`, `feature_flag_configuration`. The other
 *     11 (security_policies, scaling_policy, performance_tuning, ...)
 *     are operational-runbook forms with no list-RPC equivalent.
 *   - Multiple entities reuse one form_id where their categories are
 *     close enough that the create-form fields overlap. ListPage uses
 *     `columns` (explicit per entity) for the row shape, so the form's
 *     create-shape mismatch doesn't show up in list views.
 *
 * Wire-shape note for queue/ListQueues:
 *   The `Queue` proto's response uses `base_response` (not the canonical
 *   `base`) — confirmed in `queue.pb.go:ListQueuesResponse.BaseResponse`.
 *   This doesn't affect ListPage which only reads the rows-key and
 *   total-count-key, but a future refactor that consolidates response
 *   envelope handling needs to know about this drift.
 */
import type { DomainModule } from '../index.js';

export const platform: DomainModule = {
  id: 'platform',
  label: 'Platform',
  entities: [
    // ---- apigateway ----
    {
      slug: 'api-routes',
      label: 'API Routes',
      formId: 'api_management',
      listEndpoint: '/platform.apigateway.api.v1.APIGatewayService/ListRoutes',
      responseRowsKey: 'routes',
      responseTotalKey: 'totalCount',
      columns: ['id', 'routeName', 'method', 'path', 'status', 'rateLimit'],
    },
    {
      slug: 'api-keys',
      label: 'API Keys',
      formId: 'api_management',
      listEndpoint: '/platform.apigateway.api.v1.APIGatewayService/ListAPIKeys',
      responseRowsKey: 'apiKeys',
      responseTotalKey: 'totalCount',
      columns: ['keyName', 'keyPrefix', 'rateLimit', 'expiresAt', 'isActive'],
    },
    {
      slug: 'rate-limits',
      label: 'Rate Limits',
      formId: 'api_management',
      listEndpoint: '/platform.apigateway.api.v1.APIGatewayService/ListRateLimits',
      responseRowsKey: 'rateLimits',
      responseTotalKey: 'totalCount',
      columns: ['limitType', 'requestsPerSecond', 'requestsPerMinute', 'burstSize', 'isActive'],
    },
    {
      slug: 'circuit-breakers',
      label: 'Circuit Breakers',
      formId: 'api_management',
      listEndpoint: '/platform.apigateway.api.v1.APIGatewayService/ListCircuitBreakers',
      responseRowsKey: 'circuits',
      // No totalCount in proto — loader falls back to rows.length.
      columns: ['serviceName', 'state', 'failureCount', 'failureThreshold', 'openedAt'],
    },
    {
      slug: 'api-versions',
      label: 'API Versions',
      formId: 'api_management',
      listEndpoint: '/platform.apigateway.api.v1.APIGatewayService/ListVersions',
      responseRowsKey: 'versions',
      // No totalCount in proto — loader falls back to rows.length.
      columns: ['version', 'description', 'isDefault', 'isDeprecated', 'sunsetDate'],
    },
    // ---- barcodeqr ----
    {
      slug: 'qr-codes',
      label: 'QR Codes',
      formId: 'system_configuration',
      listEndpoint: '/platform.barcodeqr.api.v1.BarcodeQRService/ListQRCodes',
      responseRowsKey: 'qrCodes',
      responseTotalKey: 'totalCount',
      columns: ['shortCode', 'qrType', 'entityType', 'scanCount'],
    },
    {
      slug: 'label-templates',
      label: 'Label Templates',
      formId: 'system_configuration',
      listEndpoint: '/platform.barcodeqr.api.v1.BarcodeQRService/ListLabelTemplates',
      responseRowsKey: 'templates',
      responseTotalKey: 'totalCount',
      columns: ['templateName', 'templateType', 'widthMm', 'heightMm'],
    },
    // ---- batch ----
    {
      slug: 'batches',
      label: 'Batch Definitions',
      formId: 'scheduled_job_management',
      listEndpoint: '/platform.batch.api.v1.BatchService/ListBatches',
      responseRowsKey: 'batches',
      responseTotalKey: 'totalCount',
      columns: ['batchCode', 'batchName', 'batchType', 'description'],
    },
    {
      slug: 'batch-schedules',
      label: 'Batch Schedules',
      formId: 'scheduled_job_management',
      listEndpoint: '/platform.batch.api.v1.BatchService/ListSchedules',
      responseRowsKey: 'schedules',
      responseTotalKey: 'totalCount',
      columns: ['cronExpression', 'timezone', 'nextRun', 'lastRun', 'isActive'],
    },
    // ---- filestorage ----
    {
      slug: 'files',
      label: 'Files',
      formId: 'backup_recovery',
      listEndpoint: '/platform.filestorage.api.v1.FileStorageService/ListFiles',
      responseRowsKey: 'files',
      responseTotalKey: 'totalCount',
      columns: ['fileName', 'contentType', 'fileSize', 'storageBackend'],
    },
    {
      slug: 'folders',
      label: 'Folders',
      formId: 'backup_recovery',
      listEndpoint: '/platform.filestorage.api.v1.FileStorageService/ListFolders',
      responseRowsKey: 'folders',
      responseTotalKey: 'totalCount',
      columns: ['folderName', 'folderPath', 'fileCount', 'totalSize'],
    },
    // ---- integration ----
    {
      slug: 'integrations',
      label: 'Integrations',
      formId: 'api_management',
      listEndpoint: '/platform.integration.api.v1.IntegrationService/ListIntegrations',
      responseRowsKey: 'integrations',
      responseTotalKey: 'totalCount',
      columns: ['code', 'name', 'category', 'provider', 'baseUrl'],
    },
    // ---- print ----
    {
      slug: 'printers',
      label: 'Printers',
      formId: 'system_configuration',
      listEndpoint: '/platform.printservice.api.v1.PrintService/ListPrinters',
      responseRowsKey: 'printers',
      responseTotalKey: 'totalCount',
      columns: ['printerCode', 'printerName', 'printerType', 'location'],
    },
    {
      slug: 'print-jobs',
      label: 'Print Jobs',
      formId: 'system_configuration',
      listEndpoint: '/platform.printservice.api.v1.PrintService/ListJobs',
      responseRowsKey: 'jobs',
      responseTotalKey: 'totalCount',
      columns: ['jobNumber', 'documentType', 'documentNumber', 'documentName'],
    },
    {
      slug: 'print-templates',
      label: 'Print Templates',
      formId: 'system_configuration',
      listEndpoint: '/platform.printservice.api.v1.PrintService/ListTemplates',
      responseRowsKey: 'templates',
      responseTotalKey: 'totalCount',
      columns: ['templateCode', 'templateName', 'documentType', 'outputFormat'],
    },
    // ---- queue ----
    {
      slug: 'queues',
      label: 'Queues',
      formId: 'system_configuration',
      listEndpoint: '/platform.queue.api.v1.QueueService/ListQueues',
      responseRowsKey: 'queues',
      responseTotalKey: 'totalCount',
      columns: ['queueName', 'queueType', 'maxSize', 'visibilityTimeout'],
    },
    // ---- scheduler ----
    {
      slug: 'scheduled-jobs',
      label: 'Scheduled Jobs',
      formId: 'scheduled_job_management',
      listEndpoint: '/platform.scheduler.api.v1.SchedulerService/ListJobs',
      responseRowsKey: 'jobs',
      responseTotalKey: 'totalCount',
      columns: ['name', 'cronExpression', 'targetService', 'targetMethod', 'isActive'],
    },
    // ---- sla ----
    {
      slug: 'sla-calendars',
      label: 'SLA Calendars',
      formId: 'system_configuration',
      listEndpoint: '/platform.sla.api.v1.SLAService/ListCalendars',
      responseRowsKey: 'calendars',
      responseTotalKey: 'totalCount',
      columns: ['name', 'timezone', 'workingHoursStart', 'workingHoursEnd', 'isDefault'],
    },
    {
      slug: 'sla-rules',
      label: 'SLA Rules',
      formId: 'system_configuration',
      listEndpoint: '/platform.sla.api.v1.SLAService/ListRules',
      responseRowsKey: 'rules',
      responseTotalKey: 'totalCount',
      columns: ['name', 'entityType', 'targetType', 'durationValue', 'durationUnit'],
    },
    // ---- systemsettings ----
    {
      slug: 'settings',
      label: 'System Settings',
      formId: 'system_configuration',
      listEndpoint: '/platform.systemsettings.api.v1.SystemSettingsService/ListSettings',
      responseRowsKey: 'settings',
      responseTotalKey: 'totalCount',
      columns: ['category', 'settingKey', 'valueType', 'isEncrypted'],
    },
    {
      slug: 'feature-flags',
      label: 'Feature Flags',
      formId: 'feature_flag_configuration',
      listEndpoint: '/platform.systemsettings.api.v1.SystemSettingsService/ListFeatureFlags',
      responseRowsKey: 'flags',
      responseTotalKey: 'totalCount',
      columns: ['flagKey', 'flagName', 'flagType', 'isEnabled', 'rolloutPercentage'],
    },
    // ---- webhook ----
    {
      slug: 'webhooks',
      label: 'Webhooks',
      formId: 'api_management',
      listEndpoint: '/platform.webhook.api.v1.WebhookService/ListWebhooks',
      responseRowsKey: 'webhooks',
      responseTotalKey: 'totalCount',
      columns: ['webhookName', 'endpointUrl', 'events', 'timeoutSeconds'],
    },
  ],
};
