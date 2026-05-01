//! Money and currency handling

use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Supported currencies
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Currency {
    INR,
    USD,
    EUR,
    GBP,
    AED,
    SGD,
    JPY,
    CNY,
    AUD,
    CAD,
}

impl Currency {
    pub fn symbol(&self) -> &'static str {
        match self {
            Currency::INR => "₹",
            Currency::USD => "$",
            Currency::EUR => "€",
            Currency::GBP => "£",
            Currency::AED => "د.إ",
            Currency::SGD => "S$",
            Currency::JPY => "¥",
            Currency::CNY => "¥",
            Currency::AUD => "A$",
            Currency::CAD => "C$",
        }
    }

    pub fn code(&self) -> &'static str {
        match self {
            Currency::INR => "INR",
            Currency::USD => "USD",
            Currency::EUR => "EUR",
            Currency::GBP => "GBP",
            Currency::AED => "AED",
            Currency::SGD => "SGD",
            Currency::JPY => "JPY",
            Currency::CNY => "CNY",
            Currency::AUD => "AUD",
            Currency::CAD => "CAD",
        }
    }

    pub fn decimal_places(&self) -> u32 {
        match self {
            Currency::JPY => 0,
            _ => 2,
        }
    }

    pub fn from_code(code: &str) -> Option<Self> {
        match code.to_uppercase().as_str() {
            "INR" => Some(Currency::INR),
            "USD" => Some(Currency::USD),
            "EUR" => Some(Currency::EUR),
            "GBP" => Some(Currency::GBP),
            "AED" => Some(Currency::AED),
            "SGD" => Some(Currency::SGD),
            "JPY" => Some(Currency::JPY),
            "CNY" => Some(Currency::CNY),
            "AUD" => Some(Currency::AUD),
            "CAD" => Some(Currency::CAD),
            _ => None,
        }
    }
}

/// Money value with currency
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Money {
    pub amount: String,
    pub currency: String,
}

impl Money {
    pub fn new(amount: Decimal, currency: Currency) -> Self {
        Self {
            amount: amount.round_dp(currency.decimal_places()).to_string(),
            currency: currency.code().to_string(),
        }
    }

    pub fn zero(currency: Currency) -> Self {
        Self {
            amount: "0".to_string(),
            currency: currency.code().to_string(),
        }
    }
}

/// Currency exchange rate
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExchangeRate {
    pub from_currency: String,
    pub to_currency: String,
    pub rate: String,
    pub date: String,
}

/// Convert currency
#[wasm_bindgen]
pub fn convert_currency(amount: &str, rate: &str) -> String {
    let amount = amount.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let rate = rate.parse::<Decimal>().unwrap_or(dec!(1));
    (amount * rate).round_dp(2).to_string()
}

/// Format money with currency symbol
#[wasm_bindgen]
pub fn format_money(amount: &str, currency_code: &str) -> String {
    let amount = amount.parse::<Decimal>().unwrap_or(Decimal::ZERO);
    let currency = Currency::from_code(currency_code).unwrap_or(Currency::INR);
    let rounded = amount.round_dp(currency.decimal_places());

    format!("{}{}", currency.symbol(), rounded)
}

/// Get currency symbol
#[wasm_bindgen]
pub fn get_currency_symbol(currency_code: &str) -> String {
    Currency::from_code(currency_code)
        .map(|c| c.symbol().to_string())
        .unwrap_or_else(|| currency_code.to_string())
}

/// Get currency decimal places
#[wasm_bindgen]
pub fn get_currency_decimals(currency_code: &str) -> u32 {
    Currency::from_code(currency_code)
        .map(|c| c.decimal_places())
        .unwrap_or(2)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_format_money() {
        assert_eq!(format_money("1234.56", "INR"), "₹1234.56");
        assert_eq!(format_money("1234.56", "USD"), "$1234.56");
        assert_eq!(format_money("1234", "JPY"), "¥1234");
    }

    #[test]
    fn test_convert_currency() {
        assert_eq!(convert_currency("100", "83.50"), "8350.00");
    }
}
