//! Trial balance computation

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Account balance entry for trial balance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrialBalanceEntry {
    pub account_code: String,
    pub account_name: String,
    pub account_type: String,
    pub group: Option<String>,
    pub opening_debit: String,
    pub opening_credit: String,
    pub period_debit: String,
    pub period_credit: String,
    pub closing_debit: String,
    pub closing_credit: String,
}

/// Trial balance result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrialBalanceResult {
    pub entries: Vec<TrialBalanceEntry>,
    pub totals: TrialBalanceTotals,
    pub is_balanced: bool,
    pub difference: String,
    pub warnings: Vec<String>,
}

/// Trial balance totals
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrialBalanceTotals {
    pub opening_debit: String,
    pub opening_credit: String,
    pub period_debit: String,
    pub period_credit: String,
    pub closing_debit: String,
    pub closing_credit: String,
}

/// Account summary for balance sheet
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AccountSummary {
    pub account_code: String,
    pub account_name: String,
    pub account_type: String,
    pub total_debit: String,
    pub total_credit: String,
}

/// Calculate trial balance
#[wasm_bindgen]
pub fn calculate_trial_balance(entries: JsValue) -> JsValue {
    let entries: Vec<TrialBalanceEntry> = serde_wasm_bindgen::from_value(entries).unwrap_or_default();

    let result = calculate_trial_balance_internal(&entries);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal trial balance calculation
fn calculate_trial_balance_internal(entries: &[TrialBalanceEntry]) -> TrialBalanceResult {
    let mut opening_debit_total = Decimal::ZERO;
    let mut opening_credit_total = Decimal::ZERO;
    let mut period_debit_total = Decimal::ZERO;
    let mut period_credit_total = Decimal::ZERO;
    let mut closing_debit_total = Decimal::ZERO;
    let mut closing_credit_total = Decimal::ZERO;
    let mut warnings = Vec::new();

    for entry in entries {
        let opening_debit: Decimal = entry.opening_debit.parse().unwrap_or(Decimal::ZERO);
        let opening_credit: Decimal = entry.opening_credit.parse().unwrap_or(Decimal::ZERO);
        let period_debit: Decimal = entry.period_debit.parse().unwrap_or(Decimal::ZERO);
        let period_credit: Decimal = entry.period_credit.parse().unwrap_or(Decimal::ZERO);
        let closing_debit: Decimal = entry.closing_debit.parse().unwrap_or(Decimal::ZERO);
        let closing_credit: Decimal = entry.closing_credit.parse().unwrap_or(Decimal::ZERO);

        opening_debit_total += opening_debit;
        opening_credit_total += opening_credit;
        period_debit_total += period_debit;
        period_credit_total += period_credit;
        closing_debit_total += closing_debit;
        closing_credit_total += closing_credit;

        // Warn if both debit and credit have values for closing
        if closing_debit > Decimal::ZERO && closing_credit > Decimal::ZERO {
            warnings.push(format!(
                "Account {} has both debit and credit closing balance",
                entry.account_code
            ));
        }
    }

    let opening_diff = (opening_debit_total - opening_credit_total).abs();
    let period_diff = (period_debit_total - period_credit_total).abs();
    let closing_diff = (closing_debit_total - closing_credit_total).abs();

    let is_balanced = opening_diff < dec!(0.01) &&
                      period_diff < dec!(0.01) &&
                      closing_diff < dec!(0.01);

    if opening_diff >= dec!(0.01) {
        warnings.push(format!("Opening balance not balanced. Difference: {}", opening_diff));
    }
    if period_diff >= dec!(0.01) {
        warnings.push(format!("Period transactions not balanced. Difference: {}", period_diff));
    }

    TrialBalanceResult {
        entries: entries.to_vec(),
        totals: TrialBalanceTotals {
            opening_debit: opening_debit_total.round_dp(2).to_string(),
            opening_credit: opening_credit_total.round_dp(2).to_string(),
            period_debit: period_debit_total.round_dp(2).to_string(),
            period_credit: period_credit_total.round_dp(2).to_string(),
            closing_debit: closing_debit_total.round_dp(2).to_string(),
            closing_credit: closing_credit_total.round_dp(2).to_string(),
        },
        is_balanced,
        difference: closing_diff.round_dp(2).to_string(),
        warnings,
    }
}

/// Calculate closing balance from opening and transactions
#[wasm_bindgen]
pub fn calculate_closing_balance(
    opening_debit: &str,
    opening_credit: &str,
    period_debit: &str,
    period_credit: &str,
    account_nature: &str,
) -> JsValue {
    let opening_dr: Decimal = opening_debit.parse().unwrap_or(Decimal::ZERO);
    let opening_cr: Decimal = opening_credit.parse().unwrap_or(Decimal::ZERO);
    let period_dr: Decimal = period_debit.parse().unwrap_or(Decimal::ZERO);
    let period_cr: Decimal = period_credit.parse().unwrap_or(Decimal::ZERO);

    let is_debit_nature = matches!(
        account_nature.to_lowercase().as_str(),
        "asset" | "expense" | "debit"
    );

    // Calculate net opening
    let opening_balance = if opening_dr > opening_cr {
        opening_dr - opening_cr
    } else {
        -(opening_cr - opening_dr)
    };

    // Calculate net movement
    let movement = period_dr - period_cr;

    // Calculate closing
    let closing = if is_debit_nature {
        opening_balance + movement
    } else {
        opening_balance - movement
    };

    let (closing_debit, closing_credit) = if closing >= Decimal::ZERO {
        if is_debit_nature {
            (closing, Decimal::ZERO)
        } else {
            (Decimal::ZERO, closing)
        }
    } else {
        if is_debit_nature {
            (Decimal::ZERO, closing.abs())
        } else {
            (closing.abs(), Decimal::ZERO)
        }
    };

    let result = serde_json::json!({
        "closingDebit": closing_debit.round_dp(2).to_string(),
        "closingCredit": closing_credit.round_dp(2).to_string(),
        "netBalance": closing.abs().round_dp(2).to_string(),
        "balanceType": if closing >= Decimal::ZERO { "Debit" } else { "Credit" }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Group trial balance by account type
#[wasm_bindgen]
pub fn group_trial_balance(entries: JsValue) -> JsValue {
    let entries: Vec<TrialBalanceEntry> = serde_wasm_bindgen::from_value(entries).unwrap_or_default();

    let mut groups: std::collections::HashMap<String, Vec<&TrialBalanceEntry>> = std::collections::HashMap::new();

    for entry in &entries {
        let group = entry.group.as_deref().unwrap_or(&entry.account_type);
        groups.entry(group.to_string()).or_default().push(entry);
    }

    // Calculate group totals
    let group_totals: Vec<serde_json::Value> = groups.iter().map(|(group, items)| {
        let mut debit_total = Decimal::ZERO;
        let mut credit_total = Decimal::ZERO;

        for item in items {
            let debit: Decimal = item.closing_debit.parse().unwrap_or(Decimal::ZERO);
            let credit: Decimal = item.closing_credit.parse().unwrap_or(Decimal::ZERO);
            debit_total += debit;
            credit_total += credit;
        }

        serde_json::json!({
            "group": group,
            "count": items.len(),
            "debit": debit_total.round_dp(2).to_string(),
            "credit": credit_total.round_dp(2).to_string()
        })
    }).collect();

    serde_wasm_bindgen::to_value(&group_totals).unwrap_or(JsValue::NULL)
}

/// Quick balance check
#[wasm_bindgen]
pub fn is_trial_balance_balanced(total_debit: &str, total_credit: &str) -> bool {
    let debit: Decimal = total_debit.parse().unwrap_or(Decimal::ZERO);
    let credit: Decimal = total_credit.parse().unwrap_or(Decimal::ZERO);

    (debit - credit).abs() < dec!(0.01)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn sample_entries() -> Vec<TrialBalanceEntry> {
        vec![
            TrialBalanceEntry {
                account_code: "1001".to_string(),
                account_name: "Cash".to_string(),
                account_type: "Asset".to_string(),
                group: None,
                opening_debit: "10000".to_string(),
                opening_credit: "0".to_string(),
                period_debit: "5000".to_string(),
                period_credit: "3000".to_string(),
                closing_debit: "12000".to_string(),
                closing_credit: "0".to_string(),
            },
            TrialBalanceEntry {
                account_code: "2001".to_string(),
                account_name: "Accounts Payable".to_string(),
                account_type: "Liability".to_string(),
                group: None,
                opening_debit: "0".to_string(),
                opening_credit: "10000".to_string(),
                period_debit: "3000".to_string(),
                period_credit: "5000".to_string(),
                closing_debit: "0".to_string(),
                closing_credit: "12000".to_string(),
            },
        ]
    }

    #[test]
    fn test_balanced_trial_balance() {
        let entries = sample_entries();
        let result = calculate_trial_balance_internal(&entries);

        assert!(result.is_balanced);
        assert_eq!(result.totals.closing_debit, "12000.00");
        assert_eq!(result.totals.closing_credit, "12000.00");
    }

    #[test]
    fn test_is_balanced() {
        assert!(is_trial_balance_balanced("10000", "10000"));
        assert!(is_trial_balance_balanced("10000.00", "10000.005")); // Within tolerance
        assert!(!is_trial_balance_balanced("10000", "10001"));
    }
}
