//! EAN (European Article Number) barcode generation
//!
//! Supports EAN-13 and EAN-8

use wasm_bindgen::prelude::*;

use crate::{BarcodeOptions, BarcodeResult, ValidationInfo};

/// EAN encoding patterns
/// L-codes (odd parity), G-codes (even parity), R-codes (right side)
const EAN_L_PATTERNS: &[&str] = &[
    "0001101", "0011001", "0010011", "0111101", "0100011",
    "0110001", "0101111", "0111011", "0110111", "0001011",
];

const EAN_G_PATTERNS: &[&str] = &[
    "0100111", "0110011", "0011011", "0100001", "0011101",
    "0111001", "0000101", "0010001", "0001001", "0010111",
];

const EAN_R_PATTERNS: &[&str] = &[
    "1110010", "1100110", "1101100", "1000010", "1011100",
    "1001110", "1010000", "1000100", "1001000", "1110100",
];

/// First digit encoding for EAN-13 (which L/G patterns to use)
const EAN13_FIRST_DIGIT: &[&str] = &[
    "LLLLLL", "LLGLGG", "LLGGLG", "LLGGGL", "LGLLGG",
    "LGGLLG", "LGGGLL", "LGLGLG", "LGLGGL", "LGGLGL",
];

/// Generate EAN-13 barcode
pub fn generate_ean13_internal(options: &BarcodeOptions) -> BarcodeResult {
    let data = options.data.trim();

    // Validate length
    if data.len() != 12 && data.len() != 13 {
        return BarcodeResult::err("EAN13", data, "EAN-13 must be 12 or 13 digits");
    }

    // Ensure all digits
    if !data.chars().all(|c| c.is_ascii_digit()) {
        return BarcodeResult::err("EAN13", data, "EAN-13 must contain only digits");
    }

    // Calculate or validate check digit
    let mut digits: Vec<u8> = data.chars()
        .filter_map(|c| c.to_digit(10).map(|d| d as u8))
        .collect();

    if digits.len() == 12 {
        let check = calculate_ean13_check(&digits);
        digits.push(check);
    } else if digits.len() == 13 {
        let check = calculate_ean13_check(&digits[..12]);
        if check != digits[12] {
            return BarcodeResult::err(
                "EAN13",
                data,
                &format!("Invalid check digit. Expected {}, got {}", check, digits[12]),
            );
        }
    }

    // Encode
    let encoding = encode_ean13(&digits);
    let svg = generate_ean_svg(&encoding, options, &digits);

    let data_str: String = digits.iter().map(|d| char::from_digit(*d as u32, 10).unwrap()).collect();
    BarcodeResult::ok("EAN13", &data_str, svg)
}

/// Generate EAN-8 barcode
pub fn generate_ean8_internal(options: &BarcodeOptions) -> BarcodeResult {
    let data = options.data.trim();

    // Validate length
    if data.len() != 7 && data.len() != 8 {
        return BarcodeResult::err("EAN8", data, "EAN-8 must be 7 or 8 digits");
    }

    // Ensure all digits
    if !data.chars().all(|c| c.is_ascii_digit()) {
        return BarcodeResult::err("EAN8", data, "EAN-8 must contain only digits");
    }

    // Calculate or validate check digit
    let mut digits: Vec<u8> = data.chars()
        .filter_map(|c| c.to_digit(10).map(|d| d as u8))
        .collect();

    if digits.len() == 7 {
        let check = calculate_ean8_check(&digits);
        digits.push(check);
    } else if digits.len() == 8 {
        let check = calculate_ean8_check(&digits[..7]);
        if check != digits[7] {
            return BarcodeResult::err(
                "EAN8",
                data,
                &format!("Invalid check digit. Expected {}, got {}", check, digits[7]),
            );
        }
    }

    // Encode
    let encoding = encode_ean8(&digits);
    let svg = generate_ean_svg(&encoding, options, &digits);

    let data_str: String = digits.iter().map(|d| char::from_digit(*d as u32, 10).unwrap()).collect();
    BarcodeResult::ok("EAN8", &data_str, svg)
}

/// Calculate EAN-13 check digit
fn calculate_ean13_check(digits: &[u8]) -> u8 {
    let sum: u32 = digits.iter().enumerate().map(|(i, &d)| {
        let weight = if i % 2 == 0 { 1 } else { 3 };
        (d as u32) * weight
    }).sum();

    let remainder = sum % 10;
    if remainder == 0 { 0 } else { (10 - remainder) as u8 }
}

/// Calculate EAN-8 check digit
fn calculate_ean8_check(digits: &[u8]) -> u8 {
    let sum: u32 = digits.iter().enumerate().map(|(i, &d)| {
        let weight = if i % 2 == 0 { 3 } else { 1 };
        (d as u32) * weight
    }).sum();

    let remainder = sum % 10;
    if remainder == 0 { 0 } else { (10 - remainder) as u8 }
}

/// Encode EAN-13
fn encode_ean13(digits: &[u8]) -> String {
    let mut encoding = String::new();

    // Start guard
    encoding.push_str("101");

    // First digit determines L/G pattern for left side
    let first_digit = digits[0] as usize;
    let pattern = EAN13_FIRST_DIGIT[first_digit];

    // Left side (digits 1-6)
    for (i, &digit) in digits[1..7].iter().enumerate() {
        let d = digit as usize;
        if pattern.chars().nth(i) == Some('L') {
            encoding.push_str(EAN_L_PATTERNS[d]);
        } else {
            encoding.push_str(EAN_G_PATTERNS[d]);
        }
    }

    // Center guard
    encoding.push_str("01010");

    // Right side (digits 7-12)
    for &digit in &digits[7..13] {
        encoding.push_str(EAN_R_PATTERNS[digit as usize]);
    }

    // End guard
    encoding.push_str("101");

    encoding
}

/// Encode EAN-8
fn encode_ean8(digits: &[u8]) -> String {
    let mut encoding = String::new();

    // Start guard
    encoding.push_str("101");

    // Left side (digits 0-3) - all L-codes
    for &digit in &digits[0..4] {
        encoding.push_str(EAN_L_PATTERNS[digit as usize]);
    }

    // Center guard
    encoding.push_str("01010");

    // Right side (digits 4-7) - all R-codes
    for &digit in &digits[4..8] {
        encoding.push_str(EAN_R_PATTERNS[digit as usize]);
    }

    // End guard
    encoding.push_str("101");

    encoding
}

/// Generate SVG for EAN barcode
fn generate_ean_svg(encoding: &str, options: &BarcodeOptions, digits: &[u8]) -> String {
    let bar_width = 2;
    let total_width = encoding.len() * bar_width;
    let height = options.height as usize;
    let margin = options.margin as usize;
    let text_height = if options.include_text { 16 } else { 0 };

    let view_width = total_width + margin * 2;
    let view_height = height + margin * 2 + text_height;

    let mut svg = format!(
        r#"<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {} {}" width="{}" height="{}">"#,
        view_width, view_height, options.width, options.height
    );

    // Background
    let bg_color = options.background_color.as_deref().unwrap_or("#ffffff");
    svg.push_str(&format!(
        r#"<rect x="0" y="0" width="{}" height="{}" fill="{}"/>"#,
        view_width, view_height, bg_color
    ));

    // Bars
    let fg_color = options.foreground_color.as_deref().unwrap_or("#000000");
    let mut x = margin;

    for (i, c) in encoding.chars().enumerate() {
        if c == '1' {
            // Guard bars are taller
            let is_guard = i < 3 || i >= encoding.len() - 3 ||
                          (i >= 45 && i < 50); // Center guard for EAN-13

            let bar_height = if is_guard {
                height - margin + 5
            } else {
                height - margin - text_height
            };

            svg.push_str(&format!(
                r#"<rect x="{}" y="{}" width="{}" height="{}" fill="{}"/>"#,
                x, margin, bar_width, bar_height, fg_color
            ));
        }
        x += bar_width;
    }

    // Text
    if options.include_text {
        let text_y = view_height - 4;
        let data_str: String = digits.iter().map(|d| char::from_digit(*d as u32, 10).unwrap()).collect();

        if digits.len() == 13 {
            // EAN-13: First digit separate, then groups
            svg.push_str(&format!(
                r#"<text x="{}" y="{}" font-family="monospace" font-size="10">{}</text>"#,
                margin - 8, text_y, data_str.chars().next().unwrap()
            ));
            svg.push_str(&format!(
                r#"<text x="{}" y="{}" text-anchor="middle" font-family="monospace" font-size="10">{}</text>"#,
                margin + 24, text_y, &data_str[1..7]
            ));
            svg.push_str(&format!(
                r#"<text x="{}" y="{}" text-anchor="middle" font-family="monospace" font-size="10">{}</text>"#,
                margin + 73, text_y, &data_str[7..]
            ));
        } else {
            // EAN-8
            svg.push_str(&format!(
                r#"<text x="{}" y="{}" text-anchor="middle" font-family="monospace" font-size="10">{}</text>"#,
                view_width / 2, text_y, data_str
            ));
        }
    }

    svg.push_str("</svg>");
    svg
}

/// Validate EAN-13 data
pub fn validate_ean13_data(data: &str) -> ValidationInfo {
    let data = data.trim();

    if data.len() != 12 && data.len() != 13 {
        return ValidationInfo {
            valid: false,
            error: Some("EAN-13 must be 12 or 13 digits".to_string()),
            corrected_data: None,
        };
    }

    if !data.chars().all(|c| c.is_ascii_digit()) {
        return ValidationInfo {
            valid: false,
            error: Some("EAN-13 must contain only digits".to_string()),
            corrected_data: None,
        };
    }

    let digits: Vec<u8> = data.chars()
        .filter_map(|c| c.to_digit(10).map(|d| d as u8))
        .collect();

    if digits.len() == 13 {
        let check = calculate_ean13_check(&digits[..12]);
        if check != digits[12] {
            return ValidationInfo {
                valid: false,
                error: Some(format!("Invalid check digit. Expected {}", check)),
                corrected_data: Some(format!(
                    "{}{}",
                    &data[..12],
                    check
                )),
            };
        }
    }

    ValidationInfo {
        valid: true,
        error: None,
        corrected_data: if digits.len() == 12 {
            let check = calculate_ean13_check(&digits);
            Some(format!("{}{}", data, check))
        } else {
            None
        },
    }
}

/// Validate EAN-8 data
pub fn validate_ean8_data(data: &str) -> ValidationInfo {
    let data = data.trim();

    if data.len() != 7 && data.len() != 8 {
        return ValidationInfo {
            valid: false,
            error: Some("EAN-8 must be 7 or 8 digits".to_string()),
            corrected_data: None,
        };
    }

    if !data.chars().all(|c| c.is_ascii_digit()) {
        return ValidationInfo {
            valid: false,
            error: Some("EAN-8 must contain only digits".to_string()),
            corrected_data: None,
        };
    }

    ValidationInfo {
        valid: true,
        error: None,
        corrected_data: None,
    }
}

/// Validate UPC-A data
pub fn validate_upca_data(data: &str) -> ValidationInfo {
    let data = data.trim();

    if data.len() != 11 && data.len() != 12 {
        return ValidationInfo {
            valid: false,
            error: Some("UPC-A must be 11 or 12 digits".to_string()),
            corrected_data: None,
        };
    }

    if !data.chars().all(|c| c.is_ascii_digit()) {
        return ValidationInfo {
            valid: false,
            error: Some("UPC-A must contain only digits".to_string()),
            corrected_data: None,
        };
    }

    ValidationInfo {
        valid: true,
        error: None,
        corrected_data: None,
    }
}

/// Generate EAN-13 barcode (simple interface)
#[wasm_bindgen]
pub fn generate_ean13(data: &str, width: Option<u32>, height: Option<u32>) -> String {
    let options = BarcodeOptions {
        format: "EAN13".to_string(),
        data: data.to_string(),
        width: width.unwrap_or(200),
        height: height.unwrap_or(100),
        margin: 10,
        error_correction: None,
        include_text: true,
        background_color: Some("#ffffff".to_string()),
        foreground_color: Some("#000000".to_string()),
    };

    let result = generate_ean13_internal(&options);
    result.svg.unwrap_or_default()
}

/// Generate EAN-8 barcode (simple interface)
#[wasm_bindgen]
pub fn generate_ean8(data: &str, width: Option<u32>, height: Option<u32>) -> String {
    let options = BarcodeOptions {
        format: "EAN8".to_string(),
        data: data.to_string(),
        width: width.unwrap_or(150),
        height: height.unwrap_or(80),
        margin: 10,
        error_correction: None,
        include_text: true,
        background_color: Some("#ffffff".to_string()),
        foreground_color: Some("#000000".to_string()),
    };

    let result = generate_ean8_internal(&options);
    result.svg.unwrap_or_default()
}

/// Calculate and append EAN-13 check digit
#[wasm_bindgen]
pub fn ean13_with_check_digit(data: &str) -> String {
    let data = data.trim();
    if data.len() != 12 || !data.chars().all(|c| c.is_ascii_digit()) {
        return data.to_string();
    }

    let digits: Vec<u8> = data.chars()
        .filter_map(|c| c.to_digit(10).map(|d| d as u8))
        .collect();

    let check = calculate_ean13_check(&digits);
    format!("{}{}", data, check)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ean13_check_digit() {
        // Example: 590123412345 -> check digit 7
        let digits: Vec<u8> = vec![5, 9, 0, 1, 2, 3, 4, 1, 2, 3, 4, 5];
        let check = calculate_ean13_check(&digits);
        assert_eq!(check, 7);
    }

    #[test]
    fn test_ean13_generation() {
        let svg = generate_ean13("5901234123457", None, None);
        assert!(!svg.is_empty());
        assert!(svg.contains("<svg"));
    }

    #[test]
    fn test_ean8_generation() {
        let svg = generate_ean8("12345670", None, None);
        assert!(!svg.is_empty());
    }

    #[test]
    fn test_ean13_validation() {
        let result = validate_ean13_data("5901234123457");
        assert!(result.valid);

        let result = validate_ean13_data("590123412345"); // 12 digits
        assert!(result.valid);
        assert!(result.corrected_data.is_some());
    }
}
