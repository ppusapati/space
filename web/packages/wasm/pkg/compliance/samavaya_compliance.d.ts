/* tslint:disable */
/* eslint-disable */

/**
 * Calculate tax liability
 */
export function calculate_tax_liability(gstr3b: any, itc: any): any;

/**
 * Check e-invoice applicability
 */
export function check_einvoice_applicability(annual_turnover: string, transaction_type: string, is_export: boolean, is_sez: boolean): any;

/**
 * Classify transaction type based on invoice details
 */
export function classify_transaction(seller_gstin: string, buyer_gstin: string | null | undefined, place_of_supply: string, invoice_value: string): any;

/**
 * Generate document hash for e-invoice
 */
export function generate_document_hash(invoice_json: string): string;

/**
 * Generate e-Invoice JSON
 */
export function generate_einvoice_json(input: any): any;

/**
 * Generate GSTR-1 data from invoices
 */
export function generate_gstr1(gstin: string, period: string, invoices: any): any;

/**
 * Generate GSTR-3B summary from GSTR-1 data
 */
export function generate_gstr3b_from_gstr1(gstr1: any, itc_data: any): any;

/**
 * Generate HSN summary from invoice items
 */
export function generate_hsn_summary(items: any): any;

/**
 * Generate IRN (Invoice Reference Number) hash
 */
export function generate_irn_hash(seller_gstin: string, invoice_no: string, fy: string): string;

/**
 * Validate e-Invoice JSON against schema
 */
export function validate_einvoice(einvoice: any): any;

/**
 * Validate GSTIN format
 */
export function validate_gstin_format(gstin: string): any;

/**
 * Validate GSTR-1 data
 */
export function validate_gstr1(gstr1: any): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly generate_einvoice_json: (a: number) => number;
  readonly validate_einvoice: (a: number) => number;
  readonly generate_irn_hash: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly generate_gstr1: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly validate_gstr1: (a: number) => number;
  readonly generate_gstr3b_from_gstr1: (a: number, b: number) => number;
  readonly calculate_tax_liability: (a: number, b: number) => number;
  readonly classify_transaction: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
  readonly generate_document_hash: (a: number, b: number, c: number) => void;
  readonly validate_gstin_format: (a: number, b: number) => number;
  readonly check_einvoice_applicability: (a: number, b: number, c: number, d: number, e: number, f: number) => number;
  readonly generate_hsn_summary: (a: number) => number;
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
