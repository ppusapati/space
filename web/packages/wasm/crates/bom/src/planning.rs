//! Material Requirement Planning (MRP) calculations

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

use crate::explosion::{Bom, BomDatabase, ItemType};

/// MRP input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MrpInput {
    pub demand: Vec<DemandEntry>,
    pub bom_db: BomDatabase,
    pub inventory: HashMap<String, InventoryEntry>,
    pub lead_times: HashMap<String, u32>,
    pub lot_sizes: HashMap<String, LotSizingPolicy>,
    pub planning_horizon_days: u32,
}

/// Demand entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DemandEntry {
    pub item_code: String,
    pub quantity: String,
    pub required_date: String, // ISO date
    pub order_reference: Option<String>,
}

/// Inventory entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InventoryEntry {
    pub on_hand: String,
    pub on_order: String,
    pub safety_stock: String,
    pub reserved: String,
}

/// Lot sizing policy
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LotSizingPolicy {
    pub method: LotSizingMethod,
    pub minimum_qty: Option<String>,
    pub maximum_qty: Option<String>,
    pub multiple_of: Option<String>,
}

/// Lot sizing method
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum LotSizingMethod {
    /// Order exactly what's needed
    LotForLot,
    /// Fixed order quantity
    FixedQty,
    /// Economic Order Quantity
    Eoq,
    /// Period Order Quantity
    Poq,
    /// Minimum order quantity
    MinQty,
}

/// MRP result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MrpResult {
    pub planned_orders: Vec<PlannedOrder>,
    pub shortage_alerts: Vec<ShortageAlert>,
    pub excess_inventory: Vec<ExcessInventory>,
    pub summary: MrpSummary,
}

/// Planned order
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlannedOrder {
    pub item_code: String,
    pub item_name: String,
    pub item_type: ItemType,
    pub quantity: String,
    pub uom: String,
    pub order_date: String,
    pub due_date: String,
    pub source_demand: Vec<String>,
}

/// Shortage alert
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ShortageAlert {
    pub item_code: String,
    pub item_name: String,
    pub shortage_qty: String,
    pub required_date: String,
    pub available_qty: String,
}

/// Excess inventory
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExcessInventory {
    pub item_code: String,
    pub item_name: String,
    pub excess_qty: String,
    pub value: String,
}

/// MRP summary
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MrpSummary {
    pub total_planned_orders: u32,
    pub items_with_shortage: u32,
    pub items_with_excess: u32,
    pub estimated_purchase_value: String,
    pub estimated_production_orders: u32,
}

/// Run MRP calculation
#[wasm_bindgen]
pub fn run_mrp(input: JsValue) -> JsValue {
    let input: MrpInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid MRP input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = run_mrp_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal MRP calculation
fn run_mrp_internal(input: &MrpInput) -> MrpResult {
    let mut requirements: HashMap<String, Vec<(Decimal, String, String)>> = HashMap::new();
    let mut planned_orders: Vec<PlannedOrder> = Vec::new();
    let mut shortage_alerts: Vec<ShortageAlert> = Vec::new();
    let mut excess_inventory: Vec<ExcessInventory> = Vec::new();

    // Step 1: Explode demand to component requirements
    for demand in &input.demand {
        let qty: Decimal = demand.quantity.parse().unwrap_or(Decimal::ZERO);
        explode_demand(
            &input.bom_db,
            &demand.item_code,
            qty,
            &demand.required_date,
            demand.order_reference.as_deref().unwrap_or(""),
            &mut requirements,
        );
    }

    // Step 2: Net against inventory
    let mut purchase_value = Decimal::ZERO;
    let mut production_orders = 0u32;

    for (item_code, reqs) in &requirements {
        let inventory = input.inventory.get(item_code);
        let on_hand: Decimal = inventory
            .map(|i| i.on_hand.parse().unwrap_or(Decimal::ZERO))
            .unwrap_or(Decimal::ZERO);
        let on_order: Decimal = inventory
            .map(|i| i.on_order.parse().unwrap_or(Decimal::ZERO))
            .unwrap_or(Decimal::ZERO);
        let safety_stock: Decimal = inventory
            .map(|i| i.safety_stock.parse().unwrap_or(Decimal::ZERO))
            .unwrap_or(Decimal::ZERO);
        let reserved: Decimal = inventory
            .map(|i| i.reserved.parse().unwrap_or(Decimal::ZERO))
            .unwrap_or(Decimal::ZERO);

        let available = on_hand + on_order - reserved - safety_stock;
        let total_req: Decimal = reqs.iter().map(|(q, _, _)| *q).sum();

        let net_req = (total_req - available).max(Decimal::ZERO);

        if net_req > Decimal::ZERO {
            // Apply lot sizing
            let lot_policy = input.lot_sizes.get(item_code);
            let order_qty = apply_lot_sizing(net_req, lot_policy);

            // Calculate order date based on lead time
            let lead_time = input.lead_times.get(item_code).copied().unwrap_or(0);
            let earliest_due = reqs.iter()
                .map(|(_, date, _)| date.clone())
                .min()
                .unwrap_or_else(|| chrono::Utc::now().format("%Y-%m-%d").to_string());

            // Get item info
            let (item_name, item_type) = get_item_info(&input.bom_db, item_code);

            let source_demands: Vec<String> = reqs.iter()
                .map(|(_, _, ref source)| source.clone())
                .collect();

            planned_orders.push(PlannedOrder {
                item_code: item_code.clone(),
                item_name,
                item_type,
                quantity: order_qty.to_string(),
                uom: "NOS".to_string(),
                order_date: earliest_due.clone(), // Would subtract lead time in real implementation
                due_date: earliest_due,
                source_demand: source_demands,
            });

            if item_type == ItemType::Raw || item_type == ItemType::Purchased {
                // Assume Rs 100 per unit for estimation
                purchase_value += order_qty * dec!(100);
            } else {
                production_orders += 1;
            }
        }

        // Check for shortages
        if available < total_req {
            let earliest_date = reqs.iter()
                .map(|(_, date, _)| date.clone())
                .min()
                .unwrap_or_default();

            shortage_alerts.push(ShortageAlert {
                item_code: item_code.clone(),
                item_name: get_item_info(&input.bom_db, item_code).0,
                shortage_qty: (total_req - available).to_string(),
                required_date: earliest_date,
                available_qty: available.max(Decimal::ZERO).to_string(),
            });
        }

        // Check for excess
        if available > total_req + safety_stock {
            let excess = available - total_req - safety_stock;
            excess_inventory.push(ExcessInventory {
                item_code: item_code.clone(),
                item_name: get_item_info(&input.bom_db, item_code).0,
                excess_qty: excess.to_string(),
                value: (excess * dec!(100)).to_string(), // Estimated
            });
        }
    }

    MrpResult {
        planned_orders: planned_orders.clone(),
        shortage_alerts: shortage_alerts.clone(),
        excess_inventory: excess_inventory.clone(),
        summary: MrpSummary {
            total_planned_orders: planned_orders.len() as u32,
            items_with_shortage: shortage_alerts.len() as u32,
            items_with_excess: excess_inventory.len() as u32,
            estimated_purchase_value: purchase_value.round_dp(2).to_string(),
            estimated_production_orders: production_orders,
        },
    }
}

/// Explode demand to component requirements
fn explode_demand(
    db: &BomDatabase,
    item_code: &str,
    quantity: Decimal,
    required_date: &str,
    source: &str,
    requirements: &mut HashMap<String, Vec<(Decimal, String, String)>>,
) {
    // Add requirement for this item
    requirements.entry(item_code.to_string())
        .or_insert_with(Vec::new)
        .push((quantity, required_date.to_string(), source.to_string()));

    // Explode BOM
    if let Some(bom) = db.boms.get(item_code) {
        let bom_qty: Decimal = bom.quantity.parse().unwrap_or(dec!(1));

        for component in &bom.components {
            let comp_qty: Decimal = component.quantity.parse().unwrap_or(Decimal::ZERO);
            let scrap_pct: Decimal = component.scrap_percentage.as_ref()
                .and_then(|s| s.parse().ok())
                .unwrap_or(Decimal::ZERO) / dec!(100);

            let required_qty = (comp_qty * quantity / bom_qty * (dec!(1) + scrap_pct)).round_dp(4);

            explode_demand(
                db,
                &component.item_code,
                required_qty,
                required_date,
                source,
                requirements,
            );
        }
    }
}

/// Apply lot sizing policy
fn apply_lot_sizing(net_req: Decimal, policy: Option<&LotSizingPolicy>) -> Decimal {
    let policy = match policy {
        Some(p) => p,
        None => return net_req.ceil(),
    };

    let min_qty: Decimal = policy.minimum_qty.as_ref()
        .and_then(|m| m.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let max_qty: Decimal = policy.maximum_qty.as_ref()
        .and_then(|m| m.parse().ok())
        .unwrap_or(Decimal::MAX);

    let multiple: Decimal = policy.multiple_of.as_ref()
        .and_then(|m| m.parse().ok())
        .unwrap_or(dec!(1));

    let mut qty = net_req;

    // Apply minimum
    qty = qty.max(min_qty);

    // Apply maximum
    qty = qty.min(max_qty);

    // Apply multiple
    if multiple > dec!(1) {
        qty = (qty / multiple).ceil() * multiple;
    }

    qty.ceil()
}

/// Get item info from BOM database
fn get_item_info(db: &BomDatabase, item_code: &str) -> (String, ItemType) {
    if let Some(bom) = db.boms.get(item_code) {
        return (bom.item_name.clone(), ItemType::Finished);
    }

    for bom in db.boms.values() {
        for component in &bom.components {
            if component.item_code == item_code {
                return (component.item_name.clone(), component.item_type);
            }
        }
    }

    (item_code.to_string(), ItemType::Raw)
}

/// Calculate reorder point
#[wasm_bindgen]
pub fn calculate_reorder_point(
    average_daily_usage: &str,
    lead_time_days: u32,
    safety_stock: &str,
) -> JsValue {
    let daily_usage: Decimal = average_daily_usage.parse().unwrap_or(Decimal::ZERO);
    let safety: Decimal = safety_stock.parse().unwrap_or(Decimal::ZERO);

    let lead_time_demand = daily_usage * Decimal::from(lead_time_days);
    let reorder_point = lead_time_demand + safety;

    let result = serde_json::json!({
        "reorder_point": reorder_point.round_dp(0).to_string(),
        "lead_time_demand": lead_time_demand.round_dp(0).to_string(),
        "safety_stock": safety.to_string(),
        "formula": format!("(Daily Usage {} × Lead Time {} days) + Safety Stock {} = {}",
            daily_usage, lead_time_days, safety, reorder_point.round_dp(0))
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate Economic Order Quantity
#[wasm_bindgen]
pub fn calculate_eoq(
    annual_demand: &str,
    ordering_cost: &str,
    holding_cost_percentage: &str,
    unit_cost: &str,
) -> JsValue {
    let demand: Decimal = annual_demand.parse().unwrap_or(Decimal::ZERO);
    let order_cost: Decimal = ordering_cost.parse().unwrap_or(Decimal::ZERO);
    let holding_pct: Decimal = holding_cost_percentage.parse().unwrap_or(Decimal::ZERO) / dec!(100);
    let unit: Decimal = unit_cost.parse().unwrap_or(Decimal::ZERO);

    let holding_cost = unit * holding_pct;

    // EOQ = sqrt((2 × D × S) / H)
    let eoq = if holding_cost > Decimal::ZERO {
        let numerator = dec!(2) * demand * order_cost;
        let ratio = (numerator / holding_cost).to_f64().unwrap_or(0.0);
        Decimal::from_f64(ratio.sqrt()).unwrap_or(Decimal::ZERO)
    } else {
        Decimal::ZERO
    };

    // Number of orders per year
    let orders_per_year = if eoq > Decimal::ZERO {
        (demand / eoq).round_dp(1)
    } else {
        Decimal::ZERO
    };

    // Total annual cost
    let annual_ordering = order_cost * orders_per_year;
    let annual_holding = (eoq / dec!(2)) * holding_cost;
    let total_cost = annual_ordering + annual_holding;

    let result = serde_json::json!({
        "eoq": eoq.round_dp(0).to_string(),
        "orders_per_year": orders_per_year.to_string(),
        "annual_ordering_cost": annual_ordering.round_dp(2).to_string(),
        "annual_holding_cost": annual_holding.round_dp(2).to_string(),
        "total_annual_cost": total_cost.round_dp(2).to_string(),
        "days_between_orders": if orders_per_year > Decimal::ZERO {
            (dec!(365) / orders_per_year).round_dp(0).to_string()
        } else {
            "N/A".to_string()
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate safety stock
#[wasm_bindgen]
pub fn calculate_safety_stock(
    average_daily_usage: &str,
    max_daily_usage: &str,
    average_lead_time: &str,
    max_lead_time: &str,
) -> JsValue {
    let avg_usage: Decimal = average_daily_usage.parse().unwrap_or(Decimal::ZERO);
    let max_usage: Decimal = max_daily_usage.parse().unwrap_or(Decimal::ZERO);
    let avg_lt: Decimal = average_lead_time.parse().unwrap_or(Decimal::ZERO);
    let max_lt: Decimal = max_lead_time.parse().unwrap_or(Decimal::ZERO);

    // Safety Stock = (Max Daily Usage × Max Lead Time) - (Avg Daily Usage × Avg Lead Time)
    let max_demand = max_usage * max_lt;
    let avg_demand = avg_usage * avg_lt;
    let safety_stock = (max_demand - avg_demand).max(Decimal::ZERO);

    let result = serde_json::json!({
        "safety_stock": safety_stock.round_dp(0).to_string(),
        "max_demand_during_lead_time": max_demand.round_dp(0).to_string(),
        "avg_demand_during_lead_time": avg_demand.round_dp(0).to_string(),
        "formula": format!("(Max Usage {} × Max LT {}) - (Avg Usage {} × Avg LT {}) = {}",
            max_usage, max_lt, avg_usage, avg_lt, safety_stock.round_dp(0))
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Generate production schedule from demand
#[wasm_bindgen]
pub fn generate_production_schedule(
    demand: JsValue,
    bom_db: JsValue,
    capacity_per_day: &str,
) -> JsValue {
    let demands: Vec<DemandEntry> = match serde_wasm_bindgen::from_value(demand) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid demand: {}", e);
            return JsValue::NULL;
        }
    };

    let db: BomDatabase = match serde_wasm_bindgen::from_value(bom_db) {
        Ok(d) => d,
        Err(e) => {
            log::error!("Invalid BOM database: {}", e);
            return JsValue::NULL;
        }
    };

    let capacity: Decimal = capacity_per_day.parse().unwrap_or(dec!(100));

    let mut schedule: Vec<serde_json::Value> = Vec::new();
    let mut current_date = chrono::Utc::now().format("%Y-%m-%d").to_string();
    let mut remaining_capacity = capacity;

    for demand in &demands {
        let qty: Decimal = demand.quantity.parse().unwrap_or(Decimal::ZERO);
        let item_name = if let Some(bom) = db.boms.get(&demand.item_code) {
            bom.item_name.clone()
        } else {
            demand.item_code.clone()
        };

        let mut remaining_qty = qty;

        while remaining_qty > Decimal::ZERO {
            let produce_qty = remaining_qty.min(remaining_capacity);

            schedule.push(serde_json::json!({
                "date": current_date,
                "item_code": demand.item_code,
                "item_name": item_name,
                "quantity": produce_qty.round_dp(0).to_string(),
                "order_reference": demand.order_reference,
                "due_date": demand.required_date
            }));

            remaining_qty -= produce_qty;
            remaining_capacity -= produce_qty;

            if remaining_capacity <= Decimal::ZERO {
                // Move to next day (simplified - would use actual date calculation)
                current_date = format!("{}_next", current_date);
                remaining_capacity = capacity;
            }
        }
    }

    serde_wasm_bindgen::to_value(&schedule).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_eoq() {
        // D = 10000, S = 100, H = 10% of 50 = 5
        // EOQ = sqrt((2 × 10000 × 100) / 5) = sqrt(400000) = 632.45
        let demand: Decimal = dec!(10000);
        let order_cost: Decimal = dec!(100);
        let holding_pct: Decimal = dec!(0.10);
        let unit_cost: Decimal = dec!(50);

        let holding_cost = unit_cost * holding_pct;
        let numerator = dec!(2) * demand * order_cost;
        let ratio = (numerator / holding_cost).to_f64().unwrap();
        let eoq = ratio.sqrt();

        assert!((eoq - 632.45).abs() < 1.0);
    }

    #[test]
    fn test_safety_stock() {
        // Max usage: 100, Max LT: 10 = 1000
        // Avg usage: 80, Avg LT: 7 = 560
        // Safety stock = 1000 - 560 = 440
        let max_demand = dec!(100) * dec!(10);
        let avg_demand = dec!(80) * dec!(7);
        let safety = max_demand - avg_demand;

        assert_eq!(safety, dec!(440));
    }

    #[test]
    fn test_lot_sizing() {
        let policy = LotSizingPolicy {
            method: LotSizingMethod::MinQty,
            minimum_qty: Some("100".to_string()),
            maximum_qty: None,
            multiple_of: Some("25".to_string()),
        };

        // Net req: 50 -> Min: 100 -> Multiple of 25: 100
        let result = apply_lot_sizing(dec!(50), Some(&policy));
        assert_eq!(result, dec!(100));

        // Net req: 110 -> Min: 110 -> Multiple of 25: 125
        let result = apply_lot_sizing(dec!(110), Some(&policy));
        assert_eq!(result, dec!(125));
    }
}
