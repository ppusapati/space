//! GST (Goods and Services Tax) calculations
//!
//! Handles CGST, SGST, IGST, UTGST, and Cess calculations
//! based on Indian GST law.

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use samavaya_core::{is_intra_state, is_union_territory};

/// GST rates as per Indian tax law
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum GstRate {
    Exempt,
    Rate0,
    Rate0_1,
    Rate0_25,
    Rate1,
    Rate1_5,
    Rate3,
    Rate5,
    Rate6,
    Rate7_5,
    Rate12,
    Rate18,
    Rate28,
}

impl GstRate {
    pub fn value(&self) -> Decimal {
        match self {
            GstRate::Exempt => dec!(0),
            GstRate::Rate0 => dec!(0),
            GstRate::Rate0_1 => dec!(0.1),
            GstRate::Rate0_25 => dec!(0.25),
            GstRate::Rate1 => dec!(1),
            GstRate::Rate1_5 => dec!(1.5),
            GstRate::Rate3 => dec!(3),
            GstRate::Rate5 => dec!(5),
            GstRate::Rate6 => dec!(6),
            GstRate::Rate7_5 => dec!(7.5),
            GstRate::Rate12 => dec!(12),
            GstRate::Rate18 => dec!(18),
            GstRate::Rate28 => dec!(28),
        }
    }

    pub fn from_value(rate: Decimal) -> Self {
        if rate == dec!(0) {
            GstRate::Rate0
        } else if rate == dec!(0.1) {
            GstRate::Rate0_1
        } else if rate == dec!(0.25) {
            GstRate::Rate0_25
        } else if rate == dec!(1) {
            GstRate::Rate1
        } else if rate == dec!(1.5) {
            GstRate::Rate1_5
        } else if rate == dec!(3) {
            GstRate::Rate3
        } else if rate == dec!(5) {
            GstRate::Rate5
        } else if rate == dec!(6) {
            GstRate::Rate6
        } else if rate == dec!(7.5) {
            GstRate::Rate7_5
        } else if rate == dec!(12) {
            GstRate::Rate12
        } else if rate == dec!(18) {
            GstRate::Rate18
        } else if rate == dec!(28) {
            GstRate::Rate28
        } else {
            GstRate::Exempt
        }
    }
}

/// GST calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstInput {
    /// Taxable amount (base amount before tax)
    pub amount: String,
    /// GST rate as percentage (e.g., "18" for 18%)
    pub gst_rate: String,
    /// Source state code (e.g., "TG" for Telangana)
    pub source_state: String,
    /// Destination state code
    pub dest_state: String,
    /// HSN/SAC code (optional, for cess calculation)
    pub hsn_code: Option<String>,
    /// Whether this is a composition dealer
    #[serde(default)]
    pub is_composition: bool,
    /// Whether reverse charge applies
    #[serde(default)]
    pub is_reverse_charge: bool,
    /// Whether to include cess
    #[serde(default)]
    pub include_cess: bool,
    /// Custom cess rate (if applicable)
    pub cess_rate: Option<String>,
    /// Additional cess amount (specific cess like coal)
    pub additional_cess: Option<String>,
}

/// GST calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstResult {
    /// Original taxable amount
    pub taxable_amount: String,
    /// CGST amount (Central GST)
    pub cgst: String,
    /// CGST rate
    pub cgst_rate: String,
    /// SGST amount (State GST)
    pub sgst: String,
    /// SGST rate
    pub sgst_rate: String,
    /// UTGST amount (Union Territory GST)
    pub utgst: String,
    /// UTGST rate
    pub utgst_rate: String,
    /// IGST amount (Integrated GST)
    pub igst: String,
    /// IGST rate
    pub igst_rate: String,
    /// Cess amount
    pub cess: String,
    /// Cess rate
    pub cess_rate: String,
    /// Total tax amount
    pub total_tax: String,
    /// Grand total (amount + tax)
    pub grand_total: String,
    /// Whether this is inter-state (IGST) or intra-state (CGST+SGST)
    pub is_inter_state: bool,
    /// Whether UTGST applies (for Union Territories)
    pub is_ut: bool,
    /// Whether reverse charge mechanism applies
    pub is_reverse_charge: bool,
}

impl Default for GstResult {
    fn default() -> Self {
        Self {
            taxable_amount: "0".to_string(),
            cgst: "0".to_string(),
            cgst_rate: "0".to_string(),
            sgst: "0".to_string(),
            sgst_rate: "0".to_string(),
            utgst: "0".to_string(),
            utgst_rate: "0".to_string(),
            igst: "0".to_string(),
            igst_rate: "0".to_string(),
            cess: "0".to_string(),
            cess_rate: "0".to_string(),
            total_tax: "0".to_string(),
            grand_total: "0".to_string(),
            is_inter_state: false,
            is_ut: false,
            is_reverse_charge: false,
        }
    }
}

/// Calculate GST
#[wasm_bindgen]
pub fn calculate_gst(input: JsValue) -> Result<JsValue, JsError> {
    let input: GstInput = serde_wasm_bindgen::from_value(input)
        .map_err(|e| JsError::new(&format!("Invalid input: {}", e)))?;

    let result = calculate_gst_internal(&input)
        .map_err(|e| JsError::new(&e))?;

    serde_wasm_bindgen::to_value(&result)
        .map_err(|e| JsError::new(&format!("Serialization error: {}", e)))
}

/// Internal GST calculation
pub fn calculate_gst_internal(input: &GstInput) -> Result<GstResult, String> {
    let amount: Decimal = input.amount.parse()
        .map_err(|_| "Invalid amount")?;

    let gst_rate: Decimal = input.gst_rate.parse()
        .map_err(|_| "Invalid GST rate")?;

    // Validate rate
    if gst_rate < dec!(0) || gst_rate > dec!(28) {
        return Err("GST rate must be between 0 and 28".to_string());
    }

    let is_inter_state = !is_intra_state(&input.source_state, &input.dest_state);
    let is_ut = is_union_territory(&input.dest_state);

    let mut result = GstResult {
        taxable_amount: amount.round_dp(2).to_string(),
        is_inter_state,
        is_ut,
        is_reverse_charge: input.is_reverse_charge,
        ..Default::default()
    };

    // Calculate base GST
    let total_gst = (amount * gst_rate / dec!(100)).round_dp(2);

    if is_inter_state {
        // Inter-state: IGST applies
        result.igst = total_gst.to_string();
        result.igst_rate = gst_rate.to_string();
    } else if is_ut {
        // Union Territory: CGST + UTGST
        let half_rate = (gst_rate / dec!(2)).round_dp(2);
        let cgst = (amount * half_rate / dec!(100)).round_dp(2);
        let utgst = total_gst - cgst; // Ensure no rounding loss

        result.cgst = cgst.to_string();
        result.cgst_rate = half_rate.to_string();
        result.utgst = utgst.to_string();
        result.utgst_rate = half_rate.to_string();
    } else {
        // Intra-state: CGST + SGST
        let half_rate = (gst_rate / dec!(2)).round_dp(2);
        let cgst = (amount * half_rate / dec!(100)).round_dp(2);
        let sgst = total_gst - cgst; // Ensure no rounding loss

        result.cgst = cgst.to_string();
        result.cgst_rate = half_rate.to_string();
        result.sgst = sgst.to_string();
        result.sgst_rate = half_rate.to_string();
    }

    // Calculate Cess if applicable
    let mut cess = Decimal::ZERO;
    if input.include_cess {
        if let Some(ref cess_rate_str) = input.cess_rate {
            if let Ok(cess_rate) = cess_rate_str.parse::<Decimal>() {
                cess = (amount * cess_rate / dec!(100)).round_dp(2);
                result.cess_rate = cess_rate.to_string();
            }
        }
        // Add specific cess if provided
        if let Some(ref additional) = input.additional_cess {
            if let Ok(add_cess) = additional.parse::<Decimal>() {
                cess += add_cess;
            }
        }
    }
    result.cess = cess.to_string();

    // Calculate totals
    let total_tax = total_gst + cess;
    result.total_tax = total_tax.round_dp(2).to_string();
    result.grand_total = (amount + total_tax).round_dp(2).to_string();

    Ok(result)
}

/// Calculate GST from inclusive amount (reverse calculation)
#[wasm_bindgen]
pub fn calculate_gst_inclusive(inclusive_amount: &str, gst_rate: &str, source_state: &str, dest_state: &str) -> Result<JsValue, JsError> {
    let inclusive: Decimal = inclusive_amount.parse()
        .map_err(|_| JsError::new("Invalid inclusive amount"))?;
    let rate: Decimal = gst_rate.parse()
        .map_err(|_| JsError::new("Invalid GST rate"))?;

    // Calculate taxable amount from inclusive
    // Inclusive = Taxable × (1 + rate/100)
    // Taxable = Inclusive / (1 + rate/100)
    let multiplier = dec!(1) + (rate / dec!(100));
    let taxable = (inclusive / multiplier).round_dp(2);

    let input = GstInput {
        amount: taxable.to_string(),
        gst_rate: gst_rate.to_string(),
        source_state: source_state.to_string(),
        dest_state: dest_state.to_string(),
        hsn_code: None,
        is_composition: false,
        is_reverse_charge: false,
        include_cess: false,
        cess_rate: None,
        additional_cess: None,
    };

    let result = calculate_gst_internal(&input)
        .map_err(|e| JsError::new(&e))?;

    serde_wasm_bindgen::to_value(&result)
        .map_err(|e| JsError::new(&format!("Serialization error: {}", e)))
}

/// Bulk GST calculation for line items
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstLineItem {
    pub item_id: String,
    pub amount: String,
    pub gst_rate: String,
    pub hsn_code: Option<String>,
    pub cess_rate: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstLineResult {
    pub item_id: String,
    pub taxable_amount: String,
    pub cgst: String,
    pub sgst: String,
    pub utgst: String,
    pub igst: String,
    pub cess: String,
    pub total_tax: String,
    pub line_total: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BulkGstInput {
    pub source_state: String,
    pub dest_state: String,
    pub items: Vec<GstLineItem>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BulkGstResult {
    pub items: Vec<GstLineResult>,
    pub summary: GstSummary,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstSummary {
    pub total_taxable: String,
    pub total_cgst: String,
    pub total_sgst: String,
    pub total_utgst: String,
    pub total_igst: String,
    pub total_cess: String,
    pub total_tax: String,
    pub grand_total: String,
    pub is_inter_state: bool,
}

/// Calculate GST for multiple line items
#[wasm_bindgen]
pub fn calculate_bulk_gst(input: JsValue) -> Result<JsValue, JsError> {
    let input: BulkGstInput = serde_wasm_bindgen::from_value(input)
        .map_err(|e| JsError::new(&format!("Invalid input: {}", e)))?;

    let is_inter_state = !is_intra_state(&input.source_state, &input.dest_state);
    let is_ut = is_union_territory(&input.dest_state);

    let mut items = Vec::new();
    let mut total_taxable = Decimal::ZERO;
    let mut total_cgst = Decimal::ZERO;
    let mut total_sgst = Decimal::ZERO;
    let mut total_utgst = Decimal::ZERO;
    let mut total_igst = Decimal::ZERO;
    let mut total_cess = Decimal::ZERO;

    for item in &input.items {
        let gst_input = GstInput {
            amount: item.amount.clone(),
            gst_rate: item.gst_rate.clone(),
            source_state: input.source_state.clone(),
            dest_state: input.dest_state.clone(),
            hsn_code: item.hsn_code.clone(),
            is_composition: false,
            is_reverse_charge: false,
            include_cess: item.cess_rate.is_some(),
            cess_rate: item.cess_rate.clone(),
            additional_cess: None,
        };

        let result = calculate_gst_internal(&gst_input)
            .map_err(|e| JsError::new(&e))?;

        let taxable: Decimal = result.taxable_amount.parse().unwrap_or_default();
        let cgst: Decimal = result.cgst.parse().unwrap_or_default();
        let sgst: Decimal = result.sgst.parse().unwrap_or_default();
        let utgst: Decimal = result.utgst.parse().unwrap_or_default();
        let igst: Decimal = result.igst.parse().unwrap_or_default();
        let cess: Decimal = result.cess.parse().unwrap_or_default();
        let total_tax: Decimal = result.total_tax.parse().unwrap_or_default();

        total_taxable += taxable;
        total_cgst += cgst;
        total_sgst += sgst;
        total_utgst += utgst;
        total_igst += igst;
        total_cess += cess;

        items.push(GstLineResult {
            item_id: item.item_id.clone(),
            taxable_amount: result.taxable_amount,
            cgst: result.cgst,
            sgst: result.sgst,
            utgst: result.utgst,
            igst: result.igst,
            cess: result.cess,
            total_tax: total_tax.to_string(),
            line_total: result.grand_total,
        });
    }

    let total_tax = total_cgst + total_sgst + total_utgst + total_igst + total_cess;

    let result = BulkGstResult {
        items,
        summary: GstSummary {
            total_taxable: total_taxable.round_dp(2).to_string(),
            total_cgst: total_cgst.round_dp(2).to_string(),
            total_sgst: total_sgst.round_dp(2).to_string(),
            total_utgst: total_utgst.round_dp(2).to_string(),
            total_igst: total_igst.round_dp(2).to_string(),
            total_cess: total_cess.round_dp(2).to_string(),
            total_tax: total_tax.round_dp(2).to_string(),
            grand_total: (total_taxable + total_tax).round_dp(2).to_string(),
            is_inter_state,
        },
    };

    serde_wasm_bindgen::to_value(&result)
        .map_err(|e| JsError::new(&format!("Serialization error: {}", e)))
}

/// Simple GST calculation (convenience function)
#[wasm_bindgen]
pub fn simple_gst(amount: &str, rate: &str, source_state: &str, dest_state: &str) -> String {
    let input = GstInput {
        amount: amount.to_string(),
        gst_rate: rate.to_string(),
        source_state: source_state.to_string(),
        dest_state: dest_state.to_string(),
        hsn_code: None,
        is_composition: false,
        is_reverse_charge: false,
        include_cess: false,
        cess_rate: None,
        additional_cess: None,
    };

    match calculate_gst_internal(&input) {
        Ok(result) => result.total_tax,
        Err(_) => "0".to_string(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_intra_state_gst() {
        let input = GstInput {
            amount: "10000".to_string(),
            gst_rate: "18".to_string(),
            source_state: "TG".to_string(),
            dest_state: "TG".to_string(),
            hsn_code: None,
            is_composition: false,
            is_reverse_charge: false,
            include_cess: false,
            cess_rate: None,
            additional_cess: None,
        };

        let result = calculate_gst_internal(&input).unwrap();
        assert_eq!(result.cgst, "900.00");
        assert_eq!(result.sgst, "900.00");
        assert_eq!(result.igst, "0");
        assert_eq!(result.total_tax, "1800.00");
        assert_eq!(result.grand_total, "11800.00");
        assert!(!result.is_inter_state);
    }

    #[test]
    fn test_inter_state_gst() {
        let input = GstInput {
            amount: "10000".to_string(),
            gst_rate: "18".to_string(),
            source_state: "TG".to_string(),
            dest_state: "KA".to_string(),
            hsn_code: None,
            is_composition: false,
            is_reverse_charge: false,
            include_cess: false,
            cess_rate: None,
            additional_cess: None,
        };

        let result = calculate_gst_internal(&input).unwrap();
        assert_eq!(result.cgst, "0");
        assert_eq!(result.sgst, "0");
        assert_eq!(result.igst, "1800.00");
        assert_eq!(result.total_tax, "1800.00");
        assert_eq!(result.grand_total, "11800.00");
        assert!(result.is_inter_state);
    }

    #[test]
    fn test_union_territory_gst() {
        let input = GstInput {
            amount: "10000".to_string(),
            gst_rate: "18".to_string(),
            source_state: "DL".to_string(),
            dest_state: "DL".to_string(),
            hsn_code: None,
            is_composition: false,
            is_reverse_charge: false,
            include_cess: false,
            cess_rate: None,
            additional_cess: None,
        };

        let result = calculate_gst_internal(&input).unwrap();
        assert_eq!(result.cgst, "900.00");
        assert_eq!(result.utgst, "900.00");
        assert_eq!(result.sgst, "0");
        assert!(result.is_ut);
    }
}
