//! Internationalization (i18n) Utilities
//!
//! This crate provides:
//! - Indian number formatting (lakhs, crores)
//! - Currency formatting (INR, USD, etc.)
//! - Date/time formatting (multiple formats)
//! - Amount to words conversion (Indian English)
//! - Pluralization rules

use chrono::{Datelike, NaiveDate, NaiveDateTime, Timelike};
use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Locale identifier
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Locale {
    EnIn, // English (India)
    EnUs, // English (US)
    HiIn, // Hindi (India)
    TaIn, // Tamil (India)
    TeIn, // Telugu (India)
    KnIn, // Kannada (India)
    MrIn, // Marathi (India)
    GuIn, // Gujarati (India)
    BnIn, // Bengali (India)
    MlIn, // Malayalam (India)
}

/// Currency code
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CurrencyCode {
    INR,
    USD,
    EUR,
    GBP,
    AED,
    SAR,
    SGD,
    JPY,
    CNY,
    AUD,
}

/// Number format style
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum NumberStyle {
    /// Standard format (1,000,000)
    Standard,
    /// Indian format (10,00,000)
    Indian,
    /// Compact format (10L, 1Cr)
    Compact,
    /// Words format
    Words,
}

/// Date format style
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum DateStyle {
    /// Short: 01/01/24
    Short,
    /// Medium: 01-Jan-2024
    Medium,
    /// Long: January 1, 2024
    Long,
    /// Full: Monday, January 1, 2024
    Full,
    /// ISO: 2024-01-01
    Iso,
    /// Indian: 01/01/2024
    Indian,
}

/// Initialize the i18n module (called from core init)
fn i18n_init() {
    console_error_panic_hook::set_once();
}

/// Format number in Indian style (lakhs and crores)
#[wasm_bindgen]
pub fn format_indian_number(value: &str, decimal_places: Option<u32>) -> String {
    let num: Decimal = value.parse().unwrap_or(Decimal::ZERO);
    let dp = decimal_places.unwrap_or(2);
    format_indian_decimal(num, dp)
}

fn format_indian_decimal(num: Decimal, decimal_places: u32) -> String {
    let is_negative = num < Decimal::ZERO;
    let abs_num = num.abs().round_dp(decimal_places);

    let (integer_part, decimal_part) = {
        let s = abs_num.to_string();
        if let Some(pos) = s.find('.') {
            (s[..pos].to_string(), s[pos..].to_string())
        } else {
            (s, String::new())
        }
    };

    let formatted = format_indian_integer(&integer_part);

    let result = if decimal_part.is_empty() && decimal_places == 0 {
        formatted
    } else if decimal_part.is_empty() {
        format!("{}.{}", formatted, "0".repeat(decimal_places as usize))
    } else {
        format!("{}{}", formatted, decimal_part)
    };

    if is_negative {
        format!("-{}", result)
    } else {
        result
    }
}

fn format_indian_integer(num: &str) -> String {
    let chars: Vec<char> = num.chars().collect();
    let len = chars.len();

    if len <= 3 {
        return num.to_string();
    }

    let mut result = String::new();
    let mut count = 0;

    for (i, c) in chars.iter().rev().enumerate() {
        if i == 3 || (i > 3 && count == 2) {
            result.insert(0, ',');
            count = 0;
        }
        result.insert(0, *c);
        count += 1;
    }

    result
}

/// Format number in compact Indian style (L, Cr)
#[wasm_bindgen]
pub fn format_compact_indian(value: &str, decimal_places: Option<u32>) -> String {
    let num: Decimal = value.parse().unwrap_or(Decimal::ZERO);
    let dp = decimal_places.unwrap_or(2);

    let is_negative = num < Decimal::ZERO;
    let abs_num = num.abs();

    let (formatted, suffix) = if abs_num >= dec!(10000000) {
        // Crores
        let cr = abs_num / dec!(10000000);
        (cr.round_dp(dp).to_string(), "Cr")
    } else if abs_num >= dec!(100000) {
        // Lakhs
        let l = abs_num / dec!(100000);
        (l.round_dp(dp).to_string(), "L")
    } else if abs_num >= dec!(1000) {
        // Thousands
        let k = abs_num / dec!(1000);
        (k.round_dp(dp).to_string(), "K")
    } else {
        (abs_num.round_dp(dp).to_string(), "")
    };

    let result = format!("{}{}", formatted, suffix);

    if is_negative {
        format!("-{}", result)
    } else {
        result
    }
}

/// Format currency
#[wasm_bindgen]
pub fn format_currency(value: &str, currency: &str, show_symbol: Option<bool>) -> String {
    let num: Decimal = value.parse().unwrap_or(Decimal::ZERO);
    let show = show_symbol.unwrap_or(true);

    let (symbol, dp, indian_style) = match currency.to_uppercase().as_str() {
        "INR" => ("₹", 2u32, true),
        "USD" => ("$", 2, false),
        "EUR" => ("€", 2, false),
        "GBP" => ("£", 2, false),
        "AED" => ("AED ", 2, false),
        "SAR" => ("SAR ", 2, false),
        "SGD" => ("S$", 2, false),
        "JPY" => ("¥", 0, false),
        "CNY" => ("¥", 2, false),
        "AUD" => ("A$", 2, false),
        _ => ("", 2, false),
    };

    let formatted = if indian_style {
        format_indian_decimal(num, dp)
    } else {
        format_standard_number(num, dp)
    };

    if show {
        format!("{}{}", symbol, formatted)
    } else {
        formatted
    }
}

fn format_standard_number(num: Decimal, decimal_places: u32) -> String {
    let is_negative = num < Decimal::ZERO;
    let abs_num = num.abs().round_dp(decimal_places);

    let (integer_part, decimal_part) = {
        let s = abs_num.to_string();
        if let Some(pos) = s.find('.') {
            (s[..pos].to_string(), s[pos..].to_string())
        } else {
            (s, String::new())
        }
    };

    // Standard grouping (every 3 digits)
    let chars: Vec<char> = integer_part.chars().collect();
    let mut result = String::new();

    for (i, c) in chars.iter().rev().enumerate() {
        if i > 0 && i % 3 == 0 {
            result.insert(0, ',');
        }
        result.insert(0, *c);
    }

    let formatted = if decimal_part.is_empty() && decimal_places == 0 {
        result
    } else if decimal_part.is_empty() {
        format!("{}.{}", result, "0".repeat(decimal_places as usize))
    } else {
        format!("{}{}", result, decimal_part)
    };

    if is_negative {
        format!("-{}", formatted)
    } else {
        formatted
    }
}

/// Convert amount to words (Indian English)
#[wasm_bindgen]
pub fn amount_to_words(value: &str, currency: &str) -> String {
    let num: Decimal = value.parse().unwrap_or(Decimal::ZERO);
    let is_negative = num < Decimal::ZERO;
    let abs_num = num.abs();

    // Split into rupees and paise
    let rupees = abs_num.trunc().to_u64().unwrap_or(0);
    let paise = ((abs_num.fract() * dec!(100)).round_dp(0)).to_u32().unwrap_or(0);

    let rupees_words = number_to_words_indian(rupees);
    let currency_name = match currency.to_uppercase().as_str() {
        "INR" => ("Rupees", "Paise"),
        "USD" => ("Dollars", "Cents"),
        "EUR" => ("Euros", "Cents"),
        "GBP" => ("Pounds", "Pence"),
        _ => ("", ""),
    };

    let mut result = String::new();

    if is_negative {
        result.push_str("Negative ");
    }

    if rupees > 0 {
        result.push_str(&rupees_words);
        result.push(' ');
        result.push_str(currency_name.0);
    }

    if paise > 0 {
        if rupees > 0 {
            result.push_str(" and ");
        }
        result.push_str(&number_to_words_indian(paise as u64));
        result.push(' ');
        result.push_str(currency_name.1);
    }

    if rupees == 0 && paise == 0 {
        result.push_str("Zero ");
        result.push_str(currency_name.0);
    }

    result.push_str(" Only");
    result
}

fn number_to_words_indian(num: u64) -> String {
    if num == 0 {
        return "Zero".to_string();
    }

    let ones = [
        "", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
        "Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen",
        "Seventeen", "Eighteen", "Nineteen",
    ];

    let tens = [
        "", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety",
    ];

    fn convert_chunk(n: u64, ones: &[&str], tens: &[&str]) -> String {
        if n == 0 {
            return String::new();
        }

        let mut result = String::new();

        if n >= 100 {
            result.push_str(ones[(n / 100) as usize]);
            result.push_str(" Hundred ");
        }

        let remainder = n % 100;
        if remainder >= 20 {
            result.push_str(tens[(remainder / 10) as usize]);
            if remainder % 10 > 0 {
                result.push(' ');
                result.push_str(ones[(remainder % 10) as usize]);
            }
        } else if remainder > 0 {
            result.push_str(ones[remainder as usize]);
        }

        result.trim().to_string()
    }

    let mut result = String::new();
    let mut remaining = num;

    // Crores (10,000,000)
    if remaining >= 10000000 {
        let crores = remaining / 10000000;
        result.push_str(&convert_chunk(crores, &ones, &tens));
        result.push_str(" Crore ");
        remaining %= 10000000;
    }

    // Lakhs (100,000)
    if remaining >= 100000 {
        let lakhs = remaining / 100000;
        result.push_str(&convert_chunk(lakhs, &ones, &tens));
        result.push_str(" Lakh ");
        remaining %= 100000;
    }

    // Thousands
    if remaining >= 1000 {
        let thousands = remaining / 1000;
        result.push_str(&convert_chunk(thousands, &ones, &tens));
        result.push_str(" Thousand ");
        remaining %= 1000;
    }

    // Hundreds and below
    if remaining > 0 {
        result.push_str(&convert_chunk(remaining, &ones, &tens));
    }

    result.trim().to_string()
}

/// Format date
#[wasm_bindgen]
pub fn format_date(date: &str, style: &str, locale: Option<String>) -> String {
    let parsed = if date.contains('T') {
        NaiveDateTime::parse_from_str(date, "%Y-%m-%dT%H:%M:%S")
            .or_else(|_| NaiveDateTime::parse_from_str(date, "%Y-%m-%dT%H:%M:%S%.f"))
            .map(|dt| dt.date())
            .ok()
    } else {
        NaiveDate::parse_from_str(date, "%Y-%m-%d")
            .or_else(|_| NaiveDate::parse_from_str(date, "%d/%m/%Y"))
            .or_else(|_| NaiveDate::parse_from_str(date, "%d-%m-%Y"))
            .ok()
    };

    let date = match parsed {
        Some(d) => d,
        None => return date.to_string(),
    };

    let loc = locale.as_deref().unwrap_or("en-IN");

    match style.to_lowercase().as_str() {
        "short" => date.format("%d/%m/%y").to_string(),
        "medium" => date.format("%d-%b-%Y").to_string(),
        "long" => format_date_long(date, loc),
        "full" => format_date_full(date, loc),
        "iso" => date.format("%Y-%m-%d").to_string(),
        "indian" => date.format("%d/%m/%Y").to_string(),
        _ => date.format("%d/%m/%Y").to_string(),
    }
}

fn format_date_long(date: NaiveDate, _locale: &str) -> String {
    let months = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December",
    ];

    format!(
        "{} {}, {}",
        months[date.month0() as usize],
        date.day(),
        date.year()
    )
}

fn format_date_full(date: NaiveDate, _locale: &str) -> String {
    let days = [
        "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
    ];
    let months = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December",
    ];

    format!(
        "{}, {} {}, {}",
        days[date.weekday().num_days_from_sunday() as usize],
        months[date.month0() as usize],
        date.day(),
        date.year()
    )
}

/// Format datetime
#[wasm_bindgen]
pub fn format_datetime(datetime: &str, date_style: &str, time_style: &str) -> String {
    let parsed = NaiveDateTime::parse_from_str(datetime, "%Y-%m-%dT%H:%M:%S")
        .or_else(|_| NaiveDateTime::parse_from_str(datetime, "%Y-%m-%dT%H:%M:%S%.f"))
        .or_else(|_| NaiveDateTime::parse_from_str(datetime, "%Y-%m-%d %H:%M:%S"));

    let dt = match parsed {
        Ok(d) => d,
        Err(_) => return datetime.to_string(),
    };

    let date_part = format_date(&dt.date().format("%Y-%m-%d").to_string(), date_style, None);

    let time_part = match time_style.to_lowercase().as_str() {
        "short" => dt.format("%H:%M").to_string(),
        "medium" => dt.format("%H:%M:%S").to_string(),
        "long" => dt.format("%H:%M:%S").to_string(),
        "12h" => format_time_12h(dt),
        _ => dt.format("%H:%M").to_string(),
    };

    format!("{} {}", date_part, time_part)
}

fn format_time_12h(dt: NaiveDateTime) -> String {
    let hour = dt.hour();
    let (h, period) = if hour == 0 {
        (12, "AM")
    } else if hour < 12 {
        (hour, "AM")
    } else if hour == 12 {
        (12, "PM")
    } else {
        (hour - 12, "PM")
    };

    format!("{}:{:02} {}", h, dt.minute(), period)
}

/// Get relative time (e.g., "2 days ago")
#[wasm_bindgen]
pub fn format_relative_time(datetime: &str, reference: Option<String>) -> String {
    let parsed = NaiveDateTime::parse_from_str(datetime, "%Y-%m-%dT%H:%M:%S")
        .or_else(|_| NaiveDateTime::parse_from_str(datetime, "%Y-%m-%dT%H:%M:%S%.f"));

    let dt = match parsed {
        Ok(d) => d,
        Err(_) => return datetime.to_string(),
    };

    let ref_dt = reference
        .and_then(|r| NaiveDateTime::parse_from_str(&r, "%Y-%m-%dT%H:%M:%S").ok())
        .unwrap_or_else(|| chrono::Utc::now().naive_utc());

    let duration = ref_dt.signed_duration_since(dt);
    let seconds = duration.num_seconds();

    if seconds < 0 {
        // Future
        let abs_seconds = -seconds;
        if abs_seconds < 60 {
            "in a few seconds".to_string()
        } else if abs_seconds < 3600 {
            format!("in {} minutes", abs_seconds / 60)
        } else if abs_seconds < 86400 {
            format!("in {} hours", abs_seconds / 3600)
        } else if abs_seconds < 604800 {
            format!("in {} days", abs_seconds / 86400)
        } else if abs_seconds < 2592000 {
            format!("in {} weeks", abs_seconds / 604800)
        } else if abs_seconds < 31536000 {
            format!("in {} months", abs_seconds / 2592000)
        } else {
            format!("in {} years", abs_seconds / 31536000)
        }
    } else {
        // Past
        if seconds < 60 {
            "just now".to_string()
        } else if seconds < 3600 {
            format!("{} minutes ago", seconds / 60)
        } else if seconds < 86400 {
            format!("{} hours ago", seconds / 3600)
        } else if seconds < 604800 {
            format!("{} days ago", seconds / 86400)
        } else if seconds < 2592000 {
            format!("{} weeks ago", seconds / 604800)
        } else if seconds < 31536000 {
            format!("{} months ago", seconds / 2592000)
        } else {
            format!("{} years ago", seconds / 31536000)
        }
    }
}

/// Get financial year
#[wasm_bindgen]
pub fn get_financial_year(date: &str) -> String {
    let parsed = NaiveDate::parse_from_str(date, "%Y-%m-%d")
        .or_else(|_| NaiveDate::parse_from_str(date, "%d/%m/%Y"));

    let date = match parsed {
        Ok(d) => d,
        Err(_) => return String::new(),
    };

    let year = date.year();
    let month = date.month();

    if month >= 4 {
        format!("{}-{}", year, (year + 1) % 100)
    } else {
        format!("{}-{}", year - 1, year % 100)
    }
}

/// Get quarter
#[wasm_bindgen]
pub fn get_quarter(date: &str, financial_year: Option<bool>) -> String {
    let parsed = NaiveDate::parse_from_str(date, "%Y-%m-%d")
        .or_else(|_| NaiveDate::parse_from_str(date, "%d/%m/%Y"));

    let date = match parsed {
        Ok(d) => d,
        Err(_) => return String::new(),
    };

    let month = date.month();
    let fy = financial_year.unwrap_or(true);

    if fy {
        // Indian Financial Year (Apr-Mar)
        match month {
            4..=6 => "Q1".to_string(),
            7..=9 => "Q2".to_string(),
            10..=12 => "Q3".to_string(),
            1..=3 => "Q4".to_string(),
            _ => String::new(),
        }
    } else {
        // Calendar Year (Jan-Dec)
        match month {
            1..=3 => "Q1".to_string(),
            4..=6 => "Q2".to_string(),
            7..=9 => "Q3".to_string(),
            10..=12 => "Q4".to_string(),
            _ => String::new(),
        }
    }
}

/// Parse number from localized string
#[wasm_bindgen]
pub fn parse_number(value: &str) -> Option<f64> {
    // Remove currency symbols and spaces
    let cleaned = value
        .replace(['₹', '$', '€', '£', '¥', ',', ' '], "")
        .replace("Cr", "")
        .replace("L", "")
        .replace("K", "");

    cleaned.parse().ok()
}

/// Get ordinal suffix
#[wasm_bindgen]
pub fn get_ordinal(num: u32) -> String {
    let suffix = match (num % 10, num % 100) {
        (1, 11) => "th",
        (2, 12) => "th",
        (3, 13) => "th",
        (1, _) => "st",
        (2, _) => "nd",
        (3, _) => "rd",
        _ => "th",
    };

    format!("{}{}", num, suffix)
}

/// Pluralize word
#[wasm_bindgen]
pub fn pluralize(word: &str, count: i64) -> String {
    if count == 1 {
        word.to_string()
    } else {
        // Simple pluralization rules
        if word.ends_with('y') && !word.ends_with("ay") && !word.ends_with("ey") && !word.ends_with("oy") && !word.ends_with("uy") {
            format!("{}ies", &word[..word.len()-1])
        } else if word.ends_with('s') || word.ends_with('x') || word.ends_with("ch") || word.ends_with("sh") {
            format!("{}es", word)
        } else {
            format!("{}s", word)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_indian_number_format() {
        assert_eq!(format_indian_number("1000000", Some(0)), "10,00,000");
        assert_eq!(format_indian_number("12345678", Some(0)), "1,23,45,678");
        assert_eq!(format_indian_number("1234.56", Some(2)), "1,234.56");
    }

    #[test]
    fn test_compact_format() {
        assert_eq!(format_compact_indian("10000000", Some(1)), "1.0Cr");
        assert_eq!(format_compact_indian("500000", Some(1)), "5.0L");
        assert_eq!(format_compact_indian("5000", Some(1)), "5.0K");
    }

    #[test]
    fn test_amount_to_words() {
        let words = amount_to_words("12345.67", "INR");
        assert!(words.contains("Twelve Thousand"));
        assert!(words.contains("Rupees"));
    }

    #[test]
    fn test_financial_year() {
        assert_eq!(get_financial_year("2024-05-15"), "2024-25");
        assert_eq!(get_financial_year("2024-01-15"), "2023-24");
    }

    #[test]
    fn test_ordinal() {
        assert_eq!(get_ordinal(1), "1st");
        assert_eq!(get_ordinal(2), "2nd");
        assert_eq!(get_ordinal(3), "3rd");
        assert_eq!(get_ordinal(4), "4th");
        assert_eq!(get_ordinal(11), "11th");
        assert_eq!(get_ordinal(21), "21st");
    }

    #[test]
    fn test_pluralize() {
        assert_eq!(pluralize("invoice", 1), "invoice");
        assert_eq!(pluralize("invoice", 2), "invoices");
        assert_eq!(pluralize("company", 2), "companies");
        assert_eq!(pluralize("box", 2), "boxes");
    }
}
