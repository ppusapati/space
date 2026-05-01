//! BOM Costing - Cost rollup and analysis

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

use crate::explosion::{Bom, BomDatabase, ItemType};

/// Cost type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CostType {
    Material,
    Labour,
    Overhead,
    Subcontracting,
    Other,
}

/// Operation cost
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OperationCost {
    pub operation_name: String,
    pub workstation: Option<String>,
    pub time_minutes: String,
    pub hourly_rate: String,
    pub cost: String,
    pub cost_type: CostType,
}

/// Item cost breakdown
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItemCostBreakdown {
    pub item_code: String,
    pub item_name: String,
    pub material_cost: String,
    pub labour_cost: String,
    pub overhead_cost: String,
    pub subcontracting_cost: String,
    pub other_cost: String,
    pub total_cost: String,
    pub component_costs: Vec<ComponentCost>,
    pub operation_costs: Vec<OperationCost>,
}

/// Component cost
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComponentCost {
    pub item_code: String,
    pub item_name: String,
    pub quantity: String,
    pub unit_cost: String,
    pub total_cost: String,
    pub cost_type: CostType,
}

/// Cost rollup input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostRollupInput {
    pub bom_db: BomDatabase,
    pub item_costs: HashMap<String, String>,
    pub operation_costs: HashMap<String, Vec<OperationCost>>,
    pub overhead_percentage: Option<String>,
}

/// Cost comparison result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostComparisonResult {
    pub item_code: String,
    pub scenarios: Vec<CostScenario>,
    pub variance_analysis: VarianceAnalysis,
}

/// Cost scenario
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostScenario {
    pub name: String,
    pub material_cost: String,
    pub labour_cost: String,
    pub overhead_cost: String,
    pub total_cost: String,
}

/// Variance analysis
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VarianceAnalysis {
    pub material_variance: String,
    pub labour_variance: String,
    pub overhead_variance: String,
    pub total_variance: String,
    pub variance_percentage: String,
}

/// Calculate BOM cost rollup
#[wasm_bindgen]
pub fn calculate_cost_rollup(input: JsValue) -> JsValue {
    let input: CostRollupInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid cost rollup input: {}", e);
            return JsValue::NULL;
        }
    };

    let mut costs: HashMap<String, ItemCostBreakdown> = HashMap::new();

    // Process BOMs in dependency order (leaf to root)
    let sorted_items = topological_sort(&input.bom_db);

    for item_code in sorted_items {
        if let Some(bom) = input.bom_db.boms.get(&item_code) {
            let breakdown = calculate_item_cost(
                bom,
                &input.item_costs,
                &input.operation_costs,
                &costs,
                input.overhead_percentage.as_deref(),
            );
            costs.insert(item_code, breakdown);
        }
    }

    serde_wasm_bindgen::to_value(&costs).unwrap_or(JsValue::NULL)
}

/// Topological sort for dependency order
fn topological_sort(db: &BomDatabase) -> Vec<String> {
    let mut result = Vec::new();
    let mut visited: HashMap<String, bool> = HashMap::new();

    for item_code in db.boms.keys() {
        if !visited.contains_key(item_code) {
            topological_visit(db, item_code, &mut visited, &mut result);
        }
    }

    result
}

fn topological_visit(
    db: &BomDatabase,
    item_code: &str,
    visited: &mut HashMap<String, bool>,
    result: &mut Vec<String>,
) {
    if visited.get(item_code).copied().unwrap_or(false) {
        return;
    }

    visited.insert(item_code.to_string(), true);

    if let Some(bom) = db.boms.get(item_code) {
        for component in &bom.components {
            if matches!(component.item_type, ItemType::SubAssembly) {
                topological_visit(db, &component.item_code, visited, result);
            }
        }
    }

    result.push(item_code.to_string());
}

/// Calculate cost for a single item
fn calculate_item_cost(
    bom: &Bom,
    item_costs: &HashMap<String, String>,
    operation_costs: &HashMap<String, Vec<OperationCost>>,
    calculated_costs: &HashMap<String, ItemCostBreakdown>,
    overhead_pct: Option<&str>,
) -> ItemCostBreakdown {
    let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));
    let mut material_cost = Decimal::ZERO;
    let mut labour_cost = Decimal::ZERO;
    let mut overhead_cost = Decimal::ZERO;
    let mut subcontracting_cost = Decimal::ZERO;
    let mut component_costs = Vec::new();

    for component in &bom.components {
        let comp_qty: Decimal = component.quantity.parse().unwrap_or(Decimal::ZERO);
        let scrap_pct: Decimal = component.scrap_percentage.as_ref()
            .and_then(|s| s.parse().ok())
            .unwrap_or(Decimal::ZERO) / dec!(100);

        let qty_with_scrap = comp_qty * (dec!(1) + scrap_pct);

        // Get unit cost
        let unit_cost: Decimal = if matches!(component.item_type, ItemType::SubAssembly) {
            // Use calculated cost from sub-assembly
            calculated_costs.get(&component.item_code)
                .map(|c| c.total_cost.parse().unwrap_or(Decimal::ZERO))
                .unwrap_or_else(|| {
                    component.unit_cost.as_ref()
                        .and_then(|c| c.parse().ok())
                        .unwrap_or(Decimal::ZERO)
                })
        } else {
            // Use item cost or BOM component cost
            item_costs.get(&component.item_code)
                .and_then(|c| c.parse().ok())
                .or_else(|| {
                    component.unit_cost.as_ref().and_then(|c| c.parse().ok())
                })
                .unwrap_or(Decimal::ZERO)
        };

        let total_cost = (qty_with_scrap * unit_cost / bom_qty).round_dp(4);

        let cost_type = match component.item_type {
            ItemType::Service => CostType::Subcontracting,
            _ => CostType::Material,
        };

        match cost_type {
            CostType::Material | CostType::Other => material_cost += total_cost,
            CostType::Subcontracting => subcontracting_cost += total_cost,
            _ => {}
        }

        component_costs.push(ComponentCost {
            item_code: component.item_code.clone(),
            item_name: component.item_name.clone(),
            quantity: qty_with_scrap.round_dp(4).to_string(),
            unit_cost: unit_cost.round_dp(4).to_string(),
            total_cost: total_cost.round_dp(4).to_string(),
            cost_type,
        });
    }

    // Add operation costs
    let ops = operation_costs.get(&bom.item_code)
        .cloned()
        .unwrap_or_default();

    for op in &ops {
        let time: Decimal = op.time_minutes.parse().unwrap_or(Decimal::ZERO);
        let rate: Decimal = op.hourly_rate.parse().unwrap_or(Decimal::ZERO);
        let op_cost = (time / dec!(60) * rate / bom_qty).round_dp(4);

        match op.cost_type {
            CostType::Labour => labour_cost += op_cost,
            CostType::Overhead => overhead_cost += op_cost,
            CostType::Subcontracting => subcontracting_cost += op_cost,
            _ => {}
        }
    }

    // Apply overhead percentage if specified
    if let Some(pct) = overhead_pct {
        let pct_val: Decimal = pct.parse().unwrap_or(Decimal::ZERO) / dec!(100);
        overhead_cost += (material_cost + labour_cost) * pct_val;
    }

    let total = material_cost + labour_cost + overhead_cost + subcontracting_cost;

    ItemCostBreakdown {
        item_code: bom.item_code.clone(),
        item_name: bom.item_name.clone(),
        material_cost: material_cost.round_dp(2).to_string(),
        labour_cost: labour_cost.round_dp(2).to_string(),
        overhead_cost: overhead_cost.round_dp(2).to_string(),
        subcontracting_cost: subcontracting_cost.round_dp(2).to_string(),
        other_cost: "0".to_string(),
        total_cost: total.round_dp(2).to_string(),
        component_costs,
        operation_costs: ops,
    }
}

/// Calculate cost for a specific quantity
#[wasm_bindgen]
pub fn calculate_bom_cost(bom: JsValue, quantity: &str, item_costs: JsValue) -> JsValue {
    let bom: Bom = match serde_wasm_bindgen::from_value(bom) {
        Ok(b) => b,
        Err(e) => {
            log::error!("Invalid BOM: {}", e);
            return JsValue::NULL;
        }
    };

    let costs: HashMap<String, String> = match serde_wasm_bindgen::from_value(item_costs) {
        Ok(c) => c,
        Err(_) => HashMap::new(),
    };

    let qty: Decimal = quantity.parse().unwrap_or(dec!(1));
    let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));

    let mut total_cost = Decimal::ZERO;
    let mut component_details: Vec<serde_json::Value> = Vec::new();

    for component in &bom.components {
        let comp_qty: Decimal = component.quantity.parse().unwrap_or(Decimal::ZERO);
        let scrap_pct: Decimal = component.scrap_percentage.as_ref()
            .and_then(|s| s.parse().ok())
            .unwrap_or(Decimal::ZERO) / dec!(100);

        let required_qty = (comp_qty * qty / bom_qty * (dec!(1) + scrap_pct)).round_dp(4);

        let unit_cost: Decimal = costs.get(&component.item_code)
            .and_then(|c| c.parse().ok())
            .or_else(|| component.unit_cost.as_ref().and_then(|c| c.parse().ok()))
            .unwrap_or(Decimal::ZERO);

        let cost = (required_qty * unit_cost).round_dp(2);
        total_cost += cost;

        component_details.push(serde_json::json!({
            "item_code": component.item_code,
            "item_name": component.item_name,
            "quantity": required_qty.to_string(),
            "uom": component.uom,
            "unit_cost": unit_cost.to_string(),
            "total_cost": cost.to_string()
        }));
    }

    let result = serde_json::json!({
        "item_code": bom.item_code,
        "item_name": bom.item_name,
        "quantity": qty.to_string(),
        "components": component_details,
        "total_material_cost": total_cost.to_string(),
        "unit_cost": (total_cost / qty).round_dp(4).to_string()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Compare costs across different scenarios
#[wasm_bindgen]
pub fn compare_bom_costs(
    bom: JsValue,
    scenarios: JsValue,
) -> JsValue {
    let bom: Bom = match serde_wasm_bindgen::from_value(bom) {
        Ok(b) => b,
        Err(e) => {
            log::error!("Invalid BOM: {}", e);
            return JsValue::NULL;
        }
    };

    let scenarios: Vec<(String, HashMap<String, String>)> = match serde_wasm_bindgen::from_value(scenarios) {
        Ok(s) => s,
        Err(e) => {
            log::error!("Invalid scenarios: {}", e);
            return JsValue::NULL;
        }
    };

    let mut cost_scenarios: Vec<CostScenario> = Vec::new();
    let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));

    for (name, costs) in &scenarios {
        let mut material_cost = Decimal::ZERO;

        for component in &bom.components {
            let comp_qty: Decimal = component.quantity.parse().unwrap_or(Decimal::ZERO);
            let scrap_pct: Decimal = component.scrap_percentage.as_ref()
                .and_then(|s| s.parse().ok())
                .unwrap_or(Decimal::ZERO) / dec!(100);

            let qty_with_scrap = comp_qty * (dec!(1) + scrap_pct) / bom_qty;

            let unit_cost: Decimal = costs.get(&component.item_code)
                .and_then(|c| c.parse().ok())
                .or_else(|| component.unit_cost.as_ref().and_then(|c| c.parse().ok()))
                .unwrap_or(Decimal::ZERO);

            material_cost += qty_with_scrap * unit_cost;
        }

        cost_scenarios.push(CostScenario {
            name: name.clone(),
            material_cost: material_cost.round_dp(2).to_string(),
            labour_cost: "0".to_string(),
            overhead_cost: "0".to_string(),
            total_cost: material_cost.round_dp(2).to_string(),
        });
    }

    // Calculate variance if we have at least 2 scenarios
    let variance = if cost_scenarios.len() >= 2 {
        let first: Decimal = cost_scenarios[0].total_cost.parse().unwrap_or(Decimal::ZERO);
        let second: Decimal = cost_scenarios[1].total_cost.parse().unwrap_or(Decimal::ZERO);
        let variance = second - first;
        let variance_pct = if first > Decimal::ZERO {
            (variance / first * dec!(100)).round_dp(2)
        } else {
            Decimal::ZERO
        };

        VarianceAnalysis {
            material_variance: variance.round_dp(2).to_string(),
            labour_variance: "0".to_string(),
            overhead_variance: "0".to_string(),
            total_variance: variance.round_dp(2).to_string(),
            variance_percentage: format!("{}%", variance_pct),
        }
    } else {
        VarianceAnalysis {
            material_variance: "0".to_string(),
            labour_variance: "0".to_string(),
            overhead_variance: "0".to_string(),
            total_variance: "0".to_string(),
            variance_percentage: "0%".to_string(),
        }
    };

    let result = CostComparisonResult {
        item_code: bom.item_code.clone(),
        scenarios: cost_scenarios,
        variance_analysis: variance,
    };

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate make vs buy analysis
#[wasm_bindgen]
pub fn make_vs_buy_analysis(
    bom: JsValue,
    make_costs: JsValue,
    buy_cost: &str,
    quantity: &str,
) -> JsValue {
    let bom: Bom = match serde_wasm_bindgen::from_value(bom) {
        Ok(b) => b,
        Err(e) => {
            log::error!("Invalid BOM: {}", e);
            return JsValue::NULL;
        }
    };

    let costs: HashMap<String, String> = match serde_wasm_bindgen::from_value(make_costs) {
        Ok(c) => c,
        Err(_) => HashMap::new(),
    };

    let qty: Decimal = quantity.parse().unwrap_or(dec!(1));
    let buy_unit_cost: Decimal = buy_cost.parse().unwrap_or(Decimal::ZERO);
    let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));

    // Calculate make cost
    let mut make_material_cost = Decimal::ZERO;
    for component in &bom.components {
        let comp_qty: Decimal = component.quantity.parse().unwrap_or(Decimal::ZERO);
        let scrap_pct: Decimal = component.scrap_percentage.as_ref()
            .and_then(|s| s.parse().ok())
            .unwrap_or(Decimal::ZERO) / dec!(100);

        let required_qty = (comp_qty * qty / bom_qty * (dec!(1) + scrap_pct)).round_dp(4);

        let unit_cost: Decimal = costs.get(&component.item_code)
            .and_then(|c| c.parse().ok())
            .or_else(|| component.unit_cost.as_ref().and_then(|c| c.parse().ok()))
            .unwrap_or(Decimal::ZERO);

        make_material_cost += required_qty * unit_cost;
    }

    let make_unit_cost = (make_material_cost / qty).round_dp(4);
    let buy_total_cost = (buy_unit_cost * qty).round_dp(2);

    let savings = buy_total_cost - make_material_cost;
    let savings_pct = if buy_total_cost > Decimal::ZERO {
        (savings / buy_total_cost * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    let recommendation = if make_material_cost < buy_total_cost {
        "MAKE"
    } else {
        "BUY"
    };

    let result = serde_json::json!({
        "item_code": bom.item_code,
        "item_name": bom.item_name,
        "quantity": qty.to_string(),
        "make": {
            "unit_cost": make_unit_cost.to_string(),
            "total_cost": make_material_cost.round_dp(2).to_string(),
            "breakdown": "Material cost only"
        },
        "buy": {
            "unit_cost": buy_unit_cost.to_string(),
            "total_cost": buy_total_cost.to_string()
        },
        "savings": savings.round_dp(2).to_string(),
        "savings_percentage": format!("{}%", savings_pct),
        "recommendation": recommendation
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::explosion::BomComponent;

    #[test]
    fn test_cost_calculation() {
        let bom = Bom {
            item_code: "FG001".to_string(),
            item_name: "Finished Product".to_string(),
            quantity: "1".to_string(),
            uom: "NOS".to_string(),
            is_active: true,
            is_default: true,
            components: vec![
                BomComponent {
                    item_code: "RM001".to_string(),
                    item_name: "Raw Material".to_string(),
                    item_type: ItemType::Raw,
                    quantity: "5".to_string(),
                    uom: "KG".to_string(),
                    scrap_percentage: Some("2".to_string()),
                    lead_time_days: None,
                    unit_cost: Some("100".to_string()),
                    warehouse: None,
                    routing_operation: None,
                },
            ],
        };

        let mut costs = HashMap::new();
        costs.insert("RM001".to_string(), "100".to_string());

        // 5 qty * 1.02 scrap * 100 cost = 510
        let breakdown = calculate_item_cost(
            &bom,
            &costs,
            &HashMap::new(),
            &HashMap::new(),
            None,
        );

        assert_eq!(breakdown.material_cost, "510");
    }
}
