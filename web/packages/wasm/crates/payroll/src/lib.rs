//! Samavaya Payroll - Salary and statutory calculations
//!
//! This crate provides:
//! - Income tax calculation (Old and New regime)
//! - PF (Provident Fund) calculations
//! - ESI (Employee State Insurance) calculations
//! - Professional Tax
//! - CTC breakdown
//! - Gratuity calculations

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

mod income_tax;
mod statutory;
mod ctc;

pub use income_tax::*;
pub use statutory::*;
pub use ctc::*;

/// Initialize payroll module (called from core init)
fn payroll_init() {
    log::info!("Samavaya Payroll module initialized");
}

/// Tax regime
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TaxRegime {
    Old,
    New,
}

/// Financial year
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FinancialYear {
    pub year: String,
    pub start_date: String,
    pub end_date: String,
}

impl Default for FinancialYear {
    fn default() -> Self {
        Self {
            year: "2024-25".to_string(),
            start_date: "2024-04-01".to_string(),
            end_date: "2025-03-31".to_string(),
        }
    }
}
