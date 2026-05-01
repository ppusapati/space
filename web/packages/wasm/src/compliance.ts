/**
 * Samavaya Compliance - TypeScript Bindings
 * GST compliance, e-Invoice, GSTR calculations using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';
import type {
  EInvoiceData,
  EInvoiceResult,
  Gstr1Data,
  Gstr1ValidationResult,
  Gstr3bData,
  TransactionClassification,
  HsnSummary,
} from './types';

// Type for the raw WASM module
interface ComplianceWasm {
  generate_einvoice_json: (data: unknown) => unknown;
  validate_einvoice: (jsonString: string) => unknown;
  generate_irn_hash: (invoiceNumber: string, sellerGstin: string, financialYear: string) => string;
  generate_gstr1: (invoices: unknown, period: string) => unknown;
  validate_gstr1: (gstr1Data: unknown) => unknown;
  generate_gstr3b_from_gstr1: (gstr1Data: unknown) => unknown;
  calculate_tax_liability: (gstr1Data: unknown, gstr2aData: unknown) => unknown;
  classify_transaction: (sourceState: string, destState: string, placeOfSupply: string, isSez: boolean, isExport: boolean) => unknown;
  generate_document_hash: (invoiceNumber: string, sellerGstin: string, totalAmount: string) => string;
  validate_gstin_format: (gstin: string) => unknown;
  check_einvoice_applicability: (turnover: string, transactionDate: string) => unknown;
  generate_hsn_summary: (items: unknown) => unknown;
  calculate_itc_eligibility: (invoices: unknown, blockedCategories: unknown) => unknown;
  generate_eway_bill_json: (data: unknown) => unknown;
  validate_eway_bill: (jsonString: string) => unknown;
}

let wasmModule: ComplianceWasm | null = null;

/**
 * Initialize the compliance module
 */
async function ensureLoaded(): Promise<ComplianceWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<ComplianceWasm>('compliance' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// e-Invoice Functions
// ============================================================================

/**
 * Generate e-Invoice JSON as per GST schema
 * @param data - e-Invoice input data
 * @returns e-Invoice JSON and validation status
 */
export async function generateEinvoiceJson(data: EInvoiceData): Promise<EInvoiceResult> {
  const wasm = await ensureLoaded();
  return wasm.generate_einvoice_json(data) as EInvoiceResult;
}

/**
 * Validate e-Invoice JSON structure
 * @param jsonString - e-Invoice JSON string
 * @returns Validation result with errors if any
 */
export async function validateEinvoice(
  jsonString: string
): Promise<{ valid: boolean; errors: string[]; warnings: string[] }> {
  const wasm = await ensureLoaded();
  return wasm.validate_einvoice(jsonString) as { valid: boolean; errors: string[]; warnings: string[] };
}

/**
 * Generate IRN hash for e-Invoice
 * @param invoiceNumber - Invoice number
 * @param sellerGstin - Seller GSTIN
 * @param financialYear - Financial year (e.g., "2024-25")
 * @returns IRN hash
 */
export async function generateIrnHash(
  invoiceNumber: string,
  sellerGstin: string,
  financialYear: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_irn_hash(invoiceNumber, sellerGstin, financialYear);
}

/**
 * Check e-Invoice applicability based on turnover
 * @param turnover - Aggregate turnover
 * @param transactionDate - Transaction date
 * @returns Applicability details
 */
export async function checkEinvoiceApplicability(
  turnover: string,
  transactionDate: string
): Promise<{ applicable: boolean; threshold: string; effectiveFrom: string }> {
  const wasm = await ensureLoaded();
  return wasm.check_einvoice_applicability(turnover, transactionDate) as {
    applicable: boolean;
    threshold: string;
    effectiveFrom: string;
  };
}

// ============================================================================
// GSTR-1 Functions
// ============================================================================

/**
 * Generate GSTR-1 data from invoices
 * @param invoices - Array of invoice data
 * @param period - Return period (e.g., "012024" for Jan 2024)
 * @returns GSTR-1 data structure
 */
export async function generateGstr1(
  invoices: unknown[],
  period: string
): Promise<Gstr1Data> {
  const wasm = await ensureLoaded();
  return wasm.generate_gstr1(invoices, period) as Gstr1Data;
}

/**
 * Validate GSTR-1 data
 * @param gstr1Data - GSTR-1 data to validate
 * @returns Validation result
 */
export async function validateGstr1(gstr1Data: Gstr1Data): Promise<Gstr1ValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_gstr1(gstr1Data) as Gstr1ValidationResult;
}

// ============================================================================
// GSTR-3B Functions
// ============================================================================

/**
 * Generate GSTR-3B from GSTR-1 data
 * @param gstr1Data - GSTR-1 data
 * @returns GSTR-3B data
 */
export async function generateGstr3bFromGstr1(gstr1Data: Gstr1Data): Promise<Gstr3bData> {
  const wasm = await ensureLoaded();
  return wasm.generate_gstr3b_from_gstr1(gstr1Data) as Gstr3bData;
}

/**
 * Calculate tax liability
 * @param gstr1Data - GSTR-1 (outward supplies)
 * @param gstr2aData - GSTR-2A (inward supplies from suppliers)
 * @returns Tax liability calculation
 */
export async function calculateTaxLiability(
  gstr1Data: Gstr1Data,
  gstr2aData: unknown
): Promise<{
  outputTax: { igst: string; cgst: string; sgst: string; cess: string };
  inputTax: { igst: string; cgst: string; sgst: string; cess: string };
  netLiability: { igst: string; cgst: string; sgst: string; cess: string };
  cashPayable: { igst: string; cgst: string; sgst: string; cess: string };
}> {
  const wasm = await ensureLoaded();
  return wasm.calculate_tax_liability(gstr1Data, gstr2aData) as {
    outputTax: { igst: string; cgst: string; sgst: string; cess: string };
    inputTax: { igst: string; cgst: string; sgst: string; cess: string };
    netLiability: { igst: string; cgst: string; sgst: string; cess: string };
    cashPayable: { igst: string; cgst: string; sgst: string; cess: string };
  };
}

// ============================================================================
// Transaction Classification
// ============================================================================

/**
 * Classify a transaction (interstate/intrastate/export/SEZ)
 * @param sourceState - Source state code
 * @param destState - Destination state code
 * @param placeOfSupply - Place of supply code
 * @param isSez - Whether destination is SEZ
 * @param isExport - Whether it's an export
 * @returns Transaction classification
 */
export async function classifyTransaction(
  sourceState: string,
  destState: string,
  placeOfSupply: string,
  isSez = false,
  isExport = false
): Promise<TransactionClassification> {
  const wasm = await ensureLoaded();
  return wasm.classify_transaction(sourceState, destState, placeOfSupply, isSez, isExport) as TransactionClassification;
}

/**
 * Generate document hash for QR code
 * @param invoiceNumber - Invoice number
 * @param sellerGstin - Seller GSTIN
 * @param totalAmount - Total invoice amount
 * @returns Document hash
 */
export async function generateDocumentHash(
  invoiceNumber: string,
  sellerGstin: string,
  totalAmount: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_document_hash(invoiceNumber, sellerGstin, totalAmount);
}

/**
 * Validate GSTIN format
 * @param gstin - GSTIN to validate
 * @returns Validation result with details
 */
export async function validateGstinFormat(
  gstin: string
): Promise<{ valid: boolean; stateCode: string; panNumber: string; entityType: string; checkDigit: string }> {
  const wasm = await ensureLoaded();
  return wasm.validate_gstin_format(gstin) as {
    valid: boolean;
    stateCode: string;
    panNumber: string;
    entityType: string;
    checkDigit: string;
  };
}

// ============================================================================
// HSN Summary
// ============================================================================

/**
 * Generate HSN-wise summary for GSTR-1
 * @param items - Invoice line items
 * @returns HSN summary
 */
export async function generateHsnSummary(
  items: Array<{
    hsnCode: string;
    quantity: string;
    uom: string;
    taxableValue: string;
    igst: string;
    cgst: string;
    sgst: string;
    cess: string;
  }>
): Promise<HsnSummary[]> {
  const wasm = await ensureLoaded();
  return wasm.generate_hsn_summary(items) as HsnSummary[];
}

// ============================================================================
// ITC Functions
// ============================================================================

/**
 * Calculate ITC eligibility
 * @param invoices - Purchase invoices
 * @param blockedCategories - List of blocked ITC categories
 * @returns ITC eligibility calculation
 */
export async function calculateItcEligibility(
  invoices: unknown[],
  blockedCategories: string[]
): Promise<{
  eligibleItc: { igst: string; cgst: string; sgst: string; cess: string };
  blockedItc: { igst: string; cgst: string; sgst: string; cess: string };
  reversalRequired: { igst: string; cgst: string; sgst: string; cess: string };
  reasons: Array<{ invoiceId: string; reason: string; amount: string }>;
}> {
  const wasm = await ensureLoaded();
  return wasm.calculate_itc_eligibility(invoices, blockedCategories) as {
    eligibleItc: { igst: string; cgst: string; sgst: string; cess: string };
    blockedItc: { igst: string; cgst: string; sgst: string; cess: string };
    reversalRequired: { igst: string; cgst: string; sgst: string; cess: string };
    reasons: Array<{ invoiceId: string; reason: string; amount: string }>;
  };
}

// ============================================================================
// E-Way Bill Functions
// ============================================================================

/**
 * Generate E-Way Bill JSON
 * @param data - E-Way Bill input data
 * @returns E-Way Bill JSON
 */
export async function generateEwayBillJson(data: {
  supplyType: string;
  subSupplyType: string;
  docType: string;
  docNo: string;
  docDate: string;
  fromGstin: string;
  fromTradeName: string;
  fromAddr1: string;
  fromPlace: string;
  fromPincode: string;
  fromStateCode: string;
  toGstin: string;
  toTradeName: string;
  toAddr1: string;
  toPlace: string;
  toPincode: string;
  toStateCode: string;
  totalValue: string;
  cgstValue: string;
  sgstValue: string;
  igstValue: string;
  cessValue: string;
  transMode: string;
  transDistance: string;
  transporterId?: string;
  transporterName?: string;
  vehicleNo?: string;
  vehicleType?: string;
  items: Array<{
    productName: string;
    productDesc: string;
    hsnCode: string;
    quantity: string;
    qtyUnit: string;
    taxableAmount: string;
    cgstRate: string;
    sgstRate: string;
    igstRate: string;
    cessRate: string;
  }>;
}): Promise<{ json: string; valid: boolean; errors: string[] }> {
  const wasm = await ensureLoaded();
  return wasm.generate_eway_bill_json(data) as { json: string; valid: boolean; errors: string[] };
}

/**
 * Validate E-Way Bill JSON
 * @param jsonString - E-Way Bill JSON string
 * @returns Validation result
 */
export async function validateEwayBill(
  jsonString: string
): Promise<{ valid: boolean; errors: string[]; warnings: string[] }> {
  const wasm = await ensureLoaded();
  return wasm.validate_eway_bill(jsonString) as { valid: boolean; errors: string[]; warnings: string[] };
}
