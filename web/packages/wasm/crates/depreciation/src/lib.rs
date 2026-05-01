//! Asset Depreciation Calculations
//!
//! This crate provides:
//! - Straight Line Method (SLM)
//! - Written Down Value (WDV)
//! - Double Declining Balance
//! - Units of Production
//! - Sum of Years Digits
//! - Depreciation schedules for Indian Income Tax Act

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Depreciation method
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum DepreciationMethod {
    /// Straight Line Method
    Slm,
    /// Written Down Value (Reducing Balance)
    Wdv,
    /// Double Declining Balance
    DoubleDeclining,
    /// Units of Production
    UnitsOfProduction,
    /// Sum of Years Digits
    SumOfYearsDigits,
}

/// Asset input for depreciation calculation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AssetInput {
    pub asset_name: String,
    pub cost: String,
    pub salvage_value: Option<String>,
    pub useful_life_years: u32,
    pub purchase_date: String,
    pub method: DepreciationMethod,
    /// For WDV method
    pub wdv_rate: Option<String>,
    /// For units of production
    pub total_units: Option<String>,
    pub units_this_period: Option<String>,
    /// For partial year calculation
    pub fiscal_year_start_month: Option<u32>,
}

/// Depreciation result for a single period
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepreciationPeriod {
    pub period: u32,
    pub year: String,
    pub opening_value: String,
    pub depreciation: String,
    pub accumulated_depreciation: String,
    pub closing_value: String,
    pub rate: String,
}

/// Complete depreciation schedule
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepreciationSchedule {
    pub asset_name: String,
    pub method: String,
    pub cost: String,
    pub salvage_value: String,
    pub depreciable_amount: String,
    pub useful_life_years: u32,
    pub annual_rate: String,
    pub periods: Vec<DepreciationPeriod>,
    pub total_depreciation: String,
}

/// Calculate depreciation schedule
#[wasm_bindgen]
pub fn calculate_depreciation_schedule(input: JsValue) -> JsValue {
    let input: AssetInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid asset input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_schedule_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal depreciation schedule calculation
fn calculate_schedule_internal(input: &AssetInput) -> DepreciationSchedule {
    let cost: Decimal = input.cost.parse().unwrap_or(Decimal::ZERO);
    let salvage: Decimal = input.salvage_value.as_ref()
        .and_then(|s| s.parse().ok())
        .unwrap_or(Decimal::ZERO);
    let useful_life = input.useful_life_years;

    let depreciable_amount = cost - salvage;

    let (periods, annual_rate) = match input.method {
        DepreciationMethod::Slm => calculate_slm(cost, salvage, useful_life, &input.purchase_date),
        DepreciationMethod::Wdv => {
            let rate = input.wdv_rate.as_ref()
                .and_then(|r| r.parse().ok())
                .unwrap_or_else(|| calculate_wdv_rate(useful_life));
            calculate_wdv(cost, salvage, useful_life, rate, &input.purchase_date)
        }
        DepreciationMethod::DoubleDeclining => calculate_ddb(cost, salvage, useful_life, &input.purchase_date),
        DepreciationMethod::SumOfYearsDigits => calculate_syd(cost, salvage, useful_life, &input.purchase_date),
        DepreciationMethod::UnitsOfProduction => {
            let total_units: Decimal = input.total_units.as_ref()
                .and_then(|u| u.parse().ok())
                .unwrap_or(dec!(1));
            calculate_units_of_production(cost, salvage, total_units, useful_life, &input.purchase_date)
        }
    };

    let total_depreciation: Decimal = periods.iter()
        .map(|p| p.depreciation.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    DepreciationSchedule {
        asset_name: input.asset_name.clone(),
        method: format!("{:?}", input.method),
        cost: cost.to_string(),
        salvage_value: salvage.to_string(),
        depreciable_amount: depreciable_amount.to_string(),
        useful_life_years: useful_life,
        annual_rate,
        periods,
        total_depreciation: total_depreciation.round_dp(2).to_string(),
    }
}

/// Calculate Straight Line Method depreciation
fn calculate_slm(
    cost: Decimal,
    salvage: Decimal,
    useful_life: u32,
    purchase_date: &str,
) -> (Vec<DepreciationPeriod>, String) {
    let depreciable = cost - salvage;
    let annual_dep = (depreciable / Decimal::from(useful_life)).round_dp(2);
    let rate = if cost > Decimal::ZERO {
        (annual_dep / cost * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    let mut periods = Vec::new();
    let mut accumulated = Decimal::ZERO;
    let mut book_value = cost;

    // Get start year from purchase date
    let start_year: i32 = purchase_date.split('-')
        .next()
        .and_then(|y| y.parse().ok())
        .unwrap_or(2024);

    for period in 1..=useful_life {
        let dep = if period == useful_life {
            // Last period - adjust for rounding
            (book_value - salvage).max(Decimal::ZERO)
        } else {
            annual_dep.min(book_value - salvage)
        };

        accumulated += dep;
        let closing = cost - accumulated;

        periods.push(DepreciationPeriod {
            period,
            year: (start_year + period as i32 - 1).to_string(),
            opening_value: book_value.round_dp(2).to_string(),
            depreciation: dep.round_dp(2).to_string(),
            accumulated_depreciation: accumulated.round_dp(2).to_string(),
            closing_value: closing.round_dp(2).to_string(),
            rate: format!("{}%", rate),
        });

        book_value = closing;
    }

    (periods, format!("{}%", rate))
}

/// Calculate WDV rate from useful life
fn calculate_wdv_rate(useful_life: u32) -> Decimal {
    // Standard WDV rates based on useful life
    match useful_life {
        1..=3 => dec!(0.40),   // 40%
        4..=5 => dec!(0.25),   // 25%
        6..=10 => dec!(0.15),  // 15%
        _ => dec!(0.10),       // 10%
    }
}

/// Calculate Written Down Value depreciation
fn calculate_wdv(
    cost: Decimal,
    salvage: Decimal,
    useful_life: u32,
    rate: Decimal,
    purchase_date: &str,
) -> (Vec<DepreciationPeriod>, String) {
    let mut periods = Vec::new();
    let mut accumulated = Decimal::ZERO;
    let mut book_value = cost;

    let start_year: i32 = purchase_date.split('-')
        .next()
        .and_then(|y| y.parse().ok())
        .unwrap_or(2024);

    for period in 1..=useful_life {
        let dep = if period == useful_life {
            (book_value - salvage).max(Decimal::ZERO)
        } else {
            (book_value * rate).round_dp(2).min(book_value - salvage)
        };

        accumulated += dep;
        let closing = cost - accumulated;

        let period_rate = if book_value > Decimal::ZERO {
            (dep / book_value * dec!(100)).round_dp(2)
        } else {
            Decimal::ZERO
        };

        periods.push(DepreciationPeriod {
            period,
            year: (start_year + period as i32 - 1).to_string(),
            opening_value: book_value.round_dp(2).to_string(),
            depreciation: dep.round_dp(2).to_string(),
            accumulated_depreciation: accumulated.round_dp(2).to_string(),
            closing_value: closing.round_dp(2).to_string(),
            rate: format!("{}%", period_rate),
        });

        book_value = closing;
    }

    let rate_pct = (rate * dec!(100)).round_dp(2);
    (periods, format!("{}%", rate_pct))
}

/// Calculate Double Declining Balance depreciation
fn calculate_ddb(
    cost: Decimal,
    salvage: Decimal,
    useful_life: u32,
    purchase_date: &str,
) -> (Vec<DepreciationPeriod>, String) {
    let rate = dec!(2) / Decimal::from(useful_life);
    let mut periods = Vec::new();
    let mut accumulated = Decimal::ZERO;
    let mut book_value = cost;

    let start_year: i32 = purchase_date.split('-')
        .next()
        .and_then(|y| y.parse().ok())
        .unwrap_or(2024);

    for period in 1..=useful_life {
        let dep = if period == useful_life {
            (book_value - salvage).max(Decimal::ZERO)
        } else {
            (book_value * rate).round_dp(2).min(book_value - salvage)
        };

        accumulated += dep;
        let closing = cost - accumulated;

        let period_rate = if book_value > Decimal::ZERO {
            (dep / book_value * dec!(100)).round_dp(2)
        } else {
            Decimal::ZERO
        };

        periods.push(DepreciationPeriod {
            period,
            year: (start_year + period as i32 - 1).to_string(),
            opening_value: book_value.round_dp(2).to_string(),
            depreciation: dep.round_dp(2).to_string(),
            accumulated_depreciation: accumulated.round_dp(2).to_string(),
            closing_value: closing.round_dp(2).to_string(),
            rate: format!("{}%", period_rate),
        });

        book_value = closing;
    }

    let rate_pct = (rate * dec!(100)).round_dp(2);
    (periods, format!("{}%", rate_pct))
}

/// Calculate Sum of Years Digits depreciation
fn calculate_syd(
    cost: Decimal,
    salvage: Decimal,
    useful_life: u32,
    purchase_date: &str,
) -> (Vec<DepreciationPeriod>, String) {
    let depreciable = cost - salvage;
    let sum_of_years = Decimal::from(useful_life * (useful_life + 1) / 2);

    let mut periods = Vec::new();
    let mut accumulated = Decimal::ZERO;
    let mut book_value = cost;

    let start_year: i32 = purchase_date.split('-')
        .next()
        .and_then(|y| y.parse().ok())
        .unwrap_or(2024);

    for period in 1..=useful_life {
        let remaining_years = useful_life - period + 1;
        let fraction = Decimal::from(remaining_years) / sum_of_years;
        let dep = (depreciable * fraction).round_dp(2);

        accumulated += dep;
        let closing = cost - accumulated;

        let period_rate = if book_value > Decimal::ZERO {
            (dep / book_value * dec!(100)).round_dp(2)
        } else {
            Decimal::ZERO
        };

        periods.push(DepreciationPeriod {
            period,
            year: (start_year + period as i32 - 1).to_string(),
            opening_value: book_value.round_dp(2).to_string(),
            depreciation: dep.round_dp(2).to_string(),
            accumulated_depreciation: accumulated.round_dp(2).to_string(),
            closing_value: closing.round_dp(2).to_string(),
            rate: format!("{}%", period_rate),
        });

        book_value = closing;
    }

    (periods, "Variable".to_string())
}

/// Calculate Units of Production depreciation
fn calculate_units_of_production(
    cost: Decimal,
    salvage: Decimal,
    total_units: Decimal,
    useful_life: u32,
    purchase_date: &str,
) -> (Vec<DepreciationPeriod>, String) {
    let depreciable = cost - salvage;
    let dep_per_unit = if total_units > Decimal::ZERO {
        depreciable / total_units
    } else {
        Decimal::ZERO
    };

    let units_per_year = total_units / Decimal::from(useful_life);
    let mut periods = Vec::new();
    let mut accumulated = Decimal::ZERO;
    let mut book_value = cost;

    let start_year: i32 = purchase_date.split('-')
        .next()
        .and_then(|y| y.parse().ok())
        .unwrap_or(2024);

    for period in 1..=useful_life {
        let units_this_period = if period == useful_life {
            // Adjust for any remaining units
            total_units - units_per_year * Decimal::from(useful_life - 1)
        } else {
            units_per_year
        };

        let dep = (units_this_period * dep_per_unit).round_dp(2);
        accumulated += dep;
        let closing = cost - accumulated;

        let period_rate = if book_value > Decimal::ZERO {
            (dep / book_value * dec!(100)).round_dp(2)
        } else {
            Decimal::ZERO
        };

        periods.push(DepreciationPeriod {
            period,
            year: (start_year + period as i32 - 1).to_string(),
            opening_value: book_value.round_dp(2).to_string(),
            depreciation: dep.round_dp(2).to_string(),
            accumulated_depreciation: accumulated.round_dp(2).to_string(),
            closing_value: closing.round_dp(2).to_string(),
            rate: format!("{}%", period_rate),
        });

        book_value = closing;
    }

    (periods, format!("{}/unit", dep_per_unit.round_dp(4)))
}

/// Indian Income Tax depreciation rates by asset class
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItDepreciationRate {
    pub asset_class: String,
    pub description: String,
    pub wdv_rate: String,
    pub slm_rate: String,
    pub block: String,
}

/// Get Income Tax depreciation rates
#[wasm_bindgen]
pub fn get_it_depreciation_rates() -> JsValue {
    let rates = vec![
        ItDepreciationRate {
            asset_class: "Building - Residential".to_string(),
            description: "Residential buildings".to_string(),
            wdv_rate: "5%".to_string(),
            slm_rate: "1.63%".to_string(),
            block: "Block 1".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Building - Non-Residential".to_string(),
            description: "Factory, office buildings".to_string(),
            wdv_rate: "10%".to_string(),
            slm_rate: "3.17%".to_string(),
            block: "Block 2".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Furniture & Fittings".to_string(),
            description: "Furniture, fittings, electrical fittings".to_string(),
            wdv_rate: "10%".to_string(),
            slm_rate: "6.33%".to_string(),
            block: "Block 3".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Plant & Machinery - General".to_string(),
            description: "General plant and machinery".to_string(),
            wdv_rate: "15%".to_string(),
            slm_rate: "4.75%".to_string(),
            block: "Block 4".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Motor Vehicles".to_string(),
            description: "Motor cars, buses, lorries".to_string(),
            wdv_rate: "15%".to_string(),
            slm_rate: "9.50%".to_string(),
            block: "Block 5".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Computers & Software".to_string(),
            description: "Computers, software, data processing".to_string(),
            wdv_rate: "40%".to_string(),
            slm_rate: "16.21%".to_string(),
            block: "Block 6".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Intangible Assets".to_string(),
            description: "Know-how, patents, copyrights, trademarks".to_string(),
            wdv_rate: "25%".to_string(),
            slm_rate: "6.33%".to_string(),
            block: "Block 7".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Ships".to_string(),
            description: "Ocean going ships, vessels".to_string(),
            wdv_rate: "20%".to_string(),
            slm_rate: "6.33%".to_string(),
            block: "Block 8".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Air Pollution Control".to_string(),
            description: "Air, water pollution control equipment".to_string(),
            wdv_rate: "40%".to_string(),
            slm_rate: "16.21%".to_string(),
            block: "Block 9".to_string(),
        },
        ItDepreciationRate {
            asset_class: "Energy Saving Devices".to_string(),
            description: "Energy saving equipment, renewable energy".to_string(),
            wdv_rate: "40%".to_string(),
            slm_rate: "16.21%".to_string(),
            block: "Block 10".to_string(),
        },
    ];

    serde_wasm_bindgen::to_value(&rates).unwrap_or(JsValue::NULL)
}

/// Companies Act depreciation rates
#[wasm_bindgen]
pub fn get_companies_act_rates() -> JsValue {
    let rates = serde_json::json!([
        {
            "asset_class": "Buildings (Factory)",
            "useful_life_years": 30,
            "slm_rate": "3.17%",
            "wdv_rate": "9.50%"
        },
        {
            "asset_class": "Buildings (Other than factory)",
            "useful_life_years": 60,
            "slm_rate": "1.58%",
            "wdv_rate": "4.87%"
        },
        {
            "asset_class": "Plant & Machinery (General)",
            "useful_life_years": 15,
            "slm_rate": "6.33%",
            "wdv_rate": "18.10%"
        },
        {
            "asset_class": "Furniture & Fittings",
            "useful_life_years": 10,
            "slm_rate": "9.50%",
            "wdv_rate": "25.89%"
        },
        {
            "asset_class": "Motor Vehicles",
            "useful_life_years": 8,
            "slm_rate": "11.88%",
            "wdv_rate": "31.23%"
        },
        {
            "asset_class": "Computers",
            "useful_life_years": 3,
            "slm_rate": "31.67%",
            "wdv_rate": "63.16%"
        },
        {
            "asset_class": "Office Equipment",
            "useful_life_years": 5,
            "slm_rate": "19.00%",
            "wdv_rate": "45.07%"
        },
        {
            "asset_class": "Electrical Installations",
            "useful_life_years": 10,
            "slm_rate": "9.50%",
            "wdv_rate": "25.89%"
        },
        {
            "asset_class": "Servers & Networks",
            "useful_life_years": 6,
            "slm_rate": "15.83%",
            "wdv_rate": "39.30%"
        }
    ]);

    serde_wasm_bindgen::to_value(&rates).unwrap_or(JsValue::NULL)
}

/// Calculate depreciation for a partial year
#[wasm_bindgen]
pub fn calculate_partial_year_depreciation(
    cost: &str,
    salvage_value: &str,
    useful_life_years: u32,
    method: &str,
    days_used: u32,
) -> JsValue {
    let cost_val: Decimal = cost.parse().unwrap_or(Decimal::ZERO);
    let salvage: Decimal = salvage_value.parse().unwrap_or(Decimal::ZERO);
    let days_in_year = 365u32;

    let annual_dep = match method.to_lowercase().as_str() {
        "slm" => {
            let depreciable = cost_val - salvage;
            (depreciable / Decimal::from(useful_life_years)).round_dp(2)
        }
        "wdv" => {
            let rate = calculate_wdv_rate(useful_life_years);
            (cost_val * rate).round_dp(2)
        }
        _ => Decimal::ZERO,
    };

    let partial_dep = (annual_dep * Decimal::from(days_used) / Decimal::from(days_in_year)).round_dp(2);

    let result = serde_json::json!({
        "cost": cost_val.to_string(),
        "salvage_value": salvage.to_string(),
        "annual_depreciation": annual_dep.to_string(),
        "days_used": days_used,
        "partial_depreciation": partial_dep.to_string(),
        "method": method,
        "formula": format!("{} × {}/{} = {}", annual_dep, days_used, days_in_year, partial_dep)
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Compare depreciation methods
#[wasm_bindgen]
pub fn compare_depreciation_methods(cost: &str, salvage: &str, useful_life: u32) -> JsValue {
    let cost_val: Decimal = cost.parse().unwrap_or(Decimal::ZERO);
    let salvage_val: Decimal = salvage.parse().unwrap_or(Decimal::ZERO);

    let input = AssetInput {
        asset_name: "Comparison".to_string(),
        cost: cost.to_string(),
        salvage_value: Some(salvage.to_string()),
        useful_life_years: useful_life,
        purchase_date: "2024-01-01".to_string(),
        method: DepreciationMethod::Slm,
        wdv_rate: None,
        total_units: None,
        units_this_period: None,
        fiscal_year_start_month: None,
    };

    // Calculate for each method
    let slm = calculate_slm(cost_val, salvage_val, useful_life, "2024-01-01");
    let wdv = calculate_wdv(cost_val, salvage_val, useful_life, calculate_wdv_rate(useful_life), "2024-01-01");
    let ddb = calculate_ddb(cost_val, salvage_val, useful_life, "2024-01-01");
    let syd = calculate_syd(cost_val, salvage_val, useful_life, "2024-01-01");

    let result = serde_json::json!({
        "cost": cost,
        "salvage_value": salvage,
        "useful_life_years": useful_life,
        "methods": {
            "slm": {
                "first_year_depreciation": slm.0.first().map(|p| p.depreciation.clone()).unwrap_or_default(),
                "annual_rate": slm.1,
                "pattern": "Constant amount each year"
            },
            "wdv": {
                "first_year_depreciation": wdv.0.first().map(|p| p.depreciation.clone()).unwrap_or_default(),
                "annual_rate": wdv.1,
                "pattern": "Decreasing amount each year"
            },
            "double_declining": {
                "first_year_depreciation": ddb.0.first().map(|p| p.depreciation.clone()).unwrap_or_default(),
                "annual_rate": ddb.1,
                "pattern": "Accelerated - faster initial depreciation"
            },
            "sum_of_years_digits": {
                "first_year_depreciation": syd.0.first().map(|p| p.depreciation.clone()).unwrap_or_default(),
                "annual_rate": syd.1,
                "pattern": "Accelerated with systematic reduction"
            }
        },
        "recommendation": if useful_life <= 5 {
            "For short-lived assets, WDV or DDB provides better tax benefits"
        } else {
            "For long-lived assets, SLM provides consistent expense recognition"
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate accumulated depreciation
#[wasm_bindgen]
pub fn calculate_accumulated_depreciation(
    cost: &str,
    salvage: &str,
    useful_life: u32,
    method: &str,
    years_elapsed: u32,
) -> JsValue {
    let cost_val: Decimal = cost.parse().unwrap_or(Decimal::ZERO);
    let salvage_val: Decimal = salvage.parse().unwrap_or(Decimal::ZERO);

    let (periods, _) = match method.to_lowercase().as_str() {
        "slm" => calculate_slm(cost_val, salvage_val, useful_life, "2024-01-01"),
        "wdv" => calculate_wdv(cost_val, salvage_val, useful_life, calculate_wdv_rate(useful_life), "2024-01-01"),
        "ddb" | "double_declining" => calculate_ddb(cost_val, salvage_val, useful_life, "2024-01-01"),
        "syd" | "sum_of_years" => calculate_syd(cost_val, salvage_val, useful_life, "2024-01-01"),
        _ => calculate_slm(cost_val, salvage_val, useful_life, "2024-01-01"),
    };

    let years = years_elapsed.min(useful_life) as usize;
    let accumulated: Decimal = periods.iter()
        .take(years)
        .map(|p| p.depreciation.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    let book_value = cost_val - accumulated;

    let result = serde_json::json!({
        "cost": cost,
        "years_elapsed": years_elapsed,
        "accumulated_depreciation": accumulated.round_dp(2).to_string(),
        "book_value": book_value.round_dp(2).to_string(),
        "remaining_useful_life": useful_life.saturating_sub(years_elapsed),
        "depreciation_by_year": periods.iter().take(years).map(|p| {
            serde_json::json!({
                "year": p.year,
                "depreciation": p.depreciation
            })
        }).collect::<Vec<_>>()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Initialize the depreciation module (called from core init)
fn depreciation_init() {
    console_error_panic_hook::set_once();
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_slm_depreciation() {
        let (periods, rate) = calculate_slm(dec!(100000), dec!(10000), 5, "2024-01-01");

        assert_eq!(periods.len(), 5);
        // Annual depreciation = (100000 - 10000) / 5 = 18000
        assert_eq!(periods[0].depreciation, "18000");
    }

    #[test]
    fn test_wdv_depreciation() {
        let rate = dec!(0.25); // 25%
        let (periods, _) = calculate_wdv(dec!(100000), dec!(10000), 5, rate, "2024-01-01");

        assert_eq!(periods.len(), 5);
        // First year: 100000 * 0.25 = 25000
        assert_eq!(periods[0].depreciation, "25000");
        // Second year: 75000 * 0.25 = 18750
        assert_eq!(periods[1].depreciation, "18750");
    }

    #[test]
    fn test_double_declining() {
        let (periods, _) = calculate_ddb(dec!(100000), dec!(10000), 5, "2024-01-01");

        assert_eq!(periods.len(), 5);
        // Rate = 2/5 = 40%
        // First year: 100000 * 0.4 = 40000
        assert_eq!(periods[0].depreciation, "40000");
    }

    #[test]
    fn test_sum_of_years_digits() {
        let (periods, _) = calculate_syd(dec!(100000), dec!(10000), 5, "2024-01-01");

        assert_eq!(periods.len(), 5);
        // Sum = 1+2+3+4+5 = 15
        // First year: 90000 * 5/15 = 30000
        assert_eq!(periods[0].depreciation, "30000");
    }
}
