//! Decimal utilities for precise financial calculations

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Standard decimal places for different operations
pub const CURRENCY_DECIMALS: u32 = 2;
pub const RATE_DECIMALS: u32 = 4;
pub const QUANTITY_DECIMALS: u32 = 3;
pub const PERCENTAGE_DECIMALS: u32 = 2;

/// Rounding modes for financial calculations
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub enum RoundingMode {
    /// Round half up (standard banking)
    HalfUp,
    /// Round half even (banker's rounding)
    HalfEven,
    /// Always round up
    Up,
    /// Always round down
    Down,
    /// Round towards zero
    Truncate,
}

/// Parse a string to Decimal, returning zero on failure
#[wasm_bindgen]
pub fn parse_decimal(s: &str) -> String {
    s.parse::<Decimal>()
        .unwrap_or(Decimal::ZERO)
        .to_string()
}

/// Round a decimal to specified places
#[wasm_bindgen]
pub fn round_decimal(value: &str, places: u32) -> String {
    value
        .parse::<Decimal>()
        .map(|d| d.round_dp(places))
        .unwrap_or(Decimal::ZERO)
        .to_string()
}

/// Round for currency display (2 decimal places)
#[wasm_bindgen]
pub fn round_currency(value: &str) -> String {
    round_decimal(value, CURRENCY_DECIMALS)
}

/// Add two decimal values
#[wasm_bindgen]
pub fn add_decimals(a: &str, b: &str) -> String {
    let a = a.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let b = b.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    (a + b).to_string()
}

/// Subtract two decimal values
#[wasm_bindgen]
pub fn subtract_decimals(a: &str, b: &str) -> String {
    let a = a.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let b = b.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    (a - b).to_string()
}

/// Multiply two decimal values
#[wasm_bindgen]
pub fn multiply_decimals(a: &str, b: &str) -> String {
    let a = a.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let b = b.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    (a * b).to_string()
}

/// Divide two decimal values (returns "0" if divisor is zero)
#[wasm_bindgen]
pub fn divide_decimals(a: &str, b: &str) -> String {
    let a = a.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let b = b.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    if b.is_zero() {
        Decimal::ZERO.to_string()
    } else {
        (a / b).to_string()
    }
}

/// Calculate percentage of a value
#[wasm_bindgen]
pub fn percentage(value: &str, percent: &str) -> String {
    let value = value.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let percent = percent.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    (value * percent / dec!(100)).round_dp(CURRENCY_DECIMALS).to_string()
}

/// Compare two decimal values: -1 if a < b, 0 if equal, 1 if a > b
#[wasm_bindgen]
pub fn compare_decimals(a: &str, b: &str) -> i32 {
    let a = a.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let b = b.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    match a.cmp(&b) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    }
}

/// Check if decimal is zero
#[wasm_bindgen]
pub fn is_zero(value: &str) -> bool {
    value.parse::<Decimal>().map(|d| d.is_zero()).unwrap_or(true)
}

/// Get absolute value
#[wasm_bindgen]
pub fn abs_decimal(value: &str) -> String {
    value
        .parse::<Decimal>()
        .map(|d| d.abs())
        .unwrap_or(Decimal::ZERO)
        .to_string()
}

/// Internal helper to parse Decimal
pub fn parse_decimal_internal(s: &str) -> Decimal {
    s.parse().unwrap_or(Decimal::ZERO)
}

/// Internal helper to parse with error
pub fn try_parse_decimal(s: &str) -> Result<Decimal, rust_decimal::Error> {
    s.parse()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_round_currency() {
        assert_eq!(round_currency("123.456"), "123.46");
        assert_eq!(round_currency("123.454"), "123.45");
        assert_eq!(round_currency("123"), "123.00");
    }

    #[test]
    fn test_percentage() {
        assert_eq!(percentage("1000", "18"), "180.00");
        assert_eq!(percentage("1000", "9"), "90.00");
    }

    #[test]
    fn test_arithmetic() {
        assert_eq!(add_decimals("100.50", "50.25"), "150.75");
        assert_eq!(subtract_decimals("100.50", "50.25"), "50.25");
        assert_eq!(multiply_decimals("10", "5.5"), "55.0");
        assert_eq!(divide_decimals("100", "4"), "25");
        assert_eq!(divide_decimals("100", "0"), "0");
    }
}
