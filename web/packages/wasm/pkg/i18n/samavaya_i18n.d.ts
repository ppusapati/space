/* tslint:disable */
/* eslint-disable */

/**
 * Convert amount to words (Indian English)
 */
export function amount_to_words(value: string, currency: string): string;

/**
 * Format number in compact Indian style (L, Cr)
 */
export function format_compact_indian(value: string, decimal_places?: number | null): string;

/**
 * Format currency
 */
export function format_currency(value: string, currency: string, show_symbol?: boolean | null): string;

/**
 * Format date
 */
export function format_date(date: string, style: string, locale?: string | null): string;

/**
 * Format datetime
 */
export function format_datetime(datetime: string, date_style: string, time_style: string): string;

/**
 * Format number in Indian style (lakhs and crores)
 */
export function format_indian_number(value: string, decimal_places?: number | null): string;

/**
 * Get relative time (e.g., "2 days ago")
 */
export function format_relative_time(datetime: string, reference?: string | null): string;

/**
 * Get financial year
 */
export function get_financial_year(date: string): string;

/**
 * Get ordinal suffix
 */
export function get_ordinal(num: number): string;

/**
 * Get quarter
 */
export function get_quarter(date: string, financial_year?: boolean | null): string;

/**
 * Parse number from localized string
 */
export function parse_number(value: string): number | undefined;

/**
 * Pluralize word
 */
export function pluralize(word: string, count: bigint): string;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly format_indian_number: (a: number, b: number, c: number, d: number) => void;
  readonly format_compact_indian: (a: number, b: number, c: number, d: number) => void;
  readonly format_currency: (a: number, b: number, c: number, d: number, e: number, f: number) => void;
  readonly amount_to_words: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly format_date: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly format_datetime: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly format_relative_time: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly get_financial_year: (a: number, b: number, c: number) => void;
  readonly get_quarter: (a: number, b: number, c: number, d: number) => void;
  readonly parse_number: (a: number, b: number, c: number) => void;
  readonly get_ordinal: (a: number, b: number) => void;
  readonly pluralize: (a: number, b: number, c: number, d: bigint) => void;
  readonly __wbindgen_add_to_stack_pointer: (a: number) => number;
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export3: (a: number, b: number, c: number) => void;
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
