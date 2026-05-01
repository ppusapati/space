/**
 * Samavaya Core - TypeScript Bindings
 * Shared utilities, Indian locale, number formatting using WASM
 */

import { loadWasmModule } from './loader';
import type { IndianState, FinancialYear } from './types';

// Type for the raw WASM module
interface CoreWasm {
  get_state_by_code: (code: string) => unknown;
  get_state_by_gst_code: (gstCode: string) => unknown;
  get_all_states: () => unknown;
  is_intra_state: (sourceState: string, destState: string) => boolean;
  is_union_territory: (stateCode: string) => boolean;
  format_indian_number: (amount: string) => string;
  amount_to_words: (amount: string) => string;
  add_decimals: (a: string, b: string) => string;
  subtract_decimals: (a: string, b: string) => string;
  multiply_decimals: (a: string, b: string) => string;
  divide_decimals: (a: string, b: string, scale?: number) => string;
  round_decimal: (value: string, scale: number) => string;
  compare_decimals: (a: string, b: string) => number;
}

let wasmModule: CoreWasm | null = null;

/**
 * Initialize the core module
 */
async function ensureLoaded(): Promise<CoreWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<CoreWasm>('core');
  }
  return wasmModule;
}

// ============================================================================
// Indian State Functions
// ============================================================================

/**
 * Get state information by state code
 * @param code - 2-letter state code (e.g., 'MH', 'KA')
 * @returns State information or null
 */
export async function getStateByCode(code: string): Promise<IndianState | null> {
  const wasm = await ensureLoaded();
  return wasm.get_state_by_code(code) as IndianState | null;
}

/**
 * Get state information by GST code
 * @param gstCode - 2-digit GST code (e.g., '27', '29')
 * @returns State information or null
 */
export async function getStateByGstCode(gstCode: string): Promise<IndianState | null> {
  const wasm = await ensureLoaded();
  return wasm.get_state_by_gst_code(gstCode) as IndianState | null;
}

/**
 * Get all Indian states and union territories
 * @returns Array of all states
 */
export async function getAllStates(): Promise<IndianState[]> {
  const wasm = await ensureLoaded();
  return wasm.get_all_states() as IndianState[];
}

/**
 * Check if transaction is intra-state (same state)
 * @param sourceState - Source state code
 * @param destState - Destination state code
 * @returns Whether transaction is intra-state
 */
export async function isIntraState(
  sourceState: string,
  destState: string
): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_intra_state(sourceState, destState);
}

/**
 * Check if state is a Union Territory
 * @param stateCode - State code
 * @returns Whether state is a UT
 */
export async function isUnionTerritory(stateCode: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_union_territory(stateCode);
}

// ============================================================================
// Indian Number Formatting
// ============================================================================

/**
 * Format number in Indian numbering system (lakhs, crores)
 * @param amount - Amount to format
 * @returns Formatted string (e.g., "1,23,45,678.00")
 */
export async function formatIndianNumber(amount: string | number): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_indian_number(String(amount));
}

/**
 * Convert amount to words (Indian format)
 * @param amount - Amount to convert
 * @returns Amount in words (e.g., "One lakh twenty three thousand four hundred fifty six only")
 */
export async function amountToWords(amount: string | number): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.amount_to_words(String(amount));
}

// ============================================================================
// Decimal Arithmetic (Precise)
// ============================================================================

/**
 * Add two decimal numbers precisely
 * @param a - First number
 * @param b - Second number
 * @returns Sum as string
 */
export async function addDecimals(a: string, b: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.add_decimals(a, b);
}

/**
 * Subtract two decimal numbers precisely
 * @param a - First number (minuend)
 * @param b - Second number (subtrahend)
 * @returns Difference as string
 */
export async function subtractDecimals(a: string, b: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.subtract_decimals(a, b);
}

/**
 * Multiply two decimal numbers precisely
 * @param a - First number
 * @param b - Second number
 * @returns Product as string
 */
export async function multiplyDecimals(a: string, b: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.multiply_decimals(a, b);
}

/**
 * Divide two decimal numbers precisely
 * @param a - Dividend
 * @param b - Divisor
 * @param scale - Decimal places (default: 2)
 * @returns Quotient as string
 */
export async function divideDecimals(
  a: string,
  b: string,
  scale = 2
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.divide_decimals(a, b, scale);
}

/**
 * Round a decimal number
 * @param value - Value to round
 * @param scale - Decimal places
 * @returns Rounded value as string
 */
export async function roundDecimal(value: string, scale: number): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.round_decimal(value, scale);
}

/**
 * Compare two decimal numbers
 * @param a - First number
 * @param b - Second number
 * @returns -1 if a < b, 0 if a == b, 1 if a > b
 */
export async function compareDecimals(a: string, b: string): Promise<-1 | 0 | 1> {
  const wasm = await ensureLoaded();
  return wasm.compare_decimals(a, b) as -1 | 0 | 1;
}

// ============================================================================
// Financial Year Utilities (Pure TypeScript)
// ============================================================================

/**
 * Get financial year for a date (April to March in India)
 * @param date - Date to get FY for (default: current date)
 * @returns Financial year info
 */
export function getFinancialYear(date?: Date): FinancialYear {
  const d = date ?? new Date();
  const month = d.getMonth(); // 0-indexed
  const year = d.getFullYear();

  // FY starts in April (month 3)
  const fyStartYear = month >= 3 ? year : year - 1;
  const fyEndYear = fyStartYear + 1;

  return {
    year: fyStartYear,
    startDate: `${fyStartYear}-04-01`,
    endDate: `${fyEndYear}-03-31`,
    label: `FY ${fyStartYear}-${String(fyEndYear).slice(-2)}`,
  };
}

/**
 * Get quarter for a date (Indian fiscal quarters)
 * @param date - Date to get quarter for (default: current date)
 * @returns Quarter number (1-4) and label
 */
export function getQuarter(date?: Date): { quarter: number; label: string; startDate: string; endDate: string } {
  const d = date ?? new Date();
  const month = d.getMonth();
  const year = d.getFullYear();
  const fy = getFinancialYear(d);

  // Q1: Apr-Jun, Q2: Jul-Sep, Q3: Oct-Dec, Q4: Jan-Mar
  let quarter: number;
  let startMonth: number;
  let endMonth: number;
  let startYear: number;
  let endYear: number;

  if (month >= 3 && month <= 5) {
    quarter = 1;
    startMonth = 4; endMonth = 6;
    startYear = endYear = year;
  } else if (month >= 6 && month <= 8) {
    quarter = 2;
    startMonth = 7; endMonth = 9;
    startYear = endYear = year;
  } else if (month >= 9 && month <= 11) {
    quarter = 3;
    startMonth = 10; endMonth = 12;
    startYear = endYear = year;
  } else {
    quarter = 4;
    startMonth = 1; endMonth = 3;
    startYear = endYear = year;
  }

  const pad = (n: number) => String(n).padStart(2, '0');

  return {
    quarter,
    label: `Q${quarter} ${fy.label}`,
    startDate: `${startYear}-${pad(startMonth)}-01`,
    endDate: `${endYear}-${pad(endMonth)}-${endMonth === 6 || endMonth === 9 ? '30' : '31'}`,
  };
}

/**
 * Check if a date falls within a financial year
 * @param date - Date to check
 * @param fyYear - Financial year start year
 * @returns Whether date is in the FY
 */
export function isInFinancialYear(date: Date, fyYear: number): boolean {
  const fy = getFinancialYear(date);
  return fy.year === fyYear;
}
