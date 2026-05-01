//! Samavaya Tax Engine - GST, TDS, TCS calculations for Indian taxation
//!
//! This crate provides comprehensive tax calculation capabilities:
//! - GST (CGST, SGST, IGST, Cess)
//! - TDS (Tax Deducted at Source)
//! - TCS (Tax Collected at Source)
//! - Reverse Charge Mechanism

pub mod gst;
pub mod tds;
pub mod tcs;
pub mod hsn;
pub mod cess;

pub use gst::*;
pub use tds::*;
pub use tcs::*;
pub use hsn::*;
pub use cess::*;

use wasm_bindgen::prelude::*;

/// Initialize tax engine (called from core init)
fn tax_engine_init() {
    log::info!("Samavaya Tax Engine initialized");
}
