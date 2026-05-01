//! Journal entry operations and validation

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Account types in double-entry accounting
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AccountType {
    Asset,
    Liability,
    Equity,
    Revenue,
    Expense,
}

impl AccountType {
    /// Returns the natural balance side (debit or credit)
    pub fn natural_balance(&self) -> BalanceSide {
        match self {
            AccountType::Asset => BalanceSide::Debit,
            AccountType::Expense => BalanceSide::Debit,
            AccountType::Liability => BalanceSide::Credit,
            AccountType::Equity => BalanceSide::Credit,
            AccountType::Revenue => BalanceSide::Credit,
        }
    }

    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_uppercase().as_str() {
            "ASSET" | "A" => Some(AccountType::Asset),
            "LIABILITY" | "L" => Some(AccountType::Liability),
            "EQUITY" | "E" | "CAPITAL" => Some(AccountType::Equity),
            "REVENUE" | "R" | "INCOME" => Some(AccountType::Revenue),
            "EXPENSE" | "X" | "EXP" => Some(AccountType::Expense),
            _ => None,
        }
    }
}

/// Balance side
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum BalanceSide {
    Debit,
    Credit,
}

/// Journal entry line
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JournalLine {
    pub account_code: String,
    pub account_name: String,
    pub account_type: Option<String>,
    pub debit: String,
    pub credit: String,
    pub narration: Option<String>,
    pub cost_center: Option<String>,
}

/// Journal entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JournalEntry {
    pub entry_number: String,
    pub entry_date: String,
    pub lines: Vec<JournalLine>,
    pub narration: String,
    pub reference: Option<String>,
    pub voucher_type: Option<String>,
}

/// Journal validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JournalValidation {
    pub valid: bool,
    pub total_debit: String,
    pub total_credit: String,
    pub difference: String,
    pub errors: Vec<String>,
    pub warnings: Vec<String>,
}

/// Validate journal entry (debit must equal credit)
#[wasm_bindgen]
pub fn validate_journal_entry(entry: JsValue) -> JsValue {
    let entry: JournalEntry = match serde_wasm_bindgen::from_value(entry) {
        Ok(e) => e,
        Err(e) => {
            let result = JournalValidation {
                valid: false,
                total_debit: "0".to_string(),
                total_credit: "0".to_string(),
                difference: "0".to_string(),
                errors: vec![format!("Invalid entry format: {}", e)],
                warnings: vec![],
            };
            return serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL);
        }
    };

    let result = validate_journal_internal(&entry);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal journal validation
pub fn validate_journal_internal(entry: &JournalEntry) -> JournalValidation {
    let mut errors = Vec::new();
    let mut warnings = Vec::new();
    let mut total_debit = Decimal::ZERO;
    let mut total_credit = Decimal::ZERO;

    // Check for empty entry
    if entry.lines.is_empty() {
        errors.push("Journal entry must have at least one line".to_string());
    }

    // Validate each line
    for (i, line) in entry.lines.iter().enumerate() {
        let line_num = i + 1;

        // Parse amounts
        let debit: Decimal = line.debit.parse().unwrap_or(Decimal::ZERO);
        let credit: Decimal = line.credit.parse().unwrap_or(Decimal::ZERO);

        // Check for both debit and credit
        if debit > Decimal::ZERO && credit > Decimal::ZERO {
            errors.push(format!("Line {}: Cannot have both debit and credit", line_num));
        }

        // Check for neither debit nor credit
        if debit.is_zero() && credit.is_zero() {
            errors.push(format!("Line {}: Must have either debit or credit", line_num));
        }

        // Check for negative amounts
        if debit < Decimal::ZERO || credit < Decimal::ZERO {
            errors.push(format!("Line {}: Amounts cannot be negative", line_num));
        }

        // Check for account code
        if line.account_code.trim().is_empty() {
            errors.push(format!("Line {}: Account code is required", line_num));
        }

        total_debit += debit;
        total_credit += credit;
    }

    // Check balance
    let difference = (total_debit - total_credit).abs();
    let is_balanced = difference < dec!(0.01); // Allow for small rounding

    if !is_balanced {
        errors.push(format!(
            "Entry is not balanced. Debit: {}, Credit: {}, Difference: {}",
            total_debit.round_dp(2),
            total_credit.round_dp(2),
            difference.round_dp(2)
        ));
    }

    // Warnings
    if entry.lines.len() < 2 {
        warnings.push("Journal entry typically has at least 2 lines".to_string());
    }

    if entry.narration.trim().is_empty() {
        warnings.push("Journal entry should have a narration".to_string());
    }

    JournalValidation {
        valid: errors.is_empty() && is_balanced,
        total_debit: total_debit.round_dp(2).to_string(),
        total_credit: total_credit.round_dp(2).to_string(),
        difference: difference.round_dp(2).to_string(),
        errors,
        warnings,
    }
}

/// Quick balance check
#[wasm_bindgen]
pub fn is_journal_balanced(debits: JsValue, credits: JsValue) -> bool {
    let debits: Vec<String> = serde_wasm_bindgen::from_value(debits).unwrap_or_default();
    let credits: Vec<String> = serde_wasm_bindgen::from_value(credits).unwrap_or_default();

    let total_debit: Decimal = debits.iter()
        .filter_map(|s| s.parse::<Decimal>().ok())
        .sum();

    let total_credit: Decimal = credits.iter()
        .filter_map(|s| s.parse::<Decimal>().ok())
        .sum();

    (total_debit - total_credit).abs() < dec!(0.01)
}

/// Calculate journal totals
#[wasm_bindgen]
pub fn calculate_journal_totals(lines: JsValue) -> JsValue {
    let lines: Vec<JournalLine> = serde_wasm_bindgen::from_value(lines).unwrap_or_default();

    let total_debit: Decimal = lines.iter()
        .filter_map(|l| l.debit.parse::<Decimal>().ok())
        .sum();

    let total_credit: Decimal = lines.iter()
        .filter_map(|l| l.credit.parse::<Decimal>().ok())
        .sum();

    let difference = total_debit - total_credit;

    let result = serde_json::json!({
        "totalDebit": total_debit.round_dp(2).to_string(),
        "totalCredit": total_credit.round_dp(2).to_string(),
        "difference": difference.round_dp(2).to_string(),
        "isBalanced": difference.abs() < dec!(0.01),
        "balanceSide": if difference > Decimal::ZERO { "debit" } else if difference < Decimal::ZERO { "credit" } else { "balanced" }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Split amount between debit and credit
#[wasm_bindgen]
pub fn split_entry(amount: &str, entry_type: &str) -> JsValue {
    let amount: Decimal = amount.parse().unwrap_or(Decimal::ZERO);

    let (debit, credit) = match entry_type.to_lowercase().as_str() {
        "debit" | "dr" => (amount, Decimal::ZERO),
        "credit" | "cr" => (Decimal::ZERO, amount),
        _ => (Decimal::ZERO, Decimal::ZERO),
    };

    let result = serde_json::json!({
        "debit": debit.round_dp(2).to_string(),
        "credit": credit.round_dp(2).to_string()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_balanced_entry() {
        let entry = JournalEntry {
            entry_number: "JV001".to_string(),
            entry_date: "2024-01-01".to_string(),
            lines: vec![
                JournalLine {
                    account_code: "1001".to_string(),
                    account_name: "Cash".to_string(),
                    account_type: Some("Asset".to_string()),
                    debit: "1000".to_string(),
                    credit: "0".to_string(),
                    narration: None,
                    cost_center: None,
                },
                JournalLine {
                    account_code: "4001".to_string(),
                    account_name: "Sales".to_string(),
                    account_type: Some("Revenue".to_string()),
                    debit: "0".to_string(),
                    credit: "1000".to_string(),
                    narration: None,
                    cost_center: None,
                },
            ],
            narration: "Cash sale".to_string(),
            reference: None,
            voucher_type: None,
        };

        let result = validate_journal_internal(&entry);
        assert!(result.valid);
        assert_eq!(result.total_debit, "1000.00");
        assert_eq!(result.total_credit, "1000.00");
        assert_eq!(result.difference, "0.00");
    }

    #[test]
    fn test_unbalanced_entry() {
        let entry = JournalEntry {
            entry_number: "JV002".to_string(),
            entry_date: "2024-01-01".to_string(),
            lines: vec![
                JournalLine {
                    account_code: "1001".to_string(),
                    account_name: "Cash".to_string(),
                    account_type: Some("Asset".to_string()),
                    debit: "1000".to_string(),
                    credit: "0".to_string(),
                    narration: None,
                    cost_center: None,
                },
                JournalLine {
                    account_code: "4001".to_string(),
                    account_name: "Sales".to_string(),
                    account_type: Some("Revenue".to_string()),
                    debit: "0".to_string(),
                    credit: "500".to_string(),
                    narration: None,
                    cost_center: None,
                },
            ],
            narration: "Unbalanced entry".to_string(),
            reference: None,
            voucher_type: None,
        };

        let result = validate_journal_internal(&entry);
        assert!(!result.valid);
        assert_eq!(result.difference, "500.00");
    }
}
