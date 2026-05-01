/**
 * Samavaya WASM Type Definitions
 * Shared TypeScript types for WASM module interfaces
 */

// ============================================================================
// Core Types
// ============================================================================

export interface ValidationResult {
  valid: boolean;
  error?: string;
  details?: Record<string, unknown>;
}

export interface DateRange {
  start: string;
  end: string;
}

export interface FinancialYear {
  year: number;
  startDate: string;
  endDate: string;
  label: string;
}

// ============================================================================
// Tax Types
// ============================================================================

export interface GstInput {
  amount: string;
  rate: string;
  isInclusive: boolean;
  sourceState: string;
  destState: string;
  cessRate?: string;
}

export interface GstResult {
  taxableAmount: string;
  cgst: string;
  sgst: string;
  igst: string;
  utgst: string;
  cess: string;
  totalTax: string;
  totalAmount: string;
  isInterState: boolean;
  effectiveRate: string;
}

export interface GstBulkItem {
  id: string;
  amount: string;
  rate: string;
}

export interface GstBulkResult {
  items: Array<{
    id: string;
    taxableAmount: string;
    cgst: string;
    sgst: string;
    igst: string;
    totalTax: string;
    totalAmount: string;
  }>;
  summary: {
    totalTaxable: string;
    totalCgst: string;
    totalSgst: string;
    totalIgst: string;
    totalTax: string;
    grandTotal: string;
  };
}

export interface TdsResult {
  grossAmount: string;
  tdsAmount: string;
  netAmount: string;
  tdsRate: string;
  section: string;
  threshold: string;
  isAboveThreshold: boolean;
}

export interface TcsResult {
  baseAmount: string;
  tcsAmount: string;
  totalAmount: string;
  tcsRate: string;
  section: string;
  threshold: string;
  isAboveThreshold: boolean;
}

export interface HsnInfo {
  code: string;
  description: string;
  gstRate: string;
  cessRate?: string;
  chapter: string;
}

// ============================================================================
// Validation Types
// ============================================================================

export interface GstinValidationResult extends ValidationResult {
  stateCode?: string;
  stateName?: string;
  panNumber?: string;
  entityType?: string;
  checkDigit?: string;
}

export interface PanValidationResult extends ValidationResult {
  holderType?: string;
  holderTypeLabel?: string;
}

export interface IfscValidationResult extends ValidationResult {
  bankName?: string;
  bankCode?: string;
  branchType?: string;
}

export interface AadhaarValidationResult extends ValidationResult {
  maskedNumber?: string;
}

export interface PasswordValidationResult {
  valid: boolean;
  strength: string;
  score: number;
  errors: string[];
  suggestions: string[];
  hasLowercase: boolean;
  hasUppercase: boolean;
  hasDigit: boolean;
  hasSpecial: boolean;
  length: number;
}

// ============================================================================
// Barcode Types
// ============================================================================

export type BarcodeFormat =
  | 'QR'
  | 'Code128'
  | 'EAN13'
  | 'EAN8'
  | 'UPC_A'
  | 'Code39'
  | 'ITF'
  | 'DataMatrix';

export interface BarcodeOptions {
  format: BarcodeFormat;
  data: string;
  width?: number;
  height?: number;
  margin?: number;
  errorCorrection?: 'L' | 'M' | 'Q' | 'H';
  includeText?: boolean;
  backgroundColor?: string;
  foregroundColor?: string;
}

export interface BarcodeResult {
  success: boolean;
  format: string;
  data: string;
  svg?: string;
  error?: string;
}

export interface UpiQrOptions {
  payeeVpa: string;
  payeeName: string;
  amount?: string;
  transactionNote?: string;
  merchantCode?: string;
  transactionId?: string;
}

export interface GstInvoiceQrData {
  sellerGstin: string;
  buyerGstin?: string;
  invoiceNumber: string;
  invoiceDate: string;
  invoiceValue: string;
  lineItems: number;
  hsnCodes: string;
  uniqueId?: string;
}

// ============================================================================
// Ledger Types
// ============================================================================

export interface JournalLine {
  accountCode: string;
  accountName: string;
  debit: string;
  credit: string;
  narration?: string;
}

export interface JournalValidationResult {
  valid: boolean;
  totalDebit: string;
  totalCredit: string;
  difference: string;
  lineCount: number;
  errors: string[];
}

export interface AccountBalance {
  accountCode: string;
  accountName: string;
  openingDebit: string;
  openingCredit: string;
  periodDebit: string;
  periodCredit: string;
  closingDebit: string;
  closingCredit: string;
}

export interface RunningBalanceEntry {
  date: string;
  description: string;
  debit: string;
  credit: string;
  balance: string;
  balanceType: 'Dr' | 'Cr';
}

export interface AgingBucket {
  label: string;
  amount: string;
  count: number;
  percentage: string;
}

export interface AgingResult {
  buckets: AgingBucket[];
  totalOutstanding: string;
  totalOverdue: string;
  averageAge: string;
  oldestDate: string;
}

export interface ReconciliationMatch {
  bankEntry: string;
  bookEntry: string;
  confidence: number;
  matchType: 'exact' | 'partial' | 'date' | 'reference';
  difference: string;
}

export interface ReconciliationResult {
  matchedCount: number;
  unmatchedBankCount: number;
  unmatchedBookCount: number;
  matches: ReconciliationMatch[];
  bankBalance: string;
  bookBalance: string;
  difference: string;
}

export interface TrialBalanceEntry {
  accountCode: string;
  accountName: string;
  accountType: string;
  group?: string;
  openingDebit: string;
  openingCredit: string;
  periodDebit: string;
  periodCredit: string;
  closingDebit: string;
  closingCredit: string;
}

export interface TrialBalanceResult {
  entries: TrialBalanceEntry[];
  totals: {
    openingDebit: string;
    openingCredit: string;
    periodDebit: string;
    periodCredit: string;
    closingDebit: string;
    closingCredit: string;
  };
  isBalanced: boolean;
  difference: string;
  warnings: string[];
}

// ============================================================================
// Pricing Types
// ============================================================================

export type DiscountType = 'percentage' | 'amount' | 'tiered';

export interface Discount {
  discountType: DiscountType;
  value: string;
  minQuantity?: string;
  maxQuantity?: string;
  minAmount?: string;
}

export interface PriceInput {
  basePrice: string;
  quantity: string;
  discounts: Discount[];
  taxRate?: string;
  includeTax?: boolean;
}

export interface DiscountBreakdown {
  description: string;
  discountType: string;
  value: string;
  amount: string;
}

export interface PriceResult {
  basePrice: string;
  quantity: string;
  grossAmount: string;
  discountAmount: string;
  discountPercentage: string;
  netAmount: string;
  taxAmount: string;
  totalAmount: string;
  unitPriceAfterDiscount: string;
  effectiveRate: string;
  breakdown: DiscountBreakdown[];
}

export interface MarginResult {
  cost: string;
  sellingPrice: string;
  profit: string;
  marginPercentage: string;
  markupPercentage: string;
  isProfitable: boolean;
}

export interface PriceTier {
  minQty: string;
  maxQty?: string;
  price: string;
}

export interface LineTotalResult {
  grossAmount: string;
  discountAmount: string;
  netAmount: string;
  effectivePrice: string;
}

// ============================================================================
// Indian Locale Types
// ============================================================================

export interface IndianState {
  code: string;
  name: string;
  gstCode: string;
  tinCode: string;
  isUt: boolean;
}

// ============================================================================
// Payroll Types
// ============================================================================

export type TaxRegime = 'old' | 'new';

export interface IncomeTaxInput {
  grossIncome: string;
  regime: TaxRegime;
  age: number;
  deductions?: {
    section80C?: string;
    section80D?: string;
    section80E?: string;
    section80G?: string;
    section80TTA?: string;
    hra?: string;
    lta?: string;
    standardDeduction?: string;
    nps80CCD1B?: string;
    nps80CCD2?: string;
    homeLoanInterest?: string;
  };
  isSeniorCitizen?: boolean;
  isSuperSeniorCitizen?: boolean;
}

export interface IncomeTaxResult {
  grossIncome: string;
  totalDeductions: string;
  taxableIncome: string;
  taxBeforeRebate: string;
  rebate87A: string;
  taxAfterRebate: string;
  surcharge: string;
  educationCess: string;
  totalTax: string;
  effectiveRate: string;
  marginalRate: string;
  slabWiseBreakdown: Array<{
    slab: string;
    rate: string;
    taxableAmount: string;
    tax: string;
  }>;
  regime: TaxRegime;
}

export interface PfResult {
  employeeContribution: string;
  employerContribution: string;
  employerPensionContribution: string;
  employerPfContribution: string;
  totalContribution: string;
  wageForPf: string;
  isAboveCeiling: boolean;
}

export interface EsiResult {
  employeeContribution: string;
  employerContribution: string;
  totalContribution: string;
  isApplicable: boolean;
  grossWage: string;
}

export interface ProfessionalTaxResult {
  monthlyTax: string;
  annualTax: string;
  state: string;
  slabApplied: string;
}

export interface StatutoryResult {
  pf: PfResult;
  esi: EsiResult;
  professionalTax: ProfessionalTaxResult;
  lwf: {
    employeeContribution: string;
    employerContribution: string;
    frequency: string;
  };
  totalEmployeeDeduction: string;
  totalEmployerContribution: string;
}

export interface CtcBreakdown {
  annualCtc: string;
  basic: string;
  hra: string;
  specialAllowance: string;
  lta: string;
  medicalAllowance: string;
  conveyanceAllowance: string;
  employerPf: string;
  employerEsi: string;
  gratuity: string;
  bonus: string;
  variablePay: string;
  monthlyGross: string;
  annualGross: string;
}

export interface SalaryStructure {
  earnings: Array<{ component: string; monthly: string; annual: string }>;
  deductions: Array<{ component: string; monthly: string; annual: string }>;
  grossSalary: string;
  totalDeductions: string;
  netSalary: string;
  proratedAmount: string;
  workingDays: number;
  paidDays: number;
}

export interface CtcOptimization {
  originalStructure: CtcBreakdown;
  optimizedStructure: CtcBreakdown;
  taxSavings: string;
  recommendations: string[];
}

// ============================================================================
// BOM Types
// ============================================================================

export interface BomItem {
  componentId: string;
  componentName: string;
  quantity: string;
  uom: string;
  isPhantom?: boolean;
  scrapPercentage?: string;
  leadTime?: number;
  cost?: string;
}

export interface BomExplosionResult {
  productId: string;
  totalLevels: number;
  items: Array<{
    componentId: string;
    componentName: string;
    level: number;
    parentId: string;
    requiredQuantity: string;
    uom: string;
    isPhantom: boolean;
    path: string[];
  }>;
  summary: Record<string, { totalQuantity: string; uom: string }>;
}

export interface BomCostResult {
  productId: string;
  materialCost: string;
  laborCost: string;
  overheadCost: string;
  totalCost: string;
  costPerLevel: Array<{ level: number; cost: string }>;
  componentCosts: Array<{
    componentId: string;
    quantity: string;
    unitCost: string;
    extendedCost: string;
  }>;
}

export interface WhereUsedResult {
  componentId: string;
  usedIn: Array<{
    productId: string;
    productName: string;
    level: number;
    quantityPer: string;
    path: string[];
  }>;
  totalProducts: number;
}

export interface MrpResult {
  plannedOrders: Array<{
    itemId: string;
    quantity: string;
    dueDate: string;
    releaseDate: string;
    orderType: 'purchase' | 'production';
  }>;
  shortages: Array<{
    itemId: string;
    quantity: string;
    date: string;
  }>;
  projectedInventory: Record<string, Array<{ date: string; quantity: string }>>;
}

export interface ProductionSchedule {
  orders: Array<{
    orderId: string;
    productId: string;
    quantity: string;
    startDate: string;
    endDate: string;
    status: 'scheduled' | 'in_progress' | 'completed';
  }>;
  capacityUtilization: Record<string, string>;
  bottlenecks: string[];
}

// ============================================================================
// Depreciation Types
// ============================================================================

export type DepreciationMethod = 'SLM' | 'WDV' | 'DDB' | 'SYD' | 'UOP';

export interface DepreciationInput {
  cost: string;
  salvageValue: string;
  usefulLife: number;
  method: DepreciationMethod;
  rate?: string;
  totalUnits?: string;
  unitsProduced?: string[];
  startDate?: string;
  isHalfYearConvention?: boolean;
}

export interface DepreciationEntry {
  year: number;
  openingValue: string;
  depreciationAmount: string;
  accumulatedDepreciation: string;
  closingValue: string;
  rate: string;
}

export interface DepreciationSchedule {
  method: DepreciationMethod;
  cost: string;
  salvageValue: string;
  usefulLife: number;
  totalDepreciation: string;
  entries: DepreciationEntry[];
}

export interface AssetBlock {
  blockId: string;
  blockName: string;
  rate: string;
  openingWdv: string;
  assets: Array<{
    assetId: string;
    assetName: string;
    purchaseDate: string;
    cost: string;
    currentWdv: string;
  }>;
}

// ============================================================================
// Compliance Types
// ============================================================================

export interface EInvoiceData {
  version: string;
  tranDtls: {
    taxSch: string;
    supTyp: string;
    regRev: string;
    igstOnIntra: string;
  };
  docDtls: {
    typ: string;
    no: string;
    dt: string;
  };
  sellerDtls: {
    gstin: string;
    lglNm: string;
    trdNm?: string;
    addr1: string;
    addr2?: string;
    loc: string;
    pin: string;
    stcd: string;
    ph?: string;
    em?: string;
  };
  buyerDtls: {
    gstin: string;
    lglNm: string;
    trdNm?: string;
    pos: string;
    addr1: string;
    addr2?: string;
    loc: string;
    pin: string;
    stcd: string;
    ph?: string;
    em?: string;
  };
  itemList: Array<{
    slNo: string;
    prdDesc: string;
    isServc: string;
    hsnCd: string;
    qty: string;
    unit: string;
    unitPrice: string;
    totAmt: string;
    discount: string;
    preTaxVal: string;
    assAmt: string;
    gstRt: string;
    igstAmt: string;
    cgstAmt: string;
    sgstAmt: string;
    cesRt: string;
    cesAmt: string;
    totItemVal: string;
  }>;
  valDtls: {
    assVal: string;
    cgstVal: string;
    sgstVal: string;
    igstVal: string;
    cesVal: string;
    totInvVal: string;
    rndOffAmt?: string;
  };
}

export interface EInvoiceResult {
  json: string;
  valid: boolean;
  errors: string[];
  warnings: string[];
  irn?: string;
}

export interface Gstr1Data {
  gstin: string;
  fp: string;
  b2b: Array<{
    ctin: string;
    inv: Array<{
      inum: string;
      idt: string;
      val: string;
      pos: string;
      rchrg: string;
      itms: Array<{
        num: number;
        itm_det: {
          txval: string;
          rt: string;
          iamt?: string;
          camt?: string;
          samt?: string;
          csamt?: string;
        };
      }>;
    }>;
  }>;
  b2cl: Array<{
    pos: string;
    inv: Array<{
      inum: string;
      idt: string;
      val: string;
      itms: Array<{
        num: number;
        itm_det: {
          txval: string;
          rt: string;
          iamt: string;
          csamt?: string;
        };
      }>;
    }>;
  }>;
  b2cs: Array<{
    sply_ty: string;
    pos: string;
    typ: string;
    txval: string;
    rt: string;
    iamt?: string;
    camt?: string;
    samt?: string;
    csamt?: string;
  }>;
  hsn: {
    data: Array<{
      hsn_sc: string;
      desc: string;
      uqc: string;
      qty: string;
      txval: string;
      iamt?: string;
      camt?: string;
      samt?: string;
      csamt?: string;
    }>;
  };
}

export interface Gstr1ValidationResult {
  valid: boolean;
  errors: Array<{ section: string; error: string; details?: string }>;
  warnings: Array<{ section: string; warning: string }>;
  summary: {
    totalInvoices: number;
    totalValue: string;
    totalTax: string;
  };
}

export interface Gstr3bData {
  gstin: string;
  ret_period: string;
  sup_details: {
    osup_det: {
      txval: string;
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    };
    osup_zero: {
      txval: string;
      iamt: string;
      csamt: string;
    };
    osup_nil_exmp: {
      txval: string;
    };
    isup_rev: {
      txval: string;
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    };
    osup_nongst: {
      txval: string;
    };
  };
  itc_elg: {
    itc_avl: Array<{
      ty: string;
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    }>;
    itc_rev: Array<{
      ty: string;
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    }>;
    itc_net: {
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    };
    itc_inelg: Array<{
      ty: string;
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    }>;
  };
  intr_ltfee: {
    intr_details: {
      iamt: string;
      camt: string;
      samt: string;
      csamt: string;
    };
  };
}

export interface TransactionClassification {
  type: 'intrastate' | 'interstate' | 'export' | 'sez' | 'deemed_export';
  applicableTaxes: ('cgst' | 'sgst' | 'igst' | 'utgst')[];
  placeOfSupply: string;
  reverseCharge: boolean;
}

export interface HsnSummary {
  hsnCode: string;
  description: string;
  uqc: string;
  totalQuantity: string;
  totalValue: string;
  taxableValue: string;
  igst: string;
  cgst: string;
  sgst: string;
  cess: string;
}

// ============================================================================
// Offline Sync Types
// ============================================================================

export type OperationType = 'create' | 'update' | 'delete';
export type SyncStatus = 'pending' | 'syncing' | 'synced' | 'failed' | 'conflict';
export type ConflictResolutionStrategy = 'server_wins' | 'client_wins' | 'last_write_wins' | 'merge_fields' | 'manual';

export interface ChangeRecord {
  id: string;
  entityType: string;
  entityId: string;
  operation: OperationType;
  data: unknown;
  timestamp: string;
  userId: string;
  deviceId: string;
  version: number;
  syncStatus: SyncStatus;
  checksum: string;
}

export interface ConflictRecord {
  id: string;
  localChange: ChangeRecord;
  remoteChange: ChangeRecord;
  conflictType: 'update_update' | 'update_delete' | 'delete_update';
  conflictingFields: string[];
  detectedAt: string;
  resolvedAt?: string;
  resolution?: ConflictResolutionStrategy;
  resolvedData?: unknown;
}

export interface Delta {
  added: Record<string, unknown>;
  modified: Record<string, { before: unknown; after: unknown }>;
  removed: string[];
}

export interface SyncResult {
  success: boolean;
  syncedCount: number;
  failedCount: number;
  conflictCount: number;
  errors: Array<{ changeId: string; error: string }>;
  conflicts: ConflictRecord[];
  serverVersion: VersionVector;
}

export interface VersionVector {
  versions: Record<string, number>;
}

export interface MergeResult<T> {
  merged: T;
  conflicts: Array<{
    field: string;
    baseValue: unknown;
    localValue: unknown;
    remoteValue: unknown;
  }>;
  hasConflicts: boolean;
}
