/* tslint:disable */
/* eslint-disable */

/**
 * Apply delta to a value
 */
export function apply_delta(base: any, delta: any): any;

/**
 * Calculate delta between two versions
 */
export function calculate_delta(old_version: any, new_version: any): any;

/**
 * Calculate sync statistics
 */
export function calculate_sync_stats(changes: any): any;

/**
 * Compare version vectors
 */
export function compare_version_vectors(vv1: any, vv2: any): any;

/**
 * Create change record
 */
export function create_change_record(entity_type: string, entity_id: string, operation: string, data: any, previous_data: any, client_id: string, version: bigint): any;

/**
 * Create version vector
 */
export function create_version_vector(): any;

/**
 * Detect conflicts between local and server versions
 */
export function detect_conflict(local: any, server: any, base: any): any;

/**
 * Generate unique change ID
 */
export function generate_change_id(): string;

/**
 * Increment version vector
 */
export function increment_version_vector(vv: any, node_id: string): any;

/**
 * Merge version vectors
 */
export function merge_version_vectors(vv1: any, vv2: any): any;

/**
 * Resolve conflict using specified strategy
 */
export function resolve_conflict(local: any, server: any, base: any, strategy: string): any;

/**
 * Sort changes for optimal sync order
 */
export function sort_changes_for_sync(changes: any): any;

/**
 * Three-way merge
 */
export function three_way_merge(base: any, local: any, server: any): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly generate_change_id: (a: number) => void;
  readonly create_change_record: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number, i: number, j: number, k: bigint) => number;
  readonly calculate_delta: (a: number, b: number) => number;
  readonly apply_delta: (a: number, b: number) => number;
  readonly detect_conflict: (a: number, b: number, c: number) => number;
  readonly resolve_conflict: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly three_way_merge: (a: number, b: number, c: number) => number;
  readonly create_version_vector: () => number;
  readonly increment_version_vector: (a: number, b: number, c: number) => number;
  readonly merge_version_vectors: (a: number, b: number) => number;
  readonly compare_version_vectors: (a: number, b: number) => number;
  readonly sort_changes_for_sync: (a: number) => number;
  readonly calculate_sync_stats: (a: number) => number;
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
