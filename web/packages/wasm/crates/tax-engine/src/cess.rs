//! GST Cess calculations
//!
//! Compensation Cess under GST for sin goods and luxury items

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Cess types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum CessType {
    /// Ad valorem (percentage based)
    AdValorem,
    /// Specific (fixed amount per unit)
    Specific,
    /// Mixed (both ad valorem and specific)
    Mixed,
}

/// Cess configuration for a product
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CessConfig {
    pub hsn_code: String,
    pub description: String,
    pub cess_type: String,
    pub ad_valorem_rate: Option<String>,
    pub specific_rate: Option<String>,
    pub specific_unit: Option<String>,
}

/// Common cess rates for items under GST Compensation Cess
const CESS_RATES: &[(&str, &str, &str, Option<&str>, Option<&str>)] = &[
    // Motor vehicles
    ("8702", "Motor vehicles for transport (10+)", "AdValorem", Some("15"), None),
    ("8703", "Motor cars (petrol <1200cc)", "AdValorem", Some("1"), None),
    ("8703", "Motor cars (diesel <1500cc)", "AdValorem", Some("3"), None),
    ("8703", "Motor cars (1200-1500cc petrol)", "AdValorem", Some("17"), None),
    ("8703", "Motor cars (1500cc+ diesel)", "AdValorem", Some("20"), None),
    ("8703", "SUVs (length >4m, capacity 1500cc+)", "AdValorem", Some("22"), None),
    ("8703", "Hybrid vehicles", "AdValorem", Some("15"), None),

    // Tobacco products
    ("2401", "Unmanufactured tobacco", "Mixed", Some("0"), Some("4170/kg")),
    ("2402", "Cigarettes (not exceeding 65mm)", "Mixed", Some("5"), Some("4170/1000")),
    ("2402", "Cigarettes (65-70mm)", "Mixed", Some("5"), Some("4170/1000")),
    ("2402", "Cigarettes (70-75mm)", "Mixed", Some("5"), Some("4170/1000")),
    ("2402", "Cigarettes (exceeding 75mm)", "Mixed", Some("36"), Some("4170/1000")),
    ("2403", "Chewing tobacco", "Mixed", Some("0"), Some("160/kg")),
    ("2403", "Zarda", "Mixed", Some("0"), Some("4170/kg")),
    ("2403", "Pan masala containing tobacco", "Mixed", Some("0"), Some("4170/kg")),

    // Aerated beverages
    ("2202", "Aerated waters with sugar", "AdValorem", Some("12"), None),
    ("2202", "Caffeinated beverages", "AdValorem", Some("12"), None),

    // Coal and lignite
    ("2701", "Coal", "Specific", None, Some("400/t")),
    ("2702", "Lignite", "Specific", None, Some("400/t")),
    ("2703", "Peat", "Specific", None, Some("400/t")),
];

/// Calculate cess for a product
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CessInput {
    /// HSN code
    pub hsn_code: String,
    /// Taxable value
    pub value: String,
    /// Quantity (for specific cess)
    pub quantity: Option<String>,
    /// Unit (for specific cess)
    pub unit: Option<String>,
    /// Custom ad valorem rate (overrides default)
    pub ad_valorem_rate: Option<String>,
    /// Custom specific rate (overrides default)
    pub specific_rate: Option<String>,
}

/// Cess calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CessResult {
    /// HSN code
    pub hsn_code: String,
    /// Taxable value
    pub taxable_value: String,
    /// Ad valorem cess
    pub ad_valorem_cess: String,
    /// Ad valorem rate
    pub ad_valorem_rate: String,
    /// Specific cess
    pub specific_cess: String,
    /// Specific rate
    pub specific_rate: String,
    /// Total cess
    pub total_cess: String,
    /// Cess type
    pub cess_type: String,
}

/// Calculate cess
#[wasm_bindgen]
pub fn calculate_cess(input: JsValue) -> Result<JsValue, JsError> {
    let input: CessInput = serde_wasm_bindgen::from_value(input)
        .map_err(|e| JsError::new(&format!("Invalid input: {}", e)))?;

    let result = calculate_cess_internal(&input)
        .map_err(|e| JsError::new(&e))?;

    serde_wasm_bindgen::to_value(&result)
        .map_err(|e| JsError::new(&format!("Serialization error: {}", e)))
}

/// Internal cess calculation
pub fn calculate_cess_internal(input: &CessInput) -> Result<CessResult, String> {
    let value: Decimal = input.value.parse()
        .map_err(|_| "Invalid value")?;

    // Find matching cess rate
    let hsn = input.hsn_code.trim();
    let mut ad_valorem_rate = Decimal::ZERO;
    let mut specific_rate_str = String::new();
    let mut cess_type = "None".to_string();

    // Check custom rates first
    if let Some(ref rate) = input.ad_valorem_rate {
        if let Ok(r) = rate.parse::<Decimal>() {
            ad_valorem_rate = r;
            cess_type = "AdValorem".to_string();
        }
    }

    if let Some(ref rate) = input.specific_rate {
        specific_rate_str = rate.clone();
        if cess_type == "AdValorem" {
            cess_type = "Mixed".to_string();
        } else {
            cess_type = "Specific".to_string();
        }
    }

    // Look up default rates if not provided
    if ad_valorem_rate.is_zero() && specific_rate_str.is_empty() {
        for (code, _desc, c_type, ad_val, specific) in CESS_RATES {
            if hsn.starts_with(code) {
                cess_type = c_type.to_string();
                if let Some(rate) = ad_val {
                    ad_valorem_rate = rate.parse().unwrap_or_default();
                }
                if let Some(rate) = specific {
                    specific_rate_str = rate.to_string();
                }
                break;
            }
        }
    }

    // Calculate ad valorem cess
    let ad_valorem_cess = (value * ad_valorem_rate / dec!(100)).round_dp(2);

    // Calculate specific cess
    let mut specific_cess = Decimal::ZERO;
    if !specific_rate_str.is_empty() {
        if let Some(ref qty) = input.quantity {
            if let Ok(quantity) = qty.parse::<Decimal>() {
                // Parse specific rate (e.g., "400/t" or "4170/kg")
                let rate = parse_specific_rate(&specific_rate_str);
                specific_cess = (quantity * rate).round_dp(2);
            }
        }
    }

    let total_cess = ad_valorem_cess + specific_cess;

    Ok(CessResult {
        hsn_code: hsn.to_string(),
        taxable_value: value.round_dp(2).to_string(),
        ad_valorem_cess: ad_valorem_cess.to_string(),
        ad_valorem_rate: ad_valorem_rate.to_string(),
        specific_cess: specific_cess.to_string(),
        specific_rate: specific_rate_str,
        total_cess: total_cess.to_string(),
        cess_type,
    })
}

/// Parse specific rate string (e.g., "400/t", "4170/kg", "4170/1000")
fn parse_specific_rate(rate_str: &str) -> Decimal {
    // Extract numeric part
    let rate_str = rate_str.trim();
    let numeric: String = rate_str.chars()
        .take_while(|c| c.is_ascii_digit() || *c == '.')
        .collect();

    numeric.parse().unwrap_or_default()
}

/// Get cess rate for HSN code
#[wasm_bindgen]
pub fn get_cess_rate(hsn_code: &str) -> JsValue {
    let hsn = hsn_code.trim();

    for (code, desc, c_type, ad_val, specific) in CESS_RATES {
        if hsn.starts_with(code) {
            let config = CessConfig {
                hsn_code: code.to_string(),
                description: desc.to_string(),
                cess_type: c_type.to_string(),
                ad_valorem_rate: ad_val.map(|s| s.to_string()),
                specific_rate: specific.map(|s| s.to_string()),
                specific_unit: None,
            };
            return serde_wasm_bindgen::to_value(&config).unwrap_or(JsValue::NULL);
        }
    }

    JsValue::NULL
}

/// Check if HSN code has cess
#[wasm_bindgen]
pub fn has_cess(hsn_code: &str) -> bool {
    let hsn = hsn_code.trim();
    CESS_RATES.iter().any(|(code, _, _, _, _)| hsn.starts_with(code))
}

/// Simple cess calculation (ad valorem only)
#[wasm_bindgen]
pub fn simple_cess(value: &str, rate: &str) -> String {
    let value: Decimal = value.parse().unwrap_or_default();
    let rate: Decimal = rate.parse().unwrap_or_default();

    (value * rate / dec!(100)).round_dp(2).to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cess_calculation() {
        let input = CessInput {
            hsn_code: "8703".to_string(),
            value: "1000000".to_string(),
            quantity: None,
            unit: None,
            ad_valorem_rate: Some("15".to_string()),
            specific_rate: None,
        };

        let result = calculate_cess_internal(&input).unwrap();
        assert_eq!(result.ad_valorem_cess, "150000.00");
        assert_eq!(result.total_cess, "150000.00");
    }

    #[test]
    fn test_coal_cess() {
        let input = CessInput {
            hsn_code: "2701".to_string(),
            value: "100000".to_string(),
            quantity: Some("10".to_string()), // 10 tonnes
            unit: Some("t".to_string()),
            ad_valorem_rate: None,
            specific_rate: Some("400/t".to_string()),
        };

        let result = calculate_cess_internal(&input).unwrap();
        assert_eq!(result.specific_cess, "4000.00"); // 10 × 400
    }

    #[test]
    fn test_has_cess() {
        assert!(has_cess("8703")); // Cars
        assert!(has_cess("2701")); // Coal
        assert!(has_cess("2402")); // Cigarettes
        assert!(!has_cess("8471")); // Computers
    }
}
