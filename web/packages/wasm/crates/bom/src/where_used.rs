//! Where-Used Analysis - Find where components are used

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

use crate::explosion::{Bom, BomDatabase, ItemType};

/// Where-used entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WhereUsedEntry {
    pub parent_item_code: String,
    pub parent_item_name: String,
    pub parent_type: ItemType,
    pub quantity_per_assembly: String,
    pub uom: String,
    pub level: u32,
    pub path: String,
}

/// Where-used result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WhereUsedResult {
    pub item_code: String,
    pub item_name: String,
    pub usage_count: u32,
    pub direct_usage: Vec<WhereUsedEntry>,
    pub indirect_usage: Vec<WhereUsedEntry>,
    pub all_parents: Vec<String>,
}

/// Impact analysis result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImpactAnalysisResult {
    pub item_code: String,
    pub affected_assemblies: Vec<AffectedAssembly>,
    pub total_quantity_impact: String,
    pub cost_impact: String,
}

/// Affected assembly
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AffectedAssembly {
    pub item_code: String,
    pub item_name: String,
    pub quantity_affected: String,
    pub cost_impact: String,
}

/// Find where an item is used (direct parents only)
#[wasm_bindgen]
pub fn where_used_single_level(bom_db: JsValue, item_code: &str) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let direct_usage = find_direct_parents(&db, item_code);

    let result = WhereUsedResult {
        item_code: item_code.to_string(),
        item_name: get_item_name(&db, item_code),
        usage_count: direct_usage.len() as u32,
        direct_usage,
        indirect_usage: Vec::new(),
        all_parents: Vec::new(),
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Find all where an item is used (all levels)
#[wasm_bindgen]
pub fn where_used_all_levels(bom_db: JsValue, item_code: &str) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let direct_usage = find_direct_parents(&db, item_code);
    let mut indirect_usage: Vec<WhereUsedEntry> = Vec::new();
    let mut all_parents: Vec<String> = direct_usage.iter()
        .map(|e| e.parent_item_code.clone())
        .collect();

    // Find indirect usage recursively
    for entry in &direct_usage {
        find_indirect_parents(
            &db,
            &entry.parent_item_code,
            2,
            entry.path.clone(),
            &mut indirect_usage,
            &mut all_parents,
        );
    }

    let result = WhereUsedResult {
        item_code: item_code.to_string(),
        item_name: get_item_name(&db, item_code),
        usage_count: (direct_usage.len() + indirect_usage.len()) as u32,
        direct_usage,
        indirect_usage,
        all_parents,
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Find direct parents
fn find_direct_parents(db: &BomDatabase, item_code: &str) -> Vec<WhereUsedEntry> {
    let mut parents: Vec<WhereUsedEntry> = Vec::new();

    for (parent_code, bom) in &db.boms {
        for component in &bom.components {
            if component.item_code == item_code {
                let parent_type = if db.boms.values().any(|b| {
                    b.components.iter().any(|c| c.item_code == *parent_code)
                }) {
                    ItemType::SubAssembly
                } else {
                    ItemType::Finished
                };

                parents.push(WhereUsedEntry {
                    parent_item_code: parent_code.clone(),
                    parent_item_name: bom.item_name.clone(),
                    parent_type,
                    quantity_per_assembly: component.quantity.clone(),
                    uom: component.uom.clone(),
                    level: 1,
                    path: format!("{} -> {}", item_code, parent_code),
                });
            }
        }
    }

    parents
}

/// Find indirect parents recursively
fn find_indirect_parents(
    db: &BomDatabase,
    item_code: &str,
    level: u32,
    current_path: String,
    indirect_usage: &mut Vec<WhereUsedEntry>,
    all_parents: &mut Vec<String>,
) {
    if level > 10 {
        return; // Prevent infinite recursion
    }

    for (parent_code, bom) in &db.boms {
        for component in &bom.components {
            if component.item_code == item_code {
                let new_path = format!("{} -> {}", current_path, parent_code);

                if !all_parents.contains(parent_code) {
                    all_parents.push(parent_code.clone());
                }

                let parent_type = if db.boms.values().any(|b| {
                    b.components.iter().any(|c| c.item_code == *parent_code)
                }) {
                    ItemType::SubAssembly
                } else {
                    ItemType::Finished
                };

                indirect_usage.push(WhereUsedEntry {
                    parent_item_code: parent_code.clone(),
                    parent_item_name: bom.item_name.clone(),
                    parent_type,
                    quantity_per_assembly: component.quantity.clone(),
                    uom: component.uom.clone(),
                    level,
                    path: new_path.clone(),
                });

                // Continue recursively
                find_indirect_parents(
                    db,
                    parent_code,
                    level + 1,
                    new_path,
                    indirect_usage,
                    all_parents,
                );
            }
        }
    }
}

/// Get item name from database
fn get_item_name(db: &BomDatabase, item_code: &str) -> String {
    // Check if it's a BOM item
    if let Some(bom) = db.boms.get(item_code) {
        return bom.item_name.clone();
    }

    // Check if it's a component in any BOM
    for bom in db.boms.values() {
        for component in &bom.components {
            if component.item_code == item_code {
                return component.item_name.clone();
            }
        }
    }

    item_code.to_string()
}

/// Analyze impact of item change
#[wasm_bindgen]
pub fn analyze_impact(
    bom_db: JsValue,
    item_code: &str,
    price_change: &str,
    quantity_on_hand: Option<String>,
) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let price_delta: Decimal = price_change.parse().unwrap_or(Decimal::ZERO);
    let qty_on_hand: Decimal = quantity_on_hand
        .and_then(|q| q.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let where_used = find_direct_parents(&db, item_code);
    let mut affected_assemblies: Vec<AffectedAssembly> = Vec::new();
    let mut total_qty_impact = Decimal::ZERO;
    let mut total_cost_impact = Decimal::ZERO;

    for usage in &where_used {
        let qty_per: Decimal = usage.quantity_per_assembly.parse().unwrap_or(Decimal::ZERO);
        let cost_impact_per = (qty_per * price_delta).round_dp(2);

        affected_assemblies.push(AffectedAssembly {
            item_code: usage.parent_item_code.clone(),
            item_name: usage.parent_item_name.clone(),
            quantity_affected: qty_per.to_string(),
            cost_impact: cost_impact_per.to_string(),
        });

        total_qty_impact += qty_per;
        total_cost_impact += cost_impact_per;
    }

    let result = ImpactAnalysisResult {
        item_code: item_code.to_string(),
        affected_assemblies,
        total_quantity_impact: total_qty_impact.to_string(),
        cost_impact: total_cost_impact.to_string(),
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Find common components between two BOMs
#[wasm_bindgen]
pub fn find_common_components(bom_db: JsValue, item_code_1: &str, item_code_2: &str) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    // Get all components for each BOM
    let components_1 = get_all_components(&db, item_code_1);
    let components_2 = get_all_components(&db, item_code_2);

    // Find common components
    let common: Vec<&String> = components_1.iter()
        .filter(|c| components_2.contains(c))
        .collect();

    // Components only in BOM 1
    let only_in_1: Vec<&String> = components_1.iter()
        .filter(|c| !components_2.contains(c))
        .collect();

    // Components only in BOM 2
    let only_in_2: Vec<&String> = components_2.iter()
        .filter(|c| !components_1.contains(c))
        .collect();

    let result = serde_json::json!({
        "item_1": item_code_1,
        "item_2": item_code_2,
        "common_components": common,
        "only_in_item_1": only_in_1,
        "only_in_item_2": only_in_2,
        "commonality_percentage": if !components_1.is_empty() && !components_2.is_empty() {
            let total_unique = components_1.len() + components_2.len() - common.len();
            let pct = (common.len() as f64 / total_unique as f64 * 100.0).round();
            format!("{}%", pct)
        } else {
            "0%".to_string()
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Get all components (recursive)
fn get_all_components(db: &BomDatabase, item_code: &str) -> Vec<String> {
    let mut components: Vec<String> = Vec::new();
    collect_components(db, item_code, &mut components);
    components
}

fn collect_components(db: &BomDatabase, item_code: &str, components: &mut Vec<String>) {
    if let Some(bom) = db.boms.get(item_code) {
        for component in &bom.components {
            if !components.contains(&component.item_code) {
                components.push(component.item_code.clone());

                if matches!(component.item_type, ItemType::SubAssembly | ItemType::Phantom) {
                    collect_components(db, &component.item_code, components);
                }
            }
        }
    }
}

/// Find substitute/alternative components
#[wasm_bindgen]
pub fn find_substitutes(bom_db: JsValue, item_code: &str, substitute_map: JsValue) -> JsValue {
    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let substitutes: HashMap<String, Vec<String>> = match serde_wasm_bindgen::from_value(substitute_map) {
        Ok(s) => s,
        Err(_) => HashMap::new(),
    };

    let where_used = find_direct_parents(&db, item_code);
    let available_substitutes = substitutes.get(item_code).cloned().unwrap_or_default();

    let result = serde_json::json!({
        "item_code": item_code,
        "item_name": get_item_name(&db, item_code),
        "used_in": where_used.iter().map(|u| u.parent_item_code.clone()).collect::<Vec<_>>(),
        "available_substitutes": available_substitutes,
        "recommendation": if available_substitutes.is_empty() {
            "No substitutes available"
        } else {
            "Substitutes available - verify specifications before replacement"
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::explosion::BomComponent;

    fn create_test_db() -> BomDatabase {
        let mut boms = HashMap::new();

        boms.insert("FG001".to_string(), Bom {
            item_code: "FG001".to_string(),
            item_name: "Product A".to_string(),
            quantity: "1".to_string(),
            uom: "NOS".to_string(),
            is_active: true,
            is_default: true,
            components: vec![
                BomComponent {
                    item_code: "RM001".to_string(),
                    item_name: "Common Part".to_string(),
                    item_type: ItemType::Raw,
                    quantity: "2".to_string(),
                    uom: "NOS".to_string(),
                    scrap_percentage: None,
                    lead_time_days: None,
                    unit_cost: None,
                    warehouse: None,
                    routing_operation: None,
                },
            ],
        });

        boms.insert("FG002".to_string(), Bom {
            item_code: "FG002".to_string(),
            item_name: "Product B".to_string(),
            quantity: "1".to_string(),
            uom: "NOS".to_string(),
            is_active: true,
            is_default: true,
            components: vec![
                BomComponent {
                    item_code: "RM001".to_string(),
                    item_name: "Common Part".to_string(),
                    item_type: ItemType::Raw,
                    quantity: "3".to_string(),
                    uom: "NOS".to_string(),
                    scrap_percentage: None,
                    lead_time_days: None,
                    unit_cost: None,
                    warehouse: None,
                    routing_operation: None,
                },
            ],
        });

        BomDatabase { boms }
    }

    #[test]
    fn test_where_used() {
        let db = create_test_db();
        let parents = find_direct_parents(&db, "RM001");

        assert_eq!(parents.len(), 2);
    }

    #[test]
    fn test_common_components() {
        let db = create_test_db();
        let comp_1 = get_all_components(&db, "FG001");
        let comp_2 = get_all_components(&db, "FG002");

        assert!(comp_1.contains(&"RM001".to_string()));
        assert!(comp_2.contains(&"RM001".to_string()));
    }
}
