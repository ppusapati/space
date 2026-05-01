//! Account balance calculations

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Ledger entry for balance calculation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LedgerEntry {
    pub date: String,
    pub voucher_number: String,
    pub narration: String,
    pub debit: String,
    pub credit: String,
}

/// Account balance result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AccountBalance {
    pub opening_balance: String,
    pub opening_balance_type: String,
    pub total_debit: String,
    pub total_credit: String,
    pub closing_balance: String,
    pub closing_balance_type: String,
    pub transaction_count: u32,
}

/// Calculate account balance
#[wasm_bindgen]
pub fn calculate_balance(
    entries: JsValue,
    opening_balance: &str,
    opening_type: &str,
    account_nature: &str,
) -> JsValue {
    let entries: Vec<LedgerEntry> = serde_wasm_bindgen::from_value(entries).unwrap_or_default();

    let opening: Decimal = opening_balance.parse().unwrap_or(Decimal::ZERO);
    let is_debit_nature = match account_nature.to_lowercase().as_str() {
        "asset" | "expense" | "debit" => true,
        _ => false,
    };

    // Convert opening to signed based on type
    let mut running_balance = if opening_type.to_lowercase() == "credit" {
        -opening
    } else {
        opening
    };

    let mut total_debit = Decimal::ZERO;
    let mut total_credit = Decimal::ZERO;

    for entry in &entries {
        let debit: Decimal = entry.debit.parse().unwrap_or(Decimal::ZERO);
        let credit: Decimal = entry.credit.parse().unwrap_or(Decimal::ZERO);

        total_debit += debit;
        total_credit += credit;

        if is_debit_nature {
            running_balance += debit - credit;
        } else {
            running_balance += credit - debit;
        }
    }

    let (closing_balance, closing_type) = if running_balance >= Decimal::ZERO {
        (running_balance, if is_debit_nature { "Debit" } else { "Credit" })
    } else {
        (running_balance.abs(), if is_debit_nature { "Credit" } else { "Debit" })
    };

    let result = AccountBalance {
        opening_balance: opening.round_dp(2).to_string(),
        opening_balance_type: opening_type.to_string(),
        total_debit: total_debit.round_dp(2).to_string(),
        total_credit: total_credit.round_dp(2).to_string(),
        closing_balance: closing_balance.round_dp(2).to_string(),
        closing_balance_type: closing_type.to_string(),
        transaction_count: entries.len() as u32,
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate running balance for ledger display
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RunningBalanceEntry {
    pub date: String,
    pub voucher_number: String,
    pub narration: String,
    pub debit: String,
    pub credit: String,
    pub balance: String,
    pub balance_type: String,
}

/// Calculate running balances
#[wasm_bindgen]
pub fn calculate_running_balances(
    entries: JsValue,
    opening_balance: &str,
    opening_type: &str,
    account_nature: &str,
) -> JsValue {
    let entries: Vec<LedgerEntry> = serde_wasm_bindgen::from_value(entries).unwrap_or_default();

    let opening: Decimal = opening_balance.parse().unwrap_or(Decimal::ZERO);
    let is_debit_nature = matches!(
        account_nature.to_lowercase().as_str(),
        "asset" | "expense" | "debit"
    );

    let mut running_balance = if opening_type.to_lowercase() == "credit" {
        -opening
    } else {
        opening
    };

    let mut results: Vec<RunningBalanceEntry> = Vec::new();

    for entry in &entries {
        let debit: Decimal = entry.debit.parse().unwrap_or(Decimal::ZERO);
        let credit: Decimal = entry.credit.parse().unwrap_or(Decimal::ZERO);

        if is_debit_nature {
            running_balance += debit - credit;
        } else {
            running_balance += credit - debit;
        }

        let (balance, balance_type) = if running_balance >= Decimal::ZERO {
            (running_balance, if is_debit_nature { "Dr" } else { "Cr" })
        } else {
            (running_balance.abs(), if is_debit_nature { "Cr" } else { "Dr" })
        };

        results.push(RunningBalanceEntry {
            date: entry.date.clone(),
            voucher_number: entry.voucher_number.clone(),
            narration: entry.narration.clone(),
            debit: if debit > Decimal::ZERO { debit.round_dp(2).to_string() } else { "".to_string() },
            credit: if credit > Decimal::ZERO { credit.round_dp(2).to_string() } else { "".to_string() },
            balance: balance.round_dp(2).to_string(),
            balance_type: balance_type.to_string(),
        });
    }

    serde_wasm_bindgen::to_value(&results).unwrap_or(JsValue::NULL)
}

/// Calculate net balance from debit and credit totals
#[wasm_bindgen]
pub fn net_balance(total_debit: &str, total_credit: &str) -> JsValue {
    let debit: Decimal = total_debit.parse().unwrap_or(Decimal::ZERO);
    let credit: Decimal = total_credit.parse().unwrap_or(Decimal::ZERO);

    let difference = debit - credit;
    let (balance, balance_type) = if difference >= Decimal::ZERO {
        (difference, "Debit")
    } else {
        (difference.abs(), "Credit")
    };

    let result = serde_json::json!({
        "balance": balance.round_dp(2).to_string(),
        "balanceType": balance_type,
        "isDebit": difference >= Decimal::ZERO
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Check if balance is within limit
#[wasm_bindgen]
pub fn is_within_limit(balance: &str, limit: &str) -> bool {
    let balance: Decimal = balance.parse().unwrap_or(Decimal::ZERO);
    let limit: Decimal = limit.parse().unwrap_or(Decimal::ZERO);

    balance.abs() <= limit.abs()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_entries() -> Vec<LedgerEntry> {
        vec![
            LedgerEntry {
                date: "2024-01-01".to_string(),
                voucher_number: "JV001".to_string(),
                narration: "Opening".to_string(),
                debit: "1000".to_string(),
                credit: "0".to_string(),
            },
            LedgerEntry {
                date: "2024-01-15".to_string(),
                voucher_number: "JV002".to_string(),
                narration: "Sale".to_string(),
                debit: "500".to_string(),
                credit: "0".to_string(),
            },
            LedgerEntry {
                date: "2024-01-20".to_string(),
                voucher_number: "JV003".to_string(),
                narration: "Payment".to_string(),
                debit: "0".to_string(),
                credit: "300".to_string(),
            },
        ]
    }
}
