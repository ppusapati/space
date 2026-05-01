/* tslint:disable */
/* eslint-disable */

/**
 * Calculate all statutory deductions
 */
export function calculate_all_statutory(basic_salary: string, gross_salary: string, state: string, da?: string | null): any;

/**
 * Calculate CTC breakdown from annual CTC
 */
export function calculate_ctc_breakdown(input: any): any;

/**
 * Calculate ESI contributions
 */
export function calculate_esi(input: any): any;

/**
 * Calculate income tax
 */
export function calculate_income_tax(input: any): any;

/**
 * Calculate Labour Welfare Fund
 */
export function calculate_lwf(input: any): any;

/**
 * Calculate PF contributions
 */
export function calculate_pf(input: any): any;

/**
 * Calculate Professional Tax
 */
export function calculate_professional_tax(input: any): any;

/**
 * Calculate salary structure from individual components
 */
export function calculate_salary_structure(input: any): any;

/**
 * Compare tax under both regimes
 */
export function compare_tax_regimes(input: any): any;

/**
 * Get ESI rates
 */
export function get_esi_rates(): any;

/**
 * Get PF rates
 */
export function get_pf_rates(): any;

/**
 * Get tax slabs for a regime
 */
export function get_tax_slabs(regime: string, fy?: string | null): any;

/**
 * Generate optimal salary structure for tax efficiency
 */
export function optimize_salary_structure(annual_ctc: string, state: string, metro_city: boolean): any;

/**
 * Calculate reverse CTC from take-home salary
 */
export function reverse_ctc_calculation(monthly_in_hand: string, state: string, include_variable: boolean): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly calculate_income_tax: (a: number) => number;
  readonly compare_tax_regimes: (a: number) => number;
  readonly get_tax_slabs: (a: number, b: number, c: number, d: number) => number;
  readonly calculate_pf: (a: number) => number;
  readonly calculate_esi: (a: number) => number;
  readonly calculate_professional_tax: (a: number) => number;
  readonly calculate_lwf: (a: number) => number;
  readonly calculate_all_statutory: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
  readonly get_pf_rates: () => number;
  readonly get_esi_rates: () => number;
  readonly calculate_ctc_breakdown: (a: number) => number;
  readonly calculate_salary_structure: (a: number) => number;
  readonly optimize_salary_structure: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly reverse_ctc_calculation: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
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
