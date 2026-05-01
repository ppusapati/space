//! PAN (Permanent Account Number) validation
//!
//! PAN Format: AAAAA0000A
//! - Position 1-3: Alphabetic series (AAA-ZZZ)
//! - Position 4: Type of assessee (C, P, H, F, A, T, B, L, J, G)
//! - Position 5: First character of surname/name
//! - Position 6-9: Sequential number (0001-9999)
//! - Position 10: Alphabetic check digit

use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::ValidationResult;

lazy_static! {
    static ref PAN_REGEX: Regex = Regex::new(
        r"^[A-Z]{3}[ABCFGHLJPTK][A-Z][0-9]{4}[A-Z]$"
    ).unwrap();
}

/// PAN holder types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PanType {
    /// A - Association of Persons (AOP)
    AssociationOfPersons,
    /// B - Body of Individuals (BOI)
    BodyOfIndividuals,
    /// C - Company
    Company,
    /// F - Firm/LLP
    Firm,
    /// G - Government
    Government,
    /// H - Hindu Undivided Family (HUF)
    Huf,
    /// J - Artificial Juridical Person
    ArtificialJuridicalPerson,
    /// K - Krish (used for TAN)
    Krish,
    /// L - Local Authority
    LocalAuthority,
    /// P - Person/Individual
    Individual,
    /// T - Trust
    Trust,
}

impl PanType {
    pub fn from_char(c: char) -> Option<Self> {
        match c {
            'A' => Some(PanType::AssociationOfPersons),
            'B' => Some(PanType::BodyOfIndividuals),
            'C' => Some(PanType::Company),
            'F' => Some(PanType::Firm),
            'G' => Some(PanType::Government),
            'H' => Some(PanType::Huf),
            'J' => Some(PanType::ArtificialJuridicalPerson),
            'K' => Some(PanType::Krish),
            'L' => Some(PanType::LocalAuthority),
            'P' => Some(PanType::Individual),
            'T' => Some(PanType::Trust),
            _ => None,
        }
    }

    pub fn code(&self) -> char {
        match self {
            PanType::AssociationOfPersons => 'A',
            PanType::BodyOfIndividuals => 'B',
            PanType::Company => 'C',
            PanType::Firm => 'F',
            PanType::Government => 'G',
            PanType::Huf => 'H',
            PanType::ArtificialJuridicalPerson => 'J',
            PanType::Krish => 'K',
            PanType::LocalAuthority => 'L',
            PanType::Individual => 'P',
            PanType::Trust => 'T',
        }
    }

    pub fn description(&self) -> &'static str {
        match self {
            PanType::AssociationOfPersons => "Association of Persons (AOP)",
            PanType::BodyOfIndividuals => "Body of Individuals (BOI)",
            PanType::Company => "Company",
            PanType::Firm => "Firm/LLP",
            PanType::Government => "Government",
            PanType::Huf => "Hindu Undivided Family (HUF)",
            PanType::ArtificialJuridicalPerson => "Artificial Juridical Person",
            PanType::Krish => "Krish",
            PanType::LocalAuthority => "Local Authority",
            PanType::Individual => "Individual",
            PanType::Trust => "Trust",
        }
    }
}

/// PAN information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PanInfo {
    pub pan: String,
    pub holder_type: String,
    pub holder_type_code: String,
    pub name_initial: String,
    pub valid: bool,
}

/// Validate PAN format
#[wasm_bindgen]
pub fn validate_pan(pan: &str) -> bool {
    validate_pan_internal(pan).is_ok()
}

/// Full PAN validation with details
pub fn validate_pan_full(pan: &str) -> ValidationResult {
    match validate_pan_internal(pan) {
        Ok(info) => {
            let details = serde_json::to_value(&info).unwrap_or(serde_json::Value::Null);
            ValidationResult::ok_with_details(pan, details)
        }
        Err(error) => ValidationResult::err(pan, &error),
    }
}

/// Internal PAN validation
pub fn validate_pan_internal(pan: &str) -> Result<PanInfo, String> {
    let pan = pan.trim().to_uppercase();

    // Check length
    if pan.len() != 10 {
        return Err(format!("PAN must be 10 characters, got {}", pan.len()));
    }

    // Check format
    if !PAN_REGEX.is_match(&pan) {
        return Err("Invalid PAN format. Format: AAAAA0000A".to_string());
    }

    // Extract holder type (4th character)
    let type_char = pan.chars().nth(3).unwrap();
    let pan_type = PanType::from_char(type_char)
        .ok_or_else(|| format!("Invalid holder type code: {}", type_char))?;

    // Extract name initial (5th character)
    let name_initial = pan.chars().nth(4).unwrap().to_string();

    Ok(PanInfo {
        pan: pan.clone(),
        holder_type: pan_type.description().to_string(),
        holder_type_code: type_char.to_string(),
        name_initial,
        valid: true,
    })
}

/// Get holder type from PAN
#[wasm_bindgen]
pub fn get_pan_holder_type(pan: &str) -> String {
    let pan = pan.trim().to_uppercase();
    if pan.len() >= 4 {
        let type_char = pan.chars().nth(3).unwrap();
        PanType::from_char(type_char)
            .map(|t| t.description().to_string())
            .unwrap_or_else(|| "Unknown".to_string())
    } else {
        "Unknown".to_string()
    }
}

/// Check if PAN belongs to a company
#[wasm_bindgen]
pub fn is_company_pan(pan: &str) -> bool {
    let pan = pan.trim().to_uppercase();
    pan.len() >= 4 && pan.chars().nth(3) == Some('C')
}

/// Check if PAN belongs to an individual
#[wasm_bindgen]
pub fn is_individual_pan(pan: &str) -> bool {
    let pan = pan.trim().to_uppercase();
    pan.len() >= 4 && pan.chars().nth(3) == Some('P')
}

/// Parse PAN and return all components
#[wasm_bindgen]
pub fn parse_pan(pan: &str) -> JsValue {
    match validate_pan_internal(pan) {
        Ok(info) => serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL),
        Err(_) => JsValue::NULL,
    }
}

/// Format PAN with proper capitalization
#[wasm_bindgen]
pub fn format_pan(pan: &str) -> String {
    pan.trim().to_uppercase()
}

/// Mask PAN for display (e.g., ABCDE1234F -> ABCXX****F)
#[wasm_bindgen]
pub fn mask_pan(pan: &str) -> String {
    let pan = pan.trim().to_uppercase();
    if pan.len() != 10 {
        return pan;
    }

    format!("{}XX****{}", &pan[0..3], &pan[9..10])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_pan() {
        assert!(validate_pan("ABCDE1234F"));
        assert!(validate_pan("AABCP1234K")); // Individual
        assert!(validate_pan("AABCC1234K")); // Company
    }

    #[test]
    fn test_invalid_pan() {
        assert!(!validate_pan("ABCDE123F")); // Too short
        assert!(!validate_pan("12345ABCDE")); // Wrong format
        assert!(!validate_pan("ABCDE12345")); // Ends with digit
    }

    #[test]
    fn test_holder_type() {
        assert_eq!(get_pan_holder_type("AABCP1234K"), "Individual");
        assert_eq!(get_pan_holder_type("AABCC1234K"), "Company");
        assert_eq!(get_pan_holder_type("AABCF1234K"), "Firm/LLP");
        assert_eq!(get_pan_holder_type("AABCH1234K"), "Hindu Undivided Family (HUF)");
    }

    #[test]
    fn test_is_company() {
        assert!(is_company_pan("AABCC1234K"));
        assert!(!is_company_pan("AABCP1234K"));
    }

    #[test]
    fn test_mask() {
        assert_eq!(mask_pan("ABCDE1234F"), "ABCXX****F");
    }
}
