//! IFSC (Indian Financial System Code) validation
//!
//! IFSC Format: AAAA0BBBBBB
//! - Position 1-4: Bank code (4 alphabets)
//! - Position 5: Always '0' (reserved for future use)
//! - Position 6-11: Branch code (6 alphanumeric)

use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;

lazy_static! {
    static ref IFSC_REGEX: Regex = Regex::new(
        r"^[A-Z]{4}0[A-Z0-9]{6}$"
    ).unwrap();
}

/// Known bank codes (sample - in production this would be a database)
const KNOWN_BANKS: &[(&str, &str)] = &[
    ("SBIN", "State Bank of India"),
    ("HDFC", "HDFC Bank"),
    ("ICIC", "ICICI Bank"),
    ("UTIB", "Axis Bank"),
    ("KKBK", "Kotak Mahindra Bank"),
    ("PUNB", "Punjab National Bank"),
    ("BARB", "Bank of Baroda"),
    ("CNRB", "Canara Bank"),
    ("UBIN", "Union Bank of India"),
    ("IOBA", "Indian Overseas Bank"),
    ("BKID", "Bank of India"),
    ("CBIN", "Central Bank of India"),
    ("IDIB", "Indian Bank"),
    ("MAHB", "Bank of Maharashtra"),
    ("UCBA", "UCO Bank"),
    ("PSIB", "Punjab & Sind Bank"),
    ("YESB", "Yes Bank"),
    ("INDB", "IndusInd Bank"),
    ("FDRL", "Federal Bank"),
    ("IDFB", "IDFC First Bank"),
    ("RATN", "RBL Bank"),
    ("SIBL", "South Indian Bank"),
    ("KARB", "Karnataka Bank"),
    ("JAKA", "J&K Bank"),
    ("CITI", "Citibank"),
    ("HSBC", "HSBC Bank"),
    ("SCBL", "Standard Chartered Bank"),
    ("DBSS", "DBS Bank"),
    ("ABNA", "ABN Amro Bank"),
    ("DEUT", "Deutsche Bank"),
];

/// IFSC information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IfscInfo {
    pub ifsc: String,
    pub bank_code: String,
    pub bank_name: String,
    pub branch_code: String,
    pub valid: bool,
}

/// Validate IFSC format
#[wasm_bindgen]
pub fn validate_ifsc(ifsc: &str) -> bool {
    validate_ifsc_internal(ifsc).is_ok()
}

/// Full IFSC validation with details
pub fn validate_ifsc_full(ifsc: &str) -> ValidationResult {
    match validate_ifsc_internal(ifsc) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(ifsc, details)
        }
        Err(error) => ValidationResult::err(ifsc, &error),
    }
}

/// Internal IFSC validation
pub fn validate_ifsc_internal(ifsc: &str) -> Result<IfscInfo, String> {
    let ifsc = ifsc.trim().to_uppercase();

    // Check length
    if ifsc.len() != 11 {
        return Err(format!("IFSC must be 11 characters, got {}", ifsc.len()));
    }

    // Check format
    if !IFSC_REGEX.is_match(&ifsc) {
        return Err("Invalid IFSC format. Format: AAAA0BBBBBB".to_string());
    }

    // Check 5th character is '0'
    if ifsc.chars().nth(4) != Some('0') {
        return Err("5th character must be '0'".to_string());
    }

    // Extract components
    let bank_code = &ifsc[0..4];
    let branch_code = &ifsc[5..11];

    // Look up bank name
    let bank_name = KNOWN_BANKS
        .iter()
        .find(|(code, _)| *code == bank_code)
        .map(|(_, name)| name.to_string())
        .unwrap_or_else(|| "Unknown Bank".to_string());

    Ok(IfscInfo {
        ifsc: ifsc.clone(),
        bank_code: bank_code.to_string(),
        bank_name,
        branch_code: branch_code.to_string(),
        valid: true,
    })
}

/// Parse IFSC and return components
#[wasm_bindgen]
pub fn parse_ifsc(ifsc: &str) -> JsValue {
    match validate_ifsc_internal(ifsc) {
        Ok(info) => serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL),
        Err(_) => JsValue::NULL,
    }
}

/// Get bank code from IFSC
#[wasm_bindgen]
pub fn get_bank_code(ifsc: &str) -> String {
    let ifsc = ifsc.trim().to_uppercase();
    if ifsc.len() >= 4 {
        ifsc[0..4].to_string()
    } else {
        String::new()
    }
}

/// Get bank name from IFSC
#[wasm_bindgen]
pub fn get_bank_name(ifsc: &str) -> String {
    let bank_code = get_bank_code(ifsc);
    KNOWN_BANKS
        .iter()
        .find(|(code, _)| *code == bank_code)
        .map(|(_, name)| name.to_string())
        .unwrap_or_else(|| "Unknown Bank".to_string())
}

/// Get branch code from IFSC
#[wasm_bindgen]
pub fn get_branch_code(ifsc: &str) -> String {
    let ifsc = ifsc.trim().to_uppercase();
    if ifsc.len() == 11 {
        ifsc[5..11].to_string()
    } else {
        String::new()
    }
}

/// Validate bank account number format
#[wasm_bindgen]
pub fn validate_bank_account(account_number: &str) -> bool {
    let account = account_number.trim();

    // Bank accounts are typically 9-18 digits
    if account.len() < 9 || account.len() > 18 {
        return false;
    }

    // Must be all digits
    account.chars().all(|c| c.is_ascii_digit())
}

/// Mask bank account for display
#[wasm_bindgen]
pub fn mask_bank_account(account_number: &str) -> String {
    let account = account_number.trim();
    let len = account.len();

    if len <= 4 {
        return account.to_string();
    }

    let visible = 4;
    let masked = len - visible;
    format!("{}{}", "X".repeat(masked), &account[masked..])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_ifsc() {
        assert!(validate_ifsc("SBIN0001234"));
        assert!(validate_ifsc("HDFC0001234"));
        assert!(validate_ifsc("ICIC0000001"));
    }

    #[test]
    fn test_invalid_ifsc() {
        assert!(!validate_ifsc("SBIN1001234")); // 5th char not 0
        assert!(!validate_ifsc("SBI01234")); // Too short
        assert!(!validate_ifsc("12340001234")); // Starts with digits
    }

    #[test]
    fn test_parse() {
        let result = validate_ifsc_internal("SBIN0001234").unwrap();
        assert_eq!(result.bank_code, "SBIN");
        assert_eq!(result.bank_name, "State Bank of India");
        assert_eq!(result.branch_code, "001234");
    }

    #[test]
    fn test_bank_account() {
        assert!(validate_bank_account("123456789012")); // 12 digits
        assert!(validate_bank_account("12345678901234567")); // 17 digits
        assert!(!validate_bank_account("12345678")); // Too short
        assert!(!validate_bank_account("123456789ABC")); // Contains letters
    }

    #[test]
    fn test_mask_account() {
        assert_eq!(mask_bank_account("1234567890"), "XXXXXX7890");
        assert_eq!(mask_bank_account("123456789012"), "XXXXXXXX9012");
    }
}
