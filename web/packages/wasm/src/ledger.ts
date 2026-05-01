/**
 * Samavaya Ledger - TypeScript Bindings
 * Journal entries, balances, aging, reconciliation using WASM
 */

import { loadWasmModule } from './loader';
import type {
  JournalLine,
  JournalValidationResult,
  AccountBalance,
  RunningBalanceEntry,
  AgingBucket,
  AgingResult,
  ReconciliationMatch,
  ReconciliationResult,
  TrialBalanceEntry,
  TrialBalanceResult,
} from './types';

// Type for the raw WASM module
interface LedgerWasm {
  validate_journal_entry: (lines: unknown) => unknown;
  is_journal_balanced: (lines: unknown) => boolean;
  calculate_balance: (openingDebit: string, openingCredit: string, periodDebit: string, periodCredit: string, accountNature: string) => unknown;
  calculate_running_balance: (entries: unknown, openingBalance: string, accountNature: string) => unknown;
  calculate_aging: (invoices: unknown, asOfDate: string, buckets: unknown) => unknown;
  calculate_overdue_interest: (amount: string, annualRate: string, daysOverdue: number) => string;
  match_bank_entries: (bankEntries: unknown, bookEntries: unknown, tolerance: string) => unknown;
  calculate_trial_balance: (entries: unknown) => unknown;
  group_trial_balance: (entries: unknown) => unknown;
  is_trial_balance_balanced: (totalDebit: string, totalCredit: string) => boolean;
  calculate_closing_balance: (openingDebit: string, openingCredit: string, periodDebit: string, periodCredit: string, accountNature: string) => unknown;
}

let wasmModule: LedgerWasm | null = null;

/**
 * Initialize the ledger module
 */
async function ensureLoaded(): Promise<LedgerWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<LedgerWasm>('ledger');
  }
  return wasmModule;
}

// ============================================================================
// Journal Entry Functions
// ============================================================================

/**
 * Validate a journal entry (check debits = credits)
 * @param lines - Array of journal lines
 * @returns Validation result with totals and errors
 */
export async function validateJournalEntry(
  lines: JournalLine[]
): Promise<JournalValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_journal_entry(lines) as JournalValidationResult;
}

/**
 * Quick check if journal entry is balanced
 * @param lines - Array of journal lines
 * @returns Whether debits equal credits
 */
export async function isJournalBalanced(lines: JournalLine[]): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_journal_balanced(lines);
}

// ============================================================================
// Balance Calculation Functions
// ============================================================================

/**
 * Calculate account balance from opening and period transactions
 * @param openingDebit - Opening debit balance
 * @param openingCredit - Opening credit balance
 * @param periodDebit - Period debit total
 * @param periodCredit - Period credit total
 * @param accountNature - Account nature: 'asset', 'expense', 'liability', 'income', 'equity'
 * @returns Closing balance details
 */
export async function calculateBalance(
  openingDebit: string,
  openingCredit: string,
  periodDebit: string,
  periodCredit: string,
  accountNature: 'asset' | 'expense' | 'liability' | 'income' | 'equity'
): Promise<{
  closingDebit: string;
  closingCredit: string;
  netBalance: string;
  balanceType: 'Debit' | 'Credit';
}> {
  const wasm = await ensureLoaded();
  return wasm.calculate_balance(
    openingDebit,
    openingCredit,
    periodDebit,
    periodCredit,
    accountNature
  ) as {
    closingDebit: string;
    closingCredit: string;
    netBalance: string;
    balanceType: 'Debit' | 'Credit';
  };
}

/**
 * Calculate closing balance from components
 * @param openingDebit - Opening debit
 * @param openingCredit - Opening credit
 * @param periodDebit - Period debit
 * @param periodCredit - Period credit
 * @param accountNature - Account nature
 * @returns Closing balance details
 */
export async function calculateClosingBalance(
  openingDebit: string,
  openingCredit: string,
  periodDebit: string,
  periodCredit: string,
  accountNature: string
): Promise<{
  closingDebit: string;
  closingCredit: string;
  netBalance: string;
  balanceType: string;
}> {
  const wasm = await ensureLoaded();
  return wasm.calculate_closing_balance(
    openingDebit,
    openingCredit,
    periodDebit,
    periodCredit,
    accountNature
  ) as {
    closingDebit: string;
    closingCredit: string;
    netBalance: string;
    balanceType: string;
  };
}

/**
 * Calculate running balance for a list of transactions
 * @param entries - Array of transaction entries
 * @param openingBalance - Opening balance amount
 * @param accountNature - Account nature
 * @returns Array of entries with running balance
 */
export async function calculateRunningBalance(
  entries: Array<{ date: string; description: string; debit: string; credit: string }>,
  openingBalance: string,
  accountNature: 'asset' | 'expense' | 'liability' | 'income' | 'equity'
): Promise<RunningBalanceEntry[]> {
  const wasm = await ensureLoaded();
  return wasm.calculate_running_balance(entries, openingBalance, accountNature) as RunningBalanceEntry[];
}

// ============================================================================
// Aging Functions
// ============================================================================

/**
 * Default aging buckets
 */
export const DEFAULT_AGING_BUCKETS = [
  { label: 'Current', minDays: 0, maxDays: 0 },
  { label: '1-30 Days', minDays: 1, maxDays: 30 },
  { label: '31-60 Days', minDays: 31, maxDays: 60 },
  { label: '61-90 Days', minDays: 61, maxDays: 90 },
  { label: '90+ Days', minDays: 91, maxDays: null },
];

/**
 * Calculate aging analysis for outstanding invoices
 * @param invoices - Array of invoices with date, due date, and amount
 * @param asOfDate - Date to calculate aging as of (YYYY-MM-DD)
 * @param buckets - Custom aging buckets (optional)
 * @returns Aging analysis result
 */
export async function calculateAging(
  invoices: Array<{
    id: string;
    date: string;
    dueDate: string;
    amount: string;
    partyName?: string;
  }>,
  asOfDate: string,
  buckets?: Array<{ label: string; minDays: number; maxDays: number | null }>
): Promise<AgingResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_aging(invoices, asOfDate, buckets ?? DEFAULT_AGING_BUCKETS) as AgingResult;
}

/**
 * Calculate overdue interest
 * @param amount - Overdue amount
 * @param annualRate - Annual interest rate percentage
 * @param daysOverdue - Number of days overdue
 * @returns Interest amount
 */
export async function calculateOverdueInterest(
  amount: string,
  annualRate: string,
  daysOverdue: number
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.calculate_overdue_interest(amount, annualRate, daysOverdue);
}

// ============================================================================
// Bank Reconciliation Functions
// ============================================================================

/**
 * Match bank statement entries with book entries
 * @param bankEntries - Bank statement entries
 * @param bookEntries - Book/ledger entries
 * @param tolerance - Amount tolerance for matching
 * @returns Reconciliation result with matches and unmatched items
 */
export async function matchBankEntries(
  bankEntries: Array<{
    id: string;
    date: string;
    amount: string;
    reference?: string;
    description?: string;
  }>,
  bookEntries: Array<{
    id: string;
    date: string;
    amount: string;
    reference?: string;
    description?: string;
  }>,
  tolerance = '0.01'
): Promise<ReconciliationResult> {
  const wasm = await ensureLoaded();
  return wasm.match_bank_entries(bankEntries, bookEntries, tolerance) as ReconciliationResult;
}

// ============================================================================
// Trial Balance Functions
// ============================================================================

/**
 * Calculate trial balance from account entries
 * @param entries - Array of trial balance entries
 * @returns Trial balance result with totals and balance check
 */
export async function calculateTrialBalance(
  entries: TrialBalanceEntry[]
): Promise<TrialBalanceResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_trial_balance(entries) as TrialBalanceResult;
}

/**
 * Group trial balance entries by account type
 * @param entries - Array of trial balance entries
 * @returns Grouped totals by account type
 */
export async function groupTrialBalance(
  entries: TrialBalanceEntry[]
): Promise<Array<{ group: string; count: number; debit: string; credit: string }>> {
  const wasm = await ensureLoaded();
  return wasm.group_trial_balance(entries) as Array<{ group: string; count: number; debit: string; credit: string }>;
}

/**
 * Quick check if trial balance is balanced
 * @param totalDebit - Total debit amount
 * @param totalCredit - Total credit amount
 * @returns Whether debits equal credits (within tolerance)
 */
export async function isTrialBalanceBalanced(
  totalDebit: string,
  totalCredit: string
): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_trial_balance_balanced(totalDebit, totalCredit);
}
