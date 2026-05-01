//! Samavaya Validation - Indian tax identifier validation
//!
//! This crate provides validation for:
//! - GSTIN (Goods and Services Tax Identification Number)
//! - PAN (Permanent Account Number)
//! - TAN (Tax Deduction Account Number)
//! - CIN (Corporate Identity Number)
//! - UAN (Universal Account Number)
//! - Aadhaar
//! - IFSC (Indian Financial System Code)
//! - Bank Account
//! - Mobile number
//! - Email
//! - Pincode

pub mod gstin;
pub mod pan;
pub mod tan;
pub mod cin;
pub mod ifsc;
pub mod aadhaar;
pub mod contact;
pub mod password;

pub use gstin::*;
pub use pan::*;
pub use tan::*;
pub use cin::*;
pub use ifsc::*;
pub use aadhaar::*;
pub use contact::*;
pub use password::*;

use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};

/// Initialize validation module (called from core init)
fn validation_init() {
    log::info!("Samavaya Validation module initialized");
}

/// Validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationResult {
    pub valid: bool,
    pub value: String,
    pub normalized: Option<String>,
    pub error: Option<String>,
    pub details: Option<serde_json::Value>,
}

impl ValidationResult {
    pub fn ok(value: &str) -> Self {
        Self {
            valid: true,
            value: value.to_string(),
            normalized: Some(value.to_uppercase()),
            error: None,
            details: None,
        }
    }

    pub fn ok_with_details(value: &str, details: serde_json::Value) -> Self {
        Self {
            valid: true,
            value: value.to_string(),
            normalized: Some(value.to_uppercase()),
            error: None,
            details: Some(details),
        }
    }

    pub fn err(value: &str, error: &str) -> Self {
        Self {
            valid: false,
            value: value.to_string(),
            normalized: None,
            error: Some(error.to_string()),
            details: None,
        }
    }

    pub fn to_js(&self) -> JsValue {
        serde_wasm_bindgen::to_value(self).unwrap_or(JsValue::NULL)
    }
}

/// Validate any identifier based on type
#[wasm_bindgen]
pub fn validate(identifier_type: &str, value: &str) -> JsValue {
    let result = match identifier_type.to_uppercase().as_str() {
        "GSTIN" | "GST" => validate_gstin_full(value),
        "PAN" => validate_pan_full(value),
        "TAN" => validate_tan_full(value),
        "CIN" => validate_cin_full(value),
        "IFSC" => validate_ifsc_full(value),
        "AADHAAR" => validate_aadhaar_full(value),
        "MOBILE" | "PHONE" => validate_mobile_full(value),
        "EMAIL" => validate_email_full(value),
        "PINCODE" => validate_pincode_full(value),
        _ => ValidationResult::err(value, &format!("Unknown identifier type: {}", identifier_type)),
    };
    result.to_js()
}
