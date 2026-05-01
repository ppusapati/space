//! Common data types used across WASM modules

use chrono::{Datelike, NaiveDate};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Date range for queries and calculations
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DateRange {
    pub start: String,
    pub end: String,
}

impl DateRange {
    pub fn new(start: &str, end: &str) -> Self {
        Self {
            start: start.to_string(),
            end: end.to_string(),
        }
    }

    pub fn parse(&self) -> Option<(NaiveDate, NaiveDate)> {
        let start = NaiveDate::parse_from_str(&self.start, "%Y-%m-%d").ok()?;
        let end = NaiveDate::parse_from_str(&self.end, "%Y-%m-%d").ok()?;
        Some((start, end))
    }

    pub fn days(&self) -> Option<i64> {
        self.parse().map(|(start, end)| (end - start).num_days())
    }
}

/// Financial year in Indian format (April to March)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FinancialYear {
    pub year: i32,       // Starting year (e.g., 2024 for FY 2024-25)
    pub start: String,   // YYYY-MM-DD
    pub end: String,     // YYYY-MM-DD
    pub display: String, // FY 2024-25
}

impl FinancialYear {
    pub fn new(year: i32) -> Self {
        Self {
            year,
            start: format!("{}-04-01", year),
            end: format!("{}-03-31", year + 1),
            display: format!("FY {}-{}", year, (year + 1) % 100),
        }
    }

    pub fn from_date(date: &str) -> Option<Self> {
        let date = NaiveDate::parse_from_str(date, "%Y-%m-%d").ok()?;
        let year = if date.month() >= 4 {
            date.year()
        } else {
            date.year() - 1
        };
        Some(Self::new(year))
    }

    pub fn current() -> Self {
        let now = chrono::Utc::now().date_naive();
        let year = if now.month() >= 4 {
            now.year()
        } else {
            now.year() - 1
        };
        Self::new(year)
    }
}

/// Quarter within a financial year
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Quarter {
    Q1, // Apr-Jun
    Q2, // Jul-Sep
    Q3, // Oct-Dec
    Q4, // Jan-Mar
}

impl Quarter {
    pub fn from_month(month: u32) -> Self {
        match month {
            4..=6 => Quarter::Q1,
            7..=9 => Quarter::Q2,
            10..=12 => Quarter::Q3,
            _ => Quarter::Q4,
        }
    }

    pub fn months(&self) -> (u32, u32, u32) {
        match self {
            Quarter::Q1 => (4, 5, 6),
            Quarter::Q2 => (7, 8, 9),
            Quarter::Q3 => (10, 11, 12),
            Quarter::Q4 => (1, 2, 3),
        }
    }
}

/// Address structure for Indian addresses
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct IndianAddress {
    pub line1: String,
    pub line2: Option<String>,
    pub city: String,
    pub district: Option<String>,
    pub state_code: String,
    pub pincode: String,
    pub country: String,
}

/// Contact information
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Contact {
    pub name: Option<String>,
    pub phone: Option<String>,
    pub email: Option<String>,
    pub designation: Option<String>,
}

/// Entity reference (for linking)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EntityRef {
    pub id: String,
    pub code: String,
    pub name: String,
    pub entity_type: String,
}

/// Pagination parameters
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Pagination {
    pub page: u32,
    pub page_size: u32,
    pub total: u32,
    pub total_pages: u32,
}

impl Pagination {
    pub fn new(page: u32, page_size: u32, total: u32) -> Self {
        let total_pages = (total as f64 / page_size as f64).ceil() as u32;
        Self {
            page,
            page_size,
            total,
            total_pages,
        }
    }
}

/// Sort direction
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SortDirection {
    Asc,
    Desc,
}

/// Sort specification
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sort {
    pub field: String,
    pub direction: SortDirection,
}

// WASM exports for date utilities

#[wasm_bindgen]
pub fn get_financial_year(date: &str) -> JsValue {
    match FinancialYear::from_date(date) {
        Some(fy) => serde_wasm_bindgen::to_value(&fy).unwrap_or(JsValue::NULL),
        None => JsValue::NULL,
    }
}

#[wasm_bindgen]
pub fn get_current_financial_year() -> JsValue {
    let fy = FinancialYear::current();
    serde_wasm_bindgen::to_value(&fy).unwrap_or(JsValue::NULL)
}

#[wasm_bindgen]
pub fn get_quarter_from_date(date: &str) -> String {
    if let Ok(date) = NaiveDate::parse_from_str(date, "%Y-%m-%d") {
        let q = Quarter::from_month(date.month());
        match q {
            Quarter::Q1 => "Q1".to_string(),
            Quarter::Q2 => "Q2".to_string(),
            Quarter::Q3 => "Q3".to_string(),
            Quarter::Q4 => "Q4".to_string(),
        }
    } else {
        "".to_string()
    }
}

#[wasm_bindgen]
pub fn days_between(start: &str, end: &str) -> i64 {
    let range = DateRange::new(start, end);
    range.days().unwrap_or(0)
}

#[wasm_bindgen]
pub fn format_date_indian(date: &str) -> String {
    if let Ok(date) = NaiveDate::parse_from_str(date, "%Y-%m-%d") {
        date.format("%d-%m-%Y").to_string()
    } else {
        date.to_string()
    }
}

#[wasm_bindgen]
pub fn format_date_iso(date: &str) -> String {
    // Try Indian format first
    if let Ok(date) = NaiveDate::parse_from_str(date, "%d-%m-%Y") {
        return date.format("%Y-%m-%d").to_string();
    }
    // Try slash format
    if let Ok(date) = NaiveDate::parse_from_str(date, "%d/%m/%Y") {
        return date.format("%Y-%m-%d").to_string();
    }
    date.to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_financial_year() {
        let fy = FinancialYear::from_date("2024-06-15").unwrap();
        assert_eq!(fy.year, 2024);
        assert_eq!(fy.display, "FY 2024-25");

        let fy = FinancialYear::from_date("2025-02-15").unwrap();
        assert_eq!(fy.year, 2024);
        assert_eq!(fy.display, "FY 2024-25");
    }

    #[test]
    fn test_quarter() {
        assert_eq!(Quarter::from_month(4), Quarter::Q1);
        assert_eq!(Quarter::from_month(7), Quarter::Q2);
        assert_eq!(Quarter::from_month(10), Quarter::Q3);
        assert_eq!(Quarter::from_month(1), Quarter::Q4);
    }
}
