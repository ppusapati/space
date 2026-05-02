/**
 * Platform & Infrastructure Service Factories
 * Typed ConnectRPC clients for notifications, communication, banking, budget, audit, data, platform
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

// Notifications
import { NotificationService } from '@chetana/proto/gen/core/notifications/notification/proto/notifcation_pb.js';
import { TemplateService } from '@chetana/proto/gen/core/notifications/template/proto/template_pb.js';

// Communication
import { ChatService } from '@chetana/proto/gen/core/communication/chat/proto/chat_pb.js';
import { CurrencyService } from '@chetana/proto/gen/core/communication/currency/proto/currency_pb.js';
import { LocalizationService as I18nService } from '@chetana/proto/gen/core/communication/i18n/proto/i18n_pb.js';

// Banking
import { BankingService } from '@chetana/proto/gen/core/banking/banking/proto/banking_pb.js';
import { EInvoiceService } from '@chetana/proto/gen/core/banking/einvoice/proto/einvoice_pb.js';
import { EWayBillService } from '@chetana/proto/gen/core/banking/ewaybill/proto/ewaybill_pb.js';
import { GSTService } from '@chetana/proto/gen/core/banking/gst/proto/gst_pb.js';
import { TDSService } from '@chetana/proto/gen/core/banking/tds/proto/tds_pb.js';

// Budget
import { BudgetService } from '@chetana/proto/gen/core/budget/budget/proto/budget_pb.js';
import { BudgetVarianceService } from '@chetana/proto/gen/core/budget/budgetvariance/proto/budgetvariance_pb.js';
import { CAPEXService } from '@chetana/proto/gen/core/budget/capex/proto/capex_pb.js';
import { ForecastingService } from '@chetana/proto/gen/core/budget/forecasting/proto/forecasting_pb.js';

// Audit
import { AuditReadService, AuditWriteService, AuditExportService } from '@chetana/proto/gen/core/audit/audit/proto/audit_pb.js';
import type {
  AuditLog,
  GetEntityAuditLogsRequest,
  GetEntityAuditLogsResponse,
} from '@chetana/proto/gen/core/audit/audit/proto/audit_pb.js';

export type {
  AuditLog,
  GetEntityAuditLogsRequest,
  GetEntityAuditLogsResponse,
};
import { ChangelogService } from '@chetana/proto/gen/core/audit/changelog/proto/versioning_pb.js';
import { ComplianceService } from '@chetana/proto/gen/core/audit/compliance/proto/compliance_pb.js';
import { GDPRService } from '@chetana/proto/gen/core/audit/gdpr/proto/gdpr_pb.js';
import { RetentionService } from '@chetana/proto/gen/core/audit/retention/proto/retention_pb.js';

// Data
import { BackupDRService } from '@chetana/proto/gen/core/data/backupdr/proto/backupdr_pb.js';
import { DataArchiveService } from '@chetana/proto/gen/core/data/dataarchive/proto/dataarchive_pb.js';
import { DataBridgeService } from '@chetana/proto/gen/core/data/databridge/proto/databridge_pb.js';

// Platform
import { SchedulerService } from '@chetana/proto/gen/core/platform/scheduler/proto/scheduler_pb.js';
import { SLAService } from '@chetana/proto/gen/core/platform/sla/proto/sla_pb.js';
import { FileStorageService } from '@chetana/proto/gen/core/platform/filestorage/proto/filestorage_pb.js';
import { BarcodeQRService } from '@chetana/proto/gen/core/platform/barcodeqr/proto/barcodeqr_pb.js';
import { IntegrationService } from '@chetana/proto/gen/core/platform/integration/proto/integration_pb.js';
import { BatchService } from '@chetana/proto/gen/core/platform/batch/proto/batch_pb.js';
import { PrintService } from '@chetana/proto/gen/core/platform/print/proto/printservice_pb.js';
import { QueueService } from '@chetana/proto/gen/core/platform/queue/proto/queue_pb.js';
import { WebhookService } from '@chetana/proto/gen/core/platform/webhook/proto/webhook_pb.js';
import { SystemSettingsService } from '@chetana/proto/gen/core/platform/systemsettings/proto/systemsettings_pb.js';
import { APIGatewayService } from '@chetana/proto/gen/core/platform/apigateway/proto/apigateway_pb.js';

export {
  NotificationService, TemplateService, ChatService, CurrencyService, I18nService,
  BankingService, EInvoiceService, EWayBillService, GSTService, TDSService,
  BudgetService, BudgetVarianceService, CAPEXService, ForecastingService,
  AuditReadService, AuditWriteService, AuditExportService, ChangelogService, ComplianceService, GDPRService, RetentionService,
  BackupDRService, DataArchiveService, DataBridgeService,
  SchedulerService, SLAService, FileStorageService, BarcodeQRService, IntegrationService,
  BatchService, PrintService, QueueService, WebhookService, SystemSettingsService, APIGatewayService,
};

// ─── Notifications ───────────────────────────────────────────────────────────

export function getNotificationService(): Client<typeof NotificationService> {
  return getApiClient().getService(NotificationService);
}

export function getTemplateService(): Client<typeof TemplateService> {
  return getApiClient().getService(TemplateService);
}

// ─── Communication ───────────────────────────────────────────────────────────

export function getChatService(): Client<typeof ChatService> {
  return getApiClient().getService(ChatService);
}

export function getCurrencyService(): Client<typeof CurrencyService> {
  return getApiClient().getService(CurrencyService);
}

// ─── Banking ─────────────────────────────────────────────────────────────────

export function getBankingService(): Client<typeof BankingService> {
  return getApiClient().getService(BankingService);
}

export function getEInvoiceService(): Client<typeof EInvoiceService> {
  return getApiClient().getService(EInvoiceService);
}

export function getEWayBillService(): Client<typeof EWayBillService> {
  return getApiClient().getService(EWayBillService);
}

export function getGSTService(): Client<typeof GSTService> {
  return getApiClient().getService(GSTService);
}

export function getTDSService(): Client<typeof TDSService> {
  return getApiClient().getService(TDSService);
}

// ─── Budget ──────────────────────────────────────────────────────────────────

export function getBudgetService(): Client<typeof BudgetService> {
  return getApiClient().getService(BudgetService);
}

export function getBudgetVarianceService(): Client<typeof BudgetVarianceService> {
  return getApiClient().getService(BudgetVarianceService);
}

export function getCAPEXService(): Client<typeof CAPEXService> {
  return getApiClient().getService(CAPEXService);
}

export function getForecastingService(): Client<typeof ForecastingService> {
  return getApiClient().getService(ForecastingService);
}

// ─── Audit ───────────────────────────────────────────────────────────────────

export function getAuditReadService(): Client<typeof AuditReadService> {
  return getApiClient().getService(AuditReadService);
}

export function getAuditWriteService(): Client<typeof AuditWriteService> {
  return getApiClient().getService(AuditWriteService);
}

export function getAuditExportService(): Client<typeof AuditExportService> {
  return getApiClient().getService(AuditExportService);
}

export function getChangelogService(): Client<typeof ChangelogService> {
  return getApiClient().getService(ChangelogService);
}

export function getComplianceService(): Client<typeof ComplianceService> {
  return getApiClient().getService(ComplianceService);
}

// ─── Data ────────────────────────────────────────────────────────────────────

export function getBackupDRService(): Client<typeof BackupDRService> {
  return getApiClient().getService(BackupDRService);
}

export function getDataArchiveService(): Client<typeof DataArchiveService> {
  return getApiClient().getService(DataArchiveService);
}

export function getDataBridgeService(): Client<typeof DataBridgeService> {
  return getApiClient().getService(DataBridgeService);
}

// ─── Platform ────────────────────────────────────────────────────────────────

export function getSchedulerService(): Client<typeof SchedulerService> {
  return getApiClient().getService(SchedulerService);
}

export function getSLAService(): Client<typeof SLAService> {
  return getApiClient().getService(SLAService);
}

export function getFileStorageService(): Client<typeof FileStorageService> {
  return getApiClient().getService(FileStorageService);
}

export function getBarcodeQRService(): Client<typeof BarcodeQRService> {
  return getApiClient().getService(BarcodeQRService);
}

export function getIntegrationService(): Client<typeof IntegrationService> {
  return getApiClient().getService(IntegrationService);
}

export function getBatchService(): Client<typeof BatchService> {
  return getApiClient().getService(BatchService);
}

export function getPrintService(): Client<typeof PrintService> {
  return getApiClient().getService(PrintService);
}

export function getQueueService(): Client<typeof QueueService> {
  return getApiClient().getService(QueueService);
}

export function getWebhookService(): Client<typeof WebhookService> {
  return getApiClient().getService(WebhookService);
}

export function getSystemSettingsService(): Client<typeof SystemSettingsService> {
  return getApiClient().getService(SystemSettingsService);
}

export function getAPIGatewayService(): Client<typeof APIGatewayService> {
  return getApiClient().getService(APIGatewayService);
}

// ─── Audit (additional) ─────────────────────────────────────────────────────

export function getGDPRService(): Client<typeof GDPRService> {
  return getApiClient().getService(GDPRService);
}

export function getRetentionService(): Client<typeof RetentionService> {
  return getApiClient().getService(RetentionService);
}

// ─── Communication (additional) ─────────────────────────────────────────────

export function getI18nService(): Client<typeof I18nService> {
  return getApiClient().getService(I18nService);
}
