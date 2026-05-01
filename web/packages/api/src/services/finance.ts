/**
 * Finance Service Factories
 * Typed ConnectRPC clients for ledger, AR, AP, journal, billing, tax, etc.
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

// Base service descriptors
import { PayableService } from '@samavāya/proto/gen/business/finance/payable/proto/payable_pb.js';
import { ReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/receivable_pb.js';
import { JournalEntryService } from '@samavāya/proto/gen/business/finance/journal/proto/journal_pb.js';
import { BillingService } from '@samavāya/proto/gen/business/finance/billing/proto/billing_pb.js';
import { CashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/cashmanagement_pb.js';
import { CostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/costcenter_pb.js';
import { CompliancePostingsService } from '@samavāya/proto/gen/business/finance/compliancepostings/proto/compliancepostings_pb.js';
import { FinancialCloseService } from '@samavāya/proto/gen/business/finance/financialclose/proto/financialclose_pb.js';
import { ReconciliationService } from '@samavāya/proto/gen/business/finance/reconciliation/proto/reconciliation_pb.js';
import { TaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/taxengine_pb.js';
import { TransactionService } from '@samavāya/proto/gen/business/finance/transaction/proto/transaction_pb.js';
import { AccountBalanceService } from '@samavāya/proto/gen/business/finance/ledger/proto/balance_pb.js';
import { FinancialPeriodService } from '@samavāya/proto/gen/business/finance/ledger/proto/period_pb.js';
import { TrialBalanceService } from '@samavāya/proto/gen/business/finance/reports/proto/reports_pb.js';
import { GeneralLedgerService } from '@samavāya/proto/gen/business/finance/reports/proto/reports_pb.js';
import { LedgerReportService } from '@samavāya/proto/gen/business/finance/reports/proto/reports_pb.js';

// Vertical-specific — Agriculture
// CashManagement agriculture vertical retired in Phase F.4.4 — see config/class_registry/cashmanagement.yaml.
// CostCenter agriculture vertical retired in Phase F.4.5 — see config/class_registry/costcenter.yaml.
// Journal agriculture vertical retired in Phase F.4.1 — see config/class_registry/journal.yaml.
// Payable agriculture vertical retired in Phase F.4.2 — see config/class_registry/payable.yaml.
// Receivable agriculture vertical retired in Phase F.4.3 — see config/class_registry/receivable.yaml.
// TaxEngine agriculture vertical retired in Phase F.4.6 — see config/class_registry/taxengine.yaml.
// Vertical-specific — Construction
// CashManagement construction vertical retired in Phase F.4.4 — see config/class_registry/cashmanagement.yaml.
// CostCenter construction vertical retired in Phase F.4.5 — see config/class_registry/costcenter.yaml.
// Payable construction vertical retired in Phase F.4.2 — see config/class_registry/payable.yaml.
// Receivable construction vertical retired in Phase F.4.3 — see config/class_registry/receivable.yaml.
// Journal constructionvertical retired in Phase F.4.1 — see config/class_registry/journal.yaml.
// Vertical-specific — MfgVertical (Manufacturing)
// CashManagement mfgvertical retired in Phase F.4.4 — see config/class_registry/cashmanagement.yaml.
// CostCenter mfgvertical retired in Phase F.4.5 — see config/class_registry/costcenter.yaml.
// Journal mfgvertical retired in Phase F.4.1 — see config/class_registry/journal.yaml.
// Payable mfgvertical retired in Phase F.4.2 — see config/class_registry/payable.yaml.
// Receivable mfgvertical retired in Phase F.4.3 — see config/class_registry/receivable.yaml.
// TaxEngine mfgvertical retired in Phase F.4.6 — see config/class_registry/taxengine.yaml.
// Vertical-specific — Solar
// CashManagement solar vertical retired in Phase F.4.4 — see config/class_registry/cashmanagement.yaml.
// CostCenter solar vertical retired in Phase F.4.5 — see config/class_registry/costcenter.yaml.
// Journal solar vertical retired in Phase F.4.1 — see config/class_registry/journal.yaml.
// Payable solar vertical retired in Phase F.4.2 — see config/class_registry/payable.yaml.
// Receivable solar vertical retired in Phase F.4.3 — see config/class_registry/receivable.yaml.
// TaxEngine solar vertical retired in Phase F.4.6 — see config/class_registry/taxengine.yaml.
// Vertical-specific — Water
// CashManagement water vertical retired in Phase F.4.4 — see config/class_registry/cashmanagement.yaml.
// CostCenter water vertical retired in Phase F.4.5 — see config/class_registry/costcenter.yaml.
// Journal water vertical retired in Phase F.4.1 — see config/class_registry/journal.yaml.
// Payable water vertical retired in Phase F.4.2 — see config/class_registry/payable.yaml.
// Receivable water vertical retired in Phase F.4.3 — see config/class_registry/receivable.yaml.
// TaxEngine water vertical retired in Phase F.4.6 — see config/class_registry/taxengine.yaml.
// Journal workvertical retired in Phase F.4.1 — see config/class_registry/journal.yaml.

// Re-export service descriptors
export {
  PayableService, ReceivableService, JournalEntryService, BillingService,
  CashManagementService, CostCenterService, CompliancePostingsService,
  FinancialCloseService, ReconciliationService, TaxEngineService, TransactionService,
  AccountBalanceService, FinancialPeriodService,
  TrialBalanceService, GeneralLedgerService, LedgerReportService,
};

// ─── Base Service Factories ──────────────────────────────────────────────────

/** Typed client for PayableService (vendor bills, payments, debit notes) */
export function getPayableService(): Client<typeof PayableService> {
  return getApiClient().getService(PayableService);
}

/** Typed client for ReceivableService (customer invoices, receipts, credit notes) */
export function getReceivableService(): Client<typeof ReceivableService> {
  return getApiClient().getService(ReceivableService);
}

/** Typed client for JournalEntryService (journal entries, postings) */
export function getJournalEntryService(): Client<typeof JournalEntryService> {
  return getApiClient().getService(JournalEntryService);
}

/** Typed client for BillingService (subscription billing, plans) */
export function getBillingService(): Client<typeof BillingService> {
  return getApiClient().getService(BillingService);
}

/** Typed client for CashManagementService (cash positions, forecasts) */
export function getCashManagementService(): Client<typeof CashManagementService> {
  return getApiClient().getService(CashManagementService);
}

/** Typed client for CostCenterService (cost centers, allocations) */
export function getCostCenterService(): Client<typeof CostCenterService> {
  return getApiClient().getService(CostCenterService);
}

/** Typed client for CompliancePostingsService */
export function getCompliancePostingsService(): Client<typeof CompliancePostingsService> {
  return getApiClient().getService(CompliancePostingsService);
}

/** Typed client for FinancialCloseService (period close, year-end) */
export function getFinancialCloseService(): Client<typeof FinancialCloseService> {
  return getApiClient().getService(FinancialCloseService);
}

/** Typed client for ReconciliationService (bank reconciliation) */
export function getReconciliationService(): Client<typeof ReconciliationService> {
  return getApiClient().getService(ReconciliationService);
}

/** Typed client for TaxEngineService (tax calculation, filing) */
export function getTaxEngineService(): Client<typeof TaxEngineService> {
  return getApiClient().getService(TaxEngineService);
}

/** Typed client for TransactionService */
export function getTransactionService(): Client<typeof TransactionService> {
  return getApiClient().getService(TransactionService);
}

// ─── Ledger Services ────────────────────────────────────────────────────────

/** Typed client for AccountBalanceService (account balances, snapshots) */
export function getAccountBalanceService(): Client<typeof AccountBalanceService> {
  return getApiClient().getService(AccountBalanceService);
}

/** Typed client for FinancialPeriodService (fiscal periods, year management) */
export function getFinancialPeriodService(): Client<typeof FinancialPeriodService> {
  return getApiClient().getService(FinancialPeriodService);
}

// ─── Reports Services ───────────────────────────────────────────────────────

/** Typed client for TrialBalanceService (trial balance reports) */
export function getTrialBalanceService(): Client<typeof TrialBalanceService> {
  return getApiClient().getService(TrialBalanceService);
}

/** Typed client for GeneralLedgerService (GL reports, ledger queries) */
export function getGeneralLedgerService(): Client<typeof GeneralLedgerService> {
  return getApiClient().getService(GeneralLedgerService);
}

/** Typed client for LedgerReportService (consolidated ledger reports) */
export function getLedgerReportService(): Client<typeof LedgerReportService> {
  return getApiClient().getService(LedgerReportService);
}


export {
  // CashManagement vertical re-exports retired in Phase F.4.4 — callers use CashManagementService + class attributes.
  // CostCenter vertical re-exports retired in Phase F.4.5 — callers use CostCenterService + class attributes.
  // Journal vertical re-exports retired in Phase F.4.1 — callers use JournalEntryService + class attributes.
  // Payable vertical re-exports retired in Phase F.4.2 — callers use PayableService + class attributes.
  // Receivable vertical re-exports retired in Phase F.4.3 — callers use ReceivableService + class attributes.
  // TaxEngine vertical re-exports retired in Phase F.4.6 — callers use TaxEngineService + class attributes.
};

// ─── Agriculture Vertical Factories ───

// getAgricultureCashManagementService retired in Phase F.4.4 — callers use getCashManagementService()
// with class="agri_procurement_account".

// getAgricultureCostCenterService retired in Phase F.4.5 — callers use getCostCenterService()
// with class="agri_farm_center".

// getAgricultureJournalService retired in Phase F.4.1 — callers use getJournalEntryService()
// with class="agri_procurement" (see config/class_registry/journal.yaml).

// getAgriculturePayableService retired in Phase F.4.2 — callers use getPayableService()
// with class="agri_procurement_invoice".

// getAgricultureReceivableService retired in Phase F.4.3 — callers use getReceivableService()
// with class="agri_sale_invoice".

// getAgricultureTaxEngineService retired in Phase F.4.6 — callers use getTaxEngineService()
// with class="agri_mandi_cess".

// ─── Construction Vertical Factories ───

// getConstructionCashManagementService retired in Phase F.4.4 — callers use getCashManagementService()
// with class="construction_retention_escrow" or "project_dedicated_account".

// getConstructionCostCenterService retired in Phase F.4.5 — callers use getCostCenterService()
// with class="construction_site_center".

// getConstructionPayableService retired in Phase F.4.2 — callers use getPayableService()
// with class="construction_subcontractor_bill".

// getConstructionReceivableService retired in Phase F.4.3 — callers use getReceivableService()
// with class="construction_progress_bill".

// ─── Construction Vertical Vertical Factories ───
// getConstructionProjectJournalService retired in Phase F.4.1 — callers use
// getJournalEntryService() with class="construction_progress_billing".

// ─── MfgVertical (Manufacturing) Vertical Factories ───

// getMfgVerticalCashManagementService retired in Phase F.4.4 — callers use getCashManagementService()
// with class="operating_collections" or "operating_disbursements".

// getMfgVerticalCostCenterService retired in Phase F.4.5 — callers use getCostCenterService()
// with class="manufacturing_cell_center".

// getMfgVerticalJournalService retired in Phase F.4.1 — callers use getJournalEntryService()
// with class="stock_issue" / "stock_receipt".

// getMfgVerticalPayableService retired in Phase F.4.2 — callers use getPayableService()
// with class="mfg_subcontractor_bill" or "goods_receipt_matched".

// getMfgVerticalReceivableService retired in Phase F.4.3 — callers use getReceivableService()
// with class="goods_shipment_invoice".

// getMfgVerticalTaxEngineService retired in Phase F.4.6 — callers use getTaxEngineService()
// with class="mfg_input_itc_reversal" or "gst_outward_supply".

// ─── Solar Vertical Factories ───

// getSolarCashManagementService retired in Phase F.4.4 — callers use getCashManagementService()
// with class="solar_project_dsra".

// getSolarCostCenterService retired in Phase F.4.5 — callers use getCostCenterService()
// with class="solar_plant_center".

// getSolarJournalService retired in Phase F.4.1 — callers use getJournalEntryService()
// with class="solar_ppa_invoicing".

// getSolarPayableService retired in Phase F.4.2 — callers use getPayableService()
// with class="solar_epc_bill".

// getSolarReceivableService retired in Phase F.4.3 — callers use getReceivableService()
// with class="solar_ppa_invoice".

// getSolarTaxEngineService retired in Phase F.4.6 — callers use getTaxEngineService()
// with class="solar_concessional_gst".

// ─── Water Vertical Factories ───

// getWaterCashManagementService retired in Phase F.4.4 — callers use getCashManagementService()
// with class="water_utility_collection_account".

// getWaterCostCenterService retired in Phase F.4.5 — callers use getCostCenterService()
// with class="water_zone_center".

// getWaterJournalService retired in Phase F.4.1 — callers use getJournalEntryService()
// with class="water_tariff_billing".

// getWaterPayableService retired in Phase F.4.2 — callers use getPayableService()
// with class="water_om_bill".

// getWaterReceivableService retired in Phase F.4.3 — callers use getReceivableService()
// with class="water_tariff_invoice".

// getWaterTaxEngineService retired in Phase F.4.6 — callers use getTaxEngineService()
// with class="water_utility_gst_exemption".

// ─── Work Vertical Vertical Factories ───
// getWorkOrderJournalService retired in Phase F.4.1 — callers use getJournalEntryService()
// with class="construction_progress_billing" (work orders are construction progress billings).

