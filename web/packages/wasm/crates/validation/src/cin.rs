//! CIN (Corporate Identity Number) validation
//!
//! CIN Format: U12345MH2020PTC123456
//! - Position 1: Listing status (U = Unlisted, L = Listed)
//! - Position 2-6: Industry code (5 digits)
//! - Position 7-8: State code (2 letters)
//! - Position 9-12: Year of incorporation (4 digits)
//! - Position 13-15: Company type (PTC, PLC, GOV, etc.)
//! - Position 16-21: Registration number (6 digits)

use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;

lazy_static! {
    static ref CIN_REGEX: Regex = Regex::new(
        r"^[LU][0-9]{5}[A-Z]{2}[0-9]{4}[A-Z]{3}[0-9]{6}$"
    ).unwrap();

    static ref LLPIN_REGEX: Regex = Regex::new(
        r"^[A-Z]{3}-[0-9]{4}$"
    ).unwrap();

    static ref FCRN_REGEX: Regex = Regex::new(
        r"^F[0-9]{5}[A-Z]{2}[0-9]{4}[A-Z]{3}[0-9]{6}$"
    ).unwrap();
}

/// Company types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CompanyType {
    /// PTC - Private Limited Company
    PrivateLimited,
    /// PLC - Public Limited Company
    PublicLimited,
    /// GOV - Government Company
    Government,
    /// OPC - One Person Company
    OnePerson,
    /// FTC - Foreign Company
    Foreign,
    /// NPL - Not for Profit License
    NotForProfit,
    /// ULT - Unlimited Company
    Unlimited,
}

impl CompanyType {
    pub fn from_code(code: &str) -> Option<Self> {
        match code.to_uppercase().as_str() {
            "PTC" => Some(CompanyType::PrivateLimited),
            "PLC" => Some(CompanyType::PublicLimited),
            "GOV" => Some(CompanyType::Government),
            "OPC" => Some(CompanyType::OnePerson),
            "FTC" => Some(CompanyType::Foreign),
            "NPL" => Some(CompanyType::NotForProfit),
            "ULT" => Some(CompanyType::Unlimited),
            _ => None,
        }
    }

    pub fn description(&self) -> &'static str {
        match self {
            CompanyType::PrivateLimited => "Private Limited Company",
            CompanyType::PublicLimited => "Public Limited Company",
            CompanyType::Government => "Government Company",
            CompanyType::OnePerson => "One Person Company",
            CompanyType::Foreign => "Foreign Company",
            CompanyType::NotForProfit => "Not for Profit (Section 8)",
            CompanyType::Unlimited => "Unlimited Company",
        }
    }
}

/// CIN information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CinInfo {
    pub cin: String,
    pub listing_status: String,
    pub industry_code: String,
    pub state_code: String,
    pub year_of_incorporation: String,
    pub company_type: String,
    pub company_type_code: String,
    pub registration_number: String,
    pub valid: bool,
}

/// Validate CIN format
#[wasm_bindgen]
pub fn validate_cin(cin: &str) -> bool {
    validate_cin_internal(cin).is_ok()
}

/// Full CIN validation with details
pub fn validate_cin_full(cin: &str) -> ValidationResult {
    match validate_cin_internal(cin) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(cin, details)
        }
        Err(error) => ValidationResult::err(cin, &error),
    }
}

/// Internal CIN validation
pub fn validate_cin_internal(cin: &str) -> Result<CinInfo, String> {
    let cin = cin.trim().to_uppercase();

    // Check length
    if cin.len() != 21 {
        return Err(format!("CIN must be 21 characters, got {}", cin.len()));
    }

    // Check format
    if !CIN_REGEX.is_match(&cin) {
        return Err("Invalid CIN format".to_string());
    }

    // Extract components
    let listing_status = match &cin[0..1] {
        "L" => "Listed",
        "U" => "Unlisted",
        _ => "Unknown",
    };
    let industry_code = &cin[1..6];
    let state_code = &cin[6..8];
    let year = &cin[8..12];
    let company_type_code = &cin[12..15];
    let registration_number = &cin[15..21];

    // Validate year
    let year_num: u32 = year.parse().map_err(|_| "Invalid year")?;
    if year_num < 1850 || year_num > 2100 {
        return Err(format!("Invalid year of incorporation: {}", year));
    }

    // Get company type description
    let company_type = CompanyType::from_code(company_type_code)
        .map(|t| t.description())
        .unwrap_or("Unknown");

    Ok(CinInfo {
        cin: cin.clone(),
        listing_status: listing_status.to_string(),
        industry_code: industry_code.to_string(),
        state_code: state_code.to_string(),
        year_of_incorporation: year.to_string(),
        company_type: company_type.to_string(),
        company_type_code: company_type_code.to_string(),
        registration_number: registration_number.to_string(),
        valid: true,
    })
}

/// Validate LLPIN (LLP Identification Number)
#[wasm_bindgen]
pub fn validate_llpin(llpin: &str) -> bool {
    let llpin = llpin.trim().to_uppercase();
    LLPIN_REGEX.is_match(&llpin)
}

/// Parse CIN and return components
#[wasm_bindgen]
pub fn parse_cin(cin: &str) -> JsValue {
    match validate_cin_internal(cin) {
        Ok(info) => serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL),
        Err(_) => JsValue::NULL,
    }
}

/// Check if company is listed
#[wasm_bindgen]
pub fn is_listed_company(cin: &str) -> bool {
    let cin = cin.trim().to_uppercase();
    cin.starts_with('L')
}

/// Get year of incorporation from CIN
#[wasm_bindgen]
pub fn get_incorporation_year(cin: &str) -> String {
    let cin = cin.trim().to_uppercase();
    if cin.len() >= 12 {
        cin[8..12].to_string()
    } else {
        String::new()
    }
}

/// Get company type from CIN
#[wasm_bindgen]
pub fn get_company_type(cin: &str) -> String {
    let cin = cin.trim().to_uppercase();
    if cin.len() >= 15 {
        let code = &cin[12..15];
        CompanyType::from_code(code)
            .map(|t| t.description().to_string())
            .unwrap_or_else(|| "Unknown".to_string())
    } else {
        "Unknown".to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_cin() {
        assert!(validate_cin("U12345MH2020PTC123456"));
        assert!(validate_cin("L67890DL2015PLC654321"));
    }

    #[test]
    fn test_invalid_cin() {
        assert!(!validate_cin("U12345MH2020PTC12345")); // Too short
        assert!(!validate_cin("X12345MH2020PTC123456")); // Invalid prefix
    }

    #[test]
    fn test_parse() {
        let result = validate_cin_internal("U12345MH2020PTC123456").unwrap();
        assert_eq!(result.listing_status, "Unlisted");
        assert_eq!(result.industry_code, "12345");
        assert_eq!(result.state_code, "MH");
        assert_eq!(result.year_of_incorporation, "2020");
        assert_eq!(result.company_type, "Private Limited Company");
        assert_eq!(result.registration_number, "123456");
    }

    #[test]
    fn test_listed() {
        assert!(is_listed_company("L67890DL2015PLC654321"));
        assert!(!is_listed_company("U12345MH2020PTC123456"));
    }
}
