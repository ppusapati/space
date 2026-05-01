//! TCS (Tax Collected at Source) calculations
//!
//! Handles TCS computation as per Indian Income Tax Act

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// TCS Section types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TcsSection {
    /// 206C(1) - Timber, forest produce, etc.
    Sec206C1,
    /// 206C(1C) - Parking lot, toll plaza, mining
    Sec206C1C,
    /// 206C(1F) - Motor vehicles over 10 lakhs
    Sec206C1F,
    /// 206C(1G) - Foreign remittance (LRS)
    Sec206C1G,
    /// 206C(1H) - Sale of goods over 50 lakhs
    Sec206C1H,
    /// 206C(1I) - Purchase of overseas tour package
    Sec206C1I,
}

impl TcsSection {
    /// Get default rate for a section
    pub fn default_rate(&self) -> Decimal {
        match self {
            TcsSection::Sec206C1 => dec!(2.5),
            TcsSection::Sec206C1C => dec!(2),
            TcsSection::Sec206C1F => dec!(1),
            TcsSection::Sec206C1G => dec!(5), // 20% for certain cases
            TcsSection::Sec206C1H => dec!(0.1),
            TcsSection::Sec206C1I => dec!(5),
        }
    }

    /// Get threshold limit
    pub fn threshold(&self) -> Decimal {
        match self {
            TcsSection::Sec206C1 => Decimal::ZERO,
            TcsSection::Sec206C1C => Decimal::ZERO,
            TcsSection::Sec206C1F => dec!(1000000), // 10 lakhs
            TcsSection::Sec206C1G => dec!(700000),   // 7 lakhs under LRS
            TcsSection::Sec206C1H => dec!(5000000),  // 50 lakhs
            TcsSection::Sec206C1I => dec!(700000),   // 7 lakhs
        }
    }

    pub fn from_code(code: &str) -> Option<Self> {
        match code.to_uppercase().replace("-", "").replace(" ", "").as_str() {
            "206C1" | "206C(1)" => Some(TcsSection::Sec206C1),
            "206C1C" | "206C(1C)" => Some(TcsSection::Sec206C1C),
            "206C1F" | "206C(1F)" => Some(TcsSection::Sec206C1F),
            "206C1G" | "206C(1G)" => Some(TcsSection::Sec206C1G),
            "206C1H" | "206C(1H)" => Some(TcsSection::Sec206C1H),
            "206C1I" | "206C(1I)" => Some(TcsSection::Sec206C1I),
            _ => None,
        }
    }
}

/// Nature of goods for TCS under 206C(1)
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TcsGoodsNature {
    /// Alcoholic liquor for human consumption
    AlcoholicLiquor,
    /// Timber obtained under forest lease
    Timber,
    /// Tendu leaves
    TenduLeaves,
    /// Timber (other than lease)
    TimberOther,
    /// Forest produce (other than timber/tendu)
    ForestProduce,
    /// Scrap
    Scrap,
    /// Minerals (coal, lignite, iron ore)
    Minerals,
}

impl TcsGoodsNature {
    pub fn rate(&self) -> Decimal {
        match self {
            TcsGoodsNature::AlcoholicLiquor => dec!(1),
            TcsGoodsNature::Timber => dec!(2.5),
            TcsGoodsNature::TenduLeaves => dec!(5),
            TcsGoodsNature::TimberOther => dec!(2.5),
            TcsGoodsNature::ForestProduce => dec!(2.5),
            TcsGoodsNature::Scrap => dec!(1),
            TcsGoodsNature::Minerals => dec!(1),
        }
    }
}

/// TCS calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TcsInput {
    /// Sale/Collection amount
    pub amount: String,
    /// TCS section code
    pub section: String,
    /// Nature of goods (for 206C1)
    pub goods_nature: Option<String>,
    /// Custom TCS rate (overrides default)
    pub rate: Option<String>,
    /// Whether PAN is available
    #[serde(default = "default_true")]
    pub has_pan: bool,
    /// Previous sales in FY (for threshold)
    pub previous_sales: Option<String>,
    /// For LRS - purpose of remittance
    pub lrs_purpose: Option<String>,
    /// For 206C1G - whether for education loan
    #[serde(default)]
    pub is_education_loan: bool,
}

fn default_true() -> bool {
    true
}

/// TCS calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TcsResult {
    /// Original amount
    pub sale_amount: String,
    /// TCS rate applied
    pub tcs_rate: String,
    /// TCS amount
    pub tcs_amount: String,
    /// Total collectible (amount + TCS)
    pub total_collectible: String,
    /// TCS section applied
    pub section: String,
    /// Threshold limit
    pub threshold: String,
    /// Whether TCS is applicable
    pub is_applicable: bool,
    /// Reason if not applicable
    pub reason: Option<String>,
    /// Higher rate applied due to no PAN
    pub higher_rate_applied: bool,
}

/// Calculate TCS
#[wasm_bindgen]
pub fn calculate_tcs(input: JsValue) -> Result<JsValue, JsError> {
    let input: TcsInput = serde_wasm_bindgen::from_value(input)
        .map_err(|e| JsError::new(&format!("Invalid input: {}", e)))?;

    let result = calculate_tcs_internal(&input)
        .map_err(|e| JsError::new(&e))?;

    serde_wasm_bindgen::to_value(&result)
        .map_err(|e| JsError::new(&format!("Serialization error: {}", e)))
}

/// Internal TCS calculation
pub fn calculate_tcs_internal(input: &TcsInput) -> Result<TcsResult, String> {
    let amount: Decimal = input.amount.parse()
        .map_err(|_| "Invalid amount")?;

    let section = TcsSection::from_code(&input.section)
        .ok_or_else(|| format!("Invalid TCS section: {}", input.section))?;

    let threshold = section.threshold();

    // Check cumulative threshold
    let total_sales = if let Some(ref prev) = input.previous_sales {
        let prev: Decimal = prev.parse().unwrap_or_default();
        prev + amount
    } else {
        amount
    };

    // For sections with threshold, check if applicable
    if threshold > Decimal::ZERO && total_sales <= threshold {
        return Ok(TcsResult {
            sale_amount: amount.round_dp(2).to_string(),
            tcs_rate: "0".to_string(),
            tcs_amount: "0".to_string(),
            total_collectible: amount.round_dp(2).to_string(),
            section: input.section.clone(),
            threshold: threshold.to_string(),
            is_applicable: false,
            reason: Some(format!("Total sales below threshold of {}", threshold)),
            higher_rate_applied: false,
        });
    }

    // Determine rate
    let mut tcs_rate = if let Some(ref custom_rate) = input.rate {
        custom_rate.parse().unwrap_or(section.default_rate())
    } else {
        section.default_rate()
    };

    // Special handling for 206C1G (LRS)
    if matches!(section, TcsSection::Sec206C1G) {
        if input.is_education_loan {
            tcs_rate = dec!(0.5); // 0.5% for education loan
        } else {
            tcs_rate = dec!(5); // 5% for amounts above 7 lakhs
        }
    }

    // Higher rate for no PAN (double the rate)
    let mut higher_rate_applied = false;
    if !input.has_pan {
        tcs_rate = tcs_rate * dec!(2);
        if tcs_rate > dec!(5) {
            tcs_rate = dec!(5); // Cap at 5% for certain sections
        }
        higher_rate_applied = true;
    }

    // Calculate TCS on amount exceeding threshold
    let taxable_amount = if threshold > Decimal::ZERO && total_sales > threshold {
        let prev: Decimal = input.previous_sales
            .as_ref()
            .and_then(|p| p.parse().ok())
            .unwrap_or_default();

        if prev >= threshold {
            amount // All of current amount is taxable
        } else {
            total_sales - threshold // Only excess is taxable
        }
    } else {
        amount
    };

    let tcs_amount = (taxable_amount * tcs_rate / dec!(100)).round_dp(2);

    Ok(TcsResult {
        sale_amount: amount.round_dp(2).to_string(),
        tcs_rate: tcs_rate.to_string(),
        tcs_amount: tcs_amount.to_string(),
        total_collectible: (amount + tcs_amount).round_dp(2).to_string(),
        section: input.section.clone(),
        threshold: threshold.to_string(),
        is_applicable: true,
        reason: None,
        higher_rate_applied,
    })
}

/// Simple TCS calculation
#[wasm_bindgen]
pub fn simple_tcs(amount: &str, section: &str, has_pan: bool) -> String {
    let input = TcsInput {
        amount: amount.to_string(),
        section: section.to_string(),
        goods_nature: None,
        rate: None,
        has_pan,
        previous_sales: None,
        lrs_purpose: None,
        is_education_loan: false,
    };

    match calculate_tcs_internal(&input) {
        Ok(result) => result.tcs_amount,
        Err(_) => "0".to_string(),
    }
}

/// Get TCS rate for a section
#[wasm_bindgen]
pub fn get_tcs_rate(section: &str) -> String {
    TcsSection::from_code(section)
        .map(|s| s.default_rate().to_string())
        .unwrap_or_else(|| "0".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tcs_206c1h() {
        let input = TcsInput {
            amount: "10000000".to_string(), // 1 crore
            section: "206C1H".to_string(),
            goods_nature: None,
            rate: None,
            has_pan: true,
            previous_sales: Some("4500000".to_string()), // Previous 45L
            lrs_purpose: None,
            is_education_loan: false,
        };

        let result = calculate_tcs_internal(&input).unwrap();
        // TCS on amount exceeding 50L threshold
        assert!(result.is_applicable);
        assert_eq!(result.tcs_rate, "0.1");
    }

    #[test]
    fn test_tcs_below_threshold() {
        let input = TcsInput {
            amount: "4000000".to_string(), // 40L
            section: "206C1H".to_string(),
            goods_nature: None,
            rate: None,
            has_pan: true,
            previous_sales: None,
            lrs_purpose: None,
            is_education_loan: false,
        };

        let result = calculate_tcs_internal(&input).unwrap();
        assert!(!result.is_applicable);
        assert_eq!(result.tcs_amount, "0");
    }
}
