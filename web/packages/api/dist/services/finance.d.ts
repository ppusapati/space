/**
 * Finance Service Factories
 * Typed ConnectRPC clients for ledger, AR, AP, journal, billing, tax, etc.
 */
import type { Client } from '@connectrpc/connect';
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
import { AgricultureCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/agriculture/cashmanagement_agriculture_pb.js';
import { AgricultureCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/agriculture/costcenter_agriculture_pb.js';
import { AgricultureJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/agriculture/journal_agriculture_pb.js';
import { AgriculturePayableService } from '@samavāya/proto/gen/business/finance/payable/proto/agriculture/payable_agriculture_pb.js';
import { AgricultureReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/agriculture/receivable_agriculture_pb.js';
import { AgricultureTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/agriculture/taxengine_agriculture_pb.js';
import { ConstructionCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/construction/cashmanagement_construction_pb.js';
import { ConstructionCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/construction/costcenter_construction_pb.js';
import { ConstructionPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/construction/payable_construction_pb.js';
import { ConstructionReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/construction/receivable_construction_pb.js';
import { ConstructionProjectJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/constructionvertical/journal_constructionvertical_pb.js';
import { MfgVerticalCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/mfgvertical/cashmanagement_mfgvertical_pb.js';
import { MfgVerticalCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/mfgvertical/costcenter_mfgvertical_pb.js';
import { MfgVerticalJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/mfgvertical/journal_mfgvertical_pb.js';
import { MfgVerticalPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/mfgvertical/payable_mfgvertical_pb.js';
import { MfgVerticalReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/mfgvertical/receivable_mfgvertical_pb.js';
import { MfgVerticalTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/mfgvertical/taxengine_mfgvertical_pb.js';
import { solarCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/solar/cashmanagement_solar_pb.js';
import { solarCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/solar/costcenter_solar_pb.js';
import { solarJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/solar/journal_solar_pb.js';
import { solarPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/solar/payable_solar_pb.js';
import { solarReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/solar/receivable_solar_pb.js';
import { solarTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/solar/taxengine_solar_pb.js';
import { WaterCashManagementService } from '@samavāya/proto/gen/business/finance/cashmanagement/proto/water/cashmanagement_water_pb.js';
import { WaterCostCenterService } from '@samavāya/proto/gen/business/finance/costcenter/proto/water/costcenter_water_pb.js';
import { WaterJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/water/journal_water_pb.js';
import { WaterPayableService } from '@samavāya/proto/gen/business/finance/payable/proto/water/payable_water_pb.js';
import { WaterReceivableService } from '@samavāya/proto/gen/business/finance/receivable/proto/water/receivable_water_pb.js';
import { WaterTaxEngineService } from '@samavāya/proto/gen/business/finance/taxengine/proto/water/taxengine_water_pb.js';
import { WorkOrderJournalService } from '@samavāya/proto/gen/business/finance/journal/proto/workvertical/journal_workvertical_pb.js';
export { PayableService, ReceivableService, JournalEntryService, BillingService, CashManagementService, CostCenterService, CompliancePostingsService, FinancialCloseService, ReconciliationService, TaxEngineService, TransactionService, AccountBalanceService, FinancialPeriodService, TrialBalanceService, GeneralLedgerService, LedgerReportService, };
/** Typed client for PayableService (vendor bills, payments, debit notes) */
export declare function getPayableService(): Client<typeof PayableService>;
/** Typed client for ReceivableService (customer invoices, receipts, credit notes) */
export declare function getReceivableService(): Client<typeof ReceivableService>;
/** Typed client for JournalEntryService (journal entries, postings) */
export declare function getJournalEntryService(): Client<typeof JournalEntryService>;
/** Typed client for BillingService (subscription billing, plans) */
export declare function getBillingService(): Client<typeof BillingService>;
/** Typed client for CashManagementService (cash positions, forecasts) */
export declare function getCashManagementService(): Client<typeof CashManagementService>;
/** Typed client for CostCenterService (cost centers, allocations) */
export declare function getCostCenterService(): Client<typeof CostCenterService>;
/** Typed client for CompliancePostingsService */
export declare function getCompliancePostingsService(): Client<typeof CompliancePostingsService>;
/** Typed client for FinancialCloseService (period close, year-end) */
export declare function getFinancialCloseService(): Client<typeof FinancialCloseService>;
/** Typed client for ReconciliationService (bank reconciliation) */
export declare function getReconciliationService(): Client<typeof ReconciliationService>;
/** Typed client for TaxEngineService (tax calculation, filing) */
export declare function getTaxEngineService(): Client<typeof TaxEngineService>;
/** Typed client for TransactionService */
export declare function getTransactionService(): Client<typeof TransactionService>;
/** Typed client for AccountBalanceService (account balances, snapshots) */
export declare function getAccountBalanceService(): Client<typeof AccountBalanceService>;
/** Typed client for FinancialPeriodService (fiscal periods, year management) */
export declare function getFinancialPeriodService(): Client<typeof FinancialPeriodService>;
/** Typed client for TrialBalanceService (trial balance reports) */
export declare function getTrialBalanceService(): Client<typeof TrialBalanceService>;
/** Typed client for GeneralLedgerService (GL reports, ledger queries) */
export declare function getGeneralLedgerService(): Client<typeof GeneralLedgerService>;
/** Typed client for LedgerReportService (consolidated ledger reports) */
export declare function getLedgerReportService(): Client<typeof LedgerReportService>;
export { AgricultureCashManagementService, ConstructionCashManagementService, MfgVerticalCashManagementService, solarCashManagementService, WaterCashManagementService, AgricultureCostCenterService, ConstructionCostCenterService, MfgVerticalCostCenterService, solarCostCenterService, WaterCostCenterService, AgricultureJournalService, ConstructionProjectJournalService, MfgVerticalJournalService, solarJournalService, WaterJournalService, WorkOrderJournalService, AgriculturePayableService, ConstructionPayableService, MfgVerticalPayableService, solarPayableService, WaterPayableService, AgricultureReceivableService, ConstructionReceivableService, MfgVerticalReceivableService, solarReceivableService, WaterReceivableService, AgricultureTaxEngineService, MfgVerticalTaxEngineService, solarTaxEngineService, WaterTaxEngineService, };
export declare function getAgricultureCashManagementService(): Client<typeof AgricultureCashManagementService>;
export declare function getAgricultureCostCenterService(): Client<typeof AgricultureCostCenterService>;
export declare function getAgricultureJournalService(): Client<typeof AgricultureJournalService>;
export declare function getAgriculturePayableService(): Client<typeof AgriculturePayableService>;
export declare function getAgricultureReceivableService(): Client<typeof AgricultureReceivableService>;
export declare function getAgricultureTaxEngineService(): Client<typeof AgricultureTaxEngineService>;
export declare function getConstructionCashManagementService(): Client<typeof ConstructionCashManagementService>;
export declare function getConstructionCostCenterService(): Client<typeof ConstructionCostCenterService>;
export declare function getConstructionPayableService(): Client<typeof ConstructionPayableService>;
export declare function getConstructionReceivableService(): Client<typeof ConstructionReceivableService>;
export declare function getConstructionProjectJournalService(): Client<typeof ConstructionProjectJournalService>;
export declare function getMfgVerticalCashManagementService(): Client<typeof MfgVerticalCashManagementService>;
export declare function getMfgVerticalCostCenterService(): Client<typeof MfgVerticalCostCenterService>;
export declare function getMfgVerticalJournalService(): Client<typeof MfgVerticalJournalService>;
export declare function getMfgVerticalPayableService(): Client<typeof MfgVerticalPayableService>;
export declare function getMfgVerticalReceivableService(): Client<typeof MfgVerticalReceivableService>;
export declare function getMfgVerticalTaxEngineService(): Client<typeof MfgVerticalTaxEngineService>;
export declare function getSolarCashManagementService(): Client<typeof solarCashManagementService>;
export declare function getSolarCostCenterService(): Client<typeof solarCostCenterService>;
export declare function getSolarJournalService(): Client<typeof solarJournalService>;
export declare function getSolarPayableService(): Client<typeof solarPayableService>;
export declare function getSolarReceivableService(): Client<typeof solarReceivableService>;
export declare function getSolarTaxEngineService(): Client<typeof solarTaxEngineService>;
export declare function getWaterCashManagementService(): Client<typeof WaterCashManagementService>;
export declare function getWaterCostCenterService(): Client<typeof WaterCostCenterService>;
export declare function getWaterJournalService(): Client<typeof WaterJournalService>;
export declare function getWaterPayableService(): Client<typeof WaterPayableService>;
export declare function getWaterReceivableService(): Client<typeof WaterReceivableService>;
export declare function getWaterTaxEngineService(): Client<typeof WaterTaxEngineService>;
export declare function getWorkOrderJournalService(): Client<typeof WorkOrderJournalService>;
//# sourceMappingURL=finance.d.ts.map