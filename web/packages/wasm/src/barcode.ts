/**
 * Samavaya Barcode - TypeScript Bindings
 * QR code, Code128, EAN barcode generation using WASM
 */

import { loadWasmModule } from './loader';
import type {
  BarcodeFormat,
  BarcodeOptions,
  BarcodeResult,
  UpiQrOptions,
  GstInvoiceQrData,
} from './types';

// Type for the raw WASM module
interface BarcodeWasm {
  generate_barcode: (options: unknown) => unknown;
  generate_qr: (data: string, size?: number) => string;
  generate_qr_with_options: (options: unknown) => unknown;
  generate_upi_qr: (options: unknown) => string;
  generate_gst_invoice_qr: (data: unknown) => string;
  generate_vcard_qr: (name: string, phone?: string, email?: string, org?: string) => string;
  generate_code128: (data: string, width?: number, height?: number) => string;
  generate_ean13: (data: string, width?: number, height?: number) => string;
  generate_ean8: (data: string, width?: number, height?: number) => string;
  validate_barcode_data: (format: string, data: string) => unknown;
  ean13_with_check_digit: (data: string) => string;
}

let wasmModule: BarcodeWasm | null = null;

/**
 * Initialize the barcode module
 */
async function ensureLoaded(): Promise<BarcodeWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<BarcodeWasm>('barcode');
  }
  return wasmModule;
}

// ============================================================================
// Generic Barcode Generation
// ============================================================================

/**
 * Generate a barcode with custom options
 * @param options - Barcode generation options
 * @returns Barcode result with SVG data
 */
export async function generateBarcode(options: BarcodeOptions): Promise<BarcodeResult> {
  const wasm = await ensureLoaded();
  return wasm.generate_barcode(options) as BarcodeResult;
}

/**
 * Validate barcode data for a specific format
 * @param format - Barcode format
 * @param data - Data to validate
 * @returns Validation result with corrected data if applicable
 */
export async function validateBarcodeData(
  format: BarcodeFormat,
  data: string
): Promise<{ valid: boolean; error?: string; correctedData?: string }> {
  const wasm = await ensureLoaded();
  return wasm.validate_barcode_data(format, data) as { valid: boolean; error?: string; correctedData?: string };
}

// ============================================================================
// QR Code Generation
// ============================================================================

/**
 * Generate a simple QR code
 * @param data - Data to encode
 * @param size - Size in pixels (default: 200)
 * @returns SVG string
 */
export async function generateQr(data: string, size = 200): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_qr(data, size);
}

/**
 * Generate a QR code with custom options
 * @param options - QR code options
 * @returns Barcode result with SVG
 */
export async function generateQrWithOptions(
  options: Omit<BarcodeOptions, 'format'>
): Promise<BarcodeResult> {
  const wasm = await ensureLoaded();
  return wasm.generate_qr_with_options({ ...options, format: 'QR' }) as BarcodeResult;
}

/**
 * Generate a UPI payment QR code
 * @param options - UPI payment details
 * @returns SVG string
 */
export async function generateUpiQr(options: UpiQrOptions): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_upi_qr(options);
}

/**
 * Generate a GST invoice QR code (as per GST e-invoice spec)
 * @param data - Invoice data
 * @returns SVG string
 */
export async function generateGstInvoiceQr(data: GstInvoiceQrData): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_gst_invoice_qr(data);
}

/**
 * Generate a vCard QR code
 * @param name - Contact name
 * @param phone - Phone number (optional)
 * @param email - Email address (optional)
 * @param org - Organization (optional)
 * @returns SVG string
 */
export async function generateVcardQr(
  name: string,
  phone?: string,
  email?: string,
  org?: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_vcard_qr(name, phone, email, org);
}

// ============================================================================
// Code128 Barcode
// ============================================================================

/**
 * Generate a Code128 barcode
 * @param data - Data to encode (alphanumeric)
 * @param width - Width in pixels (default: 200)
 * @param height - Height in pixels (default: 80)
 * @returns SVG string
 */
export async function generateCode128(
  data: string,
  width = 200,
  height = 80
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_code128(data, width, height);
}

// ============================================================================
// EAN Barcodes
// ============================================================================

/**
 * Generate an EAN-13 barcode
 * @param data - 12 or 13 digit number
 * @param width - Width in pixels (default: 200)
 * @param height - Height in pixels (default: 100)
 * @returns SVG string
 */
export async function generateEan13(
  data: string,
  width = 200,
  height = 100
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_ean13(data, width, height);
}

/**
 * Generate an EAN-8 barcode
 * @param data - 7 or 8 digit number
 * @param width - Width in pixels (default: 150)
 * @param height - Height in pixels (default: 80)
 * @returns SVG string
 */
export async function generateEan8(
  data: string,
  width = 150,
  height = 80
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_ean8(data, width, height);
}

/**
 * Calculate and append EAN-13 check digit
 * @param data - 12 digit number
 * @returns 13 digit number with check digit
 */
export async function ean13WithCheckDigit(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.ean13_with_check_digit(data);
}
