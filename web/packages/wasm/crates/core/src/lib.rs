//! Samavaya Core - Shared types and utilities for WASM modules
//!
//! This crate provides common functionality used across all WASM modules:
//! - Decimal arithmetic with proper precision
//! - Date/time handling
//! - Error types
//! - Common data structures
//! - Indian locale utilities (states, currencies, etc.)

pub mod decimal;
pub mod error;
pub mod indian;
pub mod money;
pub mod result;
pub mod types;

pub use decimal::*;
pub use error::*;
pub use indian::*;
pub use money::*;
pub use result::*;
pub use types::*;

use wasm_bindgen::prelude::*;

#[wasm_bindgen(start)]
pub fn init() {
    console_log::init_with_level(log::Level::Debug).ok();
    log::info!("Samavaya WASM Core initialized");
}

#[wasm_bindgen]
pub fn version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}
