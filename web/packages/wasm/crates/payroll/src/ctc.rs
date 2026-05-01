//! CTC (Cost to Company) breakdown and salary structure calculations

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::statutory::{
    calculate_esi_internal, calculate_pf_internal, calculate_professional_tax_internal,
    EsiInput, PfInput, ProfessionalTaxInput,
};

/// CTC breakdown input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CtcInput {
    pub annual_ctc: String,
    pub state: Option<String>,
    // Optional component percentages
    pub basic_percentage: Option<String>,     // Default 40% or 50%
    pub hra_percentage: Option<String>,       // Default 40% or 50% of basic
    pub special_allowance: Option<bool>,      // Include remaining as special allowance
    // Optional fixed amounts
    pub fixed_basic: Option<String>,
    pub fixed_hra: Option<String>,
    pub fixed_lta: Option<String>,
    pub fixed_medical: Option<String>,
    pub fixed_conveyance: Option<String>,
    pub fixed_food_allowance: Option<String>,
    pub fixed_other_allowance: Option<String>,
    // Employer contributions (if custom)
    pub employer_pf_contribution: Option<String>,
    pub include_employer_pf_in_ctc: Option<bool>,
    pub include_employer_esi_in_ctc: Option<bool>,
    pub include_gratuity_in_ctc: Option<bool>,
    // Variable pay
    pub variable_pay_percentage: Option<String>,
    pub bonus_percentage: Option<String>,
}

/// Salary component
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SalaryComponent {
    pub name: String,
    pub monthly: String,
    pub annual: String,
    pub component_type: String, // "earning", "deduction", "employer_contribution"
}

/// CTC breakdown result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CtcBreakdownResult {
    // CTC
    pub annual_ctc: String,
    pub monthly_ctc: String,
    // Gross salary
    pub annual_gross: String,
    pub monthly_gross: String,
    // Components
    pub earnings: Vec<SalaryComponent>,
    pub deductions: Vec<SalaryComponent>,
    pub employer_contributions: Vec<SalaryComponent>,
    // Totals
    pub total_earnings: String,
    pub total_deductions: String,
    pub total_employer_contributions: String,
    // Net salary
    pub annual_net: String,
    pub monthly_net: String,
    // In-hand
    pub annual_in_hand: String,
    pub monthly_in_hand: String,
    // Effective rates
    pub effective_tax_rate: String,
    pub take_home_percentage: String,
}

/// Salary structure input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SalaryStructureInput {
    pub basic: String,
    pub hra: Option<String>,
    pub conveyance: Option<String>,
    pub medical: Option<String>,
    pub lta: Option<String>,
    pub special_allowance: Option<String>,
    pub food_allowance: Option<String>,
    pub other_allowances: Option<String>,
    pub variable_pay: Option<String>,
    pub bonus: Option<String>,
    pub state: Option<String>,
}

/// Calculate CTC breakdown from annual CTC
#[wasm_bindgen]
pub fn calculate_ctc_breakdown(input: JsValue) -> JsValue {
    let input: CtcInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid CTC input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_ctc_breakdown_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal CTC breakdown calculation
fn calculate_ctc_breakdown_internal(input: &CtcInput) -> CtcBreakdownResult {
    let annual_ctc: Decimal = input.annual_ctc.parse().unwrap_or(Decimal::ZERO);
    let monthly_ctc = annual_ctc / dec!(12);
    let state = input.state.as_deref().unwrap_or("Maharashtra");

    let include_employer_pf = input.include_employer_pf_in_ctc.unwrap_or(true);
    let include_employer_esi = input.include_employer_esi_in_ctc.unwrap_or(true);
    let include_gratuity = input.include_gratuity_in_ctc.unwrap_or(true);

    // Calculate components
    let mut earnings: Vec<SalaryComponent> = Vec::new();
    let mut deductions: Vec<SalaryComponent> = Vec::new();
    let mut employer_contributions: Vec<SalaryComponent> = Vec::new();

    // Step 1: Determine basic salary
    let basic_percentage: Decimal = input.basic_percentage.as_ref()
        .and_then(|p| p.parse().ok())
        .unwrap_or(dec!(0.40)); // Default 40%

    let annual_basic = if let Some(ref fixed) = input.fixed_basic {
        fixed.parse().unwrap_or(annual_ctc * basic_percentage)
    } else {
        (annual_ctc * basic_percentage).round_dp(0)
    };
    let monthly_basic = (annual_basic / dec!(12)).round_dp(0);

    earnings.push(SalaryComponent {
        name: "Basic Salary".to_string(),
        monthly: monthly_basic.to_string(),
        annual: annual_basic.round_dp(0).to_string(),
        component_type: "earning".to_string(),
    });

    // Step 2: HRA (typically 40-50% of basic)
    let hra_percentage: Decimal = input.hra_percentage.as_ref()
        .and_then(|p| p.parse().ok())
        .unwrap_or(dec!(0.50)); // Default 50% of basic

    let annual_hra = if let Some(ref fixed) = input.fixed_hra {
        fixed.parse().unwrap_or(annual_basic * hra_percentage)
    } else {
        (annual_basic * hra_percentage).round_dp(0)
    };
    let monthly_hra = (annual_hra / dec!(12)).round_dp(0);

    earnings.push(SalaryComponent {
        name: "House Rent Allowance".to_string(),
        monthly: monthly_hra.to_string(),
        annual: annual_hra.round_dp(0).to_string(),
        component_type: "earning".to_string(),
    });

    // Step 3: LTA (Leave Travel Allowance)
    let annual_lta = if let Some(ref fixed) = input.fixed_lta {
        fixed.parse().unwrap_or(Decimal::ZERO)
    } else {
        // Default: ~4% of basic or a fixed amount
        (annual_basic * dec!(0.04)).round_dp(0).min(dec!(50000))
    };
    let monthly_lta = (annual_lta / dec!(12)).round_dp(0);

    if annual_lta > Decimal::ZERO {
        earnings.push(SalaryComponent {
            name: "Leave Travel Allowance".to_string(),
            monthly: monthly_lta.to_string(),
            annual: annual_lta.round_dp(0).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Step 4: Medical Allowance
    let annual_medical = if let Some(ref fixed) = input.fixed_medical {
        fixed.parse().unwrap_or(Decimal::ZERO)
    } else {
        dec!(15000) // Standard Rs 15,000 per year
    };
    let monthly_medical = (annual_medical / dec!(12)).round_dp(0);

    if annual_medical > Decimal::ZERO {
        earnings.push(SalaryComponent {
            name: "Medical Allowance".to_string(),
            monthly: monthly_medical.to_string(),
            annual: annual_medical.round_dp(0).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Step 5: Conveyance Allowance
    let annual_conveyance = if let Some(ref fixed) = input.fixed_conveyance {
        fixed.parse().unwrap_or(Decimal::ZERO)
    } else {
        dec!(19200) // Standard Rs 1,600 per month
    };
    let monthly_conveyance = (annual_conveyance / dec!(12)).round_dp(0);

    if annual_conveyance > Decimal::ZERO {
        earnings.push(SalaryComponent {
            name: "Conveyance Allowance".to_string(),
            monthly: monthly_conveyance.to_string(),
            annual: annual_conveyance.round_dp(0).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Step 6: Food Allowance
    let annual_food = if let Some(ref fixed) = input.fixed_food_allowance {
        fixed.parse().unwrap_or(Decimal::ZERO)
    } else {
        Decimal::ZERO
    };
    let monthly_food = (annual_food / dec!(12)).round_dp(0);

    if annual_food > Decimal::ZERO {
        earnings.push(SalaryComponent {
            name: "Food Allowance".to_string(),
            monthly: monthly_food.to_string(),
            annual: annual_food.round_dp(0).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Step 7: Variable Pay / Performance Bonus
    let variable_percentage: Decimal = input.variable_pay_percentage.as_ref()
        .and_then(|p| p.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let annual_variable = (annual_ctc * variable_percentage).round_dp(0);
    let monthly_variable = (annual_variable / dec!(12)).round_dp(0);

    if annual_variable > Decimal::ZERO {
        earnings.push(SalaryComponent {
            name: "Variable Pay".to_string(),
            monthly: monthly_variable.to_string(),
            annual: annual_variable.round_dp(0).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Step 8: Calculate employer contributions (to subtract from CTC for gross)
    let mut total_employer_contrib = Decimal::ZERO;

    // Employer PF (12% of basic, capped at 15000)
    let pf_wages = monthly_basic.min(dec!(15000));
    let employer_pf_monthly = (pf_wages * dec!(0.12)).round_dp(0);
    let employer_pf_annual = (employer_pf_monthly * dec!(12)).round_dp(0);

    if include_employer_pf {
        total_employer_contrib += employer_pf_annual;
        employer_contributions.push(SalaryComponent {
            name: "Employer PF Contribution".to_string(),
            monthly: employer_pf_monthly.to_string(),
            annual: employer_pf_annual.to_string(),
            component_type: "employer_contribution".to_string(),
        });
    }

    // Gratuity (4.81% of basic) - if included in CTC
    let gratuity_rate = dec!(0.0481); // 15/26 * 15/12 ≈ 4.81%
    let employer_gratuity_annual = (annual_basic * gratuity_rate).round_dp(0);
    let employer_gratuity_monthly = (employer_gratuity_annual / dec!(12)).round_dp(0);

    if include_gratuity {
        total_employer_contrib += employer_gratuity_annual;
        employer_contributions.push(SalaryComponent {
            name: "Gratuity Provision".to_string(),
            monthly: employer_gratuity_monthly.to_string(),
            annual: employer_gratuity_annual.to_string(),
            component_type: "employer_contribution".to_string(),
        });
    }

    // Calculate gross salary (before ESI check)
    let gross_before_esi = annual_ctc - total_employer_contrib;
    let monthly_gross_before_esi = (gross_before_esi / dec!(12)).round_dp(0);

    // Employer ESI (3.25% of gross) - only if gross <= 21000
    let esi_applicable = monthly_gross_before_esi <= dec!(21000);
    let employer_esi_monthly = if esi_applicable {
        (monthly_gross_before_esi * dec!(0.0325)).round_dp(0)
    } else {
        Decimal::ZERO
    };
    let employer_esi_annual = (employer_esi_monthly * dec!(12)).round_dp(0);

    if include_employer_esi && esi_applicable {
        total_employer_contrib += employer_esi_annual;
        employer_contributions.push(SalaryComponent {
            name: "Employer ESI Contribution".to_string(),
            monthly: employer_esi_monthly.to_string(),
            annual: employer_esi_annual.to_string(),
            component_type: "employer_contribution".to_string(),
        });
    }

    // Calculate final gross salary
    let annual_gross = annual_ctc - total_employer_contrib;
    let monthly_gross = (annual_gross / dec!(12)).round_dp(0);

    // Calculate Special Allowance (remaining amount)
    let total_fixed_earnings = annual_basic + annual_hra + annual_lta + annual_medical +
        annual_conveyance + annual_food + annual_variable;

    let annual_special = (annual_gross - total_fixed_earnings).max(Decimal::ZERO);
    let monthly_special = (annual_special / dec!(12)).round_dp(0);

    if annual_special > Decimal::ZERO {
        earnings.push(SalaryComponent {
            name: "Special Allowance".to_string(),
            monthly: monthly_special.to_string(),
            annual: annual_special.round_dp(0).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Calculate deductions
    // Employee PF
    let employee_pf_monthly = (pf_wages * dec!(0.12)).round_dp(0);
    let employee_pf_annual = (employee_pf_monthly * dec!(12)).round_dp(0);

    deductions.push(SalaryComponent {
        name: "Employee PF Contribution".to_string(),
        monthly: employee_pf_monthly.to_string(),
        annual: employee_pf_annual.to_string(),
        component_type: "deduction".to_string(),
    });

    // Employee ESI
    let employee_esi_monthly = if esi_applicable {
        (monthly_gross * dec!(0.0075)).round_dp(0)
    } else {
        Decimal::ZERO
    };
    let employee_esi_annual = (employee_esi_monthly * dec!(12)).round_dp(0);

    if esi_applicable {
        deductions.push(SalaryComponent {
            name: "Employee ESI Contribution".to_string(),
            monthly: employee_esi_monthly.to_string(),
            annual: employee_esi_annual.to_string(),
            component_type: "deduction".to_string(),
        });
    }

    // Professional Tax
    let pt_input = ProfessionalTaxInput {
        gross_salary: monthly_gross.to_string(),
        state: state.to_string(),
    };
    let pt_result = calculate_professional_tax_internal(&pt_input);
    let pt_monthly: Decimal = pt_result.monthly_tax.parse().unwrap_or(Decimal::ZERO);
    let pt_annual = (pt_monthly * dec!(12)).min(dec!(2500)); // Capped at 2500

    if pt_monthly > Decimal::ZERO {
        deductions.push(SalaryComponent {
            name: "Professional Tax".to_string(),
            monthly: pt_monthly.to_string(),
            annual: pt_annual.to_string(),
            component_type: "deduction".to_string(),
        });
    }

    // Calculate totals
    let total_earnings: Decimal = earnings.iter()
        .map(|e| e.annual.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    let total_deductions: Decimal = deductions.iter()
        .map(|d| d.annual.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    let total_employer: Decimal = employer_contributions.iter()
        .map(|c| c.annual.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    // Net salary (before tax)
    let annual_net = (annual_gross - total_deductions).max(Decimal::ZERO);
    let monthly_net = (annual_net / dec!(12)).round_dp(0);

    // In-hand (assuming no income tax for simplicity - actual would need tax calculation)
    let annual_in_hand = annual_net;
    let monthly_in_hand = monthly_net;

    // Calculate percentages
    let take_home_pct = if annual_ctc > Decimal::ZERO {
        ((annual_in_hand / annual_ctc) * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    CtcBreakdownResult {
        annual_ctc: annual_ctc.round_dp(0).to_string(),
        monthly_ctc: monthly_ctc.round_dp(0).to_string(),
        annual_gross: annual_gross.round_dp(0).to_string(),
        monthly_gross: monthly_gross.round_dp(0).to_string(),
        earnings,
        deductions,
        employer_contributions,
        total_earnings: total_earnings.round_dp(0).to_string(),
        total_deductions: total_deductions.round_dp(0).to_string(),
        total_employer_contributions: total_employer.round_dp(0).to_string(),
        annual_net: annual_net.round_dp(0).to_string(),
        monthly_net: monthly_net.round_dp(0).to_string(),
        annual_in_hand: annual_in_hand.round_dp(0).to_string(),
        monthly_in_hand: monthly_in_hand.round_dp(0).to_string(),
        effective_tax_rate: "0%".to_string(), // Would need tax calculation
        take_home_percentage: format!("{}%", take_home_pct),
    }
}

/// Calculate salary structure from individual components
#[wasm_bindgen]
pub fn calculate_salary_structure(input: JsValue) -> JsValue {
    let input: SalaryStructureInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid salary structure input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_salary_structure_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Salary structure result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SalaryStructureResult {
    pub monthly_gross: String,
    pub annual_gross: String,
    pub components: Vec<SalaryComponent>,
    pub deductions: Vec<SalaryComponent>,
    pub monthly_net: String,
    pub annual_net: String,
    pub estimated_annual_ctc: String,
}

/// Internal salary structure calculation
fn calculate_salary_structure_internal(input: &SalaryStructureInput) -> SalaryStructureResult {
    let basic: Decimal = input.basic.parse().unwrap_or(Decimal::ZERO);
    let hra: Decimal = input.hra.as_ref().and_then(|h| h.parse().ok()).unwrap_or(Decimal::ZERO);
    let conveyance: Decimal = input.conveyance.as_ref().and_then(|c| c.parse().ok()).unwrap_or(Decimal::ZERO);
    let medical: Decimal = input.medical.as_ref().and_then(|m| m.parse().ok()).unwrap_or(Decimal::ZERO);
    let lta: Decimal = input.lta.as_ref().and_then(|l| l.parse().ok()).unwrap_or(Decimal::ZERO);
    let special: Decimal = input.special_allowance.as_ref().and_then(|s| s.parse().ok()).unwrap_or(Decimal::ZERO);
    let food: Decimal = input.food_allowance.as_ref().and_then(|f| f.parse().ok()).unwrap_or(Decimal::ZERO);
    let other: Decimal = input.other_allowances.as_ref().and_then(|o| o.parse().ok()).unwrap_or(Decimal::ZERO);
    let variable: Decimal = input.variable_pay.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
    let bonus: Decimal = input.bonus.as_ref().and_then(|b| b.parse().ok()).unwrap_or(Decimal::ZERO);
    let state = input.state.as_deref().unwrap_or("Maharashtra");

    let monthly_gross = basic + hra + conveyance + medical + lta + special + food + other + variable + bonus;
    let annual_gross = monthly_gross * dec!(12);

    let mut components = Vec::new();

    if basic > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Basic Salary".to_string(),
            monthly: basic.to_string(),
            annual: (basic * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if hra > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "House Rent Allowance".to_string(),
            monthly: hra.to_string(),
            annual: (hra * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if conveyance > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Conveyance Allowance".to_string(),
            monthly: conveyance.to_string(),
            annual: (conveyance * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if medical > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Medical Allowance".to_string(),
            monthly: medical.to_string(),
            annual: (medical * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if lta > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Leave Travel Allowance".to_string(),
            monthly: lta.to_string(),
            annual: (lta * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if special > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Special Allowance".to_string(),
            monthly: special.to_string(),
            annual: (special * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if food > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Food Allowance".to_string(),
            monthly: food.to_string(),
            annual: (food * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if other > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Other Allowances".to_string(),
            monthly: other.to_string(),
            annual: (other * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if variable > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Variable Pay".to_string(),
            monthly: variable.to_string(),
            annual: (variable * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    if bonus > Decimal::ZERO {
        components.push(SalaryComponent {
            name: "Bonus".to_string(),
            monthly: bonus.to_string(),
            annual: (bonus * dec!(12)).to_string(),
            component_type: "earning".to_string(),
        });
    }

    // Calculate deductions
    let mut deductions = Vec::new();

    // PF
    let pf_wages = basic.min(dec!(15000));
    let employee_pf = (pf_wages * dec!(0.12)).round_dp(0);
    deductions.push(SalaryComponent {
        name: "Employee PF".to_string(),
        monthly: employee_pf.to_string(),
        annual: (employee_pf * dec!(12)).to_string(),
        component_type: "deduction".to_string(),
    });

    // ESI
    let esi_applicable = monthly_gross <= dec!(21000);
    let employee_esi = if esi_applicable {
        (monthly_gross * dec!(0.0075)).round_dp(0)
    } else {
        Decimal::ZERO
    };

    if esi_applicable {
        deductions.push(SalaryComponent {
            name: "Employee ESI".to_string(),
            monthly: employee_esi.to_string(),
            annual: (employee_esi * dec!(12)).to_string(),
            component_type: "deduction".to_string(),
        });
    }

    // Professional Tax
    let pt_input = ProfessionalTaxInput {
        gross_salary: monthly_gross.to_string(),
        state: state.to_string(),
    };
    let pt_result = calculate_professional_tax_internal(&pt_input);
    let pt_monthly: Decimal = pt_result.monthly_tax.parse().unwrap_or(Decimal::ZERO);

    if pt_monthly > Decimal::ZERO {
        deductions.push(SalaryComponent {
            name: "Professional Tax".to_string(),
            monthly: pt_monthly.to_string(),
            annual: (pt_monthly * dec!(12)).min(dec!(2500)).to_string(),
            component_type: "deduction".to_string(),
        });
    }

    let total_deductions = employee_pf + employee_esi + pt_monthly;
    let monthly_net = monthly_gross - total_deductions;
    let annual_net = monthly_net * dec!(12);

    // Estimate CTC
    let employer_pf = (pf_wages * dec!(0.12)).round_dp(0);
    let employer_esi = if esi_applicable {
        (monthly_gross * dec!(0.0325)).round_dp(0)
    } else {
        Decimal::ZERO
    };
    let gratuity = (basic * dec!(0.0481)).round_dp(0);

    let monthly_ctc = monthly_gross + employer_pf + employer_esi + gratuity;
    let annual_ctc = monthly_ctc * dec!(12);

    SalaryStructureResult {
        monthly_gross: monthly_gross.round_dp(0).to_string(),
        annual_gross: annual_gross.round_dp(0).to_string(),
        components,
        deductions,
        monthly_net: monthly_net.round_dp(0).to_string(),
        annual_net: annual_net.round_dp(0).to_string(),
        estimated_annual_ctc: annual_ctc.round_dp(0).to_string(),
    }
}

/// Generate optimal salary structure for tax efficiency
#[wasm_bindgen]
pub fn optimize_salary_structure(
    annual_ctc: &str,
    state: &str,
    metro_city: bool,
) -> JsValue {
    let ctc: Decimal = annual_ctc.parse().unwrap_or(Decimal::ZERO);

    // Optimal structure for tax efficiency
    // Basic: 40% (lower basic = lower PF, but higher basic = higher HRA)
    let basic_percentage = if metro_city { dec!(0.40) } else { dec!(0.50) };
    let annual_basic = (ctc * basic_percentage).round_dp(0);
    let monthly_basic = (annual_basic / dec!(12)).round_dp(0);

    // HRA: 50% of basic for metro, 40% for non-metro
    let hra_percentage = if metro_city { dec!(0.50) } else { dec!(0.40) };
    let annual_hra = (annual_basic * hra_percentage).round_dp(0);

    // Standard allowances
    let annual_lta = dec!(50000).min(ctc * dec!(0.05));
    let annual_medical = dec!(15000);
    let annual_conveyance = dec!(19200);
    let annual_food = dec!(26400); // Rs 2200/month (tax free under 50/meal)

    // Employer contributions
    let pf_wages = monthly_basic.min(dec!(15000));
    let annual_employer_pf = (pf_wages * dec!(0.12) * dec!(12)).round_dp(0);
    let annual_gratuity = (annual_basic * dec!(0.0481)).round_dp(0);

    // ESI check
    let remaining = ctc - annual_basic - annual_hra - annual_lta - annual_medical -
        annual_conveyance - annual_food - annual_employer_pf - annual_gratuity;
    let monthly_gross_estimate = ((ctc - annual_employer_pf - annual_gratuity) / dec!(12)).round_dp(0);
    let esi_applicable = monthly_gross_estimate <= dec!(21000);
    let annual_employer_esi = if esi_applicable {
        (monthly_gross_estimate * dec!(0.0325) * dec!(12)).round_dp(0)
    } else {
        Decimal::ZERO
    };

    // Special allowance (remaining)
    let annual_special = (remaining - annual_employer_esi).max(Decimal::ZERO);

    let result = serde_json::json!({
        "recommended": {
            "basic": {
                "monthly": (annual_basic / dec!(12)).round_dp(0).to_string(),
                "annual": annual_basic.to_string(),
                "percentage": format!("{}%", (basic_percentage * dec!(100)).round_dp(0))
            },
            "hra": {
                "monthly": (annual_hra / dec!(12)).round_dp(0).to_string(),
                "annual": annual_hra.to_string(),
                "percentage": format!("{}% of basic", (hra_percentage * dec!(100)).round_dp(0))
            },
            "lta": {
                "monthly": (annual_lta / dec!(12)).round_dp(0).to_string(),
                "annual": annual_lta.round_dp(0).to_string()
            },
            "medical": {
                "monthly": (annual_medical / dec!(12)).round_dp(0).to_string(),
                "annual": annual_medical.to_string()
            },
            "conveyance": {
                "monthly": (annual_conveyance / dec!(12)).round_dp(0).to_string(),
                "annual": annual_conveyance.to_string()
            },
            "food_allowance": {
                "monthly": (annual_food / dec!(12)).round_dp(0).to_string(),
                "annual": annual_food.to_string()
            },
            "special_allowance": {
                "monthly": (annual_special / dec!(12)).round_dp(0).to_string(),
                "annual": annual_special.round_dp(0).to_string()
            }
        },
        "employer_contributions": {
            "pf": annual_employer_pf.to_string(),
            "esi": annual_employer_esi.to_string(),
            "gratuity": annual_gratuity.to_string()
        },
        "tax_saving_tips": [
            "Maximize HRA exemption by paying rent",
            "Claim LTA for actual travel",
            "Use food coupons/meal cards (tax free up to Rs 50/meal)",
            "Invest in NPS for additional Rs 50,000 deduction",
            "Claim medical insurance premium under 80D"
        ],
        "metro_city": metro_city,
        "state": state
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate reverse CTC from take-home salary
#[wasm_bindgen]
pub fn reverse_ctc_calculation(
    monthly_in_hand: &str,
    state: &str,
    include_variable: bool,
) -> JsValue {
    let in_hand: Decimal = monthly_in_hand.parse().unwrap_or(Decimal::ZERO);

    // Estimate gross from in-hand (iterative approach)
    // Assume ~80-85% take-home for estimation
    let mut estimated_gross = in_hand / dec!(0.82);

    // Iterate to converge
    for _ in 0..5 {
        let pf_wages = (estimated_gross * dec!(0.40)).min(dec!(15000));
        let employee_pf = pf_wages * dec!(0.12);
        let esi = if estimated_gross <= dec!(21000) {
            estimated_gross * dec!(0.0075)
        } else {
            Decimal::ZERO
        };
        let pt = dec!(200); // Approximate

        let total_deductions = employee_pf + esi + pt;
        estimated_gross = in_hand + total_deductions;
    }

    let annual_gross = estimated_gross * dec!(12);

    // Calculate CTC
    let basic = estimated_gross * dec!(0.40);
    let pf_wages = basic.min(dec!(15000));
    let employer_pf = pf_wages * dec!(0.12);
    let employer_esi = if estimated_gross <= dec!(21000) {
        estimated_gross * dec!(0.0325)
    } else {
        Decimal::ZERO
    };
    let gratuity = basic * dec!(0.0481);

    let monthly_ctc = estimated_gross + employer_pf + employer_esi + gratuity;
    let annual_ctc = monthly_ctc * dec!(12);

    let result = serde_json::json!({
        "monthly_in_hand": in_hand.round_dp(0).to_string(),
        "estimated_monthly_gross": estimated_gross.round_dp(0).to_string(),
        "estimated_annual_gross": annual_gross.round_dp(0).to_string(),
        "estimated_monthly_ctc": monthly_ctc.round_dp(0).to_string(),
        "estimated_annual_ctc": annual_ctc.round_dp(0).to_string(),
        "note": "This is an approximation. Actual CTC may vary based on company policy."
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ctc_breakdown() {
        let input = CtcInput {
            annual_ctc: "1200000".to_string(),
            state: Some("Maharashtra".to_string()),
            basic_percentage: None,
            hra_percentage: None,
            special_allowance: None,
            fixed_basic: None,
            fixed_hra: None,
            fixed_lta: None,
            fixed_medical: None,
            fixed_conveyance: None,
            fixed_food_allowance: None,
            fixed_other_allowance: None,
            employer_pf_contribution: None,
            include_employer_pf_in_ctc: Some(true),
            include_employer_esi_in_ctc: Some(true),
            include_gratuity_in_ctc: Some(true),
            variable_pay_percentage: None,
            bonus_percentage: None,
        };

        let result = calculate_ctc_breakdown_internal(&input);
        assert_eq!(result.annual_ctc, "1200000");
        assert!(!result.earnings.is_empty());
    }

    #[test]
    fn test_salary_structure() {
        let input = SalaryStructureInput {
            basic: "40000".to_string(),
            hra: Some("20000".to_string()),
            conveyance: Some("1600".to_string()),
            medical: Some("1250".to_string()),
            lta: Some("4000".to_string()),
            special_allowance: Some("20000".to_string()),
            food_allowance: None,
            other_allowances: None,
            variable_pay: None,
            bonus: None,
            state: Some("Maharashtra".to_string()),
        };

        let result = calculate_salary_structure_internal(&input);
        assert_eq!(result.monthly_gross, "86850");
    }
}
