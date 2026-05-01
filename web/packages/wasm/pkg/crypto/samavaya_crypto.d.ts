/* tslint:disable */
/* eslint-disable */

/**
 * Decode Base64 to string
 */
export function base64_decode(data: string): string | undefined;

/**
 * Decode Base64 to bytes
 */
export function base64_decode_bytes(data: string): Uint8Array | undefined;

/**
 * Encode string to Base64
 */
export function base64_encode(data: string): string;

/**
 * Encode bytes to Base64
 */
export function base64_encode_bytes(data: Uint8Array): string;

/**
 * Decrypt data using AES-256-GCM
 */
export function decrypt_aes_gcm(ciphertext: string, key: string, nonce: string): string | undefined;

/**
 * Derive key from password (for encryption)
 */
export function derive_key(password: string, salt: string, iterations?: number | null): string;

/**
 * Encrypt data using AES-256-GCM
 */
export function encrypt_aes_gcm(plaintext: string, key: string): any;

/**
 * Hash for file checksum
 */
export function file_checksum(data: Uint8Array, algorithm: string): string;

/**
 * Generate a cryptographically secure API key
 */
export function generate_api_key(prefix?: string | null): string;

/**
 * Generate random base64 string
 */
export function generate_random_base64(length: number): string;

/**
 * Generate random bytes
 */
export function generate_random_bytes(length: number): Uint8Array;

/**
 * Generate random hex string
 */
export function generate_random_hex(length: number): string;

/**
 * Generate a secure token
 */
export function generate_secure_token(length: number): string;

/**
 * Calculate hash of input data
 */
export function hash(data: string, algorithm: string, encoding: string): string;

/**
 * Calculate hash of binary data
 */
export function hash_bytes(data: Uint8Array, algorithm: string, encoding: string): string;

/**
 * Hash password using PBKDF2
 */
export function hash_password(password: string, salt?: string | null, iterations?: number | null): any;

/**
 * Decode Hex to string
 */
export function hex_decode(data: string): string | undefined;

/**
 * Encode string to Hex
 */
export function hex_encode(data: string): string;

/**
 * Calculate HMAC-SHA256
 */
export function hmac_sha256(data: string, key: string, encoding: string): string;

/**
 * Constant-time string comparison
 */
export function secure_compare(a: string, b: string): boolean;

/**
 * Calculate SHA-256 hash (convenience function)
 */
export function sha256(data: string): string;

/**
 * Calculate SHA-512 hash (convenience function)
 */
export function sha512(data: string): string;

/**
 * Verify file checksum
 */
export function verify_checksum(data: Uint8Array, expected: string, algorithm: string): boolean;

/**
 * Verify HMAC-SHA256
 */
export function verify_hmac_sha256(data: string, key: string, signature: string, encoding: string): boolean;

/**
 * Verify password against hash
 */
export function verify_password(password: string, stored_hash: string, salt: string, iterations: number): boolean;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
  readonly memory: WebAssembly.Memory;
  readonly hash: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly sha256: (a: number, b: number, c: number) => void;
  readonly sha512: (a: number, b: number, c: number) => void;
  readonly hmac_sha256: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly verify_hmac_sha256: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => number;
  readonly hash_password: (a: number, b: number, c: number, d: number, e: number) => number;
  readonly verify_password: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => number;
  readonly generate_random_bytes: (a: number, b: number) => void;
  readonly generate_random_hex: (a: number, b: number) => void;
  readonly generate_random_base64: (a: number, b: number) => void;
  readonly encrypt_aes_gcm: (a: number, b: number, c: number, d: number) => number;
  readonly decrypt_aes_gcm: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly base64_encode: (a: number, b: number, c: number) => void;
  readonly base64_decode: (a: number, b: number, c: number) => void;
  readonly base64_decode_bytes: (a: number, b: number, c: number) => void;
  readonly hex_encode: (a: number, b: number, c: number) => void;
  readonly hex_decode: (a: number, b: number, c: number) => void;
  readonly generate_api_key: (a: number, b: number, c: number) => void;
  readonly generate_secure_token: (a: number, b: number) => void;
  readonly secure_compare: (a: number, b: number, c: number, d: number) => number;
  readonly derive_key: (a: number, b: number, c: number, d: number, e: number, f: number) => void;
  readonly file_checksum: (a: number, b: number, c: number, d: number, e: number) => void;
  readonly verify_checksum: (a: number, b: number, c: number, d: number, e: number, f: number) => number;
  readonly base64_encode_bytes: (a: number, b: number, c: number) => void;
  readonly hash_bytes: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => void;
  readonly __wbindgen_export: (a: number) => void;
  readonly __wbindgen_add_to_stack_pointer: (a: number) => number;
  readonly __wbindgen_export2: (a: number, b: number) => number;
  readonly __wbindgen_export3: (a: number, b: number, c: number, d: number) => number;
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
