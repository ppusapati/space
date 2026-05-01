/* tslint:disable */
/* eslint-disable */

/**
 * Calculate accumulated depreciation
 */
export function calculate_accumulated_depreciation(cost: string, salvage: string, useful_life: number, method: string, years_elapsed: number): any;

/**
 * Calculate depreciation schedule
 */
export function calculate_depreciation_schedule(input: any): any;

/**
 * Calculate depreciation for a partial year
 */
export function calculate_partial_year_depreciation(cost: string, salvage_value: string, useful_life_years: number, method: string, days_used: number): any;

/**
 * Compare depreciation methods
 */
export function compare_depreciation_methods(cost: string, salvage: string, useful_life: number): any;

/**
 * Companies Act depreciation rates
 */
export function get_companies_act_rates(): any;

/**
 * Get Income Tax depreciation rates
 */
export function get_it_depreciation_rates(): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly calculate_depreciation_schedule: (a: number) => number;
  readonly get_it_depreciation_rates: () => number;
  readonly get_companies_act_rates: () => number;
  readonly calculate_partial_year_depreciation: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
  readonly compare_depreciation_methods: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly calculate_accumulated_depreciation: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
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
