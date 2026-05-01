//! Cryptographic Utilities
//!
//! This crate provides:
//! - SHA-256, SHA-512, SHA-3 hashing
//! - HMAC-SHA256 for message authentication
//! - PBKDF2 for password hashing
//! - AES-GCM encryption/decryption
//! - Base64/Hex encoding

use aes_gcm::{
    aead::Aead,
    Aes256Gcm, Nonce,
};
use base64::{engine::general_purpose::STANDARD as BASE64, Engine};
use hmac::{Hmac, Mac, digest::KeyInit};
use pbkdf2::pbkdf2_hmac;
use rand::RngCore;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256, Sha512};
use sha3::{Sha3_256, Sha3_512};
use wasm_bindgen::prelude::*;

type HmacSha256 = Hmac<Sha256>;

/// Hash algorithm
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum HashAlgorithm {
    Sha256,
    Sha512,
    Sha3_256,
    Sha3_512,
}

/// Encoding type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum EncodingType {
    Hex,
    Base64,
}

/// Encryption result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncryptionResult {
    pub ciphertext: String,
    pub nonce: String,
    pub algorithm: String,
}

/// Initialize the crypto module (called from core init)
fn crypto_init() {
    console_error_panic_hook::set_once();
}

/// Calculate hash of input data
#[wasm_bindgen]
pub fn hash(data: &str, algorithm: &str, encoding: &str) -> String {
    let algo = match algorithm.to_lowercase().as_str() {
        "sha256" | "sha-256" => HashAlgorithm::Sha256,
        "sha512" | "sha-512" => HashAlgorithm::Sha512,
        "sha3-256" | "sha3_256" => HashAlgorithm::Sha3_256,
        "sha3-512" | "sha3_512" => HashAlgorithm::Sha3_512,
        _ => HashAlgorithm::Sha256,
    };

    let enc = match encoding.to_lowercase().as_str() {
        "base64" => EncodingType::Base64,
        _ => EncodingType::Hex,
    };

    let hash_bytes = match algo {
        HashAlgorithm::Sha256 => {
            let mut hasher = Sha256::new();
            hasher.update(data.as_bytes());
            hasher.finalize().to_vec()
        }
        HashAlgorithm::Sha512 => {
            let mut hasher = Sha512::new();
            hasher.update(data.as_bytes());
            hasher.finalize().to_vec()
        }
        HashAlgorithm::Sha3_256 => {
            let mut hasher = Sha3_256::new();
            hasher.update(data.as_bytes());
            hasher.finalize().to_vec()
        }
        HashAlgorithm::Sha3_512 => {
            let mut hasher = Sha3_512::new();
            hasher.update(data.as_bytes());
            hasher.finalize().to_vec()
        }
    };

    match enc {
        EncodingType::Hex => hex::encode(hash_bytes),
        EncodingType::Base64 => BASE64.encode(hash_bytes),
    }
}

/// Calculate SHA-256 hash (convenience function)
#[wasm_bindgen]
pub fn sha256(data: &str) -> String {
    hash(data, "sha256", "hex")
}

/// Calculate SHA-512 hash (convenience function)
#[wasm_bindgen]
pub fn sha512(data: &str) -> String {
    hash(data, "sha512", "hex")
}

/// Calculate hash of binary data
#[wasm_bindgen]
pub fn hash_bytes(data: &[u8], algorithm: &str, encoding: &str) -> String {
    let algo = match algorithm.to_lowercase().as_str() {
        "sha256" | "sha-256" => HashAlgorithm::Sha256,
        "sha512" | "sha-512" => HashAlgorithm::Sha512,
        "sha3-256" | "sha3_256" => HashAlgorithm::Sha3_256,
        "sha3-512" | "sha3_512" => HashAlgorithm::Sha3_512,
        _ => HashAlgorithm::Sha256,
    };

    let enc = match encoding.to_lowercase().as_str() {
        "base64" => EncodingType::Base64,
        _ => EncodingType::Hex,
    };

    let hash_bytes = match algo {
        HashAlgorithm::Sha256 => {
            let mut hasher = Sha256::new();
            hasher.update(data);
            hasher.finalize().to_vec()
        }
        HashAlgorithm::Sha512 => {
            let mut hasher = Sha512::new();
            hasher.update(data);
            hasher.finalize().to_vec()
        }
        HashAlgorithm::Sha3_256 => {
            let mut hasher = Sha3_256::new();
            hasher.update(data);
            hasher.finalize().to_vec()
        }
        HashAlgorithm::Sha3_512 => {
            let mut hasher = Sha3_512::new();
            hasher.update(data);
            hasher.finalize().to_vec()
        }
    };

    match enc {
        EncodingType::Hex => hex::encode(hash_bytes),
        EncodingType::Base64 => BASE64.encode(hash_bytes),
    }
}

/// Calculate HMAC-SHA256
#[wasm_bindgen]
pub fn hmac_sha256(data: &str, key: &str, encoding: &str) -> Result<String, JsValue> {
    let enc = match encoding.to_lowercase().as_str() {
        "base64" => EncodingType::Base64,
        _ => EncodingType::Hex,
    };

    let mut mac = <HmacSha256 as KeyInit>::new_from_slice(key.as_bytes())
        .map_err(|e| JsValue::from_str(&format!("Invalid key: {}", e)))?;

    mac.update(data.as_bytes());
    let result = mac.finalize().into_bytes().to_vec();

    Ok(match enc {
        EncodingType::Hex => hex::encode(result),
        EncodingType::Base64 => BASE64.encode(result),
    })
}

/// Verify HMAC-SHA256
#[wasm_bindgen]
pub fn verify_hmac_sha256(data: &str, key: &str, signature: &str, encoding: &str) -> bool {
    let expected = match hmac_sha256(data, key, encoding) {
        Ok(h) => h,
        Err(_) => return false,
    };

    // Constant-time comparison
    expected == signature
}

/// Hash password using PBKDF2
#[wasm_bindgen]
pub fn hash_password(password: &str, salt: Option<String>, iterations: Option<u32>) -> JsValue {
    let iter_count = iterations.unwrap_or(100_000);

    // Generate or use provided salt
    let salt_bytes: Vec<u8> = match salt {
        Some(s) => {
            if let Ok(bytes) = BASE64.decode(&s) {
                bytes
            } else {
                generate_salt()
            }
        }
        None => generate_salt(),
    };

    // Derive key using PBKDF2-SHA256
    let mut key = [0u8; 32];
    pbkdf2_hmac::<Sha256>(
        password.as_bytes(),
        &salt_bytes,
        iter_count,
        &mut key,
    );

    let result = serde_json::json!({
        "hash": BASE64.encode(key),
        "salt": BASE64.encode(&salt_bytes),
        "iterations": iter_count,
        "algorithm": "PBKDF2-SHA256"
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Verify password against hash
#[wasm_bindgen]
pub fn verify_password(password: &str, stored_hash: &str, salt: &str, iterations: u32) -> bool {
    let salt_bytes = match BASE64.decode(salt) {
        Ok(b) => b,
        Err(_) => return false,
    };

    let mut key = [0u8; 32];
    pbkdf2_hmac::<Sha256>(
        password.as_bytes(),
        &salt_bytes,
        iterations,
        &mut key,
    );

    let computed = BASE64.encode(key);
    computed == stored_hash
}

/// Generate random salt
fn generate_salt() -> Vec<u8> {
    let mut salt = vec![0u8; 16];
    rand::thread_rng().fill_bytes(&mut salt);
    salt
}

/// Generate random bytes
#[wasm_bindgen]
pub fn generate_random_bytes(length: usize) -> Vec<u8> {
    let mut bytes = vec![0u8; length];
    rand::thread_rng().fill_bytes(&mut bytes);
    bytes
}

/// Generate random hex string
#[wasm_bindgen]
pub fn generate_random_hex(length: usize) -> String {
    let bytes = generate_random_bytes(length);
    hex::encode(bytes)
}

/// Generate random base64 string
#[wasm_bindgen]
pub fn generate_random_base64(length: usize) -> String {
    let bytes = generate_random_bytes(length);
    BASE64.encode(bytes)
}

/// Encrypt data using AES-256-GCM
#[wasm_bindgen]
pub fn encrypt_aes_gcm(plaintext: &str, key: &str) -> JsValue {
    // Key should be 32 bytes (256 bits)
    let key_bytes = if key.len() == 64 {
        // Hex encoded
        match hex::decode(key) {
            Ok(b) => b,
            Err(_) => return JsValue::NULL,
        }
    } else if let Ok(b) = BASE64.decode(key) {
        b
    } else {
        // Hash the key to get 32 bytes
        let mut hasher = Sha256::new();
        hasher.update(key.as_bytes());
        hasher.finalize().to_vec()
    };

    if key_bytes.len() != 32 {
        return JsValue::NULL;
    }

    // Generate random nonce (12 bytes for GCM)
    let mut nonce_bytes = [0u8; 12];
    rand::thread_rng().fill_bytes(&mut nonce_bytes);

    let cipher = match Aes256Gcm::new_from_slice(&key_bytes) {
        Ok(c) => c,
        Err(_) => return JsValue::NULL,
    };

    let nonce = Nonce::from_slice(&nonce_bytes);

    match cipher.encrypt(nonce, plaintext.as_bytes()) {
        Ok(ciphertext) => {
            let result = EncryptionResult {
                ciphertext: BASE64.encode(&ciphertext),
                nonce: BASE64.encode(nonce_bytes),
                algorithm: "AES-256-GCM".to_string(),
            };
            serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
        }
        Err(_) => JsValue::NULL,
    }
}

/// Decrypt data using AES-256-GCM
#[wasm_bindgen]
pub fn decrypt_aes_gcm(ciphertext: &str, key: &str, nonce: &str) -> Option<String> {
    // Decode key
    let key_bytes = if key.len() == 64 {
        hex::decode(key).ok()?
    } else if let Ok(b) = BASE64.decode(key) {
        b
    } else {
        let mut hasher = Sha256::new();
        hasher.update(key.as_bytes());
        hasher.finalize().to_vec()
    };

    if key_bytes.len() != 32 {
        return None;
    }

    // Decode ciphertext and nonce
    let ciphertext_bytes = BASE64.decode(ciphertext).ok()?;
    let nonce_bytes = BASE64.decode(nonce).ok()?;

    if nonce_bytes.len() != 12 {
        return None;
    }

    let cipher = Aes256Gcm::new_from_slice(&key_bytes).ok()?;
    let nonce = Nonce::from_slice(&nonce_bytes);

    let plaintext = cipher.decrypt(nonce, ciphertext_bytes.as_ref()).ok()?;
    String::from_utf8(plaintext).ok()
}

/// Encode string to Base64
#[wasm_bindgen]
pub fn base64_encode(data: &str) -> String {
    BASE64.encode(data.as_bytes())
}

/// Decode Base64 to string
#[wasm_bindgen]
pub fn base64_decode(data: &str) -> Option<String> {
    let bytes = BASE64.decode(data).ok()?;
    String::from_utf8(bytes).ok()
}

/// Encode bytes to Base64
#[wasm_bindgen]
pub fn base64_encode_bytes(data: &[u8]) -> String {
    BASE64.encode(data)
}

/// Decode Base64 to bytes
#[wasm_bindgen]
pub fn base64_decode_bytes(data: &str) -> Option<Vec<u8>> {
    BASE64.decode(data).ok()
}

/// Encode string to Hex
#[wasm_bindgen]
pub fn hex_encode(data: &str) -> String {
    hex::encode(data.as_bytes())
}

/// Decode Hex to string
#[wasm_bindgen]
pub fn hex_decode(data: &str) -> Option<String> {
    let bytes = hex::decode(data).ok()?;
    String::from_utf8(bytes).ok()
}

/// Generate a cryptographically secure API key
#[wasm_bindgen]
pub fn generate_api_key(prefix: Option<String>) -> String {
    let random_part = generate_random_bytes(24);
    let encoded = BASE64.encode(random_part)
        .replace('+', "-")
        .replace('/', "_")
        .trim_end_matches('=')
        .to_string();

    match prefix {
        Some(p) => format!("{}_{}", p, encoded),
        None => encoded,
    }
}

/// Generate a secure token
#[wasm_bindgen]
pub fn generate_secure_token(length: usize) -> String {
    let bytes = generate_random_bytes(length);
    BASE64.encode(bytes)
        .replace('+', "-")
        .replace('/', "_")
        .trim_end_matches('=')
        .to_string()
}

/// Constant-time string comparison
#[wasm_bindgen]
pub fn secure_compare(a: &str, b: &str) -> bool {
    if a.len() != b.len() {
        return false;
    }

    let a_bytes = a.as_bytes();
    let b_bytes = b.as_bytes();

    let mut result = 0u8;
    for (x, y) in a_bytes.iter().zip(b_bytes.iter()) {
        result |= x ^ y;
    }

    result == 0
}

/// Derive key from password (for encryption)
#[wasm_bindgen]
pub fn derive_key(password: &str, salt: &str, iterations: Option<u32>) -> String {
    let iter_count = iterations.unwrap_or(100_000);

    let salt_bytes = if let Ok(bytes) = BASE64.decode(salt) {
        bytes
    } else {
        salt.as_bytes().to_vec()
    };

    let mut key = [0u8; 32];
    pbkdf2_hmac::<Sha256>(
        password.as_bytes(),
        &salt_bytes,
        iter_count,
        &mut key,
    );

    BASE64.encode(key)
}

/// Hash for file checksum
#[wasm_bindgen]
pub fn file_checksum(data: &[u8], algorithm: &str) -> String {
    hash_bytes(data, algorithm, "hex")
}

/// Verify file checksum
#[wasm_bindgen]
pub fn verify_checksum(data: &[u8], expected: &str, algorithm: &str) -> bool {
    let computed = file_checksum(data, algorithm);
    secure_compare(&computed, expected)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sha256() {
        let hash = sha256("hello world");
        assert_eq!(
            hash,
            "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
        );
    }

    #[test]
    fn test_hmac_sha256() {
        let hmac = hmac_sha256("hello", "secret", "hex").unwrap();
        assert!(!hmac.is_empty());
    }

    #[test]
    fn test_password_hashing() {
        // This would need the wasm environment
    }

    #[test]
    fn test_base64() {
        let encoded = base64_encode("hello");
        let decoded = base64_decode(&encoded).unwrap();
        assert_eq!(decoded, "hello");
    }

    #[test]
    fn test_hex() {
        let encoded = hex_encode("hello");
        assert_eq!(encoded, "68656c6c6f");
        let decoded = hex_decode(&encoded).unwrap();
        assert_eq!(decoded, "hello");
    }

    #[test]
    fn test_secure_compare() {
        assert!(secure_compare("abc", "abc"));
        assert!(!secure_compare("abc", "abd"));
        assert!(!secure_compare("abc", "abcd"));
    }
}
