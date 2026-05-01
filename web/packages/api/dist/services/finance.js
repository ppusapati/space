/**
 * Finance Service Factories
 * Typed ConnectRPC clients for ledger, AR, AP, journal, billing, tax, etc.
 */
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
import { AgricultureCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/agriculture/cashmanagement_agriculture_pb.js';
import { AgricultureCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/agriculture/costcenter_agriculture_pb.js';
import { AgricultureJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/agriculture/journal_agriculture_pb.js';
import { AgriculturePayableService } from '@samavāya/proto/gen/business/finance/payable/proto/agriculture/payable_agriculture_pb.js';
import { AgricultureReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/agriculture/receivable_agriculture_pb.js';
import { AgricultureTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/agriculture/taxengine_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/construction/cashmanagement_construction_pb.js';
import { ConstructionCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/construction/costcenter_construction_pb.js';
import { ConstructionPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/construction/payable_construction_pb.js';
import { ConstructionReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/construction/receivable_construction_pb.js';
// Vertical-specific — Construction Vertical
import { ConstructionProjectJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/constructionvertical/journal_constructionvertical_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/mfgvertical/cashmanagement_mfgvertical_pb.js';
import { MfgVerticalCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/mfgvertical/costcenter_mfgvertical_pb.js';
import { MfgVerticalJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/mfgvertical/journal_mfgvertical_pb.js';
import { MfgVerticalPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/mfgvertical/payable_mfgvertical_pb.js';
import { MfgVerticalReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/mfgvertical/receivable_mfgvertical_pb.js';
import { MfgVerticalTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/mfgvertical/taxengine_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/solar/cashmanagement_solar_pb.js';
import { solarCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/solar/costcenter_solar_pb.js';
import { solarJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/solar/journal_solar_pb.js';
import { solarPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/solar/payable_solar_pb.js';
import { solarReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/solar/receivable_solar_pb.js';
import { solarTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/solar/taxengine_solar_pb.js';
// Vertical-specific — Water
import { WaterCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/water/cashmanagement_water_pb.js';
import { WaterCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/water/costcenter_water_pb.js';
import { WaterJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/water/journal_water_pb.js';
import { WaterPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/water/payable_water_pb.js';
import { WaterReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/water/receivable_water_pb.js';
import { WaterTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/water/taxengine_water_pb.js';
// Vertical-specific — Work Vertical
import { WorkOrderJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/workvertical/journal_workvertical_pb.js';
// Re-export service descriptors
export { PayableService, ReceivableService, JournalEntryService, BillingService, CashManagementService, CostCenterService, CompliancePostingsService, FinancialCloseService, ReconciliationService, TaxEngineService, TransactionService, AccountBalanceService, FinancialPeriodService, TrialBalanceService, GeneralLedgerService, LedgerReportService, };
// ─── Base Service Factories ──────────────────────────────────────────────────
/** Typed client for PayableService (vendor bills, payments, debit notes) */
export function getPayableService() {
    return getApiClient().getService(PayableService);
}
/** Typed client for ReceivableService (customer invoices, receipts, credit notes) */
export function getReceivableService() {
    return getApiClient().getService(ReceivableService);
}
/** Typed client for JournalEntryService (journal entries, postings) */
export function getJournalEntryService() {
    return getApiClient().getService(JournalEntryService);
}
/** Typed client for BillingService (subscription billing, plans) */
export function getBillingService() {
    return getApiClient().getService(BillingService);
}
/** Typed client for CashManagementService (cash positions, forecasts) */
export function getCashManagementService() {
    return getApiClient().getService(CashManagementService);
}
/** Typed client for CostCenterService (cost centers, allocations) */
export function getCostCenterService() {
    return getApiClient().getService(CostCenterService);
}
/** Typed client for CompliancePostingsService */
export function getCompliancePostingsService() {
    return getApiClient().getService(CompliancePostingsService);
}
/** Typed client for FinancialCloseService (period close, year-end) */
export function getFinancialCloseService() {
    return getApiClient().getService(FinancialCloseService);
}
/** Typed client for ReconciliationService (bank reconciliation) */
export function getReconciliationService() {
    return getApiClient().getService(ReconciliationService);
}
/** Typed client for TaxEngineService (tax calculation, filing) */
export function getTaxEngineService() {
    return getApiClient().getService(TaxEngineService);
}
/** Typed client for TransactionService */
export function getTransactionService() {
    return getApiClient().getService(TransactionService);
}
// ─── Ledger Services ────────────────────────────────────────────────────────
/** Typed client for AccountBalanceService (account balances, snapshots) */
export function getAccountBalanceService() {
    return getApiClient().getService(AccountBalanceService);
}
/** Typed client for FinancialPeriodService (fiscal periods, year management) */
export function getFinancialPeriodService() {
    return getApiClient().getService(FinancialPeriodService);
}
// ─── Reports Services ───────────────────────────────────────────────────────
/** Typed client for TrialBalanceService (trial balance reports) */
export function getTrialBalanceService() {
    return getApiClient().getService(TrialBalanceService);
}
/** Typed client for GeneralLedgerService (GL reports, ledger queries) */
export function getGeneralLedgerService() {
    return getApiClient().getService(GeneralLedgerService);
}
/** Typed client for LedgerReportService (consolidated ledger reports) */
export function getLedgerReportService() {
    return getApiClient().getService(LedgerReportService);
}
export { AgricultureCashManagementService, ConstructionCashManagementService, MfgVerticalCashManagementService, solarCashManagementService, WaterCashManagementService, AgricultureCostCenterService, ConstructionCostCenterService, MfgVerticalCostCenterService, solarCostCenterService, WaterCostCenterService, AgricultureJournalService, ConstructionProjectJournalService, MfgVerticalJournalService, solarJournalService, WaterJournalService, WorkOrderJournalService, AgriculturePayableService, ConstructionPayableService, MfgVerticalPayableService, solarPayableService, WaterPayableService, AgricultureReceivableService, ConstructionReceivableService, MfgVerticalReceivableService, solarReceivableService, WaterReceivableService, AgricultureTaxEngineService, MfgVerticalTaxEngineService, solarTaxEngineService, WaterTaxEngineService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureCashManagementService() {
    return getApiClient().getService(AgricultureCashManagementService);
}
export function getAgricultureCostCenterService() {
    return getApiClient().getService(AgricultureCostCenterService);
}
export function getAgricultureJournalService() {
    return getApiClient().getService(AgricultureJournalService);
}
export function getAgriculturePayableService() {
    return getApiClient().getService(AgriculturePayableService);
}
export function getAgricultureReceivableService() {
    return getApiClient().getService(AgricultureReceivableService);
}
export function getAgricultureTaxEngineService() {
    return getApiClient().getService(AgricultureTaxEngineService);
}
// ─── Construction Vertical Factories ───
export function getConstructionCashManagementService() {
    return getApiClient().getService(ConstructionCashManagementService);
}
export function getConstructionCostCenterService() {
    return getApiClient().getService(ConstructionCostCenterService);
}
export function getConstructionPayableService() {
    return getApiClient().getService(ConstructionPayableService);
}
export function getConstructionReceivableService() {
    return getApiClient().getService(ConstructionReceivableService);
}
// ─── Construction Vertical Vertical Factories ───
export function getConstructionProjectJournalService() {
    return getApiClient().getService(ConstructionProjectJournalService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalCashManagementService() {
    return getApiClient().getService(MfgVerticalCashManagementService);
}
export function getMfgVerticalCostCenterService() {
    return getApiClient().getService(MfgVerticalCostCenterService);
}
export function getMfgVerticalJournalService() {
    return getApiClient().getService(MfgVerticalJournalService);
}
export function getMfgVerticalPayableService() {
    return getApiClient().getService(MfgVerticalPayableService);
}
export function getMfgVerticalReceivableService() {
    return getApiClient().getService(MfgVerticalReceivableService);
}
export function getMfgVerticalTaxEngineService() {
    return getApiClient().getService(MfgVerticalTaxEngineService);
}
// ─── Solar Vertical Factories ───
export function getSolarCashManagementService() {
    return getApiClient().getService(solarCashManagementService);
}
export function getSolarCostCenterService() {
    return getApiClient().getService(solarCostCenterService);
}
export function getSolarJournalService() {
    return getApiClient().getService(solarJournalService);
}
export function getSolarPayableService() {
    return getApiClient().getService(solarPayableService);
}
export function getSolarReceivableService() {
    return getApiClient().getService(solarReceivableService);
}
export function getSolarTaxEngineService() {
    return getApiClient().getService(solarTaxEngineService);
}
// ─── Water Vertical Factories ───
export function getWaterCashManagementService() {
    return getApiClient().getService(WaterCashManagementService);
}
export function getWaterCostCenterService() {
    return getApiClient().getService(WaterCostCenterService);
}
export function getWaterJournalService() {
    return getApiClient().getService(WaterJournalService);
}
export function getWaterPayableService() {
    return getApiClient().getService(WaterPayableService);
}
export function getWaterReceivableService() {
    return getApiClient().getService(WaterReceivableService);
}
export function getWaterTaxEngineService() {
    return getApiClient().getService(WaterTaxEngineService);
}
// ─── Work Vertical Vertical Factories ───
export function getWorkOrderJournalService() {
    return getApiClient().getService(WorkOrderJournalService);
}
//# sourceMappingURL=finance.js.map