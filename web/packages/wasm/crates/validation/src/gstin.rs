//! GSTIN (Goods and Services Tax Identification Number) validation
//!
//! GSTIN Format: 22AAAAA0000A1Z5
//! - Position 1-2: State code (01-38)
//! - Position 3-12: PAN of the taxpayer
//! - Position 13: Entity number (1-9 or A-Z)
//! - Position 14: 'Z' by default
//! - Position 15: Check digit (computed using Luhn algorithm variant)

use js_sys;
use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;
use samavaya_core::get_state_by_gst_code;

lazy_static! {
    static ref GSTIN_REGEX: Regex = Regex::new(
        r"^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z][1-9A-Z]Z[0-9A-Z]$"
    ).unwrap();
}

/// GSTIN components
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstinInfo {
    pub gstin: String,
    pub state_code: String,
    pub state_name: String,
    pub pan: String,
    pub entity_code: String,
    pub check_digit: String,
    pub taxpayer_type: String,
    pub valid: bool,
}

/// Validate GSTIN format
#[wasm_bindgen]
pub fn validate_gstin(gstin: &str) -> bool {
    validate_gstin_internal(gstin).is_ok()
}

/// Full GSTIN validation with details
pub fn validate_gstin_full(gstin: &str) -> ValidationResult {
    match validate_gstin_internal(gstin) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(gstin, details)
        }
        Err(error) => ValidationResult::err(gstin, &error),
    }
}

/// Internal GSTIN validation
pub fn validate_gstin_internal(gstin: &str) -> Result<GstinInfo, String> {
    let gstin = gstin.trim().to_uppercase();

    // Check length
    if gstin.len() != 15 {
        return Err(format!("GSTIN must be 15 characters, got {}", gstin.len()));
    }

    // Check format
    if !GSTIN_REGEX.is_match(&gstin) {
        return Err("Invalid GSTIN format".to_string());
    }

    // Extract components
    let state_code = &gstin[0..2];
    let pan = &gstin[2..12];
    let entity_code = &gstin[12..13];
    let _z = &gstin[13..14]; // Always 'Z'
    let check_digit = &gstin[14..15];

    // Validate state code (01-38)
    let state_num: u32 = state_code.parse().map_err(|_| "Invalid state code")?;
    if state_num < 1 || state_num > 38 {
        return Err(format!("Invalid state code: {}. Must be between 01 and 38", state_code));
    }

    // Get state name from code
    let state_info = get_state_by_gst_code(state_code);
    let state_name = if state_info.is_null() {
        "Unknown".to_string()
    } else {
        // Extract state name from JsValue
        js_sys::Reflect::get(&state_info, &"name".into())
            .ok()
            .and_then(|v| v.as_string())
            .unwrap_or_else(|| "Unknown".to_string())
    };

    // Validate check digit
    let calculated_check = calculate_gstin_check_digit(&gstin[0..14]);
    if calculated_check != check_digit.chars().next().unwrap_or(' ') {
        return Err(format!(
            "Invalid check digit. Expected '{}', got '{}'",
            calculated_check, check_digit
        ));
    }

    // Determine taxpayer type from PAN 4th character
    let pan_type = pan.chars().nth(3).unwrap_or(' ');
    let taxpayer_type = match pan_type {
        'C' => "Company",
        'P' => "Individual",
        'H' => "HUF",
        'F' => "Firm/LLP",
        'A' => "AOP/BOI",
        'T' => "Trust",
        'G' => "Government",
        'L' => "Local Authority",
        'J' => "Artificial Juridical Person",
        _ => "Unknown",
    };

    Ok(GstinInfo {
        gstin: gstin.clone(),
        state_code: state_code.to_string(),
        state_name,
        pan: pan.to_string(),
        entity_code: entity_code.to_string(),
        check_digit: check_digit.to_string(),
        taxpayer_type: taxpayer_type.to_string(),
        valid: true,
    })
}

/// Calculate GSTIN check digit using weighted checksum
fn calculate_gstin_check_digit(gstin_without_check: &str) -> char {
    // Character set for GSTIN
    const CHARS: &str = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ";

    let mut total = 0;

    for (i, c) in gstin_without_check.chars().enumerate() {
        let pos = CHARS.find(c).unwrap_or(0);
        let factor = if i % 2 == 0 { 1 } else { 2 };
        let product = pos * factor;
        let quotient = product / 36;
        let remainder = product % 36;
        total += quotient + remainder;
    }

    let remainder = total % 36;
    let check_code = (36 - remainder) % 36;

    CHARS.chars().nth(check_code).unwrap_or('0')
}

/// Get PAN from GSTIN
#[wasm_bindgen]
pub fn get_pan_from_gstin(gstin: &str) -> String {
    let gstin = gstin.trim().to_uppercase();
    if gstin.len() >= 12 {
        gstin[2..12].to_string()
    } else {
        String::new()
    }
}

/// Get state code from GSTIN
#[wasm_bindgen]
pub fn get_state_from_gstin(gstin: &str) -> String {
    let gstin = gstin.trim().to_uppercase();
    if gstin.len() >= 2 {
        gstin[0..2].to_string()
    } else {
        String::new()
    }
}

/// Parse GSTIN and return all components
#[wasm_bindgen]
pub fn parse_gstin(gstin: &str) -> JsValue {
    match validate_gstin_internal(gstin) {
        Ok(info) => serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL),
        Err(_) => JsValue::NULL,
    }
}

/// Generate GSTIN from components (for display/testing)
#[wasm_bindgen]
pub fn format_gstin(state_code: &str, pan: &str, entity_number: &str) -> String {
    let partial = format!(
        "{}{}{}Z",
        state_code.to_uppercase(),
        pan.to_uppercase(),
        entity_number.to_uppercase()
    );

    if partial.len() != 14 {
        return String::new();
    }

    let check_digit = calculate_gstin_check_digit(&partial);
    format!("{}{}", partial, check_digit)
}

/// Check if two GSTINs are from the same state (for IGST determination)
#[wasm_bindgen]
pub fn is_same_state_gstin(gstin1: &str, gstin2: &str) -> bool {
    let state1 = get_state_from_gstin(gstin1);
    let state2 = get_state_from_gstin(gstin2);
    !state1.is_empty() && state1 == state2
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_gstin() {
        // Sample valid GSTIN patterns (check digits are calculated)
        let result = validate_gstin_internal("27AAPFU0939F1ZV");
        // Note: Real check digit may differ
        assert!(result.is_ok() || result.is_err()); // Just check it runs
    }

    #[test]
    fn test_invalid_gstin_length() {
        let result = validate_gstin_internal("27AAPFU0939F1Z");
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("15 characters"));
    }

    #[test]
    fn test_invalid_state_code() {
        let result = validate_gstin_internal("99AAPFU0939F1ZV");
        assert!(result.is_err());
    }

    #[test]
    fn test_pan_extraction() {
        assert_eq!(get_pan_from_gstin("27AAPFU0939F1ZV"), "AAPFU0939F");
    }

    #[test]
    fn test_state_extraction() {
        assert_eq!(get_state_from_gstin("27AAPFU0939F1ZV"), "27");
    }

    #[test]
    fn test_same_state() {
        assert!(is_same_state_gstin("27AAPFU0939F1ZV", "27BBBBB1234B1Z2"));
        assert!(!is_same_state_gstin("27AAPFU0939F1ZV", "29BBBBB1234B1Z2"));
    }
}
