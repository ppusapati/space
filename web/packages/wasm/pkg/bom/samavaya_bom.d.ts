/* tslint:disable */
/* eslint-disable */

/**
 * Analyze impact of item change
 */
export function analyze_impact(bom_db: any, item_code: string, price_change: string, quantity_on_hand?: string | null): any;

/**
 * Calculate cost for a specific quantity
 */
export function calculate_bom_cost(bom: any, quantity: string, item_costs: any): any;

/**
 * Calculate BOM cost rollup
 */
export function calculate_cost_rollup(input: any): any;

/**
 * Calculate Economic Order Quantity
 */
export function calculate_eoq(annual_demand: string, ordering_cost: string, holding_cost_percentage: string, unit_cost: string): any;

/**
 * Calculate reorder point
 */
export function calculate_reorder_point(average_daily_usage: string, lead_time_days: number, safety_stock: string): any;

/**
 * Calculate safety stock
 */
export function calculate_safety_stock(average_daily_usage: string, max_daily_usage: string, average_lead_time: string, max_lead_time: string): any;

/**
 * Compare costs across different scenarios
 */
export function compare_bom_costs(bom: any, scenarios: any): any;

/**
 * Explode a BOM to all levels
 */
export function explode_bom(bom_db: any, item_code: string, quantity: string): any;

/**
 * Explode BOM for a single level only
 */
export function explode_bom_single_level(bom: any, quantity: string): any;

/**
 * Find common components between two BOMs
 */
export function find_common_components(bom_db: any, item_code_1: string, item_code_2: string): any;

/**
 * Find substitute/alternative components
 */
export function find_substitutes(bom_db: any, item_code: string, substitute_map: any): any;

/**
 * Generate production schedule from demand
 */
export function generate_production_schedule(demand: any, bom_db: any, capacity_per_day: string): any;

/**
 * Get BOM tree structure (for display)
 */
export function get_bom_tree(bom_db: any, item_code: string): any;

/**
 * Calculate make vs buy analysis
 */
export function make_vs_buy_analysis(bom: any, make_costs: any, buy_cost: string, quantity: string): any;

/**
 * Run MRP calculation
 */
export function run_mrp(input: any): any;

/**
 * Validate BOM for circular references
 */
export function validate_bom_circular(bom_db: any, item_code: string): any;

/**
 * Find all where an item is used (all levels)
 */
export function where_used_all_levels(bom_db: any, item_code: string): any;

/**
 * Find where an item is used (direct parents only)
 */
export function where_used_single_level(bom_db: any, item_code: string): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly explode_bom: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly explode_bom_single_level: (a: number, b: number, c: number) => number;
  readonly validate_bom_circular: (a: number, b: number, c: number) => number;
  readonly get_bom_tree: (a: number, b: number, c: number) => number;
  readonly calculate_cost_rollup: (a: number) => number;
  readonly calculate_bom_cost: (a: number, b: number, c: number, d: number) => number;
  readonly compare_bom_costs: (a: number, b: number) => number;
  readonly make_vs_buy_analysis: (a: number, b: number, c: number, d: number, e: number, f: number) => number;
  readonly where_used_single_level: (a: number, b: number, c: number) => number;
  readonly where_used_all_levels: (a: number, b: number, c: number) => number;
  readonly analyze_impact: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => number;
  readonly find_common_components: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly find_substitutes: (a: number, b: number, c: number, d: number) => number;
  readonly run_mrp: (a: number) => number;
  readonly calculate_reorder_point: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly calculate_eoq: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
  readonly calculate_safety_stock: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
  readonly generate_production_schedule: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export3: (a: number) => void;
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
