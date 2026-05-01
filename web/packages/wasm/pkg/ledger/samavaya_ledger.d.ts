/* tslint:disable */
/* eslint-disable */

/**
 * Calculate aging for items
 */
export function calculate_aging(items: any, as_of_date: string | null | undefined, buckets: any): any;

/**
 * Calculate account balance
 */
export function calculate_balance(entries: any, opening_balance: string, opening_type: string, account_nature: string): any;

/**
 * Calculate closing balance from opening and transactions
 */
export function calculate_closing_balance(opening_debit: string, opening_credit: string, period_debit: string, period_credit: string, account_nature: string): any;

/**
 * Calculate journal totals
 */
export function calculate_journal_totals(lines: any): any;

/**
 * Calculate interest on overdue amount
 */
export function calculate_overdue_interest(amount: string, due_date: string, interest_rate: string, as_of_date?: string | null): string;

/**
 * Calculate running balances
 */
export function calculate_running_balances(entries: any, opening_balance: string, opening_type: string, account_nature: string): any;

/**
 * Calculate trial balance
 */
export function calculate_trial_balance(entries: any): any;

/**
 * Calculate days overdue
 */
export function days_overdue(due_date: string, as_of_date?: string | null): number;

/**
 * Find matching book entry for a bank entry
 */
export function find_match(bank_entry: any, book_entries: any): any;

/**
 * Get aging bucket for days overdue
 */
export function get_aging_bucket(days: number): string;

/**
 * Group trial balance by account type
 */
export function group_trial_balance(entries: any): any;

/**
 * Quick balance check
 */
export function is_journal_balanced(debits: any, credits: any): boolean;

/**
 * Check if document is overdue
 */
export function is_overdue(due_date: string): boolean;

/**
 * Quick balance check
 */
export function is_trial_balance_balanced(total_debit: string, total_credit: string): boolean;

/**
 * Check if balance is within limit
 */
export function is_within_limit(balance: string, limit: string): boolean;

/**
 * Calculate net balance from debit and credit totals
 */
export function net_balance(total_debit: string, total_credit: string): any;

/**
 * Perform bank reconciliation
 */
export function reconcile_bank(bank_entries: any, book_entries: any, tolerance?: number | null): any;

/**
 * Split amount between debit and credit
 */
export function split_entry(amount: string, entry_type: string): any;

/**
 * Validate journal entry (debit must equal credit)
 */
export function validate_journal_entry(entry: any): any;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly validate_journal_entry: (a: number) => number;
  readonly is_journal_balanced: (a: number, b: number) => number;
  readonly calculate_journal_totals: (a: number) => number;
  readonly split_entry: (a: number, b: number, c: number, d: number) => number;
  readonly calculate_balance: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => number;
  readonly calculate_running_balances: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => number;
  readonly net_balance: (a: number, b: number, c: number, d: number) => number;
  readonly is_within_limit: (a: number, b: number, c: number, d: number) => number;
  readonly calculate_aging: (a: number, b: number, c: number, d: number) => number;
  readonly days_overdue: (a: number, b: number, c: number, d: number) => number;
  readonly is_overdue: (a: number, b: number) => number;
  readonly get_aging_bucket: (a: number, b: number) => void;
  readonly calculate_overdue_interest: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number, i: number) => void;
  readonly reconcile_bank: (a: number, b: number, c: number, d: number) => number;
  readonly find_match: (a: number, b: number) => number;
  readonly calculate_trial_balance: (a: number) => number;
  readonly calculate_closing_balance: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number, i: number, j: number) => number;
  readonly group_trial_balance: (a: number) => number;
  readonly is_trial_balance_balanced: (a: number, b: number, c: number, d: number) => number;
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
