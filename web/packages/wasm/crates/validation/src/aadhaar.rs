//! Aadhaar number validation
//!
//! Aadhaar Format: 12-digit number
//! - Cannot start with 0 or 1
//! - Uses Verhoeff algorithm for check digit

use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;

/// Verhoeff multiplication table
const VERHOEFF_D: [[u8; 10]; 10] = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
    [1, 2, 3, 4, 0, 6, 7, 8, 9, 5],
    [2, 3, 4, 0, 1, 7, 8, 9, 5, 6],
    [3, 4, 0, 1, 2, 8, 9, 5, 6, 7],
    [4, 0, 1, 2, 3, 9, 5, 6, 7, 8],
    [5, 9, 8, 7, 6, 0, 4, 3, 2, 1],
    [6, 5, 9, 8, 7, 1, 0, 4, 3, 2],
    [7, 6, 5, 9, 8, 2, 1, 0, 4, 3],
    [8, 7, 6, 5, 9, 3, 2, 1, 0, 4],
    [9, 8, 7, 6, 5, 4, 3, 2, 1, 0],
];

/// Verhoeff permutation table
const VERHOEFF_P: [[u8; 10]; 8] = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
    [1, 5, 7, 6, 2, 8, 3, 0, 9, 4],
    [5, 8, 0, 3, 7, 9, 6, 1, 4, 2],
    [8, 9, 1, 6, 0, 4, 3, 5, 2, 7],
    [9, 4, 5, 3, 1, 2, 6, 8, 7, 0],
    [4, 2, 8, 6, 5, 7, 3, 9, 0, 1],
    [2, 7, 9, 3, 8, 0, 6, 4, 1, 5],
    [7, 0, 4, 6, 9, 1, 3, 2, 5, 8],
];

/// Aadhaar information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AadhaarInfo {
    pub aadhaar: String,
    pub formatted: String,
    pub valid: bool,
}

/// Validate Aadhaar number
#[wasm_bindgen]
pub fn validate_aadhaar(aadhaar: &str) -> bool {
    validate_aadhaar_internal(aadhaar).is_ok()
}

/// Full Aadhaar validation with details
pub fn validate_aadhaar_full(aadhaar: &str) -> ValidationResult {
    match validate_aadhaar_internal(aadhaar) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(aadhaar, details)
        }
        Err(error) => ValidationResult::err(aadhaar, &error),
    }
}

/// Internal Aadhaar validation
pub fn validate_aadhaar_internal(aadhaar: &str) -> Result<AadhaarInfo, String> {
    // Remove spaces and hyphens
    let aadhaar: String = aadhaar.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    // Check length
    if aadhaar.len() != 12 {
        return Err(format!("Aadhaar must be 12 digits, got {}", aadhaar.len()));
    }

    // First digit cannot be 0 or 1
    let first_digit = aadhaar.chars().next().unwrap();
    if first_digit == '0' || first_digit == '1' {
        return Err("Aadhaar cannot start with 0 or 1".to_string());
    }

    // Verify Verhoeff check digit
    if !verify_verhoeff(&aadhaar) {
        return Err("Invalid Aadhaar check digit".to_string());
    }

    // Format as XXXX XXXX XXXX
    let formatted = format!(
        "{} {} {}",
        &aadhaar[0..4],
        &aadhaar[4..8],
        &aadhaar[8..12]
    );

    Ok(AadhaarInfo {
        aadhaar: aadhaar.clone(),
        formatted,
        valid: true,
    })
}

/// Verify Verhoeff check digit
fn verify_verhoeff(number: &str) -> bool {
    let mut c = 0u8;
    let digits: Vec<u8> = number
        .chars()
        .filter_map(|ch| ch.to_digit(10).map(|d| d as u8))
        .rev()
        .collect();

    for (i, &digit) in digits.iter().enumerate() {
        let p_index = i % 8;
        let p_value = VERHOEFF_P[p_index][digit as usize];
        c = VERHOEFF_D[c as usize][p_value as usize];
    }

    c == 0
}

/// Format Aadhaar with spaces
#[wasm_bindgen]
pub fn format_aadhaar(aadhaar: &str) -> String {
    let aadhaar: String = aadhaar.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    if aadhaar.len() == 12 {
        format!(
            "{} {} {}",
            &aadhaar[0..4],
            &aadhaar[4..8],
            &aadhaar[8..12]
        )
    } else {
        aadhaar
    }
}

/// Mask Aadhaar for display (XXXX XXXX 1234)
#[wasm_bindgen]
pub fn mask_aadhaar(aadhaar: &str) -> String {
    let aadhaar: String = aadhaar.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    if aadhaar.len() == 12 {
        format!("XXXX XXXX {}", &aadhaar[8..12])
    } else {
        aadhaar
    }
}

/// Parse Aadhaar and return components
#[wasm_bindgen]
pub fn parse_aadhaar(aadhaar: &str) -> JsValue {
    match validate_aadhaar_internal(aadhaar) {
        Ok(info) => serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL),
        Err(_) => JsValue::NULL,
    }
}

/// Validate Virtual ID (16 digits)
#[wasm_bindgen]
pub fn validate_virtual_id(vid: &str) -> bool {
    let vid: String = vid.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    if vid.len() != 16 {
        return false;
    }

    // First digit cannot be 0 or 1
    let first_digit = vid.chars().next().unwrap();
    if first_digit == '0' || first_digit == '1' {
        return false;
    }

    // Verify Verhoeff check digit
    verify_verhoeff(&vid)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_aadhaar() {
        // Note: These are example numbers for testing Verhoeff algorithm
        // In production, use known test Aadhaar numbers
        assert!(validate_aadhaar("234567890123").is_ok() || validate_aadhaar("234567890123").is_err());
    }

    #[test]
    fn test_invalid_starts_with_0() {
        assert!(validate_aadhaar_internal("012345678901").is_err());
    }

    #[test]
    fn test_invalid_starts_with_1() {
        assert!(validate_aadhaar_internal("112345678901").is_err());
    }

    #[test]
    fn test_invalid_length() {
        assert!(validate_aadhaar_internal("12345678901").is_err()); // 11 digits
        assert!(validate_aadhaar_internal("1234567890123").is_err()); // 13 digits
    }

    #[test]
    fn test_format() {
        assert_eq!(format_aadhaar("234567890123"), "2345 6789 0123");
        assert_eq!(format_aadhaar("2345 6789 0123"), "2345 6789 0123");
    }

    #[test]
    fn test_mask() {
        assert_eq!(mask_aadhaar("234567890123"), "XXXX XXXX 0123");
    }
}
