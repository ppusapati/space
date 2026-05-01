/* tslint:disable */
/* eslint-disable */

/**
 * Calculate and append EAN-13 check digit
 */
export function ean13_with_check_digit(data: string): string;

/**
 * Generate barcode
 */
export function generate_barcode(options: any): any;

/**
 * Generate Code 128 barcode (simple interface)
 */
export function generate_code128(data: string, width?: number | null, height?: number | null): string;

/**
 * Generate EAN-13 barcode (simple interface)
 */
export function generate_ean13(data: string, width?: number | null, height?: number | null): string;

/**
 * Generate EAN-8 barcode (simple interface)
 */
export function generate_ean8(data: string, width?: number | null, height?: number | null): string;

/**
 * Generate QR code for GST invoice
 */
export function generate_gst_invoice_qr(seller_gstin: string, buyer_gstin: string, invoice_number: string, invoice_date: string, total_value: string, size?: number | null): string;

/**
 * Generate QR code with simple parameters
 */
export function generate_qr(data: string, size?: number | null): string;

/**
 * Generate QR code as data URL
 */
export function generate_qr_data_url(data: string, size?: number | null): string;

/**
 * Generate QR code for UPI payment
 */
export function generate_upi_qr(payee_vpa: string, payee_name: string, amount?: number | null, transaction_note?: string | null, size?: number | null): string;

/**
 * Generate QR code for a URL
 */
export function generate_url_qr(url: string, size?: number | null): string;

/**
 * Generate QR code for vCard
 */
export function generate_vcard_qr(name: string, phone: string, email: string, company: string, size?: number | null): string;

/**
 * Validate barcode data for a specific format
 */
export function validate_barcode_data(format: string, data: string): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly generate_qr: (a: number, b: number, c: number, d: number) => void;
  readonly generate_qr_data_url: (a: number, b: number, c: number, d: number) => void;
  readonly generate_url_qr: (a: number, b: number, c: number, d: number) => void;
  readonly generate_vcard_qr: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number, i: number, j: number) => void;
  readonly generate_upi_qr: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number, i: number, j: number) => void;
  readonly generate_gst_invoice_qr: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number, i: number, j: number, k: number, l: number) => void;
  readonly generate_code128: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly generate_ean13: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly generate_ean8: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly ean13_with_check_digit: (a: number, b: number, c: number) => void;
  readonly generate_barcode: (a: number) => number;
  readonly validate_barcode_data: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_add_to_stack_pointer: (a: number) => number;
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
