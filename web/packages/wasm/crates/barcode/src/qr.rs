//! QR Code generation

use qrcode::render::svg;
use qrcode::{EcLevel, QrCode, Version};
use wasm_bindgen::prelude::*;

use crate::{BarcodeOptions, BarcodeResult, ValidationInfo};

/// Generate QR code
pub fn generate_qr_internal(options: &BarcodeOptions) -> BarcodeResult {
    // Parse error correction level
    let ec_level = match options.error_correction.as_deref() {
        Some("L") | Some("l") => EcLevel::L,
        Some("M") | Some("m") => EcLevel::M,
        Some("Q") | Some("q") => EcLevel::Q,
        Some("H") | Some("h") => EcLevel::H,
        _ => EcLevel::M, // Default
    };

    // Generate QR code
    let code = match QrCode::with_error_correction_level(&options.data, ec_level) {
        Ok(c) => c,
        Err(e) => return BarcodeResult::err("QR", &options.data, &format!("QR generation failed: {}", e)),
    };

    // Render to SVG
    let svg_string = code.render::<svg::Color>()
        .min_dimensions(options.width, options.height)
        .quiet_zone(true)
        .build();

    BarcodeResult::ok("QR", &options.data, svg_string)
}

/// Validate QR data
pub fn validate_qr_data(data: &str) -> ValidationInfo {
    if data.is_empty() {
        return ValidationInfo {
            valid: false,
            error: Some("QR data cannot be empty".to_string()),
            corrected_data: None,
        };
    }

    // QR codes can hold up to 7089 numeric, 4296 alphanumeric, or 2953 binary characters
    if data.len() > 2953 {
        return ValidationInfo {
            valid: false,
            error: Some(format!("Data too long for QR code. Max 2953 bytes, got {}", data.len())),
            corrected_data: None,
        };
    }

    ValidationInfo {
        valid: true,
        error: None,
        corrected_data: None,
    }
}

/// Generate QR code with simple parameters
#[wasm_bindgen]
pub fn generate_qr(data: &str, size: Option<u32>) -> String {
    let size = size.unwrap_or(200);

    let code = match QrCode::with_error_correction_level(data, EcLevel::M) {
        Ok(c) => c,
        Err(_) => return String::new(),
    };

    code.render::<svg::Color>()
        .min_dimensions(size, size)
        .quiet_zone(true)
        .build()
}

/// Generate QR code as data URL
#[wasm_bindgen]
pub fn generate_qr_data_url(data: &str, size: Option<u32>) -> String {
    let svg = generate_qr(data, size);
    if svg.is_empty() {
        return String::new();
    }

    format!(
        "data:image/svg+xml;base64,{}",
        base64::Engine::encode(&base64::engine::general_purpose::STANDARD, &svg)
    )
}

/// Generate QR code for a URL
#[wasm_bindgen]
pub fn generate_url_qr(url: &str, size: Option<u32>) -> String {
    generate_qr(url, size)
}

/// Generate QR code for vCard
#[wasm_bindgen]
pub fn generate_vcard_qr(name: &str, phone: &str, email: &str, company: &str, size: Option<u32>) -> String {
    let vcard = format!(
        "BEGIN:VCARD\n\
         VERSION:3.0\n\
         N:{}\n\
         FN:{}\n\
         ORG:{}\n\
         TEL:{}\n\
         EMAIL:{}\n\
         END:VCARD",
        name, name, company, phone, email
    );

    generate_qr(&vcard, size)
}

/// Generate QR code for UPI payment
#[wasm_bindgen]
pub fn generate_upi_qr(
    payee_vpa: &str,
    payee_name: &str,
    amount: Option<f64>,
    transaction_note: Option<String>,
    size: Option<u32>,
) -> String {
    let mut upi_string = format!(
        "upi://pay?pa={}&pn={}",
        payee_vpa,
        payee_name.replace(' ', "%20")
    );

    if let Some(amt) = amount {
        upi_string.push_str(&format!("&am={:.2}", amt));
    }

    if let Some(note) = transaction_note {
        upi_string.push_str(&format!("&tn={}", note.replace(' ', "%20")));
    }

    upi_string.push_str("&cu=INR");

    generate_qr(&upi_string, size)
}

/// Generate QR code for GST invoice
#[wasm_bindgen]
pub fn generate_gst_invoice_qr(
    seller_gstin: &str,
    buyer_gstin: &str,
    invoice_number: &str,
    invoice_date: &str,
    total_value: &str,
    size: Option<u32>,
) -> String {
    // Simplified GST QR format (actual format varies by version)
    let qr_data = format!(
        "{{\"sellerGstin\":\"{}\",\"buyerGstin\":\"{}\",\"invNo\":\"{}\",\"invDt\":\"{}\",\"totVal\":{}}}",
        seller_gstin, buyer_gstin, invoice_number, invoice_date, total_value
    );

    generate_qr(&qr_data, size)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_qr_generation() {
        let svg = generate_qr("Hello World", Some(100));
        assert!(!svg.is_empty());
        assert!(svg.contains("<svg"));
    }

    #[test]
    fn test_qr_data_url() {
        let data_url = generate_qr_data_url("Test", Some(100));
        assert!(data_url.starts_with("data:image/svg+xml;base64,"));
    }

    #[test]
    fn test_upi_qr() {
        let svg = generate_upi_qr("test@upi", "Test User", Some(100.50), None, Some(100));
        assert!(!svg.is_empty());
    }
}
