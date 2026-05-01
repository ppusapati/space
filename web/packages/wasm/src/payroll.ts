/**
 * Samavaya Payroll - TypeScript Bindings
 * Income tax, PF, ESI, CTC calculations using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';
import type {
  IncomeTaxInput,
  IncomeTaxResult,
  TaxRegime,
  PfResult,
  EsiResult,
  ProfessionalTaxResult,
  StatutoryResult,
  CtcBreakdown,
  SalaryStructure,
  CtcOptimization,
} from './types';

// Type for the raw WASM module
interface PayrollWasm {
  calculate_income_tax: (input: unknown) => unknown;
  compare_tax_regimes: (income: string, deductions: unknown) => unknown;
  calculate_advance_tax: (estimatedTax: string, taxPaid: string, currentQuarter: number) => unknown;
  calculate_pf: (basicSalary: string, isNewEmployee: boolean) => unknown;
  calculate_esi: (grossSalary: string, state: string) => unknown;
  calculate_professional_tax: (grossSalary: string, state: string) => unknown;
  calculate_lwf: (state: string) => unknown;
  calculate_all_statutory: (grossSalary: string, basicSalary: string, state: string, isNewEmployee: boolean) => unknown;
  calculate_ctc_breakdown: (ctcAmount: string, config: unknown) => unknown;
  calculate_salary_structure: (ctcBreakdown: unknown, workingDays: number, paidDays: number) => unknown;
  optimize_salary_structure: (ctcAmount: string) => unknown;
  reverse_ctc_calculation: (takeHome: string, config: unknown) => unknown;
}

let wasmModule: PayrollWasm | null = null;

/**
 * Initialize the payroll module
 */
async function ensureLoaded(): Promise<PayrollWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<PayrollWasm>('payroll' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// Income Tax Functions
// ============================================================================

/**
 * Calculate income tax based on regime
 * @param input - Income tax calculation input
 * @returns Detailed income tax calculation
 */
export async function calculateIncomeTax(input: IncomeTaxInput): Promise<IncomeTaxResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_income_tax(input) as IncomeTaxResult;
}

/**
 * Compare tax liability between old and new regime
 * @param income - Gross income
 * @param deductions - Deductions under various sections
 * @returns Comparison of both regimes
 */
export async function compareTaxRegimes(
  income: string,
  deductions: Record<string, string>
): Promise<{ oldRegime: IncomeTaxResult; newRegime: IncomeTaxResult; recommendation: TaxRegime; savings: string }> {
  const wasm = await ensureLoaded();
  return wasm.compare_tax_regimes(income, deductions) as {
    oldRegime: IncomeTaxResult;
    newRegime: IncomeTaxResult;
    recommendation: TaxRegime;
    savings: string;
  };
}

/**
 * Calculate advance tax installment
 * @param estimatedTax - Total estimated tax for the year
 * @param taxPaid - Tax already paid
 * @param currentQuarter - Current quarter (1-4)
 * @returns Advance tax details
 */
export async function calculateAdvanceTax(
  estimatedTax: string,
  taxPaid: string,
  currentQuarter: number
): Promise<{ dueAmount: string; dueDate: string; cumulativeDue: string; installmentNumber: number }> {
  const wasm = await ensureLoaded();
  return wasm.calculate_advance_tax(estimatedTax, taxPaid, currentQuarter) as {
    dueAmount: string;
    dueDate: string;
    cumulativeDue: string;
    installmentNumber: number;
  };
}

// ============================================================================
// Statutory Deductions
// ============================================================================

/**
 * Calculate Provident Fund contributions
 * @param basicSalary - Basic salary amount
 * @param isNewEmployee - Whether employee is new (affects pension contribution)
 * @returns PF calculation breakdown
 */
export async function calculatePf(basicSalary: string, isNewEmployee = false): Promise<PfResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_pf(basicSalary, isNewEmployee) as PfResult;
}

/**
 * Calculate ESI contributions
 * @param grossSalary - Gross salary amount
 * @param state - State code
 * @returns ESI calculation
 */
export async function calculateEsi(grossSalary: string, state: string): Promise<EsiResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_esi(grossSalary, state) as EsiResult;
}

/**
 * Calculate Professional Tax
 * @param grossSalary - Gross salary amount
 * @param state - State code
 * @returns Professional tax amount
 */
export async function calculateProfessionalTax(
  grossSalary: string,
  state: string
): Promise<ProfessionalTaxResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_professional_tax(grossSalary, state) as ProfessionalTaxResult;
}

/**
 * Calculate Labour Welfare Fund
 * @param state - State code
 * @returns LWF contributions
 */
export async function calculateLwf(
  state: string
): Promise<{ employeeContribution: string; employerContribution: string; frequency: string }> {
  const wasm = await ensureLoaded();
  return wasm.calculate_lwf(state) as {
    employeeContribution: string;
    employerContribution: string;
    frequency: string;
  };
}

/**
 * Calculate all statutory deductions at once
 * @param grossSalary - Gross salary
 * @param basicSalary - Basic salary
 * @param state - State code
 * @param isNewEmployee - Whether employee is new
 * @returns All statutory calculations
 */
export async function calculateAllStatutory(
  grossSalary: string,
  basicSalary: string,
  state: string,
  isNewEmployee = false
): Promise<StatutoryResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_all_statutory(grossSalary, basicSalary, state, isNewEmployee) as StatutoryResult;
}

// ============================================================================
// CTC Functions
// ============================================================================

/**
 * Break down CTC into components
 * @param ctcAmount - Annual CTC
 * @param config - Configuration options
 * @returns CTC breakdown
 */
export async function calculateCtcBreakdown(
  ctcAmount: string,
  config?: {
    includeGratuity?: boolean;
    pfOnFullBasic?: boolean;
    includeBonus?: boolean;
    customBasicPercentage?: string;
  }
): Promise<CtcBreakdown> {
  const wasm = await ensureLoaded();
  return wasm.calculate_ctc_breakdown(ctcAmount, config || {}) as CtcBreakdown;
}

/**
 * Calculate monthly salary structure from CTC
 * @param ctcBreakdown - CTC breakdown
 * @param workingDays - Total working days in month
 * @param paidDays - Days to be paid
 * @returns Monthly salary structure
 */
export async function calculateSalaryStructure(
  ctcBreakdown: CtcBreakdown,
  workingDays: number,
  paidDays: number
): Promise<SalaryStructure> {
  const wasm = await ensureLoaded();
  return wasm.calculate_salary_structure(ctcBreakdown, workingDays, paidDays) as SalaryStructure;
}

/**
 * Optimize salary structure for maximum tax efficiency
 * @param ctcAmount - Annual CTC
 * @returns Optimized salary breakdown
 */
export async function optimizeSalaryStructure(ctcAmount: string): Promise<CtcOptimization> {
  const wasm = await ensureLoaded();
  return wasm.optimize_salary_structure(ctcAmount) as CtcOptimization;
}

/**
 * Reverse calculate CTC from desired take-home
 * @param takeHome - Desired monthly take-home
 * @param config - Configuration options
 * @returns Required CTC
 */
export async function reverseCtcCalculation(
  takeHome: string,
  config?: {
    includeGratuity?: boolean;
    pfOnFullBasic?: boolean;
    state?: string;
  }
): Promise<{ requiredCtc: string; breakdown: CtcBreakdown }> {
  const wasm = await ensureLoaded();
  return wasm.reverse_ctc_calculation(takeHome, config || {}) as {
    requiredCtc: string;
    breakdown: CtcBreakdown;
  };
}
