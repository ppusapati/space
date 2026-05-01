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

/**
 * Format Aadhaar with spaces
 */
export function format_aadhaar(aadhaar: string): string;

export function format_date_indian(date: string): string;

export function format_date_iso(date: string): string;

/**
 * Generate GSTIN from components (for display/testing)
 */
export function format_gstin(state_code: string, pan: string, entity_number: string): string;

/**
 * Format amount in Indian numbering system (lakhs, crores)
 */
export function format_indian_number(amount: string): string;

/**
 * Format landline with STD code
 */
export function format_landline(landline: string): string;

/**
 * Format mobile number
 */
export function format_mobile(mobile: string): string;

/**
 * Format money with currency symbol
 */
export function format_money(amount: string, currency_code: string): string;

/**
 * Format PAN with proper capitalization
 */
export function format_pan(pan: string): string;

/**
 * Format TAN with proper capitalization
 */
export function format_tan(tan: string): string;

/**
 * Get all states
 */
export function get_all_states(): any;

/**
 * Get bank code from IFSC
 */
export function get_bank_code(ifsc: string): string;

/**
 * Get bank name from IFSC
 */
export function get_bank_name(ifsc: string): string;

/**
 * Get branch code from IFSC
 */
export function get_branch_code(ifsc: string): string;

/**
 * Get company type from CIN
 */
export function get_company_type(cin: string): string;

/**
 * Get currency decimal places
 */
export function get_currency_decimals(currency_code: string): number;

/**
 * Get currency symbol
 */
export function get_currency_symbol(currency_code: string): string;

export function get_current_financial_year(): any;

/**
 * Get domain from email
 */
export function get_email_domain(email: string): string;

export function get_financial_year(date: string): any;

/**
 * Get year of incorporation from CIN
 */
export function get_incorporation_year(cin: string): string;

/**
 * Get PAN from GSTIN
 */
export function get_pan_from_gstin(gstin: string): string;

/**
 * Get holder type from PAN
 */
export function get_pan_holder_type(pan: string): string;

/**
 * Get postal region from pincode (first digit)
 */
export function get_postal_region(pincode: string): string;

export function get_quarter_from_date(date: string): string;

/**
 * Get state by code
 */
export function get_state_by_code(code: string): any;

/**
 * Get state by GST code (first 2 digits of GSTIN)
 */
export function get_state_by_gst_code(gst_code: string): any;

/**
 * Get state code from GSTIN
 */
export function get_state_from_gstin(gstin: string): string;

export function init(): void;

/**
 * Check if PAN belongs to a company
 */
export function is_company_pan(pan: string): boolean;

/**
 * Check if PAN belongs to an individual
 */
export function is_individual_pan(pan: string): boolean;

/**
 * Check if source and destination are same state (for CGST/SGST vs IGST)
 */
export function is_intra_state(source_state: string, dest_state: string): boolean;

/**
 * Check if company is listed
 */
export function is_listed_company(cin: string): boolean;

/**
 * Check if two GSTINs are from the same state (for IGST determination)
 */
export function is_same_state_gstin(gstin1: string, gstin2: string): boolean;

/**
 * Check if state is a Union Territory
 */
export function is_union_territory(state_code: string): boolean;

/**
 * Check if password meets minimum requirements
 */
export function is_valid_password(password: string): boolean;

/**
 * Check if decimal is zero
 */
export function is_zero(value: string): boolean;

/**
 * Mask Aadhaar for display (XXXX XXXX 1234)
 */
export function mask_aadhaar(aadhaar: string): string;

/**
 * Mask bank account for display
 */
export function mask_bank_account(account_number: string): string;

/**
 * Mask email for display (a***@example.com)
 */
export function mask_email(email: string): string;

/**
 * Mask mobile for display (XXXXX XX789)
 */
export function mask_mobile(mobile: string): string;

/**
 * Mask PAN for display (e.g., ABCDE1234F -> ABCXX****F)
 */
export function mask_pan(pan: string): string;

/**
 * Mask TAN for display
 */
export function mask_tan(tan: string): string;

/**
 * Multiply two decimal values
 */
export function multiply_decimals(a: string, b: string): string;

/**
 * Parse Aadhaar and return components
 */
export function parse_aadhaar(aadhaar: string): any;

/**
 * Parse CIN and return components
 */
export function parse_cin(cin: string): any;

/**
 * Parse a string to Decimal, returning zero on failure
 */
export function parse_decimal(s: string): string;

/**
 * Parse GSTIN and return all components
 */
export function parse_gstin(gstin: string): any;

/**
 * Parse IFSC and return components
 */
export function parse_ifsc(ifsc: string): any;

/**
 * Parse PAN and return all components
 */
export function parse_pan(pan: string): any;

/**
 * Parse TAN and return components
 */
export function parse_tan(tan: string): any;

/**
 * Generate password strength indicator color
 */
export function password_strength_color(password: string): string;

/**
 * Quick password strength check (returns score 1-5)
 */
export function password_strength_score(password: string): number;

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

/**
 * Validate any identifier based on type
 */
export function validate(identifier_type: string, value: string): any;

/**
 * Validate Aadhaar number
 */
export function validate_aadhaar(aadhaar: string): boolean;

/**
 * Validate bank account number format
 */
export function validate_bank_account(account_number: string): boolean;

/**
 * Validate CIN format
 */
export function validate_cin(cin: string): boolean;

/**
 * Validate email address
 */
export function validate_email(email: string): boolean;

/**
 * Validate GSTIN format
 */
export function validate_gstin(gstin: string): boolean;

/**
 * Validate IFSC format
 */
export function validate_ifsc(ifsc: string): boolean;

/**
 * Validate Indian landline number
 */
export function validate_landline(landline: string): boolean;

/**
 * Validate LLPIN (LLP Identification Number)
 */
export function validate_llpin(llpin: string): boolean;

/**
 * Validate Indian mobile number
 */
export function validate_mobile(mobile: string): boolean;

/**
 * Validate PAN format
 */
export function validate_pan(pan: string): boolean;

/**
 * Validate password strength
 */
export function validate_password(password: string, min_length?: number | null): any;

/**
 * Validate Indian pincode
 */
export function validate_pincode(pincode: string): boolean;

/**
 * Validate TAN format
 */
export function validate_tan(tan: string): boolean;

/**
 * Validate Virtual ID (16 digits)
 */
export function validate_virtual_id(vid: string): boolean;

export function version(): string;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly validate_gstin: (a: number, b: number) => number;
  readonly get_pan_from_gstin: (a: number, b: number, c: number) => void;
  readonly get_state_from_gstin: (a: number, b: number, c: number) => void;
  readonly parse_gstin: (a: number, b: number) => number;
  readonly format_gstin: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly is_same_state_gstin: (a: number, b: number, c: number, d: number) => number;
  readonly validate_pan: (a: number, b: number) => number;
  readonly get_pan_holder_type: (a: number, b: number, c: number) => void;
  readonly is_company_pan: (a: number, b: number) => number;
  readonly is_individual_pan: (a: number, b: number) => number;
  readonly parse_pan: (a: number, b: number) => number;
  readonly format_pan: (a: number, b: number, c: number) => void;
  readonly mask_pan: (a: number, b: number, c: number) => void;
  readonly validate_tan: (a: number, b: number) => number;
  readonly parse_tan: (a: number, b: number) => number;
  readonly mask_tan: (a: number, b: number, c: number) => void;
  readonly validate_cin: (a: number, b: number) => number;
  readonly validate_llpin: (a: number, b: number) => number;
  readonly parse_cin: (a: number, b: number) => number;
  readonly is_listed_company: (a: number, b: number) => number;
  readonly get_incorporation_year: (a: number, b: number, c: number) => void;
  readonly get_company_type: (a: number, b: number, c: number) => void;
  readonly validate_ifsc: (a: number, b: number) => number;
  readonly parse_ifsc: (a: number, b: number) => number;
  readonly get_bank_code: (a: number, b: number, c: number) => void;
  readonly get_bank_name: (a: number, b: number, c: number) => void;
  readonly get_branch_code: (a: number, b: number, c: number) => void;
  readonly validate_bank_account: (a: number, b: number) => number;
  readonly mask_bank_account: (a: number, b: number, c: number) => void;
  readonly validate_aadhaar: (a: number, b: number) => number;
  readonly format_aadhaar: (a: number, b: number, c: number) => void;
  readonly mask_aadhaar: (a: number, b: number, c: number) => void;
  readonly parse_aadhaar: (a: number, b: number) => number;
  readonly validate_virtual_id: (a: number, b: number) => number;
  readonly validate_mobile: (a: number, b: number) => number;
  readonly format_mobile: (a: number, b: number, c: number) => void;
  readonly mask_mobile: (a: number, b: number, c: number) => void;
  readonly validate_email: (a: number, b: number) => number;
  readonly get_email_domain: (a: number, b: number, c: number) => void;
  readonly mask_email: (a: number, b: number, c: number) => void;
  readonly validate_pincode: (a: number, b: number) => number;
  readonly get_postal_region: (a: number, b: number, c: number) => void;
  readonly validate_landline: (a: number, b: number) => number;
  readonly format_landline: (a: number, b: number, c: number) => void;
  readonly validate_password: (a: number, b: number, c: number) => number;
  readonly password_strength_score: (a: number, b: number) => number;
  readonly is_valid_password: (a: number, b: number) => number;
  readonly password_strength_color: (a: number, b: number, c: number) => void;
  readonly validate: (a: number, b: number, c: number, d: number) => number;
  readonly format_tan: (a: number, b: number, c: number) => void;
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
  readonly __wbindgen_export: (a: number, b: number) => number;
  readonly __wbindgen_export2: (a: number, b: number, c: number, d: number) => number;
  readonly __wbindgen_export3: (a: number) => void;
  readonly __wbindgen_add_to_stack_pointer: (a: number) => number;
  readonly __wbindgen_export4: (a: number, b: number, c: number) => void;
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
