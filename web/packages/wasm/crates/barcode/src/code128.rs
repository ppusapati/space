//! Code 128 barcode generation

use wasm_bindgen::prelude::*;

use crate::{BarcodeOptions, BarcodeResult, ValidationInfo};

/// Code 128 character sets
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Code128Set {
    A, // Control + uppercase + digits + special
    B, // All printable ASCII
    C, // Numeric pairs (00-99)
}

/// Code 128 encoding table
const CODE128_PATTERNS: &[&str] = &[
    "11011001100", "11001101100", "11001100110", "10010011000", "10010001100", // 0-4
    "10001001100", "10011001000", "10011000100", "10001100100", "11001001000", // 5-9
    "11001000100", "11000100100", "10110011100", "10011011100", "10011001110", // 10-14
    "10111001100", "10011101100", "10011100110", "11001110010", "11001011100", // 15-19
    "11001001110", "11011100100", "11001110100", "11101101110", "11101001100", // 20-24
    "11100101100", "11100100110", "11101100100", "11100110100", "11100110010", // 25-29
    "11011011000", "11011000110", "11000110110", "10100011000", "10001011000", // 30-34
    "10001000110", "10110001000", "10001101000", "10001100010", "11010001000", // 35-39
    "11000101000", "11000100010", "10110111000", "10110001110", "10001101110", // 40-44
    "10111011000", "10111000110", "10001110110", "11101110110", "11010001110", // 45-49
    "11000101110", "11011101000", "11011100010", "11011101110", "11101011000", // 50-54
    "11101000110", "11100010110", "11101101000", "11101100010", "11100011010", // 55-59
    "11101111010", "11001000010", "11110001010", "10100110000", "10100001100", // 60-64
    "10010110000", "10010000110", "10000101100", "10000100110", "10110010000", // 65-69
    "10110000100", "10011010000", "10011000010", "10000110100", "10000110010", // 70-74
    "11000010010", "11001010000", "11110111010", "11000010100", "10001111010", // 75-79
    "10100111100", "10010111100", "10010011110", "10111100100", "10011110100", // 80-84
    "10011110010", "11110100100", "11110010100", "11110010010", "11011011110", // 85-89
    "11011110110", "11110110110", "10101111000", "10100011110", "10001011110", // 90-94
    "10111101000", "10111100010", "11110101000", "11110100010", "10111011110", // 95-99
    "10111101110", "11101011110", "11110101110", "11010000100", "11010010000", // 100-104 (Start A, B, C)
    "11010011100", "1100011101011", // 105-106 (Stop)
];

/// Generate Code 128 barcode
pub fn generate_code128_internal(options: &BarcodeOptions) -> BarcodeResult {
    let data = &options.data;

    if data.is_empty() {
        return BarcodeResult::err("CODE128", data, "Data cannot be empty");
    }

    // Encode data
    let encoding = encode_code128(data);
    if encoding.is_empty() {
        return BarcodeResult::err("CODE128", data, "Failed to encode data");
    }

    // Generate SVG
    let svg = generate_barcode_svg(&encoding, options);

    BarcodeResult::ok("CODE128", data, svg)
}

/// Encode data to Code 128
fn encode_code128(data: &str) -> String {
    let mut result = String::new();
    let mut checksum: u32;

    // Determine best starting code set
    let all_numeric = data.chars().all(|c| c.is_ascii_digit()) && data.len() % 2 == 0;

    if all_numeric && data.len() >= 4 {
        // Use Code C for numeric data
        result.push_str(CODE128_PATTERNS[105]); // Start C
        checksum = 105;

        let mut weight = 1;
        let chars: Vec<char> = data.chars().collect();
        for chunk in chars.chunks(2) {
            if chunk.len() == 2 {
                let pair: String = chunk.iter().collect();
                let value: u32 = pair.parse().unwrap_or(0);
                result.push_str(CODE128_PATTERNS[value as usize]);
                checksum += value * weight;
                weight += 1;
            }
        }
    } else {
        // Use Code B for general data
        result.push_str(CODE128_PATTERNS[104]); // Start B
        checksum = 104;

        let mut weight = 1;
        for c in data.chars() {
            let value = get_code128b_value(c);
            if value < 0 {
                continue; // Skip unsupported characters
            }
            result.push_str(CODE128_PATTERNS[value as usize]);
            checksum += (value as u32) * weight;
            weight += 1;
        }
    }

    // Add checksum
    let check_value = (checksum % 103) as usize;
    result.push_str(CODE128_PATTERNS[check_value]);

    // Add stop code
    result.push_str(CODE128_PATTERNS[106]);

    result
}

/// Get Code 128B value for a character
fn get_code128b_value(c: char) -> i32 {
    let code = c as u32;
    if code >= 32 && code <= 127 {
        (code - 32) as i32
    } else {
        -1
    }
}

/// Generate SVG from barcode encoding
fn generate_barcode_svg(encoding: &str, options: &BarcodeOptions) -> String {
    let bar_width = 2;
    let total_width = encoding.len() * bar_width;
    let height = options.height as usize;
    let margin = options.margin as usize;

    let view_width = total_width + margin * 2;
    let view_height = height + margin * 2;

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

    for c in encoding.chars() {
        if c == '1' {
            svg.push_str(&format!(
                r#"<rect x="{}" y="{}" width="{}" height="{}" fill="{}"/>"#,
                x, margin, bar_width, height - margin, fg_color
            ));
        }
        x += bar_width;
    }

    // Optional text
    if options.include_text {
        let text_y = view_height - 5;
        let text_x = view_width / 2;
        svg.push_str(&format!(
            r#"<text x="{}" y="{}" text-anchor="middle" font-family="monospace" font-size="12">{}</text>"#,
            text_x, text_y, options.data
        ));
    }

    svg.push_str("</svg>");
    svg
}

/// Validate Code 128 data
pub fn validate_code128_data(data: &str) -> ValidationInfo {
    if data.is_empty() {
        return ValidationInfo {
            valid: false,
            error: Some("Data cannot be empty".to_string()),
            corrected_data: None,
        };
    }

    // Check for unsupported characters
    for c in data.chars() {
        let code = c as u32;
        if code < 32 || code > 127 {
            return ValidationInfo {
                valid: false,
                error: Some(format!("Unsupported character: '{}' (code {})", c, code)),
                corrected_data: None,
            };
        }
    }

    ValidationInfo {
        valid: true,
        error: None,
        corrected_data: None,
    }
}

/// Generate Code 128 barcode (simple interface)
#[wasm_bindgen]
pub fn generate_code128(data: &str, width: Option<u32>, height: Option<u32>) -> String {
    let options = BarcodeOptions {
        format: "CODE128".to_string(),
        data: data.to_string(),
        width: width.unwrap_or(300),
        height: height.unwrap_or(100),
        margin: 10,
        error_correction: None,
        include_text: true,
        background_color: Some("#ffffff".to_string()),
        foreground_color: Some("#000000".to_string()),
    };

    let result = generate_code128_internal(&options);
    result.svg.unwrap_or_default()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_code128_generation() {
        let svg = generate_code128("ABC123", None, None);
        assert!(!svg.is_empty());
        assert!(svg.contains("<svg"));
    }

    #[test]
    fn test_code128_numeric() {
        let svg = generate_code128("12345678", None, None);
        assert!(!svg.is_empty());
    }

    #[test]
    fn test_code128_validation() {
        let result = validate_code128_data("ABC123");
        assert!(result.valid);

        let result = validate_code128_data("");
        assert!(!result.valid);
    }
}
