/**
 * Samavaya Depreciation - TypeScript Bindings
 * Asset depreciation calculations using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';
import type {
  DepreciationMethod,
  DepreciationInput,
  DepreciationSchedule,
  DepreciationEntry,
  AssetBlock,
} from './types';

// Type for the raw WASM module
interface DepreciationWasm {
  calculate_depreciation_schedule: (input: unknown) => unknown;
  calculate_single_year: (input: unknown, year: number) => unknown;
  get_it_depreciation_rates: () => unknown;
  get_companies_act_rates: () => unknown;
  compare_depreciation_methods: (cost: string, salvageValue: string, usefulLife: number) => unknown;
  calculate_block_depreciation: (block: unknown, additions: unknown, disposals: unknown, year: string) => unknown;
  calculate_impairment: (bookValue: string, recoverableAmount: string) => unknown;
  calculate_revaluation: (bookValue: string, fairValue: string) => unknown;
}

let wasmModule: DepreciationWasm | null = null;

/**
 * Initialize the depreciation module
 */
async function ensureLoaded(): Promise<DepreciationWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<DepreciationWasm>('depreciation' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// Depreciation Calculation Functions
// ============================================================================

/**
 * Calculate full depreciation schedule for an asset
 * @param input - Depreciation input parameters
 * @returns Complete depreciation schedule
 */
export async function calculateDepreciationSchedule(
  input: DepreciationInput
): Promise<DepreciationSchedule> {
  const wasm = await ensureLoaded();
  return wasm.calculate_depreciation_schedule(input) as DepreciationSchedule;
}

/**
 * Calculate depreciation for a single year
 * @param input - Depreciation input parameters
 * @param year - Year number (1-based)
 * @returns Single year depreciation entry
 */
export async function calculateSingleYear(
  input: DepreciationInput,
  year: number
): Promise<DepreciationEntry> {
  const wasm = await ensureLoaded();
  return wasm.calculate_single_year(input, year) as DepreciationEntry;
}

// ============================================================================
// Rate Reference Functions
// ============================================================================

/**
 * Get Income Tax Act depreciation rates
 * @returns Depreciation rates by asset block
 */
export async function getItDepreciationRates(): Promise<
  Record<string, { rate: string; description: string; examples: string[] }>
> {
  const wasm = await ensureLoaded();
  return wasm.get_it_depreciation_rates() as Record<
    string,
    { rate: string; description: string; examples: string[] }
  >;
}

/**
 * Get Companies Act depreciation rates (Schedule II)
 * @returns Depreciation rates by asset category
 */
export async function getCompaniesActRates(): Promise<
  Record<string, { usefulLife: number; residualValue: string; category: string }>
> {
  const wasm = await ensureLoaded();
  return wasm.get_companies_act_rates() as Record<
    string,
    { usefulLife: number; residualValue: string; category: string }
  >;
}

// ============================================================================
// Comparison and Analysis Functions
// ============================================================================

/**
 * Compare depreciation across different methods
 * @param cost - Asset cost
 * @param salvageValue - Salvage value
 * @param usefulLife - Useful life in years
 * @returns Comparison of all methods
 */
export async function compareDepreciationMethods(
  cost: string,
  salvageValue: string,
  usefulLife: number
): Promise<Record<DepreciationMethod, DepreciationSchedule>> {
  const wasm = await ensureLoaded();
  return wasm.compare_depreciation_methods(cost, salvageValue, usefulLife) as Record<
    DepreciationMethod,
    DepreciationSchedule
  >;
}

// ============================================================================
// Block Depreciation (Indian IT Act)
// ============================================================================

/**
 * Calculate block depreciation as per Income Tax Act
 * @param block - Asset block details
 * @param additions - Additions during the year
 * @param disposals - Disposals during the year
 * @param year - Financial year
 * @returns Block depreciation calculation
 */
export async function calculateBlockDepreciation(
  block: AssetBlock,
  additions: Array<{ cost: string; date: string }>,
  disposals: Array<{ saleValue: string; date: string; originalCost: string }>,
  year: string
): Promise<{
  openingWdv: string;
  additions: string;
  disposals: string;
  baseForDepreciation: string;
  depreciationAmount: string;
  closingWdv: string;
  shortTermGain: string;
  longTermGain: string;
}> {
  const wasm = await ensureLoaded();
  return wasm.calculate_block_depreciation(block, additions, disposals, year) as {
    openingWdv: string;
    additions: string;
    disposals: string;
    baseForDepreciation: string;
    depreciationAmount: string;
    closingWdv: string;
    shortTermGain: string;
    longTermGain: string;
  };
}

// ============================================================================
// Impairment and Revaluation
// ============================================================================

/**
 * Calculate impairment loss
 * @param bookValue - Current book value
 * @param recoverableAmount - Recoverable amount (higher of value in use and fair value less costs to sell)
 * @returns Impairment calculation
 */
export async function calculateImpairment(
  bookValue: string,
  recoverableAmount: string
): Promise<{ impairmentLoss: string; revisedBookValue: string; isImpaired: boolean }> {
  const wasm = await ensureLoaded();
  return wasm.calculate_impairment(bookValue, recoverableAmount) as {
    impairmentLoss: string;
    revisedBookValue: string;
    isImpaired: boolean;
  };
}

/**
 * Calculate revaluation adjustment
 * @param bookValue - Current book value
 * @param fairValue - Fair value
 * @returns Revaluation calculation
 */
export async function calculateRevaluation(
  bookValue: string,
  fairValue: string
): Promise<{
  revaluationSurplus: string;
  revaluationDeficit: string;
  revisedBookValue: string;
  adjustmentType: 'surplus' | 'deficit' | 'none';
}> {
  const wasm = await ensureLoaded();
  return wasm.calculate_revaluation(bookValue, fairValue) as {
    revaluationSurplus: string;
    revaluationDeficit: string;
    revisedBookValue: string;
    adjustmentType: 'surplus' | 'deficit' | 'none';
  };
}

// ============================================================================
// Convenience Functions
// ============================================================================

/**
 * Calculate straight-line depreciation
 */
export async function calculateSlm(
  cost: string,
  salvageValue: string,
  usefulLife: number
): Promise<DepreciationSchedule> {
  return calculateDepreciationSchedule({
    cost,
    salvageValue,
    usefulLife,
    method: 'SLM',
  });
}

/**
 * Calculate written-down value depreciation
 */
export async function calculateWdv(
  cost: string,
  rate: string,
  years: number
): Promise<DepreciationSchedule> {
  return calculateDepreciationSchedule({
    cost,
    salvageValue: '0',
    usefulLife: years,
    method: 'WDV',
    rate,
  });
}

/**
 * Calculate double declining balance depreciation
 */
export async function calculateDdb(
  cost: string,
  salvageValue: string,
  usefulLife: number
): Promise<DepreciationSchedule> {
  return calculateDepreciationSchedule({
    cost,
    salvageValue,
    usefulLife,
    method: 'DDB',
  });
}

/**
 * Calculate sum of years digits depreciation
 */
export async function calculateSyd(
  cost: string,
  salvageValue: string,
  usefulLife: number
): Promise<DepreciationSchedule> {
  return calculateDepreciationSchedule({
    cost,
    salvageValue,
    usefulLife,
    method: 'SYD',
  });
}

/**
 * Calculate units of production depreciation
 */
export async function calculateUop(
  cost: string,
  salvageValue: string,
  totalUnits: string,
  unitsProduced: string[]
): Promise<DepreciationSchedule> {
  return calculateDepreciationSchedule({
    cost,
    salvageValue,
    usefulLife: unitsProduced.length,
    method: 'UOP',
    totalUnits,
    unitsProduced,
  });
}
