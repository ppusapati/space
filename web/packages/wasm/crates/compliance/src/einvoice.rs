//! e-Invoice (Electronic Invoice) generation for GST

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use wasm_bindgen::prelude::*;

/// e-Invoice transaction details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct EInvoiceTransaction {
    pub version: String,
    pub tran_dtls: TransactionDetails,
    pub doc_dtls: DocumentDetails,
    pub seller_dtls: PartyDetails,
    pub buyer_dtls: PartyDetails,
    pub dispatch_dtls: Option<DispatchDetails>,
    pub ship_dtls: Option<ShipDetails>,
    pub item_list: Vec<EInvoiceItem>,
    pub val_dtls: ValueDetails,
    pub pay_dtls: Option<PaymentDetails>,
    pub ref_dtls: Option<ReferenceDetails>,
    pub addl_doc_dtls: Option<Vec<AdditionalDocument>>,
    pub exp_dtls: Option<ExportDetails>,
    pub eway_bill_dtls: Option<EwayBillDetails>,
}

/// Transaction details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct TransactionDetails {
    pub tax_sch: String, // "GST"
    pub supply_typ: String, // B2B, SEZWP, SEZWOP, EXPWP, EXPWOP, DEXP
    pub reg_rev: Option<String>, // Y/N for reverse charge
    pub ec_om_gstin: Option<String>, // E-commerce GSTIN
    pub igst_on_intra: Option<String>, // Y/N for IGST on intra-state
}

/// Document details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DocumentDetails {
    pub typ: String, // INV, CRN, DBN
    pub no: String,
    pub dt: String, // DD/MM/YYYY
}

/// Party details (Seller/Buyer)
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct PartyDetails {
    pub gstin: String,
    pub lgl_nm: String, // Legal name
    pub trd_nm: Option<String>, // Trade name
    pub addr1: String,
    pub addr2: Option<String>,
    pub loc: String, // City
    pub pin: u32,
    pub stcd: String, // State code
    pub ph: Option<String>,
    pub em: Option<String>,
}

/// Dispatch details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DispatchDetails {
    pub nm: String,
    pub addr1: String,
    pub addr2: Option<String>,
    pub loc: String,
    pub pin: u32,
    pub stcd: String,
}

/// Ship-to details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct ShipDetails {
    pub gstin: Option<String>,
    pub lgl_nm: String,
    pub trd_nm: Option<String>,
    pub addr1: String,
    pub addr2: Option<String>,
    pub loc: String,
    pub pin: u32,
    pub stcd: String,
}

/// Invoice item
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct EInvoiceItem {
    pub sl_no: String,
    pub prd_desc: String,
    pub is_servc: String, // Y/N
    pub hsn_cd: String,
    pub barcde: Option<String>,
    pub qty: Decimal,
    pub free_qty: Option<Decimal>,
    pub unit: String,
    pub unit_price: Decimal,
    pub tot_amt: Decimal,
    pub discount: Option<Decimal>,
    pub pre_tax_val: Option<Decimal>,
    pub ass_amt: Decimal, // Assessable amount
    pub gst_rt: Decimal,
    pub igst_amt: Option<Decimal>,
    pub cgst_amt: Option<Decimal>,
    pub sgst_amt: Option<Decimal>,
    pub cess_rt: Option<Decimal>,
    pub cess_amt: Option<Decimal>,
    pub cess_non_advol_amt: Option<Decimal>,
    pub state_cess_rt: Option<Decimal>,
    pub state_cess_amt: Option<Decimal>,
    pub state_cess_non_advol_amt: Option<Decimal>,
    pub oth_chrg: Option<Decimal>,
    pub tot_item_val: Decimal,
    pub ord_line_ref: Option<String>,
    pub org_cntry: Option<String>,
    pub prd_sl_no: Option<String>,
    pub batch_dtls: Option<BatchDetails>,
    pub attrib_dtls: Option<Vec<AttributeDetails>>,
}

/// Batch details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct BatchDetails {
    pub nm: String,
    pub exp_dt: Option<String>,
    pub wrntdt: Option<String>,
}

/// Attribute details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct AttributeDetails {
    pub nm: String,
    pub val: String,
}

/// Value details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct ValueDetails {
    pub ass_val: Decimal,
    pub cgst_val: Option<Decimal>,
    pub sgst_val: Option<Decimal>,
    pub igst_val: Option<Decimal>,
    pub cess_val: Option<Decimal>,
    pub st_cess_val: Option<Decimal>,
    pub discount: Option<Decimal>,
    pub oth_chrg: Option<Decimal>,
    pub rnd_off_amt: Option<Decimal>,
    pub tot_inv_val: Decimal,
    pub tot_inv_val_fc: Option<Decimal>,
}

/// Payment details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct PaymentDetails {
    pub nm: Option<String>,
    pub accdet: Option<String>,
    pub mode: Option<String>,
    pub fin_instr_id: Option<String>,
    pub payterm: Option<String>,
    pub payinstr: Option<String>,
    pub cr_trn: Option<String>,
    pub dir_dr: Option<String>,
    pub cr_days: Option<u32>,
    pub paid_amt: Option<Decimal>,
    pub paymtdue: Option<Decimal>,
}

/// Reference details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct ReferenceDetails {
    pub inv_rm: Option<String>,
    pub doc_perd_dtls: Option<DocPeriodDetails>,
    pub prec_doc_dtls: Option<Vec<PrecedingDocDetails>>,
    pub contr_dtls: Option<Vec<ContractDetails>>,
}

/// Document period details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DocPeriodDetails {
    pub inv_st_dt: String,
    pub inv_end_dt: String,
}

/// Preceding document details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct PrecedingDocDetails {
    pub inv_no: String,
    pub inv_dt: String,
    pub oth_ref_no: Option<String>,
}

/// Contract details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct ContractDetails {
    pub rec_adv_ref: Option<String>,
    pub rec_adv_dt: Option<String>,
    pub tendref: Option<String>,
    pub contrref: Option<String>,
    pub extref: Option<String>,
    pub projref: Option<String>,
    pub poref: Option<String>,
    pub porefdt: Option<String>,
}

/// Additional document
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct AdditionalDocument {
    pub url: String,
    pub docs: String,
    pub info: Option<String>,
}

/// Export details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct ExportDetails {
    pub ship_bn: Option<String>,
    pub ship_bdt: Option<String>,
    pub port: Option<String>,
    pub ref_clm: Option<String>,
    pub for_cur: Option<String>,
    pub cnt_code: Option<String>,
    pub exp_duty: Option<Decimal>,
}

/// E-way bill details
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct EwayBillDetails {
    pub trans_id: Option<String>,
    pub trans_name: Option<String>,
    pub trans_mode: Option<String>,
    pub distance: u32,
    pub trans_doc_no: Option<String>,
    pub trans_doc_dt: Option<String>,
    pub veh_no: Option<String>,
    pub veh_type: Option<String>,
}

/// e-Invoice input (simplified)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EInvoiceInput {
    pub invoice_no: String,
    pub invoice_date: String,
    pub supply_type: String,
    pub reverse_charge: bool,
    pub seller: PartyInput,
    pub buyer: PartyInput,
    pub items: Vec<ItemInput>,
    pub discount: Option<String>,
    pub other_charges: Option<String>,
    pub round_off: Option<String>,
    pub is_export: bool,
    pub export_details: Option<ExportInput>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PartyInput {
    pub gstin: String,
    pub legal_name: String,
    pub trade_name: Option<String>,
    pub address1: String,
    pub address2: Option<String>,
    pub city: String,
    pub pincode: String,
    pub state_code: String,
    pub phone: Option<String>,
    pub email: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItemInput {
    pub description: String,
    pub hsn_code: String,
    pub quantity: String,
    pub uom: String,
    pub unit_price: String,
    pub discount: Option<String>,
    pub gst_rate: String,
    pub cess_rate: Option<String>,
    pub is_service: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExportInput {
    pub port_code: Option<String>,
    pub shipping_bill_no: Option<String>,
    pub shipping_bill_date: Option<String>,
    pub country_code: Option<String>,
    pub currency: Option<String>,
}

/// Generate e-Invoice JSON
#[wasm_bindgen]
pub fn generate_einvoice_json(input: JsValue) -> JsValue {
    let input: EInvoiceInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid e-Invoice input: {}", e);
            return JsValue::NULL;
        }
    };

    let einvoice = build_einvoice(&input);
    serde_wasm_bindgen::to_value(&einvoice).unwrap_or(JsValue::NULL)
}

fn build_einvoice(input: &EInvoiceInput) -> EInvoiceTransaction {
    let seller_state = &input.seller.state_code;
    let buyer_state = &input.buyer.state_code;
    let is_inter_state = seller_state != buyer_state || input.is_export;

    // Build items
    let mut items: Vec<EInvoiceItem> = Vec::new();
    let mut total_ass_val = Decimal::ZERO;
    let mut total_igst = Decimal::ZERO;
    let mut total_cgst = Decimal::ZERO;
    let mut total_sgst = Decimal::ZERO;
    let mut total_cess = Decimal::ZERO;

    for (idx, item) in input.items.iter().enumerate() {
        let qty: Decimal = item.quantity.parse().unwrap_or(dec!(1));
        let unit_price: Decimal = item.unit_price.parse().unwrap_or(Decimal::ZERO);
        let discount: Decimal = item.discount.as_ref()
            .and_then(|d| d.parse().ok())
            .unwrap_or(Decimal::ZERO);
        let gst_rate: Decimal = item.gst_rate.parse().unwrap_or(Decimal::ZERO);
        let cess_rate: Decimal = item.cess_rate.as_ref()
            .and_then(|c| c.parse().ok())
            .unwrap_or(Decimal::ZERO);

        let tot_amt = (qty * unit_price).round_dp(2);
        let ass_amt = (tot_amt - discount).round_dp(2);

        let (igst, cgst, sgst) = if is_inter_state {
            let igst = (ass_amt * gst_rate / dec!(100)).round_dp(2);
            (Some(igst), None, None)
        } else {
            let half_rate = gst_rate / dec!(2);
            let cgst = (ass_amt * half_rate / dec!(100)).round_dp(2);
            let sgst = (ass_amt * half_rate / dec!(100)).round_dp(2);
            (None, Some(cgst), Some(sgst))
        };

        let cess_amt = if cess_rate > Decimal::ZERO {
            Some((ass_amt * cess_rate / dec!(100)).round_dp(2))
        } else {
            None
        };

        let tot_item_val = ass_amt
            + igst.unwrap_or(Decimal::ZERO)
            + cgst.unwrap_or(Decimal::ZERO)
            + sgst.unwrap_or(Decimal::ZERO)
            + cess_amt.unwrap_or(Decimal::ZERO);

        total_ass_val += ass_amt;
        total_igst += igst.unwrap_or(Decimal::ZERO);
        total_cgst += cgst.unwrap_or(Decimal::ZERO);
        total_sgst += sgst.unwrap_or(Decimal::ZERO);
        total_cess += cess_amt.unwrap_or(Decimal::ZERO);

        items.push(EInvoiceItem {
            sl_no: (idx + 1).to_string(),
            prd_desc: item.description.clone(),
            is_servc: if item.is_service { "Y".to_string() } else { "N".to_string() },
            hsn_cd: item.hsn_code.clone(),
            barcde: None,
            qty,
            free_qty: None,
            unit: item.uom.clone(),
            unit_price,
            tot_amt,
            discount: if discount > Decimal::ZERO { Some(discount) } else { None },
            pre_tax_val: None,
            ass_amt,
            gst_rt: gst_rate,
            igst_amt: igst,
            cgst_amt: cgst,
            sgst_amt: sgst,
            cess_rt: if cess_rate > Decimal::ZERO { Some(cess_rate) } else { None },
            cess_amt,
            cess_non_advol_amt: None,
            state_cess_rt: None,
            state_cess_amt: None,
            state_cess_non_advol_amt: None,
            oth_chrg: None,
            tot_item_val,
            ord_line_ref: None,
            org_cntry: None,
            prd_sl_no: None,
            batch_dtls: None,
            attrib_dtls: None,
        });
    }

    // Invoice level adjustments
    let inv_discount: Decimal = input.discount.as_ref()
        .and_then(|d| d.parse().ok())
        .unwrap_or(Decimal::ZERO);
    let other_charges: Decimal = input.other_charges.as_ref()
        .and_then(|o| o.parse().ok())
        .unwrap_or(Decimal::ZERO);
    let round_off: Decimal = input.round_off.as_ref()
        .and_then(|r| r.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let tot_inv_val = total_ass_val + total_igst + total_cgst + total_sgst + total_cess
        - inv_discount + other_charges + round_off;

    EInvoiceTransaction {
        version: "1.1".to_string(),
        tran_dtls: TransactionDetails {
            tax_sch: "GST".to_string(),
            supply_typ: input.supply_type.clone(),
            reg_rev: if input.reverse_charge { Some("Y".to_string()) } else { Some("N".to_string()) },
            ec_om_gstin: None,
            igst_on_intra: None,
        },
        doc_dtls: DocumentDetails {
            typ: "INV".to_string(),
            no: input.invoice_no.clone(),
            dt: input.invoice_date.clone(),
        },
        seller_dtls: PartyDetails {
            gstin: input.seller.gstin.clone(),
            lgl_nm: input.seller.legal_name.clone(),
            trd_nm: input.seller.trade_name.clone(),
            addr1: input.seller.address1.clone(),
            addr2: input.seller.address2.clone(),
            loc: input.seller.city.clone(),
            pin: input.seller.pincode.parse().unwrap_or(0),
            stcd: input.seller.state_code.clone(),
            ph: input.seller.phone.clone(),
            em: input.seller.email.clone(),
        },
        buyer_dtls: PartyDetails {
            gstin: input.buyer.gstin.clone(),
            lgl_nm: input.buyer.legal_name.clone(),
            trd_nm: input.buyer.trade_name.clone(),
            addr1: input.buyer.address1.clone(),
            addr2: input.buyer.address2.clone(),
            loc: input.buyer.city.clone(),
            pin: input.buyer.pincode.parse().unwrap_or(0),
            stcd: input.buyer.state_code.clone(),
            ph: input.buyer.phone.clone(),
            em: input.buyer.email.clone(),
        },
        dispatch_dtls: None,
        ship_dtls: None,
        item_list: items,
        val_dtls: ValueDetails {
            ass_val: total_ass_val.round_dp(2),
            cgst_val: if total_cgst > Decimal::ZERO { Some(total_cgst.round_dp(2)) } else { None },
            sgst_val: if total_sgst > Decimal::ZERO { Some(total_sgst.round_dp(2)) } else { None },
            igst_val: if total_igst > Decimal::ZERO { Some(total_igst.round_dp(2)) } else { None },
            cess_val: if total_cess > Decimal::ZERO { Some(total_cess.round_dp(2)) } else { None },
            st_cess_val: None,
            discount: if inv_discount > Decimal::ZERO { Some(inv_discount.round_dp(2)) } else { None },
            oth_chrg: if other_charges > Decimal::ZERO { Some(other_charges.round_dp(2)) } else { None },
            rnd_off_amt: if round_off != Decimal::ZERO { Some(round_off.round_dp(2)) } else { None },
            tot_inv_val: tot_inv_val.round_dp(2),
            tot_inv_val_fc: None,
        },
        pay_dtls: None,
        ref_dtls: None,
        addl_doc_dtls: None,
        exp_dtls: input.export_details.as_ref().map(|e| ExportDetails {
            ship_bn: e.shipping_bill_no.clone(),
            ship_bdt: e.shipping_bill_date.clone(),
            port: e.port_code.clone(),
            ref_clm: None,
            for_cur: e.currency.clone(),
            cnt_code: e.country_code.clone(),
            exp_duty: None,
        }),
        eway_bill_dtls: None,
    }
}

/// Validate e-Invoice JSON against schema
#[wasm_bindgen]
pub fn validate_einvoice(einvoice: JsValue) -> JsValue {
    let einvoice: EInvoiceTransaction = match serde_wasm_bindgen::from_value(einvoice) {
        Ok(e) => e,
        Err(e) => {
            let result = serde_json::json!({
                "is_valid": false,
                "errors": [format!("Invalid e-Invoice structure: {}", e)]
            });
            return serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL);
        }
    };

    let mut errors: Vec<String> = Vec::new();

    // Validate seller GSTIN
    if einvoice.seller_dtls.gstin.len() != 15 {
        errors.push("Invalid seller GSTIN length".to_string());
    }

    // Validate buyer GSTIN
    if einvoice.buyer_dtls.gstin.len() != 15 && einvoice.buyer_dtls.gstin != "URP" {
        errors.push("Invalid buyer GSTIN length".to_string());
    }

    // Validate document number
    if einvoice.doc_dtls.no.is_empty() || einvoice.doc_dtls.no.len() > 16 {
        errors.push("Document number must be 1-16 characters".to_string());
    }

    // Validate items
    if einvoice.item_list.is_empty() {
        errors.push("At least one item is required".to_string());
    }

    for (idx, item) in einvoice.item_list.iter().enumerate() {
        if item.hsn_cd.len() < 4 {
            errors.push(format!("Item {}: HSN code must be at least 4 digits", idx + 1));
        }
        if item.qty <= Decimal::ZERO {
            errors.push(format!("Item {}: Quantity must be positive", idx + 1));
        }
    }

    // Validate totals
    let calc_ass_val: Decimal = einvoice.item_list.iter()
        .map(|i| i.ass_amt)
        .sum();

    if (calc_ass_val - einvoice.val_dtls.ass_val).abs() > dec!(1) {
        errors.push(format!(
            "Assessable value mismatch: calculated {} vs declared {}",
            calc_ass_val, einvoice.val_dtls.ass_val
        ));
    }

    let result = serde_json::json!({
        "is_valid": errors.is_empty(),
        "errors": errors,
        "warnings": []
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Generate IRN (Invoice Reference Number) hash
#[wasm_bindgen]
pub fn generate_irn_hash(seller_gstin: &str, invoice_no: &str, fy: &str) -> String {
    // IRN is SHA256 hash of: GSTIN + Invoice No + FY
    let input = format!("{}{}{}", seller_gstin, invoice_no, fy);
    let mut hasher = Sha256::new();
    hasher.update(input.as_bytes());
    let result = hasher.finalize();

    // Convert to hex string
    result.iter().map(|b| format!("{:02x}", b)).collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_irn_generation() {
        let irn = generate_irn_hash("27AADCB2230M1ZT", "INV001", "2024-25");
        assert_eq!(irn.len(), 64); // SHA256 produces 64 hex chars
    }
}
