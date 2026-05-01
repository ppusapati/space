//! BOM Explosion - Multi-level breakdown of assemblies

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

/// BOM item type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ItemType {
    /// Raw material
    Raw,
    /// Sub-assembly
    SubAssembly,
    /// Finished good
    Finished,
    /// Phantom assembly (exploded but not manufactured)
    Phantom,
    /// Purchased component
    Purchased,
    /// Service/Operation
    Service,
}

/// BOM component (single level)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BomComponent {
    pub item_code: String,
    pub item_name: String,
    pub item_type: ItemType,
    pub quantity: String,
    pub uom: String,
    pub scrap_percentage: Option<String>,
    pub lead_time_days: Option<u32>,
    pub unit_cost: Option<String>,
    pub warehouse: Option<String>,
    pub routing_operation: Option<String>,
}

/// BOM header
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Bom {
    pub item_code: String,
    pub item_name: String,
    pub quantity: String,
    pub uom: String,
    pub is_active: bool,
    pub is_default: bool,
    pub components: Vec<BomComponent>,
}

/// Exploded BOM line (result of explosion)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExplodedBomLine {
    pub level: u32,
    pub parent_item_code: Option<String>,
    pub item_code: String,
    pub item_name: String,
    pub item_type: ItemType,
    pub bom_quantity: String,
    pub required_quantity: String,
    pub uom: String,
    pub scrap_quantity: String,
    pub total_quantity: String,
    pub unit_cost: String,
    pub total_cost: String,
    pub lead_time_days: u32,
    pub cumulative_lead_time: u32,
    pub warehouse: Option<String>,
    pub indent_path: String,
}

/// BOM explosion result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BomExplosionResult {
    pub item_code: String,
    pub item_name: String,
    pub quantity_required: String,
    pub lines: Vec<ExplodedBomLine>,
    pub total_raw_materials: Vec<MaterialSummary>,
    pub total_cost: String,
    pub total_lead_time_days: u32,
    pub levels_deep: u32,
}

/// Material summary (aggregated)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaterialSummary {
    pub item_code: String,
    pub item_name: String,
    pub item_type: ItemType,
    pub total_quantity: String,
    pub uom: String,
    pub unit_cost: String,
    pub total_cost: String,
    pub used_in: Vec<String>,
}

/// BOM database (for multi-level explosion)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BomDatabase {
    pub boms: HashMap<String, Bom>,
}

/// Explode a BOM to all levels
#[wasm_bindgen]
pub fn explode_bom(bom_db: JsValue, item_code: &str, quantity: &str) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let qty: Decimal = quantity.parse().unwrap_or(dec!(1));
    let result = explode_bom_internal(&db, item_code, qty);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal BOM explosion
fn explode_bom_internal(db: &BomDatabase, item_code: &str, quantity: Decimal) -> BomExplosionResult {
    let mut lines: Vec<ExplodedBomLine> = Vec::new();
    let mut material_map: HashMap<String, MaterialSummary> = HashMap::new();
    let mut max_level = 0u32;
    let mut max_lead_time = 0u32;

    let bom = db.boms.get(item_code);
    let item_name = bom.map(|b| b.item_name.clone()).unwrap_or_else(|| item_code.to_string());

    // Recursive explosion
    explode_recursive(
        db,
        item_code,
        quantity,
        0,
        None,
        "".to_string(),
        0,
        &mut lines,
        &mut material_map,
        &mut max_level,
        &mut max_lead_time,
    );

    // Convert material map to vec
    let total_raw_materials: Vec<MaterialSummary> = material_map.into_values().collect();

    // Calculate total cost
    let total_cost: Decimal = total_raw_materials.iter()
        .map(|m| m.total_cost.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    BomExplosionResult {
        item_code: item_code.to_string(),
        item_name,
        quantity_required: quantity.to_string(),
        lines,
        total_raw_materials,
        total_cost: total_cost.round_dp(2).to_string(),
        total_lead_time_days: max_lead_time,
        levels_deep: max_level,
    }
}

fn explode_recursive(
    db: &BomDatabase,
    item_code: &str,
    parent_qty: Decimal,
    level: u32,
    parent_item: Option<&str>,
    indent_path: String,
    cumulative_lead_time: u32,
    lines: &mut Vec<ExplodedBomLine>,
    material_map: &mut HashMap<String, MaterialSummary>,
    max_level: &mut u32,
    max_lead_time: &mut u32,
) {
    if level > *max_level {
        *max_level = level;
    }

    if let Some(bom) = db.boms.get(item_code) {
        let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));

        for component in &bom.components {
            let comp_qty: Decimal = component.quantity.parse().unwrap_or(Decimal::ZERO);
            let scrap_pct: Decimal = component.scrap_percentage.as_ref()
                .and_then(|s| s.parse().ok())
                .unwrap_or(Decimal::ZERO) / dec!(100);

            // Calculate required quantity
            let required_qty = (comp_qty * parent_qty / bom_qty).round_dp(4);
            let scrap_qty = (required_qty * scrap_pct).round_dp(4);
            let total_qty = required_qty + scrap_qty;

            // Get unit cost
            let unit_cost: Decimal = component.unit_cost.as_ref()
                .and_then(|c| c.parse().ok())
                .unwrap_or(Decimal::ZERO);
            let total_cost = (total_qty * unit_cost).round_dp(2);

            // Lead time
            let lead_time = component.lead_time_days.unwrap_or(0);
            let cum_lead_time = cumulative_lead_time + lead_time;

            if cum_lead_time > *max_lead_time {
                *max_lead_time = cum_lead_time;
            }

            // Build indent path
            let new_indent_path = if indent_path.is_empty() {
                component.item_code.clone()
            } else {
                format!("{} > {}", indent_path, component.item_code)
            };

            // Add to lines
            lines.push(ExplodedBomLine {
                level: level + 1,
                parent_item_code: Some(item_code.to_string()),
                item_code: component.item_code.clone(),
                item_name: component.item_name.clone(),
                item_type: component.item_type,
                bom_quantity: comp_qty.to_string(),
                required_quantity: required_qty.to_string(),
                uom: component.uom.clone(),
                scrap_quantity: scrap_qty.to_string(),
                total_quantity: total_qty.to_string(),
                unit_cost: unit_cost.to_string(),
                total_cost: total_cost.to_string(),
                lead_time_days: lead_time,
                cumulative_lead_time: cum_lead_time,
                warehouse: component.warehouse.clone(),
                indent_path: new_indent_path.clone(),
            });

            // Check if this is a raw material or phantom
            let should_explode = matches!(component.item_type, ItemType::SubAssembly | ItemType::Phantom);

            if should_explode && db.boms.contains_key(&component.item_code) {
                // Recursively explode
                explode_recursive(
                    db,
                    &component.item_code,
                    total_qty,
                    level + 1,
                    Some(item_code),
                    new_indent_path,
                    cum_lead_time,
                    lines,
                    material_map,
                    max_level,
                    max_lead_time,
                );
            } else {
                // Add to material summary
                let entry = material_map.entry(component.item_code.clone())
                    .or_insert_with(|| MaterialSummary {
                        item_code: component.item_code.clone(),
                        item_name: component.item_name.clone(),
                        item_type: component.item_type,
                        total_quantity: "0".to_string(),
                        uom: component.uom.clone(),
                        unit_cost: unit_cost.to_string(),
                        total_cost: "0".to_string(),
                        used_in: Vec::new(),
                    });

                let existing_qty: Decimal = entry.total_quantity.parse().unwrap_or(Decimal::ZERO);
                let new_qty = existing_qty + total_qty;
                entry.total_quantity = new_qty.to_string();

                let existing_cost: Decimal = entry.total_cost.parse().unwrap_or(Decimal::ZERO);
                entry.total_cost = (existing_cost + total_cost).to_string();

                if !entry.used_in.contains(&item_code.to_string()) {
                    entry.used_in.push(item_code.to_string());
                }
            }
        }
    }
}

/// Explode BOM for a single level only
#[wasm_bindgen]
pub fn explode_bom_single_level(bom: JsValue, quantity: &str) -> JsValue {
    let bom: Bom = match serde_wasm_bindgen::from_value(bom) {
        Ok(b) => b,
        Err(e) => {
            log::error!("Invalid BOM: {}", e);
            return JsValue::NULL;
        }
    };

    let qty: Decimal = quantity.parse().unwrap_or(dec!(1));
    let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));

    let lines: Vec<ExplodedBomLine> = bom.components.iter().map(|c| {
        let comp_qty: Decimal = c.quantity.parse().unwrap_or(Decimal::ZERO);
        let scrap_pct: Decimal = c.scrap_percentage.as_ref()
            .and_then(|s| s.parse().ok())
            .unwrap_or(Decimal::ZERO) / dec!(100);

        let required_qty = (comp_qty * qty / bom_qty).round_dp(4);
        let scrap_qty = (required_qty * scrap_pct).round_dp(4);
        let total_qty = required_qty + scrap_qty;

        let unit_cost: Decimal = c.unit_cost.as_ref()
            .and_then(|u| u.parse().ok())
            .unwrap_or(Decimal::ZERO);
        let total_cost = (total_qty * unit_cost).round_dp(2);

        ExplodedBomLine {
            level: 1,
            parent_item_code: Some(bom.item_code.clone()),
            item_code: c.item_code.clone(),
            item_name: c.item_name.clone(),
            item_type: c.item_type,
            bom_quantity: comp_qty.to_string(),
            required_quantity: required_qty.to_string(),
            uom: c.uom.clone(),
            scrap_quantity: scrap_qty.to_string(),
            total_quantity: total_qty.to_string(),
            unit_cost: unit_cost.to_string(),
            total_cost: total_cost.to_string(),
            lead_time_days: c.lead_time_days.unwrap_or(0),
            cumulative_lead_time: c.lead_time_days.unwrap_or(0),
            warehouse: c.warehouse.clone(),
            indent_path: c.item_code.clone(),
        }
    }).collect();

    let total_cost: Decimal = lines.iter()
        .map(|l| l.total_cost.parse::<Decimal>().unwrap_or(Decimal::ZERO))
        .sum();

    let result = serde_json::json!({
        "item_code": bom.item_code,
        "item_name": bom.item_name,
        "quantity_required": qty.to_string(),
        "lines": lines,
        "total_cost": total_cost.round_dp(2).to_string()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Validate BOM for circular references
#[wasm_bindgen]
pub fn validate_bom_circular(bom_db: JsValue, item_code: &str) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let mut visited: Vec<String> = Vec::new();
    let mut circular_path: Option<Vec<String>> = None;

    check_circular(&db, item_code, &mut visited, &mut circular_path);

    let result = serde_json::json!({
        "has_circular_reference": circular_path.is_some(),
        "circular_path": circular_path,
        "item_code": item_code
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

fn check_circular(
    db: &BomDatabase,
    item_code: &str,
    visited: &mut Vec<String>,
    circular_path: &mut Option<Vec<String>>,
) -> bool {
    if visited.contains(&item_code.to_string()) {
        // Found circular reference
        let idx = visited.iter().position(|x| x == item_code).unwrap();
        let mut path = visited[idx..].to_vec();
        path.push(item_code.to_string());
        *circular_path = Some(path);
        return true;
    }

    visited.push(item_code.to_string());

    if let Some(bom) = db.boms.get(item_code) {
        for component in &bom.components {
            if matches!(component.item_type, ItemType::SubAssembly | ItemType::Phantom) {
                if check_circular(db, &component.item_code, visited, circular_path) {
                    return true;
                }
            }
        }
    }

    visited.pop();
    false
}

/// Get BOM tree structure (for display)
#[wasm_bindgen]
pub fn get_bom_tree(bom_db: JsValue, item_code: &str) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let tree = build_tree(&db, item_code, 0);
    serde_wasm_bindgen::to_value(&tree).unwrap_or(JsValue::NULL)
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct BomTreeNode {
    item_code: String,
    item_name: String,
    item_type: ItemType,
    quantity: String,
    uom: String,
    level: u32,
    children: Vec<BomTreeNode>,
}

fn build_tree(db: &BomDatabase, item_code: &str, level: u32) -> BomTreeNode {
    let bom = db.boms.get(item_code);

    let (item_name, quantity, uom, item_type) = if let Some(b) = bom {
        (b.item_name.clone(), b.quantity.clone(), b.uom.clone(), ItemType::Finished)
    } else {
        (item_code.to_string(), "1".to_string(), "NOS".to_string(), ItemType::Raw)
    };

    let children: Vec<BomTreeNode> = if let Some(b) = bom {
        b.components.iter().map(|c| {
            if matches!(c.item_type, ItemType::SubAssembly | ItemType::Phantom) && db.boms.contains_key(&c.item_code) {
                build_tree(db, &c.item_code, level + 1)
            } else {
                BomTreeNode {
                    item_code: c.item_code.clone(),
                    item_name: c.item_name.clone(),
                    item_type: c.item_type,
                    quantity: c.quantity.clone(),
                    uom: c.uom.clone(),
                    level: level + 1,
                    children: Vec::new(),
                }
            }
        }).collect()
    } else {
        Vec::new()
    };

    BomTreeNode {
        item_code: item_code.to_string(),
        item_name,
        item_type,
        quantity,
        uom,
        level,
        children,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_db() -> BomDatabase {
        let mut boms = HashMap::new();

        // Finished product
        boms.insert("FG001".to_string(), Bom {
            item_code: "FG001".to_string(),
            item_name: "Finished Product".to_string(),
            quantity: "1".to_string(),
            uom: "NOS".to_string(),
            is_active: true,
            is_default: true,
            components: vec![
                BomComponent {
                    item_code: "SA001".to_string(),
                    item_name: "Sub Assembly".to_string(),
                    item_type: ItemType::SubAssembly,
                    quantity: "2".to_string(),
                    uom: "NOS".to_string(),
                    scrap_percentage: Some("5".to_string()),
                    lead_time_days: Some(2),
                    unit_cost: Some("100".to_string()),
                    warehouse: None,
                    routing_operation: None,
                },
                BomComponent {
                    item_code: "RM001".to_string(),
                    item_name: "Raw Material 1".to_string(),
                    item_type: ItemType::Raw,
                    quantity: "5".to_string(),
                    uom: "KG".to_string(),
                    scrap_percentage: Some("2".to_string()),
                    lead_time_days: Some(5),
                    unit_cost: Some("50".to_string()),
                    warehouse: None,
                    routing_operation: None,
                },
            ],
        });

        // Sub assembly
        boms.insert("SA001".to_string(), Bom {
            item_code: "SA001".to_string(),
            item_name: "Sub Assembly".to_string(),
            quantity: "1".to_string(),
            uom: "NOS".to_string(),
            is_active: true,
            is_default: true,
            components: vec![
                BomComponent {
                    item_code: "RM002".to_string(),
                    item_name: "Raw Material 2".to_string(),
                    item_type: ItemType::Raw,
                    quantity: "3".to_string(),
                    uom: "KG".to_string(),
                    scrap_percentage: None,
                    lead_time_days: Some(3),
                    unit_cost: Some("25".to_string()),
                    warehouse: None,
                    routing_operation: None,
                },
            ],
        });

        BomDatabase { boms }
    }

    #[test]
    fn test_bom_explosion() {
        let db = create_test_db();
        let result = explode_bom_internal(&db, "FG001", dec!(10));

        assert_eq!(result.item_code, "FG001");
        assert!(!result.lines.is_empty());
        assert_eq!(result.levels_deep, 2);
    }

    #[test]
    fn test_no_circular_reference() {
        let db = create_test_db();
        let mut visited = Vec::new();
        let mut circular_path = None;

        let has_circular = check_circular(&db, "FG001", &mut visited, &mut circular_path);
        assert!(!has_circular);
    }
}
