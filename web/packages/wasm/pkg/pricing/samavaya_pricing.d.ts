/* tslint:disable */
/* eslint-disable */

/**
 * Calculate simple percentage discount
 */
export function apply_percentage_discount(amount: string, percentage: string): string;

/**
 * Calculate discount amount from percentage
 */
export function calculate_discount_amount(amount: string, percentage: string): string;

/**
 * Calculate percentage from discount amount
 */
export function calculate_discount_percentage(amount: string, discount: string): string;

/**
 * Calculate line total
 */
export function calculate_line_total(unit_price: string, quantity: string, discount_percentage: string): any;

/**
 * Calculate margin percentage
 */
export function calculate_margin(cost: string, selling_price: string): any;

/**
 * Calculate price with discounts
 */
export function calculate_price(input: any): any;

/**
 * Get price for quantity based on tiers
 */
export function get_tiered_price(quantity: string, tiers: any): string;

/**
 * Calculate selling price from cost and margin
 */
export function price_from_margin(cost: string, margin_percentage: string): string;

/**
 * Calculate selling price from cost and markup
 */
export function price_from_markup(cost: string, markup_percentage: string): string;

/**
 * Round price to nearest standard value
 */
export function round_price(price: string, rounding: string): string;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly calculate_price: (a: number) => number;
  readonly apply_percentage_discount: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly calculate_discount_amount: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly calculate_discount_percentage: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly calculate_margin: (a: number, b: number, c: number, d: number) => number;
  readonly price_from_margin: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly price_from_markup: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly get_tiered_price: (a: number, b: number, c: number, d: number) => void;
  readonly calculate_line_total: (a: number, b: number, c: number, d: number, e: number, f: number) => number;
  readonly round_price: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export3: (a: number) => void;
  readonly __wbindgen_add_to_stack_pointer: (a: number) => number;
  readonly __wbindgen_export4: (a: number, b: number, c: number) => void;
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
