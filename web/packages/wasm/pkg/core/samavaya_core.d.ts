/* tslint:disable */
/* eslint-disable */

/**
 * Get absolute value
 */
export function abs_decimal(value: string): string;

/**
 * Add two decimal values
 */
export function add_decimals(a: string, b: string): string;

/**
 * Convert amount to words (Indian format)
 */
export function amount_to_words(amount: string): string;

/**
 * Compare two decimal values: -1 if a < b, 0 if equal, 1 if a > b
 */
export function compare_decimals(a: string, b: string): number;

/**
 * Convert currency
 */
export function convert_currency(amount: string, rate: string): string;

export function create_validation_error(field: string, message: string, code: string): any;

export function days_between(start: string, end: string): bigint;

/**
 * Divide two decimal values (returns "0" if divisor is zero)
 */
export function divide_decimals(a: string, b: string): string;

export function format_date_indian(date: string): string;

export function format_date_iso(date: string): string;

/**
 * Format amount in Indian numbering system (lakhs, crores)
 */
export function format_indian_number(amount: string): string;

/**
 * Format money with currency symbol
 */
export function format_money(amount: string, currency_code: string): string;

/**
 * Get all states
 */
export function get_all_states(): any;

/**
 * Get currency decimal places
 */
export function get_currency_decimals(currency_code: string): number;

/**
 * Get currency symbol
 */
export function get_currency_symbol(currency_code: string): string;

export function get_current_financial_year(): any;

export function get_financial_year(date: string): any;

export function get_quarter_from_date(date: string): string;

/**
 * Get state by code
 */
export function get_state_by_code(code: string): any;

/**
 * Get state by GST code (first 2 digits of GSTIN)
 */
export function get_state_by_gst_code(gst_code: string): any;

export function init(): void;

/**
 * Check if source and destination are same state (for CGST/SGST vs IGST)
 */
export function is_intra_state(source_state: string, dest_state: string): boolean;

/**
 * Check if state is a Union Territory
 */
export function is_union_territory(state_code: string): boolean;

/**
 * Check if decimal is zero
 */
export function is_zero(value: string): boolean;

/**
 * Multiply two decimal values
 */
export function multiply_decimals(a: string, b: string): string;

/**
 * Parse a string to Decimal, returning zero on failure
 */
export function parse_decimal(s: string): string;

/**
 * Calculate percentage of a value
 */
export function percentage(value: string, percent: string): string;

/**
 * Round for currency display (2 decimal places)
 */
export function round_currency(value: string): string;

/**
 * Round a decimal to specified places
 */
export function round_decimal(value: string, places: number): string;

/**
 * Subtract two decimal values
 */
export function subtract_decimals(a: string, b: string): string;

export function version(): string;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly parse_decimal: (a: number, b: number, c: number) => void;
  readonly round_decimal: (a: number, b: number, c: number, d: number) => void;
  readonly round_currency: (a: number, b: number, c: number) => void;
  readonly add_decimals: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly subtract_decimals: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly multiply_decimals: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly divide_decimals: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly percentage: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly compare_decimals: (a: number, b: number, c: number, d: number) => number;
  readonly is_zero: (a: number, b: number) => number;
  readonly abs_decimal: (a: number, b: number, c: number) => void;
  readonly create_validation_error: (a: number, b: number, c: number, d: number, e: number, f: number) => number;
  readonly get_state_by_code: (a: number, b: number) => number;
  readonly get_state_by_gst_code: (a: number, b: number) => number;
  readonly get_all_states: () => number;
  readonly is_intra_state: (a: number, b: number, c: number, d: number) => number;
  readonly is_union_territory: (a: number, b: number) => number;
  readonly format_indian_number: (a: number, b: number, c: number) => void;
  readonly amount_to_words: (a: number, b: number, c: number) => void;
  readonly convert_currency: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly format_money: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly get_currency_symbol: (a: number, b: number, c: number) => void;
  readonly get_currency_decimals: (a: number, b: number) => number;
  readonly get_financial_year: (a: number, b: number) => number;
  readonly get_current_financial_year: () => number;
  readonly get_quarter_from_date: (a: number, b: number, c: number) => void;
  readonly days_between: (a: number, b: number, c: number, d: number) => bigint;
  readonly format_date_indian: (a: number, b: number, c: number) => void;
  readonly format_date_iso: (a: number, b: number, c: number) => void;
  readonly init: () => void;
  readonly version: (a: number) => void;
  readonly __wbindgen_add_to_stack_pointer: (a: number) => number;
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export3: (a: number, b: number, c: number) => void;
  readonly __wbindgen_start: () => void;
}

export type SyncInitInput = BufferSource | WebAssembly.Module;

/**
* Instantiates the given `module`, which can either be bytes or
* a precompiled `WebAssembly.Module`.
*
* @param {{ module: SyncInitInput }} module - Passing `SyncInitInput` directly is deprecated.
*
* @returns {InitOutput}
*/
export function initSync(module: { module: SyncInitInput } | SyncInitInput): InitOutput;

/**
* If `module_or_path` is {RequestInfo} or {URL}, makes a request and
* for everything else, calls `WebAssembly.instantiate` directly.
*
* @param {{ module_or_path: InitInput | Promise<InitInput> }} module_or_path - Passing `InitInput` directly is deprecated.
*
* @returns {Promise<InitOutput>}
*/
export default function __wbg_init (module_or_path?: { module_or_path: InitInput | Promise<InitInput> } | InitInput | Promise<InitInput>): Promise<InitOutput>;
