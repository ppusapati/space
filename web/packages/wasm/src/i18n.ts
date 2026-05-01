/**
 * Samavaya i18n - TypeScript Bindings
 * Indian localization utilities using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';

// Type for the raw WASM module
interface I18nWasm {
  format_indian_number: (value: string, decimals?: number) => string;
  format_compact_indian: (value: string) => string;
  format_currency: (value: string, currencyCode: string, showSymbol: boolean) => string;
  format_currency_words: (value: string, currencyCode: string) => string;
  amount_to_words: (value: string, currencyCode?: string) => string;
  amount_to_words_hindi: (value: string) => string;
  format_date: (isoDate: string, format: string, locale?: string) => string;
  format_datetime: (isoDatetime: string, format: string, locale?: string) => string;
  format_relative_time: (isoDate: string, baseDate?: string) => string;
  get_financial_year: (date: string) => unknown;
  get_quarter: (date: string) => unknown;
  get_ordinal: (number: number, locale?: string) => string;
  pluralize: (count: number, singular: string, plural: string) => string;
  format_percentage: (value: string, decimals?: number) => string;
  parse_indian_number: (formatted: string) => string;
  format_file_size: (bytes: number) => string;
  format_duration: (seconds: number) => string;
  transliterate: (text: string, from: string, to: string) => string;
}

let wasmModule: I18nWasm | null = null;

/**
 * Initialize the i18n module
 */
async function ensureLoaded(): Promise<I18nWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<I18nWasm>('i18n' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// Number Formatting Functions
// ============================================================================

/**
 * Format number in Indian numbering system (lakhs, crores)
 * @param value - Number to format
 * @param decimals - Decimal places (default: 2)
 * @returns Formatted string (e.g., "1,23,45,678.00")
 */
export async function formatIndianNumber(value: string, decimals = 2): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_indian_number(value, decimals);
}

/**
 * Format number in compact Indian notation
 * @param value - Number to format
 * @returns Compact string (e.g., "1.23 Cr", "45.6 L")
 */
export async function formatCompactIndian(value: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_compact_indian(value);
}

/**
 * Parse Indian formatted number back to decimal
 * @param formatted - Formatted string
 * @returns Decimal string
 */
export async function parseIndianNumber(formatted: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.parse_indian_number(formatted);
}

/**
 * Format percentage
 * @param value - Decimal value (0.15 = 15%)
 * @param decimals - Decimal places
 * @returns Formatted percentage
 */
export async function formatPercentage(value: string, decimals = 2): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_percentage(value, decimals);
}

// ============================================================================
// Currency Formatting Functions
// ============================================================================

/**
 * Format currency amount
 * @param value - Amount to format
 * @param currencyCode - Currency code (INR, USD, EUR, etc.)
 * @param showSymbol - Whether to show currency symbol
 * @returns Formatted currency string
 */
export async function formatCurrency(
  value: string,
  currencyCode = 'INR',
  showSymbol = true
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_currency(value, currencyCode, showSymbol);
}

/**
 * Convert amount to words with currency
 * @param value - Amount
 * @param currencyCode - Currency code
 * @returns Amount in words (e.g., "Rupees One Lakh Twenty Three Thousand Only")
 */
export async function formatCurrencyWords(value: string, currencyCode = 'INR'): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_currency_words(value, currencyCode);
}

/**
 * Convert amount to words
 * @param value - Amount
 * @param currencyCode - Optional currency code
 * @returns Amount in words
 */
export async function amountToWords(value: string, currencyCode?: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.amount_to_words(value, currencyCode);
}

/**
 * Convert amount to Hindi words
 * @param value - Amount
 * @returns Amount in Hindi (e.g., "एक लाख तेईस हज़ार")
 */
export async function amountToWordsHindi(value: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.amount_to_words_hindi(value);
}

// ============================================================================
// Date Formatting Functions
// ============================================================================

/**
 * Format date
 * @param isoDate - ISO date string
 * @param format - Format string (e.g., "DD/MM/YYYY", "D MMMM YYYY")
 * @param locale - Locale (default: "en-IN")
 * @returns Formatted date
 */
export async function formatDate(
  isoDate: string,
  format: string,
  locale = 'en-IN'
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_date(isoDate, format, locale);
}

/**
 * Format datetime
 * @param isoDatetime - ISO datetime string
 * @param format - Format string
 * @param locale - Locale
 * @returns Formatted datetime
 */
export async function formatDatetime(
  isoDatetime: string,
  format: string,
  locale = 'en-IN'
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_datetime(isoDatetime, format, locale);
}

/**
 * Format relative time (e.g., "2 days ago", "in 3 hours")
 * @param isoDate - ISO date to compare
 * @param baseDate - Base date (defaults to now)
 * @returns Relative time string
 */
export async function formatRelativeTime(isoDate: string, baseDate?: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_relative_time(isoDate, baseDate);
}

// ============================================================================
// Financial Year Functions
// ============================================================================

/**
 * Get financial year for a date
 * @param date - ISO date string
 * @returns Financial year info
 */
export async function getFinancialYear(
  date: string
): Promise<{ year: number; label: string; startDate: string; endDate: string }> {
  const wasm = await ensureLoaded();
  return wasm.get_financial_year(date) as {
    year: number;
    label: string;
    startDate: string;
    endDate: string;
  };
}

/**
 * Get quarter for a date
 * @param date - ISO date string
 * @returns Quarter info
 */
export async function getQuarter(
  date: string
): Promise<{ quarter: number; label: string; startDate: string; endDate: string; financialYear: string }> {
  const wasm = await ensureLoaded();
  return wasm.get_quarter(date) as {
    quarter: number;
    label: string;
    startDate: string;
    endDate: string;
    financialYear: string;
  };
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Get ordinal suffix for a number
 * @param number - Number
 * @param locale - Locale
 * @returns Ordinal (e.g., "1st", "2nd", "3rd")
 */
export async function getOrdinal(number: number, locale = 'en'): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.get_ordinal(number, locale);
}

/**
 * Pluralize a word based on count
 * @param count - Count
 * @param singular - Singular form
 * @param plural - Plural form
 * @returns Appropriate form with count
 */
export async function pluralize(count: number, singular: string, plural: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.pluralize(count, singular, plural);
}

/**
 * Format file size
 * @param bytes - Size in bytes
 * @returns Human-readable size (e.g., "1.5 MB")
 */
export async function formatFileSize(bytes: number): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_file_size(bytes);
}

/**
 * Format duration
 * @param seconds - Duration in seconds
 * @returns Human-readable duration (e.g., "2h 30m")
 */
export async function formatDuration(seconds: number): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.format_duration(seconds);
}

/**
 * Transliterate text between scripts
 * @param text - Text to transliterate
 * @param from - Source script (latin, devanagari, etc.)
 * @param to - Target script
 * @returns Transliterated text
 */
export async function transliterate(text: string, from: string, to: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.transliterate(text, from, to);
}

// ============================================================================
// Convenience Exports
// ============================================================================

/**
 * Format as Indian Rupees
 */
export async function formatInr(value: string): Promise<string> {
  return formatCurrency(value, 'INR', true);
}

/**
 * Format as compact Indian Rupees (e.g., ₹1.23 Cr)
 */
export async function formatInrCompact(value: string): Promise<string> {
  const compact = await formatCompactIndian(value);
  return `₹${compact}`;
}

/**
 * Format date in Indian format (DD/MM/YYYY)
 */
export async function formatDateIndian(isoDate: string): Promise<string> {
  return formatDate(isoDate, 'DD/MM/YYYY');
}

/**
 * Format date in long Indian format (1 January 2024)
 */
export async function formatDateLong(isoDate: string): Promise<string> {
  return formatDate(isoDate, 'D MMMM YYYY');
}
