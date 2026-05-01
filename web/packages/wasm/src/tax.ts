/**
 * Samavaya Tax Engine - TypeScript Bindings
 * GST, TDS, TCS calculations using WASM
 */

import { loadWasmModule } from './loader';
import type {
  GstInput,
  GstResult,
  GstBulkItem,
  GstBulkResult,
  TdsResult,
  TcsResult,
  HsnInfo,
} from './types';

// Type for the raw WASM module
interface TaxEngineWasm {
  calculate_gst: (input: unknown) => unknown;
  calculate_gst_bulk: (items: unknown, sourceState: string, destState: string, isInclusive: boolean) => unknown;
  calculate_tds: (section: string, amount: string, hasPan: boolean) => unknown;
  calculate_tcs: (section: string, amount: string) => unknown;
  extract_tax_from_inclusive: (totalAmount: string, taxRate: string) => unknown;
  get_hsn_rate: (hsnCode: string) => unknown;
  validate_hsn: (hsnCode: string) => boolean;
  get_tds_section_info: (section: string) => unknown;
  get_tcs_section_info: (section: string) => unknown;
  calculate_cess: (taxableAmount: string, category: string) => string;
}

let wasmModule: TaxEngineWasm | null = null;

/**
 * Initialize the tax engine module
 */
async function ensureLoaded(): Promise<TaxEngineWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<TaxEngineWasm>('tax-engine');
  }
  return wasmModule;
}

// ============================================================================
// GST Functions
// ============================================================================

/**
 * Calculate GST for a given amount
 * @param input - GST calculation input
 * @returns GST calculation result with all components
 */
export async function calculateGst(input: GstInput): Promise<GstResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_gst(input) as GstResult;
}

/**
 * Calculate GST for multiple items in bulk
 * @param items - Array of items with amounts and rates
 * @param sourceState - Source state code
 * @param destState - Destination state code
 * @param isInclusive - Whether amounts are tax-inclusive
 * @returns Bulk GST calculation results with summary
 */
export async function calculateGstBulk(
  items: GstBulkItem[],
  sourceState: string,
  destState: string,
  isInclusive = false
): Promise<GstBulkResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_gst_bulk(items, sourceState, destState, isInclusive) as GstBulkResult;
}

/**
 * Extract tax amount from a tax-inclusive price
 * @param totalAmount - Total amount including tax
 * @param taxRate - Tax rate percentage
 * @returns Object with taxable amount and tax amount
 */
export async function extractTaxFromInclusive(
  totalAmount: string,
  taxRate: string
): Promise<{ taxableAmount: string; taxAmount: string }> {
  const wasm = await ensureLoaded();
  return wasm.extract_tax_from_inclusive(totalAmount, taxRate) as { taxableAmount: string; taxAmount: string };
}

// ============================================================================
// TDS Functions
// ============================================================================

/**
 * Calculate TDS for a given section and amount
 * @param section - TDS section code (e.g., "194A", "194C")
 * @param amount - Gross amount
 * @param hasPan - Whether deductee has PAN
 * @returns TDS calculation result
 */
export async function calculateTds(
  section: string,
  amount: string,
  hasPan = true
): Promise<TdsResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_tds(section, amount, hasPan) as TdsResult;
}

/**
 * Get TDS section information
 * @param section - TDS section code
 * @returns Section details including rate and threshold
 */
export async function getTdsSectionInfo(
  section: string
): Promise<{ section: string; description: string; rate: string; threshold: string } | null> {
  const wasm = await ensureLoaded();
  const result = wasm.get_tds_section_info(section);
  return result as { section: string; description: string; rate: string; threshold: string } | null;
}

// ============================================================================
// TCS Functions
// ============================================================================

/**
 * Calculate TCS for a given section and amount
 * @param section - TCS section code
 * @param amount - Base amount
 * @returns TCS calculation result
 */
export async function calculateTcs(
  section: string,
  amount: string
): Promise<TcsResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_tcs(section, amount) as TcsResult;
}

/**
 * Get TCS section information
 * @param section - TCS section code
 * @returns Section details including rate and threshold
 */
export async function getTcsSectionInfo(
  section: string
): Promise<{ section: string; description: string; rate: string; threshold: string } | null> {
  const wasm = await ensureLoaded();
  const result = wasm.get_tcs_section_info(section);
  return result as { section: string; description: string; rate: string; threshold: string } | null;
}

// ============================================================================
// HSN Functions
// ============================================================================

/**
 * Get GST rate for an HSN/SAC code
 * @param hsnCode - HSN or SAC code
 * @returns HSN information including rate
 */
export async function getHsnRate(hsnCode: string): Promise<HsnInfo | null> {
  const wasm = await ensureLoaded();
  return wasm.get_hsn_rate(hsnCode) as HsnInfo | null;
}

/**
 * Validate an HSN code format
 * @param hsnCode - HSN code to validate
 * @returns Whether the HSN code is valid
 */
export async function validateHsn(hsnCode: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.validate_hsn(hsnCode);
}

// ============================================================================
// Cess Functions
// ============================================================================

/**
 * Calculate compensation cess
 * @param taxableAmount - Taxable amount
 * @param category - Product category (e.g., "tobacco", "luxury_car")
 * @returns Cess amount
 */
export async function calculateCess(
  taxableAmount: string,
  category: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.calculate_cess(taxableAmount, category);
}
