/**
 * Samavaya Crypto - TypeScript Bindings
 * Cryptographic utilities using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';

// Type for the raw WASM module
interface CryptoWasm {
  hash: (algorithm: string, data: string) => string;
  sha256: (data: string) => string;
  sha512: (data: string) => string;
  hmac_sha256: (key: string, data: string) => string;
  hmac_sha512: (key: string, data: string) => string;
  hash_password: (password: string, iterations?: number) => unknown;
  verify_password: (password: string, hash: string, salt: string, iterations: number) => boolean;
  encrypt_aes_gcm: (plaintext: string, key: string) => unknown;
  decrypt_aes_gcm: (ciphertext: string, key: string, nonce: string, tag: string) => unknown;
  generate_key: (length: number) => string;
  generate_api_key: (prefix?: string) => string;
  generate_otp: (length: number) => string;
  generate_uuid: () => string;
  base64_encode: (data: string) => string;
  base64_decode: (data: string) => string;
  hex_encode: (data: string) => string;
  hex_decode: (data: string) => string;
  secure_compare: (a: string, b: string) => boolean;
  generate_checksum: (data: string) => string;
  verify_checksum: (data: string, checksum: string) => boolean;
}

let wasmModule: CryptoWasm | null = null;

/**
 * Initialize the crypto module
 */
async function ensureLoaded(): Promise<CryptoWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<CryptoWasm>('crypto' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// Hash Functions
// ============================================================================

/**
 * Hash data using specified algorithm
 * @param algorithm - Hash algorithm (sha256, sha512)
 * @param data - Data to hash
 * @returns Hex-encoded hash
 */
export async function hash(algorithm: 'sha256' | 'sha512', data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.hash(algorithm, data);
}

/**
 * SHA-256 hash
 * @param data - Data to hash
 * @returns Hex-encoded SHA-256 hash
 */
export async function sha256(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.sha256(data);
}

/**
 * SHA-512 hash
 * @param data - Data to hash
 * @returns Hex-encoded SHA-512 hash
 */
export async function sha512(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.sha512(data);
}

// ============================================================================
// HMAC Functions
// ============================================================================

/**
 * HMAC-SHA256
 * @param key - HMAC key
 * @param data - Data to authenticate
 * @returns Hex-encoded HMAC
 */
export async function hmacSha256(key: string, data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.hmac_sha256(key, data);
}

/**
 * HMAC-SHA512
 * @param key - HMAC key
 * @param data - Data to authenticate
 * @returns Hex-encoded HMAC
 */
export async function hmacSha512(key: string, data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.hmac_sha512(key, data);
}

// ============================================================================
// Password Functions
// ============================================================================

/**
 * Hash a password using PBKDF2
 * @param password - Password to hash
 * @param iterations - Number of iterations (default: 100000)
 * @returns Password hash with salt
 */
export async function hashPassword(
  password: string,
  iterations = 100000
): Promise<{ hash: string; salt: string; iterations: number }> {
  const wasm = await ensureLoaded();
  return wasm.hash_password(password, iterations) as { hash: string; salt: string; iterations: number };
}

/**
 * Verify a password against stored hash
 * @param password - Password to verify
 * @param hash - Stored hash
 * @param salt - Stored salt
 * @param iterations - Number of iterations used
 * @returns Whether password matches
 */
export async function verifyPassword(
  password: string,
  hash: string,
  salt: string,
  iterations: number
): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.verify_password(password, hash, salt, iterations);
}

// ============================================================================
// Encryption Functions
// ============================================================================

/**
 * Encrypt data using AES-GCM
 * @param plaintext - Data to encrypt
 * @param key - Encryption key (256-bit hex)
 * @returns Encrypted data with nonce and tag
 */
export async function encryptAesGcm(
  plaintext: string,
  key: string
): Promise<{ ciphertext: string; nonce: string; tag: string }> {
  const wasm = await ensureLoaded();
  return wasm.encrypt_aes_gcm(plaintext, key) as { ciphertext: string; nonce: string; tag: string };
}

/**
 * Decrypt data using AES-GCM
 * @param ciphertext - Encrypted data
 * @param key - Decryption key (256-bit hex)
 * @param nonce - Nonce used during encryption
 * @param tag - Authentication tag
 * @returns Decrypted plaintext
 */
export async function decryptAesGcm(
  ciphertext: string,
  key: string,
  nonce: string,
  tag: string
): Promise<{ plaintext: string; success: boolean }> {
  const wasm = await ensureLoaded();
  return wasm.decrypt_aes_gcm(ciphertext, key, nonce, tag) as { plaintext: string; success: boolean };
}

// ============================================================================
// Key Generation Functions
// ============================================================================

/**
 * Generate a random key
 * @param length - Key length in bytes
 * @returns Hex-encoded random key
 */
export async function generateKey(length: number): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_key(length);
}

/**
 * Generate an API key
 * @param prefix - Optional prefix for the key
 * @returns API key
 */
export async function generateApiKey(prefix?: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_api_key(prefix);
}

/**
 * Generate a numeric OTP
 * @param length - OTP length (4-8 digits)
 * @returns Numeric OTP
 */
export async function generateOtp(length = 6): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_otp(length);
}

/**
 * Generate a UUID v4
 * @returns UUID string
 */
export async function generateUuid(): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_uuid();
}

// ============================================================================
// Encoding Functions
// ============================================================================

/**
 * Base64 encode
 * @param data - Data to encode
 * @returns Base64-encoded string
 */
export async function base64Encode(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.base64_encode(data);
}

/**
 * Base64 decode
 * @param data - Base64 string to decode
 * @returns Decoded string
 */
export async function base64Decode(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.base64_decode(data);
}

/**
 * Hex encode
 * @param data - Data to encode
 * @returns Hex-encoded string
 */
export async function hexEncode(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.hex_encode(data);
}

/**
 * Hex decode
 * @param data - Hex string to decode
 * @returns Decoded string
 */
export async function hexDecode(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.hex_decode(data);
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Constant-time string comparison
 * @param a - First string
 * @param b - Second string
 * @returns Whether strings are equal
 */
export async function secureCompare(a: string, b: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.secure_compare(a, b);
}

/**
 * Generate checksum for data
 * @param data - Data to checksum
 * @returns Checksum string
 */
export async function generateChecksum(data: string): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_checksum(data);
}

/**
 * Verify checksum
 * @param data - Original data
 * @param checksum - Checksum to verify
 * @returns Whether checksum is valid
 */
export async function verifyChecksum(data: string, checksum: string): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.verify_checksum(data, checksum);
}
