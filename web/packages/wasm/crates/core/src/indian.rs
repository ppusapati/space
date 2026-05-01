//! Indian locale utilities - States, GST codes, etc.

use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Indian state with GST code
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IndianState {
    pub code: String,
    pub name: String,
    pub gst_code: String,
    pub tin_code: String,
    pub is_ut: bool, // Union Territory
}

/// List of all Indian states and UTs with their GST codes
pub const INDIAN_STATES: &[(&str, &str, &str, &str, bool)] = &[
    ("AN", "Andaman and Nicobar Islands", "35", "35", true),
    ("AP", "Andhra Pradesh", "37", "37", false),
    ("AR", "Arunachal Pradesh", "12", "12", false),
    ("AS", "Assam", "18", "18", false),
    ("BR", "Bihar", "10", "10", false),
    ("CH", "Chandigarh", "04", "04", true),
    ("CT", "Chhattisgarh", "22", "22", false),
    ("DD", "Dadra Nagar Haveli and Daman Diu", "26", "26", true),
    ("DL", "Delhi", "07", "07", true),
    ("GA", "Goa", "30", "30", false),
    ("GJ", "Gujarat", "24", "24", false),
    ("HP", "Himachal Pradesh", "02", "02", false),
    ("HR", "Haryana", "06", "06", false),
    ("JH", "Jharkhand", "20", "20", false),
    ("JK", "Jammu and Kashmir", "01", "01", true),
    ("KA", "Karnataka", "29", "29", false),
    ("KL", "Kerala", "32", "32", false),
    ("LA", "Ladakh", "38", "38", true),
    ("LD", "Lakshadweep", "31", "31", true),
    ("MH", "Maharashtra", "27", "27", false),
    ("ML", "Meghalaya", "17", "17", false),
    ("MN", "Manipur", "14", "14", false),
    ("MP", "Madhya Pradesh", "23", "23", false),
    ("MZ", "Mizoram", "15", "15", false),
    ("NL", "Nagaland", "13", "13", false),
    ("OD", "Odisha", "21", "21", false),
    ("PB", "Punjab", "03", "03", false),
    ("PY", "Puducherry", "34", "34", true),
    ("RJ", "Rajasthan", "08", "08", false),
    ("SK", "Sikkim", "11", "11", false),
    ("TG", "Telangana", "36", "36", false),
    ("TN", "Tamil Nadu", "33", "33", false),
    ("TR", "Tripura", "16", "16", false),
    ("UK", "Uttarakhand", "05", "05", false),
    ("UP", "Uttar Pradesh", "09", "09", false),
    ("WB", "West Bengal", "19", "19", false),
    ("OT", "Other Territory", "97", "97", false),
];

/// Get state by code
#[wasm_bindgen]
pub fn get_state_by_code(code: &str) -> JsValue {
    let code_upper = code.to_uppercase();
    for (c, name, gst, tin, is_ut) in INDIAN_STATES {
        if *c == code_upper {
            let state = IndianState {
                code: c.to_string(),
                name: name.to_string(),
                gst_code: gst.to_string(),
                tin_code: tin.to_string(),
                is_ut: *is_ut,
            };
            return serde_wasm_bindgen::to_value(&state).unwrap_or(JsValue::NULL);
        }
    }
    JsValue::NULL
}

/// Get state by GST code (first 2 digits of GSTIN)
#[wasm_bindgen]
pub fn get_state_by_gst_code(gst_code: &str) -> JsValue {
    for (code, name, gst, tin, is_ut) in INDIAN_STATES {
        if *gst == gst_code {
            let state = IndianState {
                code: code.to_string(),
                name: name.to_string(),
                gst_code: gst.to_string(),
                tin_code: tin.to_string(),
                is_ut: *is_ut,
            };
            return serde_wasm_bindgen::to_value(&state).unwrap_or(JsValue::NULL);
        }
    }
    JsValue::NULL
}

/// Get all states
#[wasm_bindgen]
pub fn get_all_states() -> JsValue {
    let states: Vec<IndianState> = INDIAN_STATES
        .iter()
        .map(|(code, name, gst, tin, is_ut)| IndianState {
            code: code.to_string(),
            name: name.to_string(),
            gst_code: gst.to_string(),
            tin_code: tin.to_string(),
            is_ut: *is_ut,
        })
        .collect();
    serde_wasm_bindgen::to_value(&states).unwrap_or(JsValue::NULL)
}

/// Check if source and destination are same state (for CGST/SGST vs IGST)
#[wasm_bindgen]
pub fn is_intra_state(source_state: &str, dest_state: &str) -> bool {
    source_state.to_uppercase() == dest_state.to_uppercase()
}

/// Check if state is a Union Territory
#[wasm_bindgen]
pub fn is_union_territory(state_code: &str) -> bool {
    let code_upper = state_code.to_uppercase();
    INDIAN_STATES
        .iter()
        .find(|(c, _, _, _, _)| *c == code_upper)
        .map(|(_, _, _, _, is_ut)| *is_ut)
        .unwrap_or(false)
}

/// Format amount in Indian numbering system (lakhs, crores)
#[wasm_bindgen]
pub fn format_indian_number(amount: &str) -> String {
    let amount = amount.parse::<f64>().unwrap_or(0.0);
    let is_negative = amount < 0.0;
    let amount = amount.abs();

    let integer_part = amount.trunc() as i64;
    let decimal_part = ((amount.fract() * 100.0).round() as i64).abs();

    let formatted = format_indian_integer(integer_part);

    let result = if decimal_part > 0 {
        format!("{}.{:02}", formatted, decimal_part)
    } else {
        formatted
    };

    if is_negative {
        format!("-{}", result)
    } else {
        result
    }
}

fn format_indian_integer(n: i64) -> String {
    if n < 1000 {
        return n.to_string();
    }

    let s = n.to_string();
    let len = s.len();
    let mut result = String::new();

    // Last 3 digits
    let last_three = &s[len.saturating_sub(3)..];
    let remaining = &s[..len.saturating_sub(3)];

    // Format remaining in pairs from right
    let mut chars: Vec<char> = remaining.chars().collect();
    chars.reverse();

    for (i, c) in chars.iter().enumerate() {
        if i > 0 && i % 2 == 0 {
            result.push(',');
        }
        result.push(*c);
    }

    let mut remaining_formatted: String = result.chars().rev().collect();
    if !remaining_formatted.is_empty() {
        remaining_formatted.push(',');
    }
    remaining_formatted.push_str(last_three);

    remaining_formatted
}

/// Convert amount to words (Indian format)
#[wasm_bindgen]
pub fn amount_to_words(amount: &str) -> String {
    let amount = amount.parse::<f64>().unwrap_or(0.0);
    let is_negative = amount < 0.0;
    let amount = amount.abs();

    let integer_part = amount.trunc() as i64;
    let decimal_part = ((amount.fract() * 100.0).round() as i64).abs();

    let mut result = number_to_words_indian(integer_part);

    if decimal_part > 0 {
        result.push_str(" and ");
        result.push_str(&number_to_words_indian(decimal_part));
        result.push_str(" paise");
    }

    result.push_str(" only");

    if is_negative {
        format!("Minus {}", result)
    } else {
        capitalize_first(&result)
    }
}

fn number_to_words_indian(n: i64) -> String {
    if n == 0 {
        return "zero".to_string();
    }

    let ones = ["", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
                "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen",
                "seventeen", "eighteen", "nineteen"];
    let tens = ["", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"];

    let mut result = String::new();
    let mut n = n;

    // Crores
    if n >= 10_000_000 {
        result.push_str(&number_to_words_indian(n / 10_000_000));
        result.push_str(" crore ");
        n %= 10_000_000;
    }

    // Lakhs
    if n >= 100_000 {
        result.push_str(&number_to_words_indian(n / 100_000));
        result.push_str(" lakh ");
        n %= 100_000;
    }

    // Thousands
    if n >= 1000 {
        result.push_str(&number_to_words_indian(n / 1000));
        result.push_str(" thousand ");
        n %= 1000;
    }

    // Hundreds
    if n >= 100 {
        result.push_str(ones[(n / 100) as usize]);
        result.push_str(" hundred ");
        n %= 100;
    }

    // Tens and ones
    if n >= 20 {
        result.push_str(tens[(n / 10) as usize]);
        if n % 10 > 0 {
            result.push(' ');
            result.push_str(ones[(n % 10) as usize]);
        }
    } else if n > 0 {
        result.push_str(ones[n as usize]);
    }

    result.trim().to_string()
}

fn capitalize_first(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        None => String::new(),
        Some(c) => c.to_uppercase().chain(chars).collect(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_indian_number_format() {
        assert_eq!(format_indian_number("1234567890"), "1,23,45,67,890");
        assert_eq!(format_indian_number("100000"), "1,00,000");
        assert_eq!(format_indian_number("1000"), "1,000");
        assert_eq!(format_indian_number("100"), "100");
    }

    #[test]
    fn test_amount_to_words() {
        assert_eq!(amount_to_words("100"), "One hundred only");
        assert_eq!(amount_to_words("1000"), "One thousand only");
        assert_eq!(amount_to_words("100000"), "One lakh only");
        assert_eq!(amount_to_words("10000000"), "One crore only");
    }
}
