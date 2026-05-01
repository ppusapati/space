/**
 * Samavaya WASM - Main Entry Point
 *
 * High-performance WebAssembly modules for Indian ERP operations
 *
 * @packageDocumentation
 */

// Re-export all types
export * from './types';

// Re-export loader utilities
export {
  loadWasmModule,
  preloadWasmModules,
  isModuleLoaded,
  unloadModule,
  unloadAllModules,
  getLoadedModules,
  isWasmSupported,
  type WasmModuleName,
} from './loader';

// Re-export module-specific functions
export * as tax from './tax';
export * as validation from './validation';
export * as barcode from './barcode';
export * as ledger from './ledger';
export * as pricing from './pricing';
export * as core from './core';
export * as payroll from './payroll';
export * as bom from './bom';
export * as depreciation from './depreciation';
export * as compliance from './compliance';
export * as crypto from './crypto';
export * as i18n from './i18n';
export * as offline from './offline';

// Convenience exports for most commonly used functions
export {
  calculateGst,
  calculateGstBulk,
  calculateTds,
  calculateTcs,
  getHsnRate,
  validateHsn,
} from './tax';

export {
  validateGstin,
  validatePan,
  validateTan,
  validateIfsc,
  validateAadhaar,
  validateMobile,
  validateEmail,
  validatePassword,
  isValidGstin,
  isValidPan,
  extractPanFromGstin,
} from './validation';

export {
  generateQr,
  generateUpiQr,
  generateGstInvoiceQr,
  generateCode128,
  generateEan13,
} from './barcode';

export {
  validateJournalEntry,
  calculateBalance,
  calculateAging,
  matchBankEntries,
  calculateTrialBalance,
} from './ledger';

export {
  calculatePrice,
  calculateLineTotal,
  calculateMargin,
  applyPercentageDiscount,
  priceFromMargin,
  priceFromMarkup,
} from './pricing';

export {
  formatIndianNumber,
  amountToWords,
  getAllStates,
  getStateByCode,
  isIntraState,
  getFinancialYear,
  getQuarter,
} from './core';

export {
  calculateIncomeTax,
  compareTaxRegimes,
  calculatePf,
  calculateEsi,
  calculateProfessionalTax,
  calculateAllStatutory,
  calculateCtcBreakdown,
  optimizeSalaryStructure,
} from './payroll';

export {
  explodeBom,
  calculateCostRollup,
  whereUsedSingleLevel,
  whereUsedAllLevels,
  runMrp,
  calculateEoq,
} from './bom';

export {
  calculateDepreciationSchedule,
  calculateSlm,
  calculateWdv,
  getItDepreciationRates,
  compareDepreciationMethods,
} from './depreciation';

export {
  generateEinvoiceJson,
  validateEinvoice,
  generateGstr1,
  validateGstr1,
  generateGstr3bFromGstr1,
  classifyTransaction,
} from './compliance';

export {
  sha256,
  sha512,
  hmacSha256,
  hashPassword,
  verifyPassword,
  encryptAesGcm,
  decryptAesGcm,
  generateApiKey,
  generateOtp,
} from './crypto';

export {
  formatIndianNumber as formatNumber,
  formatCompactIndian,
  formatCurrency,
  amountToWords as numberToWords,
  formatDate,
  formatRelativeTime,
  getFinancialYear as getFY,
} from './i18n';

export {
  createChangeRecord,
  calculateDelta,
  applyDelta,
  detectConflict,
  resolveConflict,
  threeWayMerge,
} from './offline';

/**
 * Initialize all critical WASM modules
 * Call this early in your app to preload modules
 */
export async function initializeWasm(): Promise<void> {
  const { preloadWasmModules } = await import('./loader');
  await preloadWasmModules(['core', 'tax-engine', 'validation']);
}

/**
 * Initialize all WASM modules
 * Use this if you need all modules loaded upfront
 */
export async function initializeAllWasm(): Promise<void> {
  const { preloadWasmModules } = await import('./loader');
  await preloadWasmModules([
    'core',
    'tax-engine',
    'validation',
    'barcode',
    'ledger',
    'pricing',
    'payroll',
    'bom',
    'depreciation',
    'compliance',
    'crypto',
    'i18n',
    'offline',
  ]);
}
