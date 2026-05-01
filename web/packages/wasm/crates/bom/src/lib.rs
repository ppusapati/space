//! Bill of Materials (BOM) calculations for manufacturing
//!
//! This crate provides:
//! - BOM explosion (multi-level breakdown)
//! - Material requirement planning
//! - Cost rollup calculations
//! - Where-used analysis
//! - Phantom assembly handling

use wasm_bindgen::prelude::*;

mod explosion;
mod costing;
mod where_used;
mod planning;

pub use explosion::*;
pub use costing::*;
pub use where_used::*;
pub use planning::*;

/// Initialize the BOM module (called from core init)
fn bom_init() {
    console_error_panic_hook::set_once();
}
