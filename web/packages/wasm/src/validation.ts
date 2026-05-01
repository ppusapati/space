/**
 * Samavaya Validation - TypeScript Bindings
 * GSTIN, PAN, TAN, CIN, IFSC, Aadhaar validation using WASM
 */

import { loadWasmModule } from './loader';
import type {
  ValidationResult,
  GstinValidationResult,
  PanValidationResult,
  IfscValidationResult,
  AadhaarValidationResult,
  PasswordValidationResult,
} from './types';

// Type for the raw WASM module
interface ValidationWasm {
  validate_gstin: (gstin: string) => unknown;
  validate_pan: (pan: string) => unknown;
  validate_tan: (tan: string) => unknown;
  validate_cin: (cin: string) => unknown;
  validate_ifsc: (ifsc: string) => unknown;
  validate_aadhaar: (aadhaar: string) => unknown;
  validate_mobile: (mobile: string) => unknown;
  validate_email: (email: string) => unknown;
  validate_pincode: (pincode: string) => unknown;
  validate_password: (password: string, minLength?: number) => unknown;
  extract_state_from_gstin: (gstin: string) => unknown;
  extract_pan_from_gstin: (gstin: string) => string | null;
  get_pan_holder_type: (pan: string) => unknown;
  is_valid_gstin: (gstin: string) => boolean;
  is_valid_pan: (pan: string) => boolean;
  is_valid_tan: (tan: string) => boolean;
  is_valid_ifsc: (ifsc: string) => boolean;
  is_valid_aadhaar: (aadhaar: string) => boolean;
  is_valid_mobile: (mobile: string) => boolean;
  is_valid_email: (email: string) => boolean;
  password_strength_score: (password: string) => number;
  password_strength_color: (password: string) => string;
}

let wasmModule: ValidationWasm | null = null;

/**
 * Initialize the validation module
 */
async function ensureLoaded(): Promise<ValidationWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<ValidationWasm>('validation');
  }
  return wasmModule;
}

// ============================================================================
// GSTIN Validation
// ============================================================================

/**
 * Validate a GSTIN (Goods and Services Tax Identification Number)
 * @param gstin - 15-character GSTIN
 * @returns Validation result with state and PAN info
 */
export async function validateGstin(gstin: string): Promise<GstinValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_gstin(gstin) as GstinValidationResult;
}

/**
 * Quick check if GSTIN is valid
 * @param gstin - GSTIN to validate
 * @returns Whether the GSTIN is valid
 */
export async function isValidGstin(gstin: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_gstin(gstin);
}

/**
 * Extract state information from GSTIN
 * @param gstin - Valid GSTIN
 * @returns State code and name
 */
export async function extractStateFromGstin(
  gstin: string
): Promise<{ code: string; name: string } | null> {
  const wasm = await ensureLoaded();
  return wasm.extract_state_from_gstin(gstin) as { code: string; name: string } | null;
}

/**
 * Extract PAN from GSTIN
 * @param gstin - Valid GSTIN
 * @returns 10-character PAN
 */
export async function extractPanFromGstin(gstin: string): Promise<string | null> {
  const wasm = await ensureLoaded();
  return wasm.extract_pan_from_gstin(gstin);
}

// ============================================================================
// PAN Validation
// ============================================================================

/**
 * Validate a PAN (Permanent Account Number)
 * @param pan - 10-character PAN
 * @returns Validation result with holder type
 */
export async function validatePan(pan: string): Promise<PanValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_pan(pan) as PanValidationResult;
}

/**
 * Quick check if PAN is valid
 * @param pan - PAN to validate
 * @returns Whether the PAN is valid
 */
export async function isValidPan(pan: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_pan(pan);
}

/**
 * Get PAN holder type information
 * @param pan - Valid PAN
 * @returns Holder type code and label
 */
export async function getPanHolderType(
  pan: string
): Promise<{ code: string; label: string } | null> {
  const wasm = await ensureLoaded();
  return wasm.get_pan_holder_type(pan) as { code: string; label: string } | null;
}

// ============================================================================
// TAN Validation
// ============================================================================

/**
 * Validate a TAN (Tax Deduction Account Number)
 * @param tan - 10-character TAN
 * @returns Validation result
 */
export async function validateTan(tan: string): Promise<ValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_tan(tan) as ValidationResult;
}

/**
 * Quick check if TAN is valid
 * @param tan - TAN to validate
 * @returns Whether the TAN is valid
 */
export async function isValidTan(tan: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_tan(tan);
}

// ============================================================================
// CIN Validation
// ============================================================================

/**
 * Validate a CIN (Corporate Identification Number)
 * @param cin - 21-character CIN
 * @returns Validation result with company type info
 */
export async function validateCin(cin: string): Promise<ValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_cin(cin) as ValidationResult;
}

// ============================================================================
// IFSC Validation
// ============================================================================

/**
 * Validate an IFSC (Indian Financial System Code)
 * @param ifsc - 11-character IFSC
 * @returns Validation result with bank info
 */
export async function validateIfsc(ifsc: string): Promise<IfscValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_ifsc(ifsc) as IfscValidationResult;
}

/**
 * Quick check if IFSC is valid
 * @param ifsc - IFSC to validate
 * @returns Whether the IFSC is valid
 */
export async function isValidIfsc(ifsc: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_ifsc(ifsc);
}

// ============================================================================
// Aadhaar Validation
// ============================================================================

/**
 * Validate an Aadhaar number (uses Verhoeff algorithm)
 * @param aadhaar - 12-digit Aadhaar number
 * @returns Validation result with masked number
 */
export async function validateAadhaar(aadhaar: string): Promise<AadhaarValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_aadhaar(aadhaar) as AadhaarValidationResult;
}

/**
 * Quick check if Aadhaar is valid
 * @param aadhaar - Aadhaar to validate
 * @returns Whether the Aadhaar is valid
 */
export async function isValidAadhaar(aadhaar: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_aadhaar(aadhaar);
}

// ============================================================================
// Contact Validation
// ============================================================================

/**
 * Validate an Indian mobile number
 * @param mobile - Mobile number (with or without country code)
 * @returns Validation result
 */
export async function validateMobile(mobile: string): Promise<ValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_mobile(mobile) as ValidationResult;
}

/**
 * Quick check if mobile is valid
 * @param mobile - Mobile to validate
 * @returns Whether the mobile is valid
 */
export async function isValidMobile(mobile: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_mobile(mobile);
}

/**
 * Validate an email address
 * @param email - Email address
 * @returns Validation result
 */
export async function validateEmail(email: string): Promise<ValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_email(email) as ValidationResult;
}

/**
 * Quick check if email is valid
 * @param email - Email to validate
 * @returns Whether the email is valid
 */
export async function isValidEmail(email: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.is_valid_email(email);
}

/**
 * Validate an Indian pincode
 * @param pincode - 6-digit pincode
 * @returns Validation result
 */
export async function validatePincode(pincode: string): Promise<ValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_pincode(pincode) as ValidationResult;
}

// ============================================================================
// Password Validation
// ============================================================================

/**
 * Validate password strength
 * @param password - Password to validate
 * @param minLength - Minimum required length (default: 8)
 * @returns Detailed validation result
 */
export async function validatePassword(
  password: string,
  minLength = 8
): Promise<PasswordValidationResult> {
  const wasm = await ensureLoaded();
  return wasm.validate_password(password, minLength) as PasswordValidationResult;
}

/**
 * Get password strength score (1-5)
 * @param password - Password to check
 * @returns Score from 1 (very weak) to 5 (very strong)
 */
export async function getPasswordStrengthScore(password: string): Promise<number> {
  const wasm = await ensureLoaded();
  return wasm.password_strength_score(password);
}

/**
 * Get password strength indicator color
 * @param password - Password to check
 * @returns Hex color code
 */
export async function getPasswordStrengthColor(password: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.password_strength_color(password);
}
