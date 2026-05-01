//! Statutory deductions - PF, ESI, Professional Tax, LWF

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// PF calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PfInput {
    pub basic_salary: String,
    pub da: Option<String>,
    pub is_international_worker: Option<bool>,
    pub is_excluded_establishment: Option<bool>,
    pub employer_contribution_on_higher: Option<bool>,
}

/// PF calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PfResult {
    pub pf_wages: String,
    pub employee_pf: String,
    pub employer_pf: String,
    pub employer_eps: String,
    pub employer_pf_admin: String,
    pub employer_edli: String,
    pub employer_edli_admin: String,
    pub total_employee_contribution: String,
    pub total_employer_contribution: String,
    pub total_contribution: String,
}

/// ESI calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EsiInput {
    pub gross_salary: String,
    pub state: Option<String>,
}

/// ESI calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EsiResult {
    pub gross_wages: String,
    pub is_applicable: bool,
    pub employee_contribution: String,
    pub employer_contribution: String,
    pub total_contribution: String,
    pub employee_rate: String,
    pub employer_rate: String,
}

/// Professional Tax input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProfessionalTaxInput {
    pub gross_salary: String,
    pub state: String,
}

/// Professional Tax result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProfessionalTaxResult {
    pub gross_salary: String,
    pub monthly_tax: String,
    pub annual_tax: String,
    pub state: String,
    pub slab_description: String,
}

/// LWF (Labour Welfare Fund) input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LwfInput {
    pub gross_salary: String,
    pub state: String,
}

/// LWF result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LwfResult {
    pub employee_contribution: String,
    pub employer_contribution: String,
    pub total_contribution: String,
    pub frequency: String,
    pub state: String,
}

/// All statutory deductions combined
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StatutoryDeductionsResult {
    pub pf: Option<PfResult>,
    pub esi: Option<EsiResult>,
    pub professional_tax: Option<ProfessionalTaxResult>,
    pub lwf: Option<LwfResult>,
    pub total_employee_deductions: String,
    pub total_employer_contributions: String,
}

// Constants
const PF_WAGE_CEILING: u64 = 15000;
const PF_EMPLOYEE_RATE: &str = "0.12";
const PF_EMPLOYER_RATE: &str = "0.0367"; // 3.67% to PF account
const EPS_RATE: &str = "0.0833"; // 8.33% to EPS
const PF_ADMIN_RATE: &str = "0.005"; // 0.5% admin charges
const EDLI_RATE: &str = "0.005"; // 0.5% EDLI
const EDLI_ADMIN_RATE: &str = "0.0005"; // 0.05% EDLI admin (removed from 01-04-2017, but kept for historical)

const ESI_WAGE_CEILING: u64 = 21000;
const ESI_EMPLOYEE_RATE: &str = "0.0075"; // 0.75%
const ESI_EMPLOYER_RATE: &str = "0.0325"; // 3.25%

/// Calculate PF contributions
#[wasm_bindgen]
pub fn calculate_pf(input: JsValue) -> JsValue {
    let input: PfInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid PF input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_pf_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal PF calculation
pub fn calculate_pf_internal(input: &PfInput) -> PfResult {
    let basic: Decimal = input.basic_salary.parse().unwrap_or(Decimal::ZERO);
    let da: Decimal = input.da.as_ref()
        .and_then(|d| d.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let pf_wages_uncapped = basic + da;
    let is_higher_contribution = input.employer_contribution_on_higher.unwrap_or(false);

    // PF wage ceiling is Rs 15,000 unless employer opts for higher contribution
    let pf_wages = if is_higher_contribution {
        pf_wages_uncapped
    } else {
        pf_wages_uncapped.min(Decimal::from(PF_WAGE_CEILING))
    };

    // Employee PF: 12% of PF wages
    let employee_pf_rate: Decimal = PF_EMPLOYEE_RATE.parse().unwrap();
    let employee_pf = (pf_wages * employee_pf_rate).round_dp(0);

    // Employer PF: 3.67% to PF account
    let employer_pf_rate: Decimal = PF_EMPLOYER_RATE.parse().unwrap();
    let employer_pf = (pf_wages * employer_pf_rate).round_dp(0);

    // Employer EPS: 8.33% to Pension Scheme (capped at 15000)
    let eps_wages = pf_wages.min(Decimal::from(PF_WAGE_CEILING));
    let eps_rate: Decimal = EPS_RATE.parse().unwrap();
    let employer_eps = (eps_wages * eps_rate).round_dp(0);

    // Admin charges: 0.5% of PF wages
    let admin_rate: Decimal = PF_ADMIN_RATE.parse().unwrap();
    let employer_pf_admin = (pf_wages * admin_rate).round_dp(0);

    // EDLI: 0.5% of PF wages (capped at 15000)
    let edli_wages = pf_wages.min(Decimal::from(PF_WAGE_CEILING));
    let edli_rate: Decimal = EDLI_RATE.parse().unwrap();
    let employer_edli = (edli_wages * edli_rate).round_dp(0);

    // EDLI Admin: Currently 0 (discontinued)
    let employer_edli_admin = Decimal::ZERO;

    let total_employee = employee_pf;
    let total_employer = employer_pf + employer_eps + employer_pf_admin + employer_edli + employer_edli_admin;

    PfResult {
        pf_wages: pf_wages.round_dp(0).to_string(),
        employee_pf: employee_pf.to_string(),
        employer_pf: employer_pf.to_string(),
        employer_eps: employer_eps.to_string(),
        employer_pf_admin: employer_pf_admin.to_string(),
        employer_edli: employer_edli.to_string(),
        employer_edli_admin: employer_edli_admin.to_string(),
        total_employee_contribution: total_employee.to_string(),
        total_employer_contribution: total_employer.round_dp(0).to_string(),
        total_contribution: (total_employee + total_employer).round_dp(0).to_string(),
    }
}

/// Calculate ESI contributions
#[wasm_bindgen]
pub fn calculate_esi(input: JsValue) -> JsValue {
    let input: EsiInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid ESI input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_esi_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal ESI calculation
pub fn calculate_esi_internal(input: &EsiInput) -> EsiResult {
    let gross: Decimal = input.gross_salary.parse().unwrap_or(Decimal::ZERO);
    let gross_u64 = gross.to_u64().unwrap_or(0);

    // ESI is applicable only if gross salary <= 21000
    let is_applicable = gross_u64 <= ESI_WAGE_CEILING;

    if !is_applicable {
        return EsiResult {
            gross_wages: gross.round_dp(0).to_string(),
            is_applicable: false,
            employee_contribution: "0".to_string(),
            employer_contribution: "0".to_string(),
            total_contribution: "0".to_string(),
            employee_rate: "0.75%".to_string(),
            employer_rate: "3.25%".to_string(),
        };
    }

    let employee_rate: Decimal = ESI_EMPLOYEE_RATE.parse().unwrap();
    let employer_rate: Decimal = ESI_EMPLOYER_RATE.parse().unwrap();

    let employee_contribution = (gross * employee_rate).round_dp(0);
    let employer_contribution = (gross * employer_rate).round_dp(0);

    EsiResult {
        gross_wages: gross.round_dp(0).to_string(),
        is_applicable: true,
        employee_contribution: employee_contribution.to_string(),
        employer_contribution: employer_contribution.to_string(),
        total_contribution: (employee_contribution + employer_contribution).to_string(),
        employee_rate: "0.75%".to_string(),
        employer_rate: "3.25%".to_string(),
    }
}

/// Professional Tax slabs by state
/// Returns (monthly_tax, slab_description)
fn get_professional_tax(gross: Decimal, state: &str) -> (Decimal, String) {
    let gross_monthly = gross.to_u64().unwrap_or(0);

    match state.to_uppercase().as_str() {
        "MAHARASHTRA" | "MH" => {
            if gross_monthly <= 7500 {
                (Decimal::ZERO, "Up to Rs 7,500 - Nil".to_string())
            } else if gross_monthly <= 10000 {
                (dec!(175), "Rs 7,501 to Rs 10,000 - Rs 175".to_string())
            } else {
                // Rs 200 for 11 months, Rs 300 for Feb
                (dec!(200), "Above Rs 10,000 - Rs 200/month (Rs 300 in Feb)".to_string())
            }
        }
        "KARNATAKA" | "KA" => {
            if gross_monthly <= 15000 {
                (Decimal::ZERO, "Up to Rs 15,000 - Nil".to_string())
            } else {
                (dec!(200), "Above Rs 15,000 - Rs 200".to_string())
            }
        }
        "WEST BENGAL" | "WB" => {
            if gross_monthly <= 10000 {
                (Decimal::ZERO, "Up to Rs 10,000 - Nil".to_string())
            } else if gross_monthly <= 15000 {
                (dec!(110), "Rs 10,001 to Rs 15,000 - Rs 110".to_string())
            } else if gross_monthly <= 25000 {
                (dec!(130), "Rs 15,001 to Rs 25,000 - Rs 130".to_string())
            } else if gross_monthly <= 40000 {
                (dec!(150), "Rs 25,001 to Rs 40,000 - Rs 150".to_string())
            } else {
                (dec!(200), "Above Rs 40,000 - Rs 200".to_string())
            }
        }
        "ANDHRA PRADESH" | "AP" => {
            if gross_monthly <= 15000 {
                (Decimal::ZERO, "Up to Rs 15,000 - Nil".to_string())
            } else if gross_monthly <= 20000 {
                (dec!(150), "Rs 15,001 to Rs 20,000 - Rs 150".to_string())
            } else {
                (dec!(200), "Above Rs 20,000 - Rs 200".to_string())
            }
        }
        "TELANGANA" | "TG" | "TS" => {
            if gross_monthly <= 15000 {
                (Decimal::ZERO, "Up to Rs 15,000 - Nil".to_string())
            } else if gross_monthly <= 20000 {
                (dec!(150), "Rs 15,001 to Rs 20,000 - Rs 150".to_string())
            } else {
                (dec!(200), "Above Rs 20,000 - Rs 200".to_string())
            }
        }
        "TAMIL NADU" | "TN" => {
            if gross_monthly <= 21000 {
                (Decimal::ZERO, "Up to Rs 21,000 - Nil".to_string())
            } else if gross_monthly <= 30000 {
                (dec!(135), "Rs 21,001 to Rs 30,000 - Rs 135".to_string())
            } else if gross_monthly <= 45000 {
                (dec!(315), "Rs 30,001 to Rs 45,000 - Rs 315".to_string())
            } else if gross_monthly <= 60000 {
                (dec!(690), "Rs 45,001 to Rs 60,000 - Rs 690".to_string())
            } else if gross_monthly <= 75000 {
                (dec!(1025), "Rs 60,001 to Rs 75,000 - Rs 1,025".to_string())
            } else {
                (dec!(1250), "Above Rs 75,000 - Rs 1,250".to_string())
            }
        }
        "GUJARAT" | "GJ" => {
            if gross_monthly <= 12000 {
                (Decimal::ZERO, "Up to Rs 12,000 - Nil".to_string())
            } else {
                (dec!(200), "Above Rs 12,000 - Rs 200".to_string())
            }
        }
        "MADHYA PRADESH" | "MP" => {
            if gross_monthly <= 18750 {
                (Decimal::ZERO, "Up to Rs 18,750 - Nil".to_string())
            } else if gross_monthly <= 25000 {
                (dec!(125), "Rs 18,751 to Rs 25,000 - Rs 125".to_string())
            } else if gross_monthly <= 33333 {
                (dec!(167), "Rs 25,001 to Rs 33,333 - Rs 167".to_string())
            } else {
                (dec!(208), "Above Rs 33,333 - Rs 208".to_string())
            }
        }
        "KERALA" | "KL" => {
            if gross_monthly <= 11999 {
                (Decimal::ZERO, "Up to Rs 11,999 - Nil".to_string())
            } else if gross_monthly <= 17999 {
                (dec!(120), "Rs 12,000 to Rs 17,999 - Rs 120".to_string())
            } else if gross_monthly <= 29999 {
                (dec!(180), "Rs 18,000 to Rs 29,999 - Rs 180".to_string())
            } else {
                (dec!(250), "Rs 30,000 and above - Rs 250".to_string())
            }
        }
        "ODISHA" | "OR" => {
            if gross_monthly <= 13304 {
                (Decimal::ZERO, "Up to Rs 13,304 - Nil".to_string())
            } else if gross_monthly <= 25000 {
                (dec!(125), "Rs 13,305 to Rs 25,000 - Rs 125".to_string())
            } else {
                (dec!(200), "Above Rs 25,000 - Rs 200".to_string())
            }
        }
        "ASSAM" | "AS" => {
            if gross_monthly <= 10000 {
                (Decimal::ZERO, "Up to Rs 10,000 - Nil".to_string())
            } else if gross_monthly <= 15000 {
                (dec!(150), "Rs 10,001 to Rs 15,000 - Rs 150".to_string())
            } else if gross_monthly <= 25000 {
                (dec!(180), "Rs 15,001 to Rs 25,000 - Rs 180".to_string())
            } else {
                (dec!(208), "Above Rs 25,000 - Rs 208".to_string())
            }
        }
        "MEGHALAYA" | "ML" => {
            if gross_monthly <= 4166 {
                (Decimal::ZERO, "Up to Rs 4,166 - Nil".to_string())
            } else if gross_monthly <= 6250 {
                (dec!(16.50), "Rs 4,167 to Rs 6,250 - Rs 16.50".to_string())
            } else if gross_monthly <= 8333 {
                (dec!(25), "Rs 6,251 to Rs 8,333 - Rs 25".to_string())
            } else if gross_monthly <= 12500 {
                (dec!(41.50), "Rs 8,334 to Rs 12,500 - Rs 41.50".to_string())
            } else if gross_monthly <= 16666 {
                (dec!(62.50), "Rs 12,501 to Rs 16,666 - Rs 62.50".to_string())
            } else if gross_monthly <= 20833 {
                (dec!(83.33), "Rs 16,667 to Rs 20,833 - Rs 83.33".to_string())
            } else {
                (dec!(208.33), "Above Rs 20,833 - Rs 208.33".to_string())
            }
        }
        "JHARKHAND" | "JH" => {
            if gross_monthly <= 25000 {
                (Decimal::ZERO, "Up to Rs 25,000 - Nil".to_string())
            } else if gross_monthly <= 41666 {
                (dec!(100), "Rs 25,001 to Rs 41,666 - Rs 100".to_string())
            } else if gross_monthly <= 66666 {
                (dec!(150), "Rs 41,667 to Rs 66,666 - Rs 150".to_string())
            } else if gross_monthly <= 83333 {
                (dec!(175), "Rs 66,667 to Rs 83,333 - Rs 175".to_string())
            } else {
                (dec!(200), "Above Rs 83,333 - Rs 200".to_string())
            }
        }
        "BIHAR" | "BR" => {
            if gross_monthly <= 25000 {
                (Decimal::ZERO, "Up to Rs 25,000 - Nil".to_string())
            } else if gross_monthly <= 41666 {
                (dec!(83.33), "Rs 25,001 to Rs 41,666 - Rs 83.33".to_string())
            } else if gross_monthly <= 66666 {
                (dec!(166.67), "Rs 41,667 to Rs 66,666 - Rs 166.67".to_string())
            } else if gross_monthly <= 83333 {
                (dec!(208.33), "Rs 66,667 to Rs 83,333 - Rs 208.33".to_string())
            } else {
                (dec!(250), "Above Rs 83,333 - Rs 250".to_string())
            }
        }
        "CHHATTISGARH" | "CG" => {
            if gross_monthly <= 12500 {
                (Decimal::ZERO, "Up to Rs 12,500 - Nil".to_string())
            } else if gross_monthly <= 16666 {
                (dec!(30), "Rs 12,501 to Rs 16,666 - Rs 30".to_string())
            } else if gross_monthly <= 25000 {
                (dec!(60), "Rs 16,667 to Rs 25,000 - Rs 60".to_string())
            } else if gross_monthly <= 33333 {
                (dec!(135), "Rs 25,001 to Rs 33,333 - Rs 135".to_string())
            } else if gross_monthly <= 41666 {
                (dec!(150), "Rs 33,334 to Rs 41,666 - Rs 150".to_string())
            } else {
                (dec!(175), "Above Rs 41,666 - Rs 175".to_string())
            }
        }
        // States without Professional Tax
        "RAJASTHAN" | "RJ" | "DELHI" | "DL" | "HARYANA" | "HR" | "UTTAR PRADESH" | "UP" |
        "UTTARAKHAND" | "UK" | "PUNJAB" | "PB" | "HIMACHAL PRADESH" | "HP" |
        "JAMMU AND KASHMIR" | "JK" | "GOA" | "GA" | "ARUNACHAL PRADESH" | "AR" |
        "MANIPUR" | "MN" | "MIZORAM" | "MZ" | "NAGALAND" | "NL" | "SIKKIM" | "SK" |
        "TRIPURA" | "TR" | "ANDAMAN AND NICOBAR" | "AN" | "CHANDIGARH" | "CH" |
        "DADRA AND NAGAR HAVELI" | "DN" | "DAMAN AND DIU" | "DD" | "LAKSHADWEEP" | "LD" |
        "PUDUCHERRY" | "PY" => {
            (Decimal::ZERO, "Professional Tax not applicable in this state".to_string())
        }
        _ => {
            // Default to 200 for unknown states with PT
            (dec!(200), "Default rate - Rs 200".to_string())
        }
    }
}

/// Calculate Professional Tax
#[wasm_bindgen]
pub fn calculate_professional_tax(input: JsValue) -> JsValue {
    let input: ProfessionalTaxInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid Professional Tax input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_professional_tax_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal Professional Tax calculation
pub fn calculate_professional_tax_internal(input: &ProfessionalTaxInput) -> ProfessionalTaxResult {
    let gross: Decimal = input.gross_salary.parse().unwrap_or(Decimal::ZERO);
    let (monthly_tax, slab_desc) = get_professional_tax(gross, &input.state);

    // Annual cap is Rs 2,500
    let annual_tax = (monthly_tax * dec!(12)).min(dec!(2500));

    ProfessionalTaxResult {
        gross_salary: gross.round_dp(0).to_string(),
        monthly_tax: monthly_tax.round_dp(0).to_string(),
        annual_tax: annual_tax.round_dp(0).to_string(),
        state: input.state.clone(),
        slab_description: slab_desc,
    }
}

/// Labour Welfare Fund rates by state
fn get_lwf_rates(state: &str) -> Option<(Decimal, Decimal, String)> {
    match state.to_uppercase().as_str() {
        "MAHARASHTRA" | "MH" => {
            Some((dec!(6), dec!(18), "Semi-Annual (June & December)".to_string()))
        }
        "KARNATAKA" | "KA" => {
            Some((dec!(20), dec!(40), "Annual (January)".to_string()))
        }
        "TAMIL NADU" | "TN" => {
            Some((dec!(5), dec!(10), "Annual".to_string()))
        }
        "KERALA" | "KL" => {
            Some((dec!(12), dec!(24), "Monthly".to_string()))
        }
        "ANDHRA PRADESH" | "AP" | "TELANGANA" | "TG" | "TS" => {
            Some((dec!(2), dec!(2), "Monthly".to_string()))
        }
        "WEST BENGAL" | "WB" => {
            Some((dec!(0.50), dec!(1.50), "Monthly".to_string()))
        }
        "GUJARAT" | "GJ" => {
            Some((dec!(6), dec!(12), "Semi-Annual (June & December)".to_string()))
        }
        "MADHYA PRADESH" | "MP" => {
            Some((dec!(10), dec!(30), "Semi-Annual (June & December)".to_string()))
        }
        "CHHATTISGARH" | "CG" => {
            Some((dec!(15), dec!(45), "Semi-Annual".to_string()))
        }
        "CHANDIGARH" | "CH" | "PUNJAB" | "PB" => {
            Some((dec!(5), dec!(20), "Monthly".to_string()))
        }
        "HARYANA" | "HR" => {
            Some((dec!(10), dec!(25), "Monthly".to_string()))
        }
        "GOA" | "GA" => {
            Some((dec!(60), dec!(180), "Annual".to_string()))
        }
        "ODISHA" | "OR" => {
            Some((dec!(10), dec!(30), "Semi-Annual".to_string()))
        }
        "DELHI" | "DL" => {
            Some((dec!(0.75), dec!(2.25), "Monthly".to_string()))
        }
        _ => None,
    }
}

/// Calculate Labour Welfare Fund
#[wasm_bindgen]
pub fn calculate_lwf(input: JsValue) -> JsValue {
    let input: LwfInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid LWF input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_lwf_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal LWF calculation
fn calculate_lwf_internal(input: &LwfInput) -> LwfResult {
    match get_lwf_rates(&input.state) {
        Some((employee, employer, frequency)) => {
            LwfResult {
                employee_contribution: employee.to_string(),
                employer_contribution: employer.to_string(),
                total_contribution: (employee + employer).to_string(),
                frequency,
                state: input.state.clone(),
            }
        }
        None => {
            LwfResult {
                employee_contribution: "0".to_string(),
                employer_contribution: "0".to_string(),
                total_contribution: "0".to_string(),
                frequency: "Not Applicable".to_string(),
                state: input.state.clone(),
            }
        }
    }
}

/// Calculate all statutory deductions
#[wasm_bindgen]
pub fn calculate_all_statutory(
    basic_salary: &str,
    gross_salary: &str,
    state: &str,
    da: Option<String>,
) -> JsValue {
    let basic: Decimal = basic_salary.parse().unwrap_or(Decimal::ZERO);
    let gross: Decimal = gross_salary.parse().unwrap_or(Decimal::ZERO);

    // PF calculation
    let pf_input = PfInput {
        basic_salary: basic_salary.to_string(),
        da,
        is_international_worker: None,
        is_excluded_establishment: None,
        employer_contribution_on_higher: None,
    };
    let pf_result = calculate_pf_internal(&pf_input);

    // ESI calculation
    let esi_input = EsiInput {
        gross_salary: gross_salary.to_string(),
        state: Some(state.to_string()),
    };
    let esi_result = calculate_esi_internal(&esi_input);

    // Professional Tax
    let pt_input = ProfessionalTaxInput {
        gross_salary: gross_salary.to_string(),
        state: state.to_string(),
    };
    let pt_result = calculate_professional_tax_internal(&pt_input);

    // LWF
    let lwf_input = LwfInput {
        gross_salary: gross_salary.to_string(),
        state: state.to_string(),
    };
    let lwf_result = calculate_lwf_internal(&lwf_input);

    // Total deductions
    let employee_pf: Decimal = pf_result.total_employee_contribution.parse().unwrap_or(Decimal::ZERO);
    let employee_esi: Decimal = esi_result.employee_contribution.parse().unwrap_or(Decimal::ZERO);
    let employee_pt: Decimal = pt_result.monthly_tax.parse().unwrap_or(Decimal::ZERO);
    let employee_lwf: Decimal = lwf_result.employee_contribution.parse().unwrap_or(Decimal::ZERO);

    let employer_pf: Decimal = pf_result.total_employer_contribution.parse().unwrap_or(Decimal::ZERO);
    let employer_esi: Decimal = esi_result.employer_contribution.parse().unwrap_or(Decimal::ZERO);
    let employer_lwf: Decimal = lwf_result.employer_contribution.parse().unwrap_or(Decimal::ZERO);

    let total_employee = employee_pf + employee_esi + employee_pt + employee_lwf;
    let total_employer = employer_pf + employer_esi + employer_lwf;

    let result = StatutoryDeductionsResult {
        pf: Some(pf_result),
        esi: if esi_result.is_applicable { Some(esi_result) } else { None },
        professional_tax: Some(pt_result),
        lwf: Some(lwf_result),
        total_employee_deductions: total_employee.round_dp(0).to_string(),
        total_employer_contributions: total_employer.round_dp(0).to_string(),
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Get PF rates
#[wasm_bindgen]
pub fn get_pf_rates() -> JsValue {
    let rates = serde_json::json!({
        "employeeContribution": "12%",
        "employerPf": "3.67%",
        "employerEps": "8.33%",
        "pfAdmin": "0.5%",
        "edli": "0.5%",
        "wageCeiling": PF_WAGE_CEILING,
    });

    serde_wasm_bindgen::to_value(&rates).unwrap_or(JsValue::NULL)
}

/// Get ESI rates
#[wasm_bindgen]
pub fn get_esi_rates() -> JsValue {
    let rates = serde_json::json!({
        "employeeContribution": "0.75%",
        "employerContribution": "3.25%",
        "wageCeiling": ESI_WAGE_CEILING,
    });

    serde_wasm_bindgen::to_value(&rates).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_pf_calculation() {
        let input = PfInput {
            basic_salary: "30000".to_string(),
            da: None,
            is_international_worker: None,
            is_excluded_establishment: None,
            employer_contribution_on_higher: None,
        };

        let result = calculate_pf_internal(&input);
        // PF wages capped at 15000
        assert_eq!(result.pf_wages, "15000");
        // Employee PF = 15000 * 12% = 1800
        assert_eq!(result.employee_pf, "1800");
    }

    #[test]
    fn test_esi_calculation() {
        let input = EsiInput {
            gross_salary: "18000".to_string(),
            state: None,
        };

        let result = calculate_esi_internal(&input);
        assert!(result.is_applicable);
        // Employee ESI = 18000 * 0.75% = 135
        assert_eq!(result.employee_contribution, "135");
        // Employer ESI = 18000 * 3.25% = 585
        assert_eq!(result.employer_contribution, "585");
    }

    #[test]
    fn test_esi_not_applicable() {
        let input = EsiInput {
            gross_salary: "25000".to_string(),
            state: None,
        };

        let result = calculate_esi_internal(&input);
        assert!(!result.is_applicable);
        assert_eq!(result.employee_contribution, "0");
    }

    #[test]
    fn test_professional_tax_maharashtra() {
        let input = ProfessionalTaxInput {
            gross_salary: "50000".to_string(),
            state: "Maharashtra".to_string(),
        };

        let result = calculate_professional_tax_internal(&input);
        assert_eq!(result.monthly_tax, "200");
    }
}
