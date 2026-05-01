//! Aging calculations for Accounts Receivable and Payable

use chrono::{NaiveDate, Utc};
use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Invoice/Bill for aging
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgingItem {
    pub document_number: String,
    pub document_date: String,
    pub due_date: String,
    pub party_name: String,
    pub original_amount: String,
    pub outstanding_amount: String,
}

/// Aging bucket configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgingBucket {
    pub label: String,
    pub from_days: i32,
    pub to_days: Option<i32>, // None means no upper limit
}

/// Aging result for a single item
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgingResult {
    pub document_number: String,
    pub document_date: String,
    pub due_date: String,
    pub party_name: String,
    pub original_amount: String,
    pub outstanding_amount: String,
    pub days_overdue: i32,
    pub bucket: String,
    pub is_overdue: bool,
}

/// Aging summary
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgingSummary {
    pub items: Vec<AgingResult>,
    pub buckets: Vec<BucketTotal>,
    pub total_outstanding: String,
    pub total_overdue: String,
    pub overdue_count: u32,
    pub as_of_date: String,
}

/// Bucket total
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BucketTotal {
    pub label: String,
    pub count: u32,
    pub amount: String,
    pub percentage: String,
}

/// Default aging buckets
pub fn default_buckets() -> Vec<AgingBucket> {
    vec![
        AgingBucket { label: "Current".to_string(), from_days: i32::MIN, to_days: Some(0) },
        AgingBucket { label: "1-30 Days".to_string(), from_days: 1, to_days: Some(30) },
        AgingBucket { label: "31-60 Days".to_string(), from_days: 31, to_days: Some(60) },
        AgingBucket { label: "61-90 Days".to_string(), from_days: 61, to_days: Some(90) },
        AgingBucket { label: "90+ Days".to_string(), from_days: 91, to_days: None },
    ]
}

/// Calculate aging for items
#[wasm_bindgen]
pub fn calculate_aging(items: JsValue, as_of_date: Option<String>, buckets: JsValue) -> JsValue {
    let items: Vec<AgingItem> = serde_wasm_bindgen::from_value(items).unwrap_or_default();
    let buckets: Vec<AgingBucket> = serde_wasm_bindgen::from_value(buckets)
        .unwrap_or_else(|_| default_buckets());

    let as_of = as_of_date
        .and_then(|d| NaiveDate::parse_from_str(&d, "%Y-%m-%d").ok())
        .unwrap_or_else(|| Utc::now().date_naive());

    let mut results: Vec<AgingResult> = Vec::new();
    let mut bucket_totals: Vec<(String, u32, Decimal)> = buckets.iter()
        .map(|b| (b.label.clone(), 0, Decimal::ZERO))
        .collect();

    let mut total_outstanding = Decimal::ZERO;
    let mut total_overdue = Decimal::ZERO;
    let mut overdue_count = 0u32;

    for item in &items {
        let due_date = NaiveDate::parse_from_str(&item.due_date, "%Y-%m-%d")
            .unwrap_or(as_of);

        let days_overdue = (as_of - due_date).num_days() as i32;
        let outstanding: Decimal = item.outstanding_amount.parse().unwrap_or(Decimal::ZERO);

        // Find bucket
        let bucket_label = buckets.iter()
            .find(|b| {
                days_overdue >= b.from_days &&
                b.to_days.map(|to| days_overdue <= to).unwrap_or(true)
            })
            .map(|b| b.label.clone())
            .unwrap_or_else(|| "Unknown".to_string());

        // Update bucket totals
        for (label, count, amount) in &mut bucket_totals {
            if *label == bucket_label {
                *count += 1;
                *amount += outstanding;
                break;
            }
        }

        let is_overdue = days_overdue > 0;
        if is_overdue {
            total_overdue += outstanding;
            overdue_count += 1;
        }
        total_outstanding += outstanding;

        results.push(AgingResult {
            document_number: item.document_number.clone(),
            document_date: item.document_date.clone(),
            due_date: item.due_date.clone(),
            party_name: item.party_name.clone(),
            original_amount: item.original_amount.clone(),
            outstanding_amount: outstanding.round_dp(2).to_string(),
            days_overdue,
            bucket: bucket_label,
            is_overdue,
        });
    }

    // Calculate percentages
    let bucket_results: Vec<BucketTotal> = bucket_totals.iter()
        .map(|(label, count, amount)| {
            let percentage = if total_outstanding > Decimal::ZERO {
                (*amount / total_outstanding * dec!(100)).round_dp(2)
            } else {
                Decimal::ZERO
            };
            BucketTotal {
                label: label.clone(),
                count: *count,
                amount: amount.round_dp(2).to_string(),
                percentage: percentage.to_string(),
            }
        })
        .collect();

    let summary = AgingSummary {
        items: results,
        buckets: bucket_results,
        total_outstanding: total_outstanding.round_dp(2).to_string(),
        total_overdue: total_overdue.round_dp(2).to_string(),
        overdue_count,
        as_of_date: as_of.format("%Y-%m-%d").to_string(),
    };

    serde_wasm_bindgen::to_value(&summary).unwrap_or(JsValue::NULL)
}

/// Calculate days overdue
#[wasm_bindgen]
pub fn days_overdue(due_date: &str, as_of_date: Option<String>) -> i32 {
    let due = NaiveDate::parse_from_str(due_date, "%Y-%m-%d")
        .unwrap_or_else(|_| Utc::now().date_naive());

    let as_of = as_of_date
        .and_then(|d| NaiveDate::parse_from_str(&d, "%Y-%m-%d").ok())
        .unwrap_or_else(|| Utc::now().date_naive());

    (as_of - due).num_days() as i32
}

/// Check if document is overdue
#[wasm_bindgen]
pub fn is_overdue(due_date: &str) -> bool {
    days_overdue(due_date, None) > 0
}

/// Get aging bucket for days overdue
#[wasm_bindgen]
pub fn get_aging_bucket(days: i32) -> String {
    let buckets = default_buckets();
    buckets.iter()
        .find(|b| {
            days >= b.from_days &&
            b.to_days.map(|to| days <= to).unwrap_or(true)
        })
        .map(|b| b.label.clone())
        .unwrap_or_else(|| "Unknown".to_string())
}

/// Calculate interest on overdue amount
#[wasm_bindgen]
pub fn calculate_overdue_interest(
    amount: &str,
    due_date: &str,
    interest_rate: &str,
    as_of_date: Option<String>,
) -> String {
    let amount: Decimal = amount.parse().unwrap_or(Decimal::ZERO);
    let rate: Decimal = interest_rate.parse().unwrap_or(Decimal::ZERO);

    let days = days_overdue(due_date, as_of_date);

    if days <= 0 {
        return "0".to_string();
    }

    // Simple interest: P × R × T / 365 / 100
    let interest = amount * rate * Decimal::from(days) / dec!(365) / dec!(100);
    interest.round_dp(2).to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_days_overdue() {
        let today = Utc::now().date_naive();
        let past = (today - chrono::Duration::days(30)).format("%Y-%m-%d").to_string();
        let future = (today + chrono::Duration::days(30)).format("%Y-%m-%d").to_string();

        assert!(days_overdue(&past, None) > 0);
        assert!(days_overdue(&future, None) < 0);
    }

    #[test]
    fn test_aging_bucket() {
        assert_eq!(get_aging_bucket(-5), "Current");
        assert_eq!(get_aging_bucket(15), "1-30 Days");
        assert_eq!(get_aging_bucket(45), "31-60 Days");
        assert_eq!(get_aging_bucket(75), "61-90 Days");
        assert_eq!(get_aging_bucket(100), "90+ Days");
    }

    #[test]
    fn test_overdue_interest() {
        // 10000 @ 18% for 30 days
        // Interest = 10000 × 18 × 30 / 365 / 100 = 147.95
        let today = Utc::now().date_naive();
        let past = (today - chrono::Duration::days(30)).format("%Y-%m-%d").to_string();

        let interest = calculate_overdue_interest("10000", &past, "18", None);
        let interest_val: f64 = interest.parse().unwrap();
        assert!(interest_val > 140.0 && interest_val < 150.0);
    }
}
