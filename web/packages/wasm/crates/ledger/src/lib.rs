//! Samavaya Ledger - Double-entry accounting operations
//!
//! This crate provides:
//! - Journal entry validation (debit = credit)
//! - Account balance calculations
//! - Trial balance computation
//! - Financial statement preparations
//! - Aging calculations (AR/AP)
//! - Bank reconciliation matching

pub mod journal;
pub mod balance;
pub mod aging;
pub mod reconciliation;
pub mod trial_balance;

pub use journal::*;
pub use balance::*;
pub use aging::*;
pub use reconciliation::*;
pub use trial_balance::*;

use wasm_bindgen::prelude::*;

/// Initialize ledger module (called from core init)
fn ledger_init() {
    log::info!("Samavaya Ledger module initialized");
}
