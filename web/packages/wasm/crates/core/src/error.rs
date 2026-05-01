//! Error types for WASM modules

use serde::{Deserialize, Serialize};
use thiserror::Error;
use wasm_bindgen::prelude::*;

/// Core error types used across all WASM modules
#[derive(Error, Debug, Clone, Serialize, Deserialize)]
pub enum CoreError {
    #[error("Invalid input: {0}")]
    InvalidInput(String),

    #[error("Validation failed: {0}")]
    ValidationError(String),

    #[error("Calculation error: {0}")]
    CalculationError(String),

    #[error("Parse error: {0}")]
    ParseError(String),

    #[error("Invalid state: {0}")]
    InvalidState(String),

    #[error("Not found: {0}")]
    NotFound(String),

    #[error("Invalid date: {0}")]
    InvalidDate(String),

    #[error("Invalid amount: {0}")]
    InvalidAmount(String),

    #[error("Division by zero")]
    DivisionByZero,

    #[error("Overflow error")]
    Overflow,

    #[error("Configuration error: {0}")]
    ConfigError(String),

    #[error("Internal error: {0}")]
    InternalError(String),
}

// Note: wasm-bindgen provides automatic From<E> for JsError for any E: std::error::Error
// So we don't need to implement it manually

impl From<rust_decimal::Error> for CoreError {
    fn from(e: rust_decimal::Error) -> Self {
        CoreError::ParseError(format!("Decimal parse error: {}", e))
    }
}

impl From<chrono::ParseError> for CoreError {
    fn from(e: chrono::ParseError) -> Self {
        CoreError::InvalidDate(format!("Date parse error: {}", e))
    }
}

impl From<serde_json::Error> for CoreError {
    fn from(e: serde_json::Error) -> Self {
        CoreError::ParseError(format!("JSON error: {}", e))
    }
}

/// Validation error with field information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FieldError {
    pub field: String,
    pub message: String,
    pub code: String,
}

impl FieldError {
    pub fn new(field: &str, message: &str, code: &str) -> Self {
        Self {
            field: field.to_string(),
            message: message.to_string(),
            code: code.to_string(),
        }
    }
}

/// Validation result containing multiple field errors
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationResult {
    pub valid: bool,
    pub errors: Vec<FieldError>,
}

impl ValidationResult {
    pub fn ok() -> Self {
        Self {
            valid: true,
            errors: vec![],
        }
    }

    pub fn error(errors: Vec<FieldError>) -> Self {
        Self {
            valid: false,
            errors,
        }
    }

    pub fn single_error(field: &str, message: &str, code: &str) -> Self {
        Self {
            valid: false,
            errors: vec![FieldError::new(field, message, code)],
        }
    }

    pub fn add_error(&mut self, field: &str, message: &str, code: &str) {
        self.valid = false;
        self.errors.push(FieldError::new(field, message, code));
    }

    pub fn merge(&mut self, other: ValidationResult) {
        if !other.valid {
            self.valid = false;
            self.errors.extend(other.errors);
        }
    }
}

#[wasm_bindgen]
pub fn create_validation_error(field: &str, message: &str, code: &str) -> JsValue {
    let result = ValidationResult::single_error(field, message, code);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}
