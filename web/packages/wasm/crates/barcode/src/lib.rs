//! Samavaya Barcode - Barcode and QR code generation
//!
//! This crate provides:
//! - QR code generation (with logo support)
//! - Code128 barcode generation
//! - Code39 barcode generation
//! - EAN-13/EAN-8 barcode generation
//! - UPC-A barcode generation
//! - ITF (Interleaved 2 of 5) barcode generation

pub mod qr;
pub mod code128;
pub mod ean;

pub use qr::*;
pub use code128::*;
pub use ean::*;

use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Initialize barcode module (called from core init)
fn barcode_init() {
    log::info!("Samavaya Barcode module initialized");
}

/// Barcode format types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BarcodeFormat {
    /// QR Code
    QRCode,
    /// Code 128
    Code128,
    /// Code 39
    Code39,
    /// EAN-13
    Ean13,
    /// EAN-8
    Ean8,
    /// UPC-A
    UpcA,
    /// ITF (Interleaved 2 of 5)
    Itf,
}

impl BarcodeFormat {
    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_uppercase().as_str() {
            "QR" | "QRCODE" | "QR_CODE" => Some(BarcodeFormat::QRCode),
            "CODE128" | "CODE_128" => Some(BarcodeFormat::Code128),
            "CODE39" | "CODE_39" => Some(BarcodeFormat::Code39),
            "EAN13" | "EAN_13" | "EAN-13" => Some(BarcodeFormat::Ean13),
            "EAN8" | "EAN_8" | "EAN-8" => Some(BarcodeFormat::Ean8),
            "UPCA" | "UPC_A" | "UPC-A" => Some(BarcodeFormat::UpcA),
            "ITF" | "INTERLEAVED_2_OF_5" => Some(BarcodeFormat::Itf),
            _ => None,
        }
    }
}

/// Barcode generation options
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BarcodeOptions {
    /// Barcode format
    pub format: String,
    /// Data to encode
    pub data: String,
    /// Width in pixels
    #[serde(default = "default_width")]
    pub width: u32,
    /// Height in pixels
    #[serde(default = "default_height")]
    pub height: u32,
    /// Margin in modules/pixels
    #[serde(default = "default_margin")]
    pub margin: u32,
    /// Error correction level for QR (L, M, Q, H)
    pub error_correction: Option<String>,
    /// Include human-readable text
    #[serde(default = "default_true")]
    pub include_text: bool,
    /// Background color (hex)
    pub background_color: Option<String>,
    /// Foreground color (hex)
    pub foreground_color: Option<String>,
}

fn default_width() -> u32 { 200 }
fn default_height() -> u32 { 100 }
fn default_margin() -> u32 { 4 }
fn default_true() -> bool { true }

/// Barcode generation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BarcodeResult {
    pub success: bool,
    pub format: String,
    pub data: String,
    pub svg: Option<String>,
    pub data_url: Option<String>,
    pub error: Option<String>,
}

impl BarcodeResult {
    pub fn ok(format: &str, data: &str, svg: String) -> Self {
        let data_url = format!(
            "data:image/svg+xml;base64,{}",
            base64::Engine::encode(&base64::engine::general_purpose::STANDARD, &svg)
        );

        Self {
            success: true,
            format: format.to_string(),
            data: data.to_string(),
            svg: Some(svg),
            data_url: Some(data_url),
            error: None,
        }
    }

    pub fn err(format: &str, data: &str, error: &str) -> Self {
        Self {
            success: false,
            format: format.to_string(),
            data: data.to_string(),
            svg: None,
            data_url: None,
            error: Some(error.to_string()),
        }
    }
}

/// Generate barcode
#[wasm_bindgen]
pub fn generate_barcode(options: JsValue) -> JsValue {
    let options: BarcodeOptions = match serde_wasm_bindgen::from_value(options) {
        Ok(o) => o,
        Err(e) => {
            let result = BarcodeResult::err("UNKNOWN", "", &format!("Invalid options: {}", e));
            return serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL);
        }
    };

    let format = BarcodeFormat::from_str(&options.format);

    let result = match format {
        Some(BarcodeFormat::QRCode) => generate_qr_internal(&options),
        Some(BarcodeFormat::Code128) => generate_code128_internal(&options),
        Some(BarcodeFormat::Ean13) => generate_ean13_internal(&options),
        Some(BarcodeFormat::Ean8) => generate_ean8_internal(&options),
        Some(f) => BarcodeResult::err(
            &options.format,
            &options.data,
            &format!("{:?} format not yet implemented", f),
        ),
        None => BarcodeResult::err(
            &options.format,
            &options.data,
            &format!("Unknown barcode format: {}", options.format),
        ),
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Validate barcode data for a specific format
#[wasm_bindgen]
pub fn validate_barcode_data(format: &str, data: &str) -> JsValue {
    let result = match BarcodeFormat::from_str(format) {
        Some(BarcodeFormat::QRCode) => validate_qr_data(data),
        Some(BarcodeFormat::Code128) => validate_code128_data(data),
        Some(BarcodeFormat::Ean13) => validate_ean13_data(data),
        Some(BarcodeFormat::Ean8) => validate_ean8_data(data),
        Some(BarcodeFormat::UpcA) => validate_upca_data(data),
        _ => ValidationInfo {
            valid: false,
            error: Some(format!("Unknown format: {}", format)),
            corrected_data: None,
        },
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Validation result for barcode data
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationInfo {
    pub valid: bool,
    pub error: Option<String>,
    pub corrected_data: Option<String>,
}
