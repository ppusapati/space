//! TDS (Tax Deducted at Source) calculations
//!
//! Handles TDS computation as per Indian Income Tax Act

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// TDS Section types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TdsSection {
    /// 192 - Salary
    Sec192,
    /// 192A - Premature withdrawal from EPF
    Sec192A,
    /// 193 - Interest on securities
    Sec193,
    /// 194 - Dividends
    Sec194,
    /// 194A - Interest other than securities
    Sec194A,
    /// 194B - Winnings from lottery
    Sec194B,
    /// 194BB - Winnings from horse race
    Sec194BB,
    /// 194C - Contractor payments
    Sec194C,
    /// 194D - Insurance commission
    Sec194D,
    /// 194DA - Life insurance maturity
    Sec194DA,
    /// 194E - Payments to non-resident sportsmen
    Sec194E,
    /// 194EE - NSS deposits
    Sec194EE,
    /// 194F - Repurchase of units by MF
    Sec194F,
    /// 194G - Commission on lottery tickets
    Sec194G,
    /// 194H - Commission or brokerage
    Sec194H,
    /// 194I - Rent
    Sec194I,
    /// 194IA - Property sale
    Sec194IA,
    /// 194IB - Rent by individual/HUF
    Sec194IB,
    /// 194IC - JDA payments
    Sec194IC,
    /// 194J - Professional/Technical fees
    Sec194J,
    /// 194K - Income from units
    Sec194K,
    /// 194LA - Compensation on land acquisition
    Sec194LA,
    /// 194LB - Income from infrastructure debt fund
    Sec194LB,
    /// 194LC - Income from specified company bonds
    Sec194LC,
    /// 194LD - Interest on bonds
    Sec194LD,
    /// 194M - Payments by individuals
    Sec194M,
    /// 194N - Cash withdrawal
    Sec194N,
    /// 194O - E-commerce participants
    Sec194O,
    /// 194P - Senior citizen exemption
    Sec194P,
    /// 194Q - Purchase of goods
    Sec194Q,
    /// 194R - Business perquisites
    Sec194R,
    /// 194S - Crypto/VDA
    Sec194S,
    /// 195 - Other non-resident payments
    Sec195,
    /// 196A - NRNR income
    Sec196A,
    /// 196B - LIC on units from offshore fund
    Sec196B,
    /// 196C - Income from foreign currency bonds
    Sec196C,
    /// 196D - Income from FII
    Sec196D,
}

impl TdsSection {
    /// Get default rate for a section
    pub fn default_rate(&self) -> Decimal {
        match self {
            TdsSection::Sec192 => dec!(0),    // Based on slab
            TdsSection::Sec192A => dec!(10),
            TdsSection::Sec193 => dec!(10),
            TdsSection::Sec194 => dec!(10),
            TdsSection::Sec194A => dec!(10),
            TdsSection::Sec194B => dec!(30),
            TdsSection::Sec194BB => dec!(30),
            TdsSection::Sec194C => dec!(2),    // 1% for individual/HUF, 2% for others
            TdsSection::Sec194D => dec!(5),
            TdsSection::Sec194DA => dec!(5),
            TdsSection::Sec194E => dec!(20),
            TdsSection::Sec194EE => dec!(10),
            TdsSection::Sec194F => dec!(20),
            TdsSection::Sec194G => dec!(5),
            TdsSection::Sec194H => dec!(5),
            TdsSection::Sec194I => dec!(10),   // 2% for plant/machinery, 10% for land/building
            TdsSection::Sec194IA => dec!(1),
            TdsSection::Sec194IB => dec!(5),
            TdsSection::Sec194IC => dec!(10),
            TdsSection::Sec194J => dec!(10),
            TdsSection::Sec194K => dec!(10),
            TdsSection::Sec194LA => dec!(10),
            TdsSection::Sec194LB => dec!(5),
            TdsSection::Sec194LC => dec!(5),
            TdsSection::Sec194LD => dec!(5),
            TdsSection::Sec194M => dec!(5),
            TdsSection::Sec194N => dec!(2),    // Higher rates for non-filers
            TdsSection::Sec194O => dec!(1),
            TdsSection::Sec194P => dec!(0),
            TdsSection::Sec194Q => dec!(0.1),
            TdsSection::Sec194R => dec!(10),
            TdsSection::Sec194S => dec!(1),
            TdsSection::Sec195 => dec!(20),    // Varies by nature
            TdsSection::Sec196A => dec!(20),
            TdsSection::Sec196B => dec!(10),
            TdsSection::Sec196C => dec!(10),
            TdsSection::Sec196D => dec!(20),
        }
    }

    /// Get threshold limit for TDS applicability
    pub fn threshold(&self) -> Decimal {
        match self {
            TdsSection::Sec192 => dec!(250000),  // Basic exemption limit
            TdsSection::Sec193 => dec!(10000),
            TdsSection::Sec194 => dec!(5000),
            TdsSection::Sec194A => dec!(40000),  // For senior citizens: 50000
            TdsSection::Sec194B => dec!(10000),
            TdsSection::Sec194BB => dec!(10000),
            TdsSection::Sec194C => dec!(30000),  // Single payment, 100000 aggregate
            TdsSection::Sec194D => dec!(15000),
            TdsSection::Sec194DA => dec!(100000),
            TdsSection::Sec194H => dec!(15000),
            TdsSection::Sec194I => dec!(240000),
            TdsSection::Sec194IA => dec!(5000000),
            TdsSection::Sec194IB => dec!(50000),
            TdsSection::Sec194J => dec!(30000),
            TdsSection::Sec194K => dec!(5000),
            TdsSection::Sec194M => dec!(5000000),
            TdsSection::Sec194N => dec!(10000000), // 20L/1Cr for filers/non-filers
            TdsSection::Sec194O => dec!(500000),
            TdsSection::Sec194Q => dec!(5000000),
            _ => Decimal::ZERO,
        }
    }

    pub fn from_code(code: &str) -> Option<Self> {
        match code.to_uppercase().as_str() {
            "192" => Some(TdsSection::Sec192),
            "192A" => Some(TdsSection::Sec192A),
            "193" => Some(TdsSection::Sec193),
            "194" => Some(TdsSection::Sec194),
            "194A" => Some(TdsSection::Sec194A),
            "194B" => Some(TdsSection::Sec194B),
            "194BB" => Some(TdsSection::Sec194BB),
            "194C" => Some(TdsSection::Sec194C),
            "194D" => Some(TdsSection::Sec194D),
            "194DA" => Some(TdsSection::Sec194DA),
            "194E" => Some(TdsSection::Sec194E),
            "194EE" => Some(TdsSection::Sec194EE),
            "194F" => Some(TdsSection::Sec194F),
            "194G" => Some(TdsSection::Sec194G),
            "194H" => Some(TdsSection::Sec194H),
            "194I" => Some(TdsSection::Sec194I),
            "194IA" => Some(TdsSection::Sec194IA),
            "194IB" => Some(TdsSection::Sec194IB),
            "194IC" => Some(TdsSection::Sec194IC),
            "194J" => Some(TdsSection::Sec194J),
            "194K" => Some(TdsSection::Sec194K),
            "194LA" => Some(TdsSection::Sec194LA),
            "194LB" => Some(TdsSection::Sec194LB),
            "194LC" => Some(TdsSection::Sec194LC),
            "194LD" => Some(TdsSection::Sec194LD),
            "194M" => Some(TdsSection::Sec194M),
            "194N" => Some(TdsSection::Sec194N),
            "194O" => Some(TdsSection::Sec194O),
            "194P" => Some(TdsSection::Sec194P),
            "194Q" => Some(TdsSection::Sec194Q),
            "194R" => Some(TdsSection::Sec194R),
            "194S" => Some(TdsSection::Sec194S),
            "195" => Some(TdsSection::Sec195),
            "196A" => Some(TdsSection::Sec196A),
            "196B" => Some(TdsSection::Sec196B),
            "196C" => Some(TdsSection::Sec196C),
            "196D" => Some(TdsSection::Sec196D),
            _ => None,
        }
    }
}

/// TDS calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TdsInput {
    /// Payment amount
    pub amount: String,
    /// TDS section code (e.g., "194C", "194J")
    pub section: String,
    /// Custom TDS rate (overrides default)
    pub rate: Option<String>,
    /// Whether PAN is available
    #[serde(default = "default_true")]
    pub has_pan: bool,
    /// Whether payee has filed returns (for 194N)
    #[serde(default = "default_true")]
    pub is_filer: bool,
    /// Whether lower deduction certificate applies
    #[serde(default)]
    pub has_ldc: bool,
    /// LDC rate if applicable
    pub ldc_rate: Option<String>,
    /// Certificate number if LDC
    pub ldc_certificate: Option<String>,
    /// Previous payments in FY (for threshold calculation)
    pub previous_payments: Option<String>,
    /// Whether surcharge applies
    #[serde(default)]
    pub include_surcharge: bool,
    /// Whether cess applies
    #[serde(default)]
    pub include_cess: bool,
}

fn default_true() -> bool {
    true
}

/// TDS calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TdsResult {
    /// Original amount
    pub gross_amount: String,
    /// TDS base rate applied
    pub tds_rate: String,
    /// TDS amount before surcharge/cess
    pub tds_amount: String,
    /// Surcharge amount
    pub surcharge: String,
    /// Surcharge rate
    pub surcharge_rate: String,
    /// Health and Education Cess
    pub cess: String,
    /// Cess rate (typically 4%)
    pub cess_rate: String,
    /// Total TDS (including surcharge and cess)
    pub total_tds: String,
    /// Net payable to vendor
    pub net_payable: String,
    /// TDS section applied
    pub section: String,
    /// Threshold limit for the section
    pub threshold: String,
    /// Whether TDS is applicable (above threshold)
    pub is_applicable: bool,
    /// Reason if not applicable
    pub reason: Option<String>,
    /// Higher rate applied due to no PAN
    pub higher_rate_applied: bool,
}

/// Calculate TDS
#[wasm_bindgen]
pub fn calculate_tds(input: JsValue) -> Result<JsValue, JsError> {
    let input: TdsInput = serde_wasm_bindgen::from_value(input)
        .map_err(|e| JsError::new(&format!("Invalid input: {}", e)))?;

    let result = calculate_tds_internal(&input)
        .map_err(|e| JsError::new(&e))?;

    serde_wasm_bindgen::to_value(&result)
        .map_err(|e| JsError::new(&format!("Serialization error: {}", e)))
}

/// Internal TDS calculation
pub fn calculate_tds_internal(input: &TdsInput) -> Result<TdsResult, String> {
    let amount: Decimal = input.amount.parse()
        .map_err(|_| "Invalid amount")?;

    let section = TdsSection::from_code(&input.section)
        .ok_or_else(|| format!("Invalid TDS section: {}", input.section))?;

    let threshold = section.threshold();

    // Check previous payments for cumulative threshold
    let total_payments = if let Some(ref prev) = input.previous_payments {
        let prev: Decimal = prev.parse().unwrap_or_default();
        prev + amount
    } else {
        amount
    };

    // Check if TDS is applicable
    if total_payments <= threshold {
        return Ok(TdsResult {
            gross_amount: amount.round_dp(2).to_string(),
            tds_rate: "0".to_string(),
            tds_amount: "0".to_string(),
            surcharge: "0".to_string(),
            surcharge_rate: "0".to_string(),
            cess: "0".to_string(),
            cess_rate: "0".to_string(),
            total_tds: "0".to_string(),
            net_payable: amount.round_dp(2).to_string(),
            section: input.section.clone(),
            threshold: threshold.to_string(),
            is_applicable: false,
            reason: Some(format!("Amount below threshold of {}", threshold)),
            higher_rate_applied: false,
        });
    }

    // Determine applicable rate
    let mut tds_rate = if let Some(ref custom_rate) = input.rate {
        custom_rate.parse().unwrap_or(section.default_rate())
    } else {
        section.default_rate()
    };

    // Apply LDC rate if available
    if input.has_ldc {
        if let Some(ref ldc_rate) = input.ldc_rate {
            tds_rate = ldc_rate.parse().unwrap_or(tds_rate);
        }
    }

    // Higher rate for no PAN (20% or rate, whichever is higher)
    let mut higher_rate_applied = false;
    if !input.has_pan {
        let higher_rate = dec!(20);
        if higher_rate > tds_rate {
            tds_rate = higher_rate;
            higher_rate_applied = true;
        }
    }

    // Calculate base TDS
    let tds_amount = (amount * tds_rate / dec!(100)).round_dp(2);

    // Calculate surcharge if applicable (for high value transactions)
    let mut surcharge = Decimal::ZERO;
    let mut surcharge_rate = Decimal::ZERO;
    if input.include_surcharge && amount >= dec!(10000000) {
        surcharge_rate = if amount >= dec!(50000000) { dec!(15) } else { dec!(10) };
        surcharge = (tds_amount * surcharge_rate / dec!(100)).round_dp(2);
    }

    // Calculate cess (Health and Education Cess - 4%)
    let mut cess = Decimal::ZERO;
    let cess_rate = dec!(4);
    if input.include_cess {
        cess = ((tds_amount + surcharge) * cess_rate / dec!(100)).round_dp(2);
    }

    let total_tds = tds_amount + surcharge + cess;

    Ok(TdsResult {
        gross_amount: amount.round_dp(2).to_string(),
        tds_rate: tds_rate.to_string(),
        tds_amount: tds_amount.to_string(),
        surcharge: surcharge.round_dp(2).to_string(),
        surcharge_rate: surcharge_rate.to_string(),
        cess: cess.round_dp(2).to_string(),
        cess_rate: if input.include_cess { cess_rate.to_string() } else { "0".to_string() },
        total_tds: total_tds.round_dp(2).to_string(),
        net_payable: (amount - total_tds).round_dp(2).to_string(),
        section: input.section.clone(),
        threshold: threshold.to_string(),
        is_applicable: true,
        reason: None,
        higher_rate_applied,
    })
}

/// Simple TDS calculation
#[wasm_bindgen]
pub fn simple_tds(amount: &str, section: &str, has_pan: bool) -> String {
    let input = TdsInput {
        amount: amount.to_string(),
        section: section.to_string(),
        rate: None,
        has_pan,
        is_filer: true,
        has_ldc: false,
        ldc_rate: None,
        ldc_certificate: None,
        previous_payments: None,
        include_surcharge: false,
        include_cess: false,
    };

    match calculate_tds_internal(&input) {
        Ok(result) => result.total_tds,
        Err(_) => "0".to_string(),
    }
}

/// Get TDS rate for a section
#[wasm_bindgen]
pub fn get_tds_rate(section: &str) -> String {
    TdsSection::from_code(section)
        .map(|s| s.default_rate().to_string())
        .unwrap_or_else(|| "0".to_string())
}

/// Get TDS threshold for a section
#[wasm_bindgen]
pub fn get_tds_threshold(section: &str) -> String {
    TdsSection::from_code(section)
        .map(|s| s.threshold().to_string())
        .unwrap_or_else(|| "0".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tds_194c() {
        let input = TdsInput {
            amount: "100000".to_string(),
            section: "194C".to_string(),
            rate: None,
            has_pan: true,
            is_filer: true,
            has_ldc: false,
            ldc_rate: None,
            ldc_certificate: None,
            previous_payments: None,
            include_surcharge: false,
            include_cess: false,
        };

        let result = calculate_tds_internal(&input).unwrap();
        assert_eq!(result.tds_rate, "2");
        assert_eq!(result.tds_amount, "2000.00");
        assert!(result.is_applicable);
    }

    #[test]
    fn test_tds_no_pan() {
        let input = TdsInput {
            amount: "100000".to_string(),
            section: "194J".to_string(),
            rate: None,
            has_pan: false,
            is_filer: true,
            has_ldc: false,
            ldc_rate: None,
            ldc_certificate: None,
            previous_payments: None,
            include_surcharge: false,
            include_cess: false,
        };

        let result = calculate_tds_internal(&input).unwrap();
        assert_eq!(result.tds_rate, "20"); // Higher rate for no PAN
        assert!(result.higher_rate_applied);
    }

    #[test]
    fn test_tds_below_threshold() {
        let input = TdsInput {
            amount: "20000".to_string(),
            section: "194C".to_string(),
            rate: None,
            has_pan: true,
            is_filer: true,
            has_ldc: false,
            ldc_rate: None,
            ldc_certificate: None,
            previous_payments: None,
            include_surcharge: false,
            include_cess: false,
        };

        let result = calculate_tds_internal(&input).unwrap();
        assert!(!result.is_applicable);
        assert_eq!(result.total_tds, "0");
    }
}
