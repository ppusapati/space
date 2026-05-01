//! Bank reconciliation matching algorithms

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Bank statement entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BankEntry {
    pub id: String,
    pub date: String,
    pub description: String,
    pub reference: Option<String>,
    pub amount: String,
    pub entry_type: String, // "credit" or "debit"
}

/// Book entry (from accounting system)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BookEntry {
    pub id: String,
    pub date: String,
    pub voucher_number: String,
    pub narration: String,
    pub reference: Option<String>,
    pub amount: String,
    pub entry_type: String, // "credit" or "debit"
    pub cheque_number: Option<String>,
}

/// Match result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MatchResult {
    pub bank_entry_id: String,
    pub book_entry_id: String,
    pub bank_amount: String,
    pub book_amount: String,
    pub match_type: String,
    pub confidence: f64,
    pub difference: String,
}

/// Unmatched entries
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnmatchedEntry {
    pub id: String,
    pub source: String, // "bank" or "book"
    pub date: String,
    pub description: String,
    pub amount: String,
    pub entry_type: String,
}

/// Reconciliation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReconciliationResult {
    pub matches: Vec<MatchResult>,
    pub unmatched_bank: Vec<UnmatchedEntry>,
    pub unmatched_book: Vec<UnmatchedEntry>,
    pub bank_balance: String,
    pub book_balance: String,
    pub reconciled_balance: String,
    pub difference: String,
    pub match_count: u32,
    pub match_percentage: String,
}

/// Perform bank reconciliation
#[wasm_bindgen]
pub fn reconcile_bank(bank_entries: JsValue, book_entries: JsValue, tolerance: Option<f64>) -> JsValue {
    let bank_entries: Vec<BankEntry> = serde_wasm_bindgen::from_value(bank_entries).unwrap_or_default();
    let book_entries: Vec<BookEntry> = serde_wasm_bindgen::from_value(book_entries).unwrap_or_default();
    let tolerance = Decimal::from_f64(tolerance.unwrap_or(0.01)).unwrap_or(dec!(0.01));

    let result = reconcile_internal(&bank_entries, &book_entries, tolerance);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal reconciliation logic
fn reconcile_internal(
    bank_entries: &[BankEntry],
    book_entries: &[BookEntry],
    tolerance: Decimal,
) -> ReconciliationResult {
    let mut matches = Vec::new();
    let mut matched_bank_ids: Vec<String> = Vec::new();
    let mut matched_book_ids: Vec<String> = Vec::new();

    // Try to match each bank entry with book entries
    for bank in bank_entries {
        let bank_amount: Decimal = bank.amount.parse().unwrap_or(Decimal::ZERO);

        // Find potential matches
        let mut best_match: Option<(String, f64, Decimal)> = None;

        for book in book_entries {
            // Skip already matched
            if matched_book_ids.contains(&book.id) {
                continue;
            }

            // Entry types must match (both credit or both debit)
            if bank.entry_type != book.entry_type {
                continue;
            }

            let book_amount: Decimal = book.amount.parse().unwrap_or(Decimal::ZERO);
            let difference = (bank_amount - book_amount).abs();

            // Calculate confidence score
            let mut confidence = 0.0;

            // Exact amount match
            if difference <= tolerance {
                confidence += 50.0;
            } else if difference <= dec!(1) {
                confidence += 30.0;
            } else if difference <= dec!(10) {
                confidence += 10.0;
            }

            // Reference match
            if let (Some(bank_ref), Some(book_ref)) = (&bank.reference, &book.reference) {
                if bank_ref.to_lowercase() == book_ref.to_lowercase() {
                    confidence += 30.0;
                }
            }

            // Cheque number in description
            if let Some(cheque) = &book.cheque_number {
                if bank.description.contains(cheque) {
                    confidence += 20.0;
                }
            }

            // Date proximity (within 3 days)
            if dates_within_days(&bank.date, &book.date, 3) {
                confidence += 10.0;
            } else if dates_within_days(&bank.date, &book.date, 7) {
                confidence += 5.0;
            }

            // Narration similarity
            let narration_similarity = text_similarity(&bank.description, &book.narration);
            confidence += narration_similarity * 10.0;

            // Keep best match above threshold
            if confidence >= 50.0 {
                if best_match.is_none() || confidence > best_match.as_ref().unwrap().1 {
                    best_match = Some((book.id.clone(), confidence, difference));
                }
            }
        }

        // Record match
        if let Some((book_id, confidence, difference)) = best_match {
            let book = book_entries.iter().find(|b| b.id == book_id).unwrap();

            matches.push(MatchResult {
                bank_entry_id: bank.id.clone(),
                book_entry_id: book_id.clone(),
                bank_amount: bank.amount.clone(),
                book_amount: book.amount.clone(),
                match_type: if difference <= tolerance { "exact" } else { "partial" }.to_string(),
                confidence,
                difference: difference.round_dp(2).to_string(),
            });

            matched_bank_ids.push(bank.id.clone());
            matched_book_ids.push(book_id);
        }
    }

    // Collect unmatched entries
    let unmatched_bank: Vec<UnmatchedEntry> = bank_entries.iter()
        .filter(|b| !matched_bank_ids.contains(&b.id))
        .map(|b| UnmatchedEntry {
            id: b.id.clone(),
            source: "bank".to_string(),
            date: b.date.clone(),
            description: b.description.clone(),
            amount: b.amount.clone(),
            entry_type: b.entry_type.clone(),
        })
        .collect();

    let unmatched_book: Vec<UnmatchedEntry> = book_entries.iter()
        .filter(|b| !matched_book_ids.contains(&b.id))
        .map(|b| UnmatchedEntry {
            id: b.id.clone(),
            source: "book".to_string(),
            date: b.date.clone(),
            description: b.narration.clone(),
            amount: b.amount.clone(),
            entry_type: b.entry_type.clone(),
        })
        .collect();

    // Calculate balances
    let bank_balance = calculate_balance(bank_entries);
    let book_balance = calculate_book_balance(book_entries);

    let match_count = matches.len() as u32;
    let total_entries = bank_entries.len() as f64;
    let match_percentage = if total_entries > 0.0 {
        (match_count as f64 / total_entries * 100.0).round()
    } else {
        0.0
    };

    ReconciliationResult {
        matches,
        unmatched_bank,
        unmatched_book,
        bank_balance: bank_balance.round_dp(2).to_string(),
        book_balance: book_balance.round_dp(2).to_string(),
        reconciled_balance: (bank_balance - book_balance).round_dp(2).to_string(),
        difference: (bank_balance - book_balance).abs().round_dp(2).to_string(),
        match_count,
        match_percentage: format!("{}%", match_percentage),
    }
}

/// Calculate bank balance
fn calculate_balance(entries: &[BankEntry]) -> Decimal {
    let mut balance = Decimal::ZERO;
    for entry in entries {
        let amount: Decimal = entry.amount.parse().unwrap_or(Decimal::ZERO);
        if entry.entry_type == "credit" {
            balance += amount;
        } else {
            balance -= amount;
        }
    }
    balance
}

/// Calculate book balance
fn calculate_book_balance(entries: &[BookEntry]) -> Decimal {
    let mut balance = Decimal::ZERO;
    for entry in entries {
        let amount: Decimal = entry.amount.parse().unwrap_or(Decimal::ZERO);
        if entry.entry_type == "credit" {
            balance += amount;
        } else {
            balance -= amount;
        }
    }
    balance
}

/// Check if two dates are within N days
fn dates_within_days(date1: &str, date2: &str, days: i64) -> bool {
    use chrono::NaiveDate;

    let d1 = NaiveDate::parse_from_str(date1, "%Y-%m-%d").ok();
    let d2 = NaiveDate::parse_from_str(date2, "%Y-%m-%d").ok();

    match (d1, d2) {
        (Some(d1), Some(d2)) => (d1 - d2).num_days().abs() <= days,
        _ => false,
    }
}

/// Simple text similarity (Jaccard index on words)
fn text_similarity(text1: &str, text2: &str) -> f64 {
    let text1_lower = text1.to_lowercase();
    let text2_lower = text2.to_lowercase();

    let words1: std::collections::HashSet<&str> = text1_lower
        .split_whitespace()
        .collect();
    let words2: std::collections::HashSet<&str> = text2_lower
        .split_whitespace()
        .collect();

    let intersection = words1.intersection(&words2).count();
    let union = words1.union(&words2).count();

    if union == 0 {
        0.0
    } else {
        intersection as f64 / union as f64
    }
}

/// Find matching book entry for a bank entry
#[wasm_bindgen]
pub fn find_match(bank_entry: JsValue, book_entries: JsValue) -> JsValue {
    let bank: BankEntry = match serde_wasm_bindgen::from_value(bank_entry) {
        Ok(b) => b,
        Err(_) => return JsValue::NULL,
    };
    let book_entries: Vec<BookEntry> = serde_wasm_bindgen::from_value(book_entries).unwrap_or_default();

    let bank_amount: Decimal = bank.amount.parse().unwrap_or(Decimal::ZERO);

    // Find entries with matching amount and type
    let candidates: Vec<&BookEntry> = book_entries.iter()
        .filter(|b| {
            let book_amount: Decimal = b.amount.parse().unwrap_or(Decimal::ZERO);
            let diff = (bank_amount - book_amount).abs();
            b.entry_type == bank.entry_type && diff <= dec!(0.01)
        })
        .collect();

    if candidates.is_empty() {
        return JsValue::NULL;
    }

    // Return best match
    serde_wasm_bindgen::to_value(&candidates[0]).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_text_similarity() {
        let sim1 = text_similarity("payment to vendor", "vendor payment received");
        assert!(sim1 > 0.0);

        let sim2 = text_similarity("abc", "xyz");
        assert_eq!(sim2, 0.0);
    }

    #[test]
    fn test_dates_within_days() {
        assert!(dates_within_days("2024-01-01", "2024-01-03", 3));
        assert!(!dates_within_days("2024-01-01", "2024-01-10", 3));
    }
}
