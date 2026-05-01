//! GSTR (GST Returns) validation and generation

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

/// GSTR-1 B2B Invoice
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Gstr1B2bInvoice {
    pub ctin: String, // Customer GSTIN
    pub inv: Vec<B2bInvoiceDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct B2bInvoiceDetails {
    pub inum: String, // Invoice number
    pub idt: String,  // Invoice date
    pub val: String,  // Invoice value
    pub pos: String,  // Place of supply
    pub rchrg: String, // Reverse charge (Y/N)
    pub inv_typ: String, // R (Regular), SEWP, SEWOP, DE
    pub itms: Vec<B2bItemDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct B2bItemDetails {
    pub num: u32,
    pub itm_det: ItemTaxDetails,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItemTaxDetails {
    pub rt: String,     // Rate
    pub txval: String,  // Taxable value
    pub iamt: Option<String>, // IGST amount
    pub camt: Option<String>, // CGST amount
    pub samt: Option<String>, // SGST amount
    pub csamt: Option<String>, // Cess amount
}

/// GSTR-1 B2CL Invoice (Large)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Gstr1B2clInvoice {
    pub pos: String, // Place of supply
    pub inv: Vec<B2clInvoiceDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct B2clInvoiceDetails {
    pub inum: String,
    pub idt: String,
    pub val: String,
    pub itms: Vec<B2bItemDetails>,
}

/// GSTR-1 B2CS Summary
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Gstr1B2csSummary {
    pub sply_ty: String, // INTRA/INTER
    pub pos: String,
    pub rt: String,
    pub typ: String, // OE/E (with/without E-commerce)
    pub txval: String,
    pub iamt: Option<String>,
    pub camt: Option<String>,
    pub samt: Option<String>,
    pub csamt: Option<String>,
}

/// GSTR-1 Complete structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Gstr1Data {
    pub gstin: String,
    pub fp: String, // Filing period MMYYYY
    pub b2b: Vec<Gstr1B2bInvoice>,
    pub b2cl: Vec<Gstr1B2clInvoice>,
    pub b2cs: Vec<Gstr1B2csSummary>,
    pub cdnr: Vec<CreditDebitNote>,
    pub nil: Option<NilRatedSupplies>,
    pub hsn: HsnSummary,
    pub doc_issue: DocumentIssuance,
}

/// Credit/Debit Note
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreditDebitNote {
    pub ctin: String,
    pub nt: Vec<NoteDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NoteDetails {
    pub ntty: String, // C/D
    pub nt_num: String,
    pub nt_dt: String,
    pub val: String,
    pub pos: String,
    pub rchrg: String,
    pub inv_typ: String,
    pub itms: Vec<B2bItemDetails>,
}

/// Nil Rated Supplies
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NilRatedSupplies {
    pub inv: Vec<NilInvoice>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NilInvoice {
    pub sply_ty: String,
    pub nil_amt: String,
    pub expt_amt: String,
    pub ngsup_amt: String,
}

/// HSN Summary
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HsnSummary {
    pub data: Vec<HsnEntry>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HsnEntry {
    pub hsn_sc: String,
    pub desc: String,
    pub uqc: String,
    pub qty: String,
    pub val: String,
    pub txval: String,
    pub iamt: Option<String>,
    pub camt: Option<String>,
    pub samt: Option<String>,
    pub csamt: Option<String>,
}

/// Document Issuance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DocumentIssuance {
    pub doc_det: Vec<DocDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DocDetails {
    pub doc_num: u32,
    pub docs: Vec<DocRange>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DocRange {
    pub num: u32,
    pub from: String,
    pub to: String,
    pub totnum: u32,
    pub cancel: u32,
    pub net_issue: u32,
}

/// GSTR-3B Summary
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Gstr3bData {
    pub gstin: String,
    pub ret_period: String,
    pub sup_details: SupplyDetails,
    pub itc_elg: ItcDetails,
    pub intr_ltfee: InterestLateFee,
    pub inward_sup: InwardSupplies,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupplyDetails {
    pub osup_det: OutwardSupplyDetails,
    pub osup_zero: OutwardZeroRated,
    pub osup_nil_exmp: OutwardNilExempt,
    pub isup_rev: InwardReverseCharge,
    pub osup_nongst: OutwardNonGst,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutwardSupplyDetails {
    pub txval: String,
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutwardZeroRated {
    pub txval: String,
    pub iamt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutwardNilExempt {
    pub txval: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InwardReverseCharge {
    pub txval: String,
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutwardNonGst {
    pub txval: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItcDetails {
    pub itc_avl: Vec<ItcAvailable>,
    pub itc_rev: Vec<ItcReversed>,
    pub itc_net: ItcNet,
    pub itc_inelg: Vec<ItcIneligible>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItcAvailable {
    pub ty: String,
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItcReversed {
    pub ty: String,
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItcNet {
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItcIneligible {
    pub ty: String,
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InterestLateFee {
    pub intr_details: IntrDetails,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IntrDetails {
    pub iamt: String,
    pub camt: String,
    pub samt: String,
    pub csamt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InwardSupplies {
    pub isup_details: Vec<IsupDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IsupDetails {
    pub ty: String,
    pub inter: String,
    pub intra: String,
}

/// Invoice input for GSTR generation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstrInvoiceInput {
    pub invoice_no: String,
    pub invoice_date: String,
    pub buyer_gstin: Option<String>,
    pub place_of_supply: String,
    pub reverse_charge: bool,
    pub invoice_type: String,
    pub items: Vec<GstrItemInput>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GstrItemInput {
    pub hsn_code: String,
    pub description: String,
    pub quantity: String,
    pub uom: String,
    pub taxable_value: String,
    pub gst_rate: String,
    pub igst: Option<String>,
    pub cgst: Option<String>,
    pub sgst: Option<String>,
    pub cess: Option<String>,
}

/// Generate GSTR-1 data from invoices
#[wasm_bindgen]
pub fn generate_gstr1(gstin: &str, period: &str, invoices: JsValue) -> JsValue {
    let invoices: Vec<GstrInvoiceInput> = match serde_wasm_bindgen::from_value(invoices) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid invoices: {}", e);
            return JsValue::NULL;
        }
    };

    let seller_state = &gstin[0..2];
    let mut b2b_map: HashMap<String, Vec<B2bInvoiceDetails>> = HashMap::new();
    let mut b2cl_map: HashMap<String, Vec<B2clInvoiceDetails>> = HashMap::new();
    let mut b2cs_map: HashMap<String, Gstr1B2csSummary> = HashMap::new();
    let mut hsn_map: HashMap<String, HsnEntry> = HashMap::new();

    for inv in &invoices {
        let pos_state = &inv.place_of_supply[0..2];
        let is_inter = seller_state != pos_state;
        let has_gstin = inv.buyer_gstin.as_ref().map(|g| g.len() == 15).unwrap_or(false);

        let total_value: Decimal = inv.items.iter()
            .map(|i| {
                let txval: Decimal = i.taxable_value.parse().unwrap_or(Decimal::ZERO);
                let igst: Decimal = i.igst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                let cgst: Decimal = i.cgst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                let sgst: Decimal = i.sgst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                txval + igst + cgst + sgst
            })
            .sum();

        // Build item details
        let items: Vec<B2bItemDetails> = inv.items.iter().enumerate().map(|(idx, item)| {
            B2bItemDetails {
                num: (idx + 1) as u32,
                itm_det: ItemTaxDetails {
                    rt: item.gst_rate.clone(),
                    txval: item.taxable_value.clone(),
                    iamt: item.igst.clone(),
                    camt: item.cgst.clone(),
                    samt: item.sgst.clone(),
                    csamt: item.cess.clone(),
                },
            }
        }).collect();

        if has_gstin {
            // B2B
            let ctin = inv.buyer_gstin.clone().unwrap_or_default();
            let details = B2bInvoiceDetails {
                inum: inv.invoice_no.clone(),
                idt: inv.invoice_date.clone(),
                val: total_value.round_dp(2).to_string(),
                pos: inv.place_of_supply.clone(),
                rchrg: if inv.reverse_charge { "Y".to_string() } else { "N".to_string() },
                inv_typ: inv.invoice_type.clone(),
                itms: items,
            };

            b2b_map.entry(ctin).or_insert_with(Vec::new).push(details);
        } else if is_inter && total_value > dec!(250000) {
            // B2CL
            let details = B2clInvoiceDetails {
                inum: inv.invoice_no.clone(),
                idt: inv.invoice_date.clone(),
                val: total_value.round_dp(2).to_string(),
                itms: items,
            };

            b2cl_map.entry(inv.place_of_supply.clone())
                .or_insert_with(Vec::new)
                .push(details);
        } else {
            // B2CS - aggregate by rate and POS
            for item in &inv.items {
                let key = format!("{}_{}_{}",
                    inv.place_of_supply,
                    item.gst_rate,
                    if is_inter { "INTER" } else { "INTRA" }
                );

                let entry = b2cs_map.entry(key).or_insert_with(|| Gstr1B2csSummary {
                    sply_ty: if is_inter { "INTER".to_string() } else { "INTRA".to_string() },
                    pos: inv.place_of_supply.clone(),
                    rt: item.gst_rate.clone(),
                    typ: "OE".to_string(),
                    txval: "0".to_string(),
                    iamt: if is_inter { Some("0".to_string()) } else { None },
                    camt: if !is_inter { Some("0".to_string()) } else { None },
                    samt: if !is_inter { Some("0".to_string()) } else { None },
                    csamt: Some("0".to_string()),
                });

                let txval: Decimal = item.taxable_value.parse().unwrap_or(Decimal::ZERO);
                let existing: Decimal = entry.txval.parse().unwrap_or(Decimal::ZERO);
                entry.txval = (existing + txval).round_dp(2).to_string();

                if is_inter {
                    let igst: Decimal = item.igst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                    let existing: Decimal = entry.iamt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                    entry.iamt = Some((existing + igst).round_dp(2).to_string());
                } else {
                    let cgst: Decimal = item.cgst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                    let sgst: Decimal = item.sgst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                    let ex_cgst: Decimal = entry.camt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                    let ex_sgst: Decimal = entry.samt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                    entry.camt = Some((ex_cgst + cgst).round_dp(2).to_string());
                    entry.samt = Some((ex_sgst + sgst).round_dp(2).to_string());
                }
            }
        }

        // HSN Summary
        for item in &inv.items {
            let hsn_entry = hsn_map.entry(item.hsn_code.clone()).or_insert_with(|| HsnEntry {
                hsn_sc: item.hsn_code.clone(),
                desc: item.description.clone(),
                uqc: item.uom.clone(),
                qty: "0".to_string(),
                val: "0".to_string(),
                txval: "0".to_string(),
                iamt: Some("0".to_string()),
                camt: Some("0".to_string()),
                samt: Some("0".to_string()),
                csamt: Some("0".to_string()),
            });

            let qty: Decimal = item.quantity.parse().unwrap_or(Decimal::ZERO);
            let txval: Decimal = item.taxable_value.parse().unwrap_or(Decimal::ZERO);
            let igst: Decimal = item.igst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
            let cgst: Decimal = item.cgst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
            let sgst: Decimal = item.sgst.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
            let total = txval + igst + cgst + sgst;

            let ex_qty: Decimal = hsn_entry.qty.parse().unwrap_or(Decimal::ZERO);
            let ex_val: Decimal = hsn_entry.val.parse().unwrap_or(Decimal::ZERO);
            let ex_txval: Decimal = hsn_entry.txval.parse().unwrap_or(Decimal::ZERO);
            let ex_iamt: Decimal = hsn_entry.iamt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
            let ex_camt: Decimal = hsn_entry.camt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
            let ex_samt: Decimal = hsn_entry.samt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);

            hsn_entry.qty = (ex_qty + qty).to_string();
            hsn_entry.val = (ex_val + total).round_dp(2).to_string();
            hsn_entry.txval = (ex_txval + txval).round_dp(2).to_string();
            hsn_entry.iamt = Some((ex_iamt + igst).round_dp(2).to_string());
            hsn_entry.camt = Some((ex_camt + cgst).round_dp(2).to_string());
            hsn_entry.samt = Some((ex_samt + sgst).round_dp(2).to_string());
        }
    }

    // Build GSTR-1 structure
    let b2b: Vec<Gstr1B2bInvoice> = b2b_map.into_iter().map(|(ctin, inv)| {
        Gstr1B2bInvoice { ctin, inv }
    }).collect();

    let b2cl: Vec<Gstr1B2clInvoice> = b2cl_map.into_iter().map(|(pos, inv)| {
        Gstr1B2clInvoice { pos, inv }
    }).collect();

    let b2cs: Vec<Gstr1B2csSummary> = b2cs_map.into_values().collect();

    let gstr1 = Gstr1Data {
        gstin: gstin.to_string(),
        fp: period.to_string(),
        b2b,
        b2cl,
        b2cs,
        cdnr: Vec::new(),
        nil: None,
        hsn: HsnSummary {
            data: hsn_map.into_values().collect(),
        },
        doc_issue: DocumentIssuance {
            doc_det: Vec::new(),
        },
    };

    serde_wasm_bindgen::to_value(&gstr1).unwrap_or(JsValue::NULL)
}

/// Validate GSTR-1 data
#[wasm_bindgen]
pub fn validate_gstr1(gstr1: JsValue) -> JsValue {
    let data: Gstr1Data = match serde_wasm_bindgen::from_value(gstr1) {
        Ok(d) => d,
        Err(e) => {
            let result = serde_json::json!({
                "is_valid": false,
                "errors": [format!("Invalid GSTR-1 structure: {}", e)]
            });
            return serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL);
        }
    };

    let mut errors: Vec<String> = Vec::new();
    let mut warnings: Vec<String> = Vec::new();

    // Validate GSTIN
    if data.gstin.len() != 15 {
        errors.push("Invalid GSTIN format".to_string());
    }

    // Validate period format (MMYYYY)
    if data.fp.len() != 6 {
        errors.push("Invalid filing period format (expected MMYYYY)".to_string());
    }

    // Validate B2B invoices
    for b2b in &data.b2b {
        if b2b.ctin.len() != 15 {
            errors.push(format!("Invalid customer GSTIN: {}", b2b.ctin));
        }

        for inv in &b2b.inv {
            if inv.inum.is_empty() || inv.inum.len() > 16 {
                errors.push(format!("Invalid invoice number: {}", inv.inum));
            }

            // Validate tax amounts match rate
            for item in &inv.itms {
                let txval: Decimal = item.itm_det.txval.parse().unwrap_or(Decimal::ZERO);
                let rate: Decimal = item.itm_det.rt.parse().unwrap_or(Decimal::ZERO);

                if let Some(ref iamt) = item.itm_det.iamt {
                    let igst: Decimal = iamt.parse().unwrap_or(Decimal::ZERO);
                    let expected = (txval * rate / dec!(100)).round_dp(2);
                    if (igst - expected).abs() > dec!(1) {
                        warnings.push(format!(
                            "Invoice {}: IGST amount {} doesn't match rate {} on taxable {}",
                            inv.inum, igst, rate, txval
                        ));
                    }
                }
            }
        }
    }

    // Validate HSN codes
    for hsn in &data.hsn.data {
        if hsn.hsn_sc.len() < 4 {
            warnings.push(format!("HSN code {} should be at least 4 digits", hsn.hsn_sc));
        }
    }

    let result = serde_json::json!({
        "is_valid": errors.is_empty(),
        "errors": errors,
        "warnings": warnings,
        "summary": {
            "b2b_invoices": data.b2b.iter().map(|b| b.inv.len()).sum::<usize>(),
            "b2cl_invoices": data.b2cl.iter().map(|b| b.inv.len()).sum::<usize>(),
            "b2cs_records": data.b2cs.len(),
            "hsn_codes": data.hsn.data.len()
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Generate GSTR-3B summary from GSTR-1 data
#[wasm_bindgen]
pub fn generate_gstr3b_from_gstr1(gstr1: JsValue, itc_data: JsValue) -> JsValue {
    let gstr1_data: Gstr1Data = match serde_wasm_bindgen::from_value(gstr1) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid GSTR-1: {}", e);
            return JsValue::NULL;
        }
    };

    // Calculate outward supplies from GSTR-1
    let mut total_txval = Decimal::ZERO;
    let mut total_igst = Decimal::ZERO;
    let mut total_cgst = Decimal::ZERO;
    let mut total_sgst = Decimal::ZERO;
    let mut total_cess = Decimal::ZERO;

    // B2B
    for b2b in &gstr1_data.b2b {
        for inv in &b2b.inv {
            for item in &inv.itms {
                let txval: Decimal = item.itm_det.txval.parse().unwrap_or(Decimal::ZERO);
                let igst: Decimal = item.itm_det.iamt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                let cgst: Decimal = item.itm_det.camt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                let sgst: Decimal = item.itm_det.samt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
                let cess: Decimal = item.itm_det.csamt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);

                total_txval += txval;
                total_igst += igst;
                total_cgst += cgst;
                total_sgst += sgst;
                total_cess += cess;
            }
        }
    }

    // B2CS
    for b2cs in &gstr1_data.b2cs {
        let txval: Decimal = b2cs.txval.parse().unwrap_or(Decimal::ZERO);
        let igst: Decimal = b2cs.iamt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
        let cgst: Decimal = b2cs.camt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
        let sgst: Decimal = b2cs.samt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);
        let cess: Decimal = b2cs.csamt.as_ref().and_then(|v| v.parse().ok()).unwrap_or(Decimal::ZERO);

        total_txval += txval;
        total_igst += igst;
        total_cgst += cgst;
        total_sgst += sgst;
        total_cess += cess;
    }

    let gstr3b = Gstr3bData {
        gstin: gstr1_data.gstin.clone(),
        ret_period: gstr1_data.fp.clone(),
        sup_details: SupplyDetails {
            osup_det: OutwardSupplyDetails {
                txval: total_txval.round_dp(2).to_string(),
                iamt: total_igst.round_dp(2).to_string(),
                camt: total_cgst.round_dp(2).to_string(),
                samt: total_sgst.round_dp(2).to_string(),
                csamt: total_cess.round_dp(2).to_string(),
            },
            osup_zero: OutwardZeroRated {
                txval: "0".to_string(),
                iamt: "0".to_string(),
                csamt: "0".to_string(),
            },
            osup_nil_exmp: OutwardNilExempt {
                txval: "0".to_string(),
            },
            isup_rev: InwardReverseCharge {
                txval: "0".to_string(),
                iamt: "0".to_string(),
                camt: "0".to_string(),
                samt: "0".to_string(),
                csamt: "0".to_string(),
            },
            osup_nongst: OutwardNonGst {
                txval: "0".to_string(),
            },
        },
        itc_elg: ItcDetails {
            itc_avl: vec![],
            itc_rev: vec![],
            itc_net: ItcNet {
                iamt: "0".to_string(),
                camt: "0".to_string(),
                samt: "0".to_string(),
                csamt: "0".to_string(),
            },
            itc_inelg: vec![],
        },
        intr_ltfee: InterestLateFee {
            intr_details: IntrDetails {
                iamt: "0".to_string(),
                camt: "0".to_string(),
                samt: "0".to_string(),
                csamt: "0".to_string(),
            },
        },
        inward_sup: InwardSupplies {
            isup_details: vec![],
        },
    };

    serde_wasm_bindgen::to_value(&gstr3b).unwrap_or(JsValue::NULL)
}

/// Calculate tax liability
#[wasm_bindgen]
pub fn calculate_tax_liability(gstr3b: JsValue, itc: JsValue) -> JsValue {
    let gstr3b_data: Gstr3bData = match serde_wasm_bindgen::from_value(gstr3b) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid GSTR-3B: {}", e);
            return JsValue::NULL;
        }
    };

    // Output tax
    let output_igst: Decimal = gstr3b_data.sup_details.osup_det.iamt.parse().unwrap_or(Decimal::ZERO);
    let output_cgst: Decimal = gstr3b_data.sup_details.osup_det.camt.parse().unwrap_or(Decimal::ZERO);
    let output_sgst: Decimal = gstr3b_data.sup_details.osup_det.samt.parse().unwrap_or(Decimal::ZERO);
    let output_cess: Decimal = gstr3b_data.sup_details.osup_det.csamt.parse().unwrap_or(Decimal::ZERO);

    // ITC
    let itc_igst: Decimal = gstr3b_data.itc_elg.itc_net.iamt.parse().unwrap_or(Decimal::ZERO);
    let itc_cgst: Decimal = gstr3b_data.itc_elg.itc_net.camt.parse().unwrap_or(Decimal::ZERO);
    let itc_sgst: Decimal = gstr3b_data.itc_elg.itc_net.samt.parse().unwrap_or(Decimal::ZERO);

    // Net liability (simplified - actual utilization has priority rules)
    let net_igst = (output_igst - itc_igst).max(Decimal::ZERO);
    let net_cgst = (output_cgst - itc_cgst).max(Decimal::ZERO);
    let net_sgst = (output_sgst - itc_sgst).max(Decimal::ZERO);
    let net_cess = output_cess; // Cess ITC can only be used for Cess

    let total_liability = net_igst + net_cgst + net_sgst + net_cess;

    let result = serde_json::json!({
        "output_tax": {
            "igst": output_igst.round_dp(2).to_string(),
            "cgst": output_cgst.round_dp(2).to_string(),
            "sgst": output_sgst.round_dp(2).to_string(),
            "cess": output_cess.round_dp(2).to_string(),
            "total": (output_igst + output_cgst + output_sgst + output_cess).round_dp(2).to_string()
        },
        "itc_utilized": {
            "igst": itc_igst.round_dp(2).to_string(),
            "cgst": itc_cgst.round_dp(2).to_string(),
            "sgst": itc_sgst.round_dp(2).to_string()
        },
        "net_liability": {
            "igst": net_igst.round_dp(2).to_string(),
            "cgst": net_cgst.round_dp(2).to_string(),
            "sgst": net_sgst.round_dp(2).to_string(),
            "cess": net_cess.round_dp(2).to_string(),
            "total": total_liability.round_dp(2).to_string()
        },
        "payment_due": total_liability.round_dp(2).to_string()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_gstr1_b2b_creation() {
        let inv = B2bInvoiceDetails {
            inum: "INV001".to_string(),
            idt: "01/01/2024".to_string(),
            val: "10000".to_string(),
            pos: "27".to_string(),
            rchrg: "N".to_string(),
            inv_typ: "R".to_string(),
            itms: vec![],
        };

        assert_eq!(inv.inum, "INV001");
    }
}
