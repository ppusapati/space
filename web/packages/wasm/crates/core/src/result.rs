//! Result types for WASM operations

use crate::error::CoreError;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Standard result type for core operations
pub type CoreResult<T> = Result<T, CoreError>;

/// Generic API response wrapper
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiResult<T> {
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
    pub code: Option<String>,
}

impl<T: Serialize> ApiResult<T> {
    pub fn ok(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
            code: None,
        }
    }

    pub fn err(error: &str, code: &str) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(error.to_string()),
            code: Some(code.to_string()),
        }
    }

    pub fn to_js(&self) -> JsValue
    where
        T: Serialize,
    {
        serde_wasm_bindgen::to_value(self).unwrap_or(JsValue::NULL)
    }
}

/// Calculation result with breakdown
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CalculationResult {
    pub result: String,
    pub breakdown: Vec<CalculationStep>,
    pub warnings: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CalculationStep {
    pub label: String,
    pub value: String,
    pub formula: Option<String>,
}

impl CalculationResult {
    pub fn new(result: &str) -> Self {
        Self {
            result: result.to_string(),
            breakdown: vec![],
            warnings: vec![],
        }
    }

    pub fn with_step(mut self, label: &str, value: &str, formula: Option<&str>) -> Self {
        self.breakdown.push(CalculationStep {
            label: label.to_string(),
            value: value.to_string(),
            formula: formula.map(|s| s.to_string()),
        });
        self
    }

    pub fn with_warning(mut self, warning: &str) -> Self {
        self.warnings.push(warning.to_string());
        self
    }
}

/// Helper to convert Result to JsValue
pub fn to_js_result<T: Serialize, E: std::fmt::Display>(result: Result<T, E>) -> JsValue {
    match result {
        Ok(data) => {
            let response = ApiResult::<T>::ok(data);
            serde_wasm_bindgen::to_value(&response).unwrap_or(JsValue::NULL)
        }
        Err(e) => {
            let response = ApiResult::<()>::err(&e.to_string(), "ERROR");
            serde_wasm_bindgen::to_value(&response).unwrap_or(JsValue::NULL)
        }
    }
}

/// Helper to parse JsValue input
pub fn from_js<T: for<'de> Deserialize<'de>>(value: JsValue) -> CoreResult<T> {
    serde_wasm_bindgen::from_value(value).map_err(|e| CoreError::ParseError(e.to_string()))
}
