//! Income tax calculations for Indian tax laws

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::TaxRegime;

/// Income tax calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IncomeTaxInput {
    pub gross_salary: String,
    pub regime: String,  // "old" or "new"
    pub fy: Option<String>,
    // Exemptions (Old Regime)
    pub hra: Option<String>,
    pub lta: Option<String>,
    pub food_allowance: Option<String>,
    // Deductions (Old Regime)
    pub section_80c: Option<String>,
    pub section_80d: Option<String>,
    pub section_80e: Option<String>,
    pub section_80g: Option<String>,
    pub home_loan_interest: Option<String>,
    pub nps_80ccd: Option<String>,
    // Other income
    pub other_income: Option<String>,
    // Age for rebate
    pub age: Option<u32>,
}

/// Tax slab
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaxSlab {
    pub min: String,
    pub max: Option<String>,
    pub rate: String,
    pub tax_amount: String,
}

/// Income tax calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IncomeTaxResult {
    pub gross_salary: String,
    pub exemptions: String,
    pub deductions: String,
    pub taxable_income: String,
    pub tax_on_income: String,
    pub surcharge: String,
    pub cess: String,
    pub total_tax: String,
    pub rebate_87a: String,
    pub net_tax: String,
    pub monthly_tds: String,
    pub effective_rate: String,
    pub regime: String,
    pub slab_breakdown: Vec<TaxSlab>,
}

/// New regime tax slabs (FY 2024-25)
const NEW_REGIME_SLABS: &[(u64, u64, u32)] = &[
    (0, 300000, 0),
    (300001, 700000, 5),
    (700001, 1000000, 10),
    (1000001, 1200000, 15),
    (1200001, 1500000, 20),
    (1500001, u64::MAX, 30),
];

/// Old regime tax slabs (FY 2024-25)
const OLD_REGIME_SLABS: &[(u64, u64, u32)] = &[
    (0, 250000, 0),
    (250001, 500000, 5),
    (500001, 1000000, 20),
    (1000001, u64::MAX, 30),
];

/// Calculate income tax
#[wasm_bindgen]
pub fn calculate_income_tax(input: JsValue) -> JsValue {
    let input: IncomeTaxInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid income tax input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_income_tax_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal income tax calculation
fn calculate_income_tax_internal(input: &IncomeTaxInput) -> IncomeTaxResult {
    let gross_salary: Decimal = input.gross_salary.parse().unwrap_or(Decimal::ZERO);
    let regime = if input.regime.to_lowercase() == "old" {
        TaxRegime::Old
    } else {
        TaxRegime::New
    };

    let age = input.age.unwrap_or(30);

    // Calculate exemptions and deductions based on regime
    let (exemptions, deductions) = if regime == TaxRegime::Old {
        calculate_old_regime_benefits(input)
    } else {
        // New regime has standard deduction of 75,000 (FY 2024-25)
        (dec!(75000), Decimal::ZERO)
    };

    // Add other income
    let other_income: Decimal = input.other_income.as_ref()
        .and_then(|o| o.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let total_income = gross_salary + other_income;
    let taxable_income = (total_income - exemptions - deductions).max(Decimal::ZERO);

    // Calculate tax based on slabs
    let (tax_on_income, slab_breakdown) = calculate_tax_on_slabs(taxable_income, regime);

    // Apply rebate u/s 87A
    let rebate_87a = calculate_rebate_87a(taxable_income, tax_on_income, regime);
    let tax_after_rebate = (tax_on_income - rebate_87a).max(Decimal::ZERO);

    // Calculate surcharge
    let surcharge = calculate_surcharge(taxable_income, tax_after_rebate);

    // Calculate cess (4%)
    let cess = ((tax_after_rebate + surcharge) * dec!(0.04)).round_dp(0);

    let total_tax = tax_after_rebate + surcharge + cess;
    let net_tax = total_tax.round_dp(0);

    // Monthly TDS
    let monthly_tds = (net_tax / dec!(12)).round_dp(0);

    // Effective rate
    let effective_rate = if gross_salary > Decimal::ZERO {
        (net_tax / gross_salary * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    IncomeTaxResult {
        gross_salary: gross_salary.round_dp(0).to_string(),
        exemptions: exemptions.round_dp(0).to_string(),
        deductions: deductions.round_dp(0).to_string(),
        taxable_income: taxable_income.round_dp(0).to_string(),
        tax_on_income: tax_on_income.round_dp(0).to_string(),
        surcharge: surcharge.round_dp(0).to_string(),
        cess: cess.round_dp(0).to_string(),
        total_tax: total_tax.round_dp(0).to_string(),
        rebate_87a: rebate_87a.round_dp(0).to_string(),
        net_tax: net_tax.to_string(),
        monthly_tds: monthly_tds.to_string(),
        effective_rate: format!("{}%", effective_rate),
        regime: if regime == TaxRegime::Old { "old".to_string() } else { "new".to_string() },
        slab_breakdown,
    }
}

/// Calculate exemptions and deductions for old regime
fn calculate_old_regime_benefits(input: &IncomeTaxInput) -> (Decimal, Decimal) {
    // Exemptions
    let hra: Decimal = input.hra.as_ref()
        .and_then(|h| h.parse().ok())
        .unwrap_or(Decimal::ZERO);
    let lta: Decimal = input.lta.as_ref()
        .and_then(|l| l.parse().ok())
        .unwrap_or(Decimal::ZERO);
    let food: Decimal = input.food_allowance.as_ref()
        .and_then(|f| f.parse().ok())
        .unwrap_or(Decimal::ZERO);

    // Standard deduction of 50,000 for old regime
    let standard_deduction = dec!(50000);
    let exemptions = hra + lta + food + standard_deduction;

    // Deductions
    let sec_80c: Decimal = input.section_80c.as_ref()
        .and_then(|s| s.parse().ok())
        .unwrap_or(Decimal::ZERO)
        .min(dec!(150000)); // Max 1.5L

    let sec_80d: Decimal = input.section_80d.as_ref()
        .and_then(|s| s.parse().ok())
        .unwrap_or(Decimal::ZERO)
        .min(dec!(100000)); // Max 1L with parents

    let sec_80e: Decimal = input.section_80e.as_ref()
        .and_then(|s| s.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let sec_80g: Decimal = input.section_80g.as_ref()
        .and_then(|s| s.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let home_loan: Decimal = input.home_loan_interest.as_ref()
        .and_then(|h| h.parse().ok())
        .unwrap_or(Decimal::ZERO)
        .min(dec!(200000)); // Max 2L

    let nps: Decimal = input.nps_80ccd.as_ref()
        .and_then(|n| n.parse().ok())
        .unwrap_or(Decimal::ZERO)
        .min(dec!(50000)); // Additional 50K under 80CCD(1B)

    let deductions = sec_80c + sec_80d + sec_80e + sec_80g + home_loan + nps;

    (exemptions, deductions)
}

/// Calculate tax on income slabs
fn calculate_tax_on_slabs(taxable_income: Decimal, regime: TaxRegime) -> (Decimal, Vec<TaxSlab>) {
    let slabs = if regime == TaxRegime::New {
        NEW_REGIME_SLABS
    } else {
        OLD_REGIME_SLABS
    };

    let income = taxable_income.to_u64().unwrap_or(0);
    let mut total_tax = Decimal::ZERO;
    let mut breakdown = Vec::new();
    let mut remaining = income;

    for (min, max, rate) in slabs {
        if remaining == 0 {
            break;
        }

        let slab_min = *min;
        let slab_max = *max;
        let slab_rate = *rate;

        if income > slab_min {
            let taxable_in_slab = if income > slab_max {
                slab_max - slab_min + 1
            } else {
                income - slab_min + 1
            }.min(remaining);

            let tax_in_slab = Decimal::from(taxable_in_slab) * Decimal::from(slab_rate) / dec!(100);
            total_tax += tax_in_slab;

            breakdown.push(TaxSlab {
                min: slab_min.to_string(),
                max: if slab_max == u64::MAX { None } else { Some(slab_max.to_string()) },
                rate: format!("{}%", slab_rate),
                tax_amount: tax_in_slab.round_dp(0).to_string(),
            });

            remaining = remaining.saturating_sub(taxable_in_slab);
        }
    }

    (total_tax, breakdown)
}

/// Calculate rebate under section 87A
fn calculate_rebate_87a(taxable_income: Decimal, tax: Decimal, regime: TaxRegime) -> Decimal {
    if regime == TaxRegime::New {
        // New regime: Full rebate if income <= 7L (tax up to 25,000)
        if taxable_income <= dec!(700000) {
            tax.min(dec!(25000))
        } else {
            Decimal::ZERO
        }
    } else {
        // Old regime: Rebate if income <= 5L (tax up to 12,500)
        if taxable_income <= dec!(500000) {
            tax.min(dec!(12500))
        } else {
            Decimal::ZERO
        }
    }
}

/// Calculate surcharge
fn calculate_surcharge(taxable_income: Decimal, tax: Decimal) -> Decimal {
    let income = taxable_income.to_u64().unwrap_or(0);

    let surcharge_rate = if income > 50000000 {
        dec!(0.37) // 37% for income > 5 Cr
    } else if income > 20000000 {
        dec!(0.25) // 25% for income > 2 Cr
    } else if income > 10000000 {
        dec!(0.15) // 15% for income > 1 Cr
    } else if income > 5000000 {
        dec!(0.10) // 10% for income > 50L
    } else {
        Decimal::ZERO
    };

    (tax * surcharge_rate).round_dp(0)
}

/// Compare tax under both regimes
#[wasm_bindgen]
pub fn compare_tax_regimes(input: JsValue) -> JsValue {
    let mut input: IncomeTaxInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid input: {}", e);
            return JsValue::NULL;
        }
    };

    // Calculate for old regime
    input.regime = "old".to_string();
    let old_result = calculate_income_tax_internal(&input);

    // Calculate for new regime
    input.regime = "new".to_string();
    let new_result = calculate_income_tax_internal(&input);

    let old_tax: Decimal = old_result.net_tax.parse().unwrap_or(Decimal::ZERO);
    let new_tax: Decimal = new_result.net_tax.parse().unwrap_or(Decimal::ZERO);

    let recommended = if new_tax <= old_tax { "new" } else { "old" };
    let savings = (old_tax - new_tax).abs();

    let result = serde_json::json!({
        "oldRegime": old_result,
        "newRegime": new_result,
        "recommended": recommended,
        "savings": savings.round_dp(0).to_string(),
        "savingsInFavorOf": if new_tax < old_tax { "new" } else { "old" }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Get tax slabs for a regime
#[wasm_bindgen]
pub fn get_tax_slabs(regime: &str, fy: Option<String>) -> JsValue {
    let slabs = if regime.to_lowercase() == "old" {
        OLD_REGIME_SLABS
    } else {
        NEW_REGIME_SLABS
    };

    let result: Vec<serde_json::Value> = slabs.iter().map(|(min, max, rate)| {
        serde_json::json!({
            "min": min,
            "max": if *max == u64::MAX { None::<u64> } else { Some(*max) },
            "rate": rate
        })
    }).collect();

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_regime_no_tax() {
        let input = IncomeTaxInput {
            gross_salary: "700000".to_string(),
            regime: "new".to_string(),
            fy: None,
            hra: None,
            lta: None,
            food_allowance: None,
            section_80c: None,
            section_80d: None,
            section_80e: None,
            section_80g: None,
            home_loan_interest: None,
            nps_80ccd: None,
            other_income: None,
            age: None,
        };

        let result = calculate_income_tax_internal(&input);
        // With standard deduction of 75K, taxable = 625000, which is < 7L
        // So full rebate applies
        assert_eq!(result.net_tax, "0");
    }

    #[test]
    fn test_new_regime_with_tax() {
        let input = IncomeTaxInput {
            gross_salary: "1200000".to_string(),
            regime: "new".to_string(),
            fy: None,
            hra: None,
            lta: None,
            food_allowance: None,
            section_80c: None,
            section_80d: None,
            section_80e: None,
            section_80g: None,
            home_loan_interest: None,
            nps_80ccd: None,
            other_income: None,
            age: None,
        };

        let result = calculate_income_tax_internal(&input);
        // Taxable = 1200000 - 75000 = 1125000
        // Tax should be calculated on this
        assert!(result.net_tax.parse::<i64>().unwrap() > 0);
    }
}
