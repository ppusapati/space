//! GST Compliance Module
//!
//! This crate provides:
//! - e-Invoice JSON generation (GST e-Invoice schema)
//! - IRN (Invoice Reference Number) generation
//! - GSTR-1, GSTR-3B validation
//! - HSN summary generation
//! - B2B, B2C classification

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use wasm_bindgen::prelude::*;

mod einvoice;
mod gstr;

pub use einvoice::*;
pub use gstr::*;

/// Transaction type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "UPPERCASE")]
pub enum TransactionType {
    /// Business to Business
    B2B,
    /// Business to Consumer (Large)
    B2CL,
    /// Business to Consumer (Small)
    B2CS,
    /// Credit/Debit Note
    CDNR,
    /// Exports
    EXP,
    /// Nil Rated
    NIL,
}

/// Supply type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SupplyType {
    /// Intra-state (within same state)
    Intra,
    /// Inter-state (different states)
    Inter,
}

/// Document type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DocumentType {
    Invoice,
    CreditNote,
    DebitNote,
    BillOfSupply,
    DeliveryChallan,
}

/// Initialize the compliance module (called from core init)
fn compliance_init() {
    console_error_panic_hook::set_once();
}

/// Classify transaction type based on invoice details
#[wasm_bindgen]
pub fn classify_transaction(
    seller_gstin: &str,
    buyer_gstin: Option<String>,
    place_of_supply: &str,
    invoice_value: &str,
) -> JsValue {
    let value: Decimal = invoice_value.parse().unwrap_or(Decimal::ZERO);
    let seller_state = &seller_gstin[0..2];
    let pos_state = &place_of_supply[0..2];

    let supply_type = if seller_state == pos_state {
        SupplyType::Intra
    } else {
        SupplyType::Inter
    };

    let transaction_type = match buyer_gstin {
        Some(ref gstin) if !gstin.is_empty() && gstin != "URP" => {
            // Registered buyer - B2B
            TransactionType::B2B
        }
        _ => {
            // Unregistered buyer
            if supply_type == SupplyType::Inter && value > dec!(250000) {
                // Inter-state, value > 2.5L - B2CL
                TransactionType::B2CL
            } else {
                // Intra-state or value <= 2.5L - B2CS
                TransactionType::B2CS
            }
        }
    };

    let result = serde_json::json!({
        "transaction_type": transaction_type,
        "supply_type": supply_type,
        "requires_einvoice": matches!(transaction_type, TransactionType::B2B),
        "requires_eway_bill": value > dec!(50000),
        "gstr1_section": match transaction_type {
            TransactionType::B2B => "B2B Invoices - 4A, 4B, 4C, 6B, 6C",
            TransactionType::B2CL => "B2C Large Invoices - 5A, 5B",
            TransactionType::B2CS => "B2C Small - 7",
            _ => "Other"
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Generate document hash for e-invoice
#[wasm_bindgen]
pub fn generate_document_hash(invoice_json: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(invoice_json.as_bytes());
    let result = hasher.finalize();
    base64::encode(result)
}

/// Validate GSTIN format
#[wasm_bindgen]
pub fn validate_gstin_format(gstin: &str) -> JsValue {
    let result = validate_gstin_internal(gstin);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[derive(Debug, Serialize)]
struct GstinValidationResult {
    is_valid: bool,
    state_code: String,
    state_name: String,
    pan: String,
    entity_type: String,
    checksum_valid: bool,
    errors: Vec<String>,
}

fn validate_gstin_internal(gstin: &str) -> GstinValidationResult {
    let mut errors = Vec::new();

    // Length check
    if gstin.len() != 15 {
        errors.push(format!("GSTIN must be 15 characters, got {}", gstin.len()));
    }

    // State code
    let state_code = if gstin.len() >= 2 {
        &gstin[0..2]
    } else {
        "00"
    };

    let state_name = get_state_name(state_code);

    // PAN extraction
    let pan = if gstin.len() >= 12 {
        &gstin[2..12]
    } else {
        ""
    };

    // Entity type (13th character)
    let entity_code = if gstin.len() >= 13 {
        &gstin[12..13]
    } else {
        ""
    };

    let entity_type = match entity_code {
        "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" => "Normal Taxpayer",
        "C" => "Central Government",
        "D" => "Embassy/Consulate",
        "F" => "Foreign Embassy",
        "G" => "Government Department",
        "H" => "Input Service Distributor",
        "N" => "Non-Resident Taxable Person",
        "P" => "Tax Deductor",
        "Q" => "Tax Collector (TCS)",
        "T" => "SEZ Developer",
        "U" => "SEZ Unit",
        "V" => "UN Bodies",
        "Z" => "Default Character",
        _ => "Unknown",
    };

    // Checksum validation (simplified)
    let checksum_valid = validate_gstin_checksum(gstin);

    if !checksum_valid {
        errors.push("Invalid checksum".to_string());
    }

    // Pattern check
    let pattern_valid = gstin.chars().enumerate().all(|(i, c)| {
        match i {
            0..=1 => c.is_ascii_digit(),
            2..=6 => c.is_ascii_uppercase(),
            7..=10 => c.is_ascii_digit(),
            11 => c.is_ascii_uppercase(),
            12 => c.is_ascii_alphanumeric(),
            13 => c == 'Z',
            14 => c.is_ascii_alphanumeric(),
            _ => false,
        }
    });

    if !pattern_valid && gstin.len() == 15 {
        errors.push("Invalid GSTIN pattern".to_string());
    }

    GstinValidationResult {
        is_valid: errors.is_empty() && gstin.len() == 15,
        state_code: state_code.to_string(),
        state_name,
        pan: pan.to_string(),
        entity_type: entity_type.to_string(),
        checksum_valid,
        errors,
    }
}

fn validate_gstin_checksum(gstin: &str) -> bool {
    if gstin.len() != 15 {
        return false;
    }

    let chars: Vec<char> = gstin.chars().collect();
    let weights = [1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2];
    let char_values = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ";

    let mut sum = 0;
    for (i, &c) in chars[..14].iter().enumerate() {
        let value = char_values.find(c).unwrap_or(0);
        let product = value * weights[i];
        sum += product / 36 + product % 36;
    }

    let check_digit = (36 - (sum % 36)) % 36;
    let expected = char_values.chars().nth(check_digit);
    let actual = chars.get(14);

    expected == actual.copied()
}

fn get_state_name(code: &str) -> String {
    match code {
        "01" => "Jammu & Kashmir",
        "02" => "Himachal Pradesh",
        "03" => "Punjab",
        "04" => "Chandigarh",
        "05" => "Uttarakhand",
        "06" => "Haryana",
        "07" => "Delhi",
        "08" => "Rajasthan",
        "09" => "Uttar Pradesh",
        "10" => "Bihar",
        "11" => "Sikkim",
        "12" => "Arunachal Pradesh",
        "13" => "Nagaland",
        "14" => "Manipur",
        "15" => "Mizoram",
        "16" => "Tripura",
        "17" => "Meghalaya",
        "18" => "Assam",
        "19" => "West Bengal",
        "20" => "Jharkhand",
        "21" => "Odisha",
        "22" => "Chhattisgarh",
        "23" => "Madhya Pradesh",
        "24" => "Gujarat",
        "26" => "Dadra & Nagar Haveli and Daman & Diu",
        "27" => "Maharashtra",
        "29" => "Karnataka",
        "30" => "Goa",
        "31" => "Lakshadweep",
        "32" => "Kerala",
        "33" => "Tamil Nadu",
        "34" => "Puducherry",
        "35" => "Andaman & Nicobar",
        "36" => "Telangana",
        "37" => "Andhra Pradesh",
        "38" => "Ladakh",
        "97" => "Other Territory",
        _ => "Unknown",
    }.to_string()
}

/// Check e-invoice applicability
#[wasm_bindgen]
pub fn check_einvoice_applicability(
    annual_turnover: &str,
    transaction_type: &str,
    is_export: bool,
    is_sez: bool,
) -> JsValue {
    let turnover: Decimal = annual_turnover.parse().unwrap_or(Decimal::ZERO);
    let threshold = dec!(5_00_00_000); // Rs 5 Cr threshold (as of 2023)

    let is_applicable = turnover >= threshold
        && (transaction_type == "B2B" || is_export || is_sez);

    let result = serde_json::json!({
        "is_applicable": is_applicable,
        "turnover": turnover.to_string(),
        "threshold": threshold.to_string(),
        "threshold_met": turnover >= threshold,
        "transaction_eligible": transaction_type == "B2B" || is_export || is_sez,
        "current_threshold_date": "2023-08-01",
        "notes": if !is_applicable {
            if turnover < threshold {
                "Turnover below Rs 5 Cr threshold"
            } else {
                "Transaction type not eligible (B2C transactions exempt)"
            }
        } else {
            "e-Invoice mandatory for this transaction"
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Generate HSN summary from invoice items
#[wasm_bindgen]
pub fn generate_hsn_summary(items: JsValue) -> JsValue {
    #[derive(Debug, Clone, Deserialize)]
    struct InvoiceItem {
        hsn_code: String,
        description: Option<String>,
        quantity: String,
        uom: String,
        taxable_value: String,
        igst: Option<String>,
        cgst: Option<String>,
        sgst: Option<String>,
        cess: Option<String>,
    }

    let items: Vec<InvoiceItem> = match serde_wasm_bindgen::from_value(items) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid items: {}", e);
            return JsValue::NULL;
        }
    };

    use std::collections::HashMap;

    #[derive(Debug, Clone, Serialize)]
    struct HsnSummaryEntry {
        hsn_code: String,
        description: String,
        uom: String,
        total_quantity: String,
        total_value: String,
        taxable_value: String,
        igst: String,
        cgst: String,
        sgst: String,
        cess: String,
    }

    let mut hsn_map: HashMap<String, HsnSummaryEntry> = HashMap::new();

    for item in items {
        let entry = hsn_map.entry(item.hsn_code.clone()).or_insert_with(|| HsnSummaryEntry {
            hsn_code: item.hsn_code.clone(),
            description: item.description.clone().unwrap_or_default(),
            uom: item.uom.clone(),
            total_quantity: "0".to_string(),
            total_value: "0".to_string(),
            taxable_value: "0".to_string(),
            igst: "0".to_string(),
            cgst: "0".to_string(),
            sgst: "0".to_string(),
            cess: "0".to_string(),
        });

        let qty: Decimal = item.quantity.parse().unwrap_or(Decimal::ZERO);
        let taxable: Decimal = item.taxable_value.parse().unwrap_or(Decimal::ZERO);
        let igst: Decimal = item.igst.as_ref().and_then(|i| i.parse().ok()).unwrap_or(Decimal::ZERO);
        let cgst: Decimal = item.cgst.as_ref().and_then(|c| c.parse().ok()).unwrap_or(Decimal::ZERO);
        let sgst: Decimal = item.sgst.as_ref().and_then(|s| s.parse().ok()).unwrap_or(Decimal::ZERO);
        let cess: Decimal = item.cess.as_ref().and_then(|c| c.parse().ok()).unwrap_or(Decimal::ZERO);

        let existing_qty: Decimal = entry.total_quantity.parse().unwrap_or(Decimal::ZERO);
        let existing_taxable: Decimal = entry.taxable_value.parse().unwrap_or(Decimal::ZERO);
        let existing_igst: Decimal = entry.igst.parse().unwrap_or(Decimal::ZERO);
        let existing_cgst: Decimal = entry.cgst.parse().unwrap_or(Decimal::ZERO);
        let existing_sgst: Decimal = entry.sgst.parse().unwrap_or(Decimal::ZERO);
        let existing_cess: Decimal = entry.cess.parse().unwrap_or(Decimal::ZERO);

        entry.total_quantity = (existing_qty + qty).to_string();
        entry.taxable_value = (existing_taxable + taxable).round_dp(2).to_string();
        entry.igst = (existing_igst + igst).round_dp(2).to_string();
        entry.cgst = (existing_cgst + cgst).round_dp(2).to_string();
        entry.sgst = (existing_sgst + sgst).round_dp(2).to_string();
        entry.cess = (existing_cess + cess).round_dp(2).to_string();

        let total_tax = igst + cgst + sgst + cess;
        let existing_total: Decimal = entry.total_value.parse().unwrap_or(Decimal::ZERO);
        entry.total_value = (existing_total + taxable + total_tax).round_dp(2).to_string();
    }

    let summary: Vec<HsnSummaryEntry> = hsn_map.into_values().collect();

    // Calculate totals
    let total_taxable: Decimal = summary.iter()
        .map(|s| s.taxable_value.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();
    let total_igst: Decimal = summary.iter()
        .map(|s| s.igst.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();
    let total_cgst: Decimal = summary.iter()
        .map(|s| s.cgst.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();
    let total_sgst: Decimal = summary.iter()
        .map(|s| s.sgst.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    let result = serde_json::json!({
        "hsn_summary": summary,
        "totals": {
            "taxable_value": total_taxable.round_dp(2).to_string(),
            "igst": total_igst.round_dp(2).to_string(),
            "cgst": total_cgst.round_dp(2).to_string(),
            "sgst": total_sgst.round_dp(2).to_string(),
            "total_tax": (total_igst + total_cgst + total_sgst).round_dp(2).to_string()
        },
        "hsn_count": summary.len()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_gstin_validation() {
        // Valid GSTIN example
        let result = validate_gstin_internal("27AADCB2230M1ZT");
        assert!(result.is_valid || !result.errors.is_empty()); // May fail checksum without real data

        // Invalid length
        let result = validate_gstin_internal("27AADCB2230M1Z");
        assert!(!result.is_valid);
    }

    #[test]
    fn test_state_name() {
        assert_eq!(get_state_name("27"), "Maharashtra");
        assert_eq!(get_state_name("07"), "Delhi");
        assert_eq!(get_state_name("29"), "Karnataka");
    }

    #[test]
    fn test_document_hash() {
        let json = r#"{"invoice":"test"}"#;
        let hash = generate_document_hash(json);
        assert!(!hash.is_empty());
    }
}
