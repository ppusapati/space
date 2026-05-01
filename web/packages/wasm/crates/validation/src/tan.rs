//! TAN (Tax Deduction Account Number) validation
//!
//! TAN Format: AAAA00000A
//! - Position 1-4: Alphabetic code (state + first 3 letters of name)
//! - Position 5-9: Numeric sequence (00001-99999)
//! - Position 10: Alphabetic check character

use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;

lazy_static! {
    static ref TAN_REGEX: Regex = Regex::new(
        r"^[A-Z]{4}[0-9]{5}[A-Z]$"
    ).unwrap();
}

/// TAN information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TanInfo {
    pub tan: String,
    pub state_code: String,
    pub name_prefix: String,
    pub sequence: String,
    pub check_char: String,
    pub valid: bool,
}

/// Validate TAN format
#[wasm_bindgen]
pub fn validate_tan(tan: &str) -> bool {
    validate_tan_internal(tan).is_ok()
}

/// Full TAN validation with details
pub fn validate_tan_full(tan: &str) -> ValidationResult {
    match validate_tan_internal(tan) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(tan, details)
        }
        Err(error) => ValidationResult::err(tan, &error),
    }
}

/// Internal TAN validation
pub fn validate_tan_internal(tan: &str) -> Result<TanInfo, String> {
    let tan = tan.trim().to_uppercase();

    // Check length
    if tan.len() != 10 {
        return Err(format!("TAN must be 10 characters, got {}", tan.len()));
    }

    // Check format
    if !TAN_REGEX.is_match(&tan) {
        return Err("Invalid TAN format. Format: AAAA00000A".to_string());
    }

    // Extract components
    let state_code = &tan[0..1];
    let name_prefix = &tan[1..4];
    let sequence = &tan[4..9];
    let check_char = &tan[9..10];

    Ok(TanInfo {
        tan: tan.clone(),
        state_code: state_code.to_string(),
        name_prefix: name_prefix.to_string(),
        sequence: sequence.to_string(),
        check_char: check_char.to_string(),
        valid: true,
    })
}

/// Parse TAN and return components
#[wasm_bindgen]
pub fn parse_tan(tan: &str) -> JsValue {
    match validate_tan_internal(tan) {
        Ok(info) => serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL),
        Err(_) => JsValue::NULL,
    }
}

/// Format TAN with proper capitalization
#[wasm_bindgen]
pub fn format_tan(tan: &str) -> String {
    tan.trim().to_uppercase()
}

/// Mask TAN for display
#[wasm_bindgen]
pub fn mask_tan(tan: &str) -> String {
    let tan = tan.trim().to_uppercase();
    if tan.len() != 10 {
        return tan;
    }

    format!("{}XXXXX{}", &tan[0..4], &tan[9..10])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_tan() {
        assert!(validate_tan("MUMA12345A"));
        assert!(validate_tan("DELC99999Z"));
    }

    #[test]
    fn test_invalid_tan() {
        assert!(!validate_tan("MUMA1234A")); // Too short
        assert!(!validate_tan("1234ABCDEF")); // Wrong format
        assert!(!validate_tan("MUMA123456")); // Ends with digit
    }

    #[test]
    fn test_parse() {
        let result = validate_tan_internal("MUMA12345A").unwrap();
        assert_eq!(result.state_code, "M");
        assert_eq!(result.name_prefix, "UMA");
        assert_eq!(result.sequence, "12345");
        assert_eq!(result.check_char, "A");
    }
}
