//! Contact information validation (mobile, email, pincode)

use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;

lazy_static! {
    // Indian mobile: 10 digits starting with 6-9
    static ref MOBILE_REGEX: Regex = Regex::new(
        r"^[6-9][0-9]{9}$"
    ).unwrap();

    // Email regex
    static ref EMAIL_REGEX: Regex = Regex::new(
        r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    ).unwrap();

    // Indian pincode: 6 digits, first digit 1-9
    static ref PINCODE_REGEX: Regex = Regex::new(
        r"^[1-9][0-9]{5}$"
    ).unwrap();

    // Landline with STD code
    static ref LANDLINE_REGEX: Regex = Regex::new(
        r"^0[1-9][0-9]{1,4}-?[0-9]{6,8}$"
    ).unwrap();
}

/// Mobile validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MobileInfo {
    pub number: String,
    pub formatted: String,
    pub country_code: String,
    pub valid: bool,
}

/// Email validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmailInfo {
    pub email: String,
    pub local_part: String,
    pub domain: String,
    pub valid: bool,
}

/// Pincode validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PincodeInfo {
    pub pincode: String,
    pub valid: bool,
}

// Mobile validation

/// Validate Indian mobile number
#[wasm_bindgen]
pub fn validate_mobile(mobile: &str) -> bool {
    validate_mobile_internal(mobile).is_ok()
}

/// Full mobile validation with details
pub fn validate_mobile_full(mobile: &str) -> ValidationResult {
    match validate_mobile_internal(mobile) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(mobile, details)
        }
        Err(error) => ValidationResult::err(mobile, &error),
    }
}

/// Internal mobile validation
pub fn validate_mobile_internal(mobile: &str) -> Result<MobileInfo, String> {
    // Remove common prefixes and non-digits
    let mut cleaned: String = mobile.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    // Remove country code if present
    if cleaned.starts_with("91") && cleaned.len() == 12 {
        cleaned = cleaned[2..].to_string();
    } else if cleaned.starts_with("0") && cleaned.len() == 11 {
        cleaned = cleaned[1..].to_string();
    }

    // Check format
    if !MOBILE_REGEX.is_match(&cleaned) {
        return Err("Invalid mobile number. Must be 10 digits starting with 6-9".to_string());
    }

    let formatted = format!("+91 {} {} {}",
        &cleaned[0..5],
        &cleaned[5..7],
        &cleaned[7..10]
    );

    Ok(MobileInfo {
        number: cleaned.clone(),
        formatted,
        country_code: "+91".to_string(),
        valid: true,
    })
}

/// Format mobile number
#[wasm_bindgen]
pub fn format_mobile(mobile: &str) -> String {
    validate_mobile_internal(mobile)
        .map(|info| info.formatted)
        .unwrap_or_else(|_| mobile.to_string())
}

/// Mask mobile for display (XXXXX XX789)
#[wasm_bindgen]
pub fn mask_mobile(mobile: &str) -> String {
    let cleaned: String = mobile.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    if cleaned.len() >= 10 {
        let last_10 = &cleaned[cleaned.len() - 10..];
        format!("XXXXX XX{}", &last_10[7..])
    } else {
        mobile.to_string()
    }
}

// Email validation

/// Validate email address
#[wasm_bindgen]
pub fn validate_email(email: &str) -> bool {
    validate_email_internal(email).is_ok()
}

/// Full email validation with details
pub fn validate_email_full(email: &str) -> ValidationResult {
    match validate_email_internal(email) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(email, details)
        }
        Err(error) => ValidationResult::err(email, &error),
    }
}

/// Internal email validation
pub fn validate_email_internal(email: &str) -> Result<EmailInfo, String> {
    let email = email.trim().to_lowercase();

    if email.is_empty() {
        return Err("Email cannot be empty".to_string());
    }

    if !EMAIL_REGEX.is_match(&email) {
        return Err("Invalid email format".to_string());
    }

    // Split into local and domain parts
    let parts: Vec<&str> = email.split('@').collect();
    if parts.len() != 2 {
        return Err("Invalid email format".to_string());
    }

    Ok(EmailInfo {
        email: email.clone(),
        local_part: parts[0].to_string(),
        domain: parts[1].to_string(),
        valid: true,
    })
}

/// Get domain from email
#[wasm_bindgen]
pub fn get_email_domain(email: &str) -> String {
    validate_email_internal(email)
        .map(|info| info.domain)
        .unwrap_or_default()
}

/// Mask email for display (a***@example.com)
#[wasm_bindgen]
pub fn mask_email(email: &str) -> String {
    if let Ok(info) = validate_email_internal(email) {
        let local = &info.local_part;
        if local.len() <= 2 {
            format!("{}***@{}", local, info.domain)
        } else {
            format!("{}***@{}", &local[0..1], info.domain)
        }
    } else {
        email.to_string()
    }
}

// Pincode validation

/// Validate Indian pincode
#[wasm_bindgen]
pub fn validate_pincode(pincode: &str) -> bool {
    validate_pincode_internal(pincode).is_ok()
}

/// Full pincode validation with details
pub fn validate_pincode_full(pincode: &str) -> ValidationResult {
    match validate_pincode_internal(pincode) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(pincode, details)
        }
        Err(error) => ValidationResult::err(pincode, &error),
    }
}

/// Internal pincode validation
pub fn validate_pincode_internal(pincode: &str) -> Result<PincodeInfo, String> {
    let pincode = pincode.trim();

    if pincode.len() != 6 {
        return Err(format!("Pincode must be 6 digits, got {}", pincode.len()));
    }

    if !PINCODE_REGEX.is_match(pincode) {
        return Err("Invalid pincode. Must be 6 digits and cannot start with 0".to_string());
    }

    Ok(PincodeInfo {
        pincode: pincode.to_string(),
        valid: true,
    })
}

/// Get postal region from pincode (first digit)
#[wasm_bindgen]
pub fn get_postal_region(pincode: &str) -> String {
    let pincode = pincode.trim();
    if pincode.is_empty() {
        return String::new();
    }

    let first_digit = pincode.chars().next().unwrap_or('0');
    match first_digit {
        '1' => "Delhi, Haryana, Punjab, Himachal Pradesh, J&K".to_string(),
        '2' => "Uttar Pradesh, Uttarakhand".to_string(),
        '3' => "Rajasthan, Gujarat".to_string(),
        '4' => "Maharashtra, Madhya Pradesh, Chhattisgarh".to_string(),
        '5' => "Andhra Pradesh, Telangana, Karnataka".to_string(),
        '6' => "Tamil Nadu, Kerala".to_string(),
        '7' => "West Bengal, Odisha, NE States".to_string(),
        '8' => "Bihar, Jharkhand".to_string(),
        '9' => "Army Post Office".to_string(),
        _ => "Unknown".to_string(),
    }
}

// Landline validation

/// Validate Indian landline number
#[wasm_bindgen]
pub fn validate_landline(landline: &str) -> bool {
    let cleaned: String = landline.chars()
        .filter(|c| c.is_ascii_digit() || *c == '-')
        .collect();

    LANDLINE_REGEX.is_match(&cleaned)
}

/// Format landline with STD code
#[wasm_bindgen]
pub fn format_landline(landline: &str) -> String {
    let digits: String = landline.chars()
        .filter(|c| c.is_ascii_digit())
        .collect();

    if digits.len() < 10 || digits.len() > 12 {
        return landline.to_string();
    }

    // Assume 4-digit STD code for major cities, else varies
    if digits.starts_with("0") {
        let std_len = if digits.len() == 11 { 4 } else { 3 };
        format!("{}-{}", &digits[..std_len], &digits[std_len..])
    } else {
        landline.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_mobile() {
        assert!(validate_mobile("9876543210"));
        assert!(validate_mobile("919876543210")); // With country code
        assert!(validate_mobile("+91 98765 43210")); // Formatted
        assert!(!validate_mobile("1234567890")); // Starts with 1
        assert!(!validate_mobile("987654321")); // 9 digits
    }

    #[test]
    fn test_email() {
        assert!(validate_email("test@example.com"));
        assert!(validate_email("user.name+tag@domain.co.in"));
        assert!(!validate_email("invalid"));
        assert!(!validate_email("@nodomain.com"));
    }

    #[test]
    fn test_pincode() {
        assert!(validate_pincode("110001")); // Delhi
        assert!(validate_pincode("500001")); // Hyderabad
        assert!(!validate_pincode("000001")); // Starts with 0
        assert!(!validate_pincode("11000")); // 5 digits
    }

    #[test]
    fn test_mask_mobile() {
        assert_eq!(mask_mobile("9876543210"), "XXXXX XX210");
    }

    #[test]
    fn test_mask_email() {
        assert_eq!(mask_email("test@example.com"), "t***@example.com");
    }
}
