//! Samavaya Pricing - Price calculations and discount evaluation
//!
//! This crate provides:
//! - Price list evaluation
//! - Discount calculations (percentage, amount, tiered)
//! - Margin calculations
//! - Currency conversion
//! - Trade/Cash discount handling

use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Initialize pricing module (called from core init)
fn pricing_init() {
    log::info!("Samavaya Pricing module initialized");
}

/// Discount type
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum DiscountType {
    Percentage,
    Amount,
    Tiered,
}

/// Discount configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Discount {
    pub discount_type: String,
    pub value: String,
    pub min_quantity: Option<String>,
    pub max_quantity: Option<String>,
    pub min_amount: Option<String>,
}

/// Price calculation input
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceInput {
    pub base_price: String,
    pub quantity: String,
    pub discounts: Vec<Discount>,
    pub tax_rate: Option<String>,
    pub include_tax: Option<bool>,
}

/// Price calculation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceResult {
    pub base_price: String,
    pub quantity: String,
    pub gross_amount: String,
    pub discount_amount: String,
    pub discount_percentage: String,
    pub net_amount: String,
    pub tax_amount: String,
    pub total_amount: String,
    pub unit_price_after_discount: String,
    pub effective_rate: String,
    pub breakdown: Vec<DiscountBreakdown>,
}

/// Discount breakdown
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscountBreakdown {
    pub description: String,
    pub discount_type: String,
    pub value: String,
    pub amount: String,
}

/// Calculate price with discounts
#[wasm_bindgen]
pub fn calculate_price(input: JsValue) -> JsValue {
    let input: PriceInput = match serde_wasm_bindgen::from_value(input) {
        Ok(i) => i,
        Err(e) => {
            log::error!("Invalid price input: {}", e);
            return JsValue::NULL;
        }
    };

    let result = calculate_price_internal(&input);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal price calculation
fn calculate_price_internal(input: &PriceInput) -> PriceResult {
    let base_price: Decimal = input.base_price.parse().unwrap_or(Decimal::ZERO);
    let quantity: Decimal = input.quantity.parse().unwrap_or(dec!(1));

    let gross_amount = base_price * quantity;
    let mut net_amount = gross_amount;
    let mut total_discount = Decimal::ZERO;
    let mut breakdown = Vec::new();

    // Apply discounts
    for discount in &input.discounts {
        // Check quantity limits
        if let Some(ref min_qty) = discount.min_quantity {
            let min: Decimal = min_qty.parse().unwrap_or(Decimal::ZERO);
            if quantity < min {
                continue;
            }
        }
        if let Some(ref max_qty) = discount.max_quantity {
            let max: Decimal = max_qty.parse().unwrap_or(Decimal::MAX);
            if quantity > max {
                continue;
            }
        }

        // Check amount limits
        if let Some(ref min_amt) = discount.min_amount {
            let min: Decimal = min_amt.parse().unwrap_or(Decimal::ZERO);
            if net_amount < min {
                continue;
            }
        }

        let discount_value: Decimal = discount.value.parse().unwrap_or(Decimal::ZERO);
        let discount_amount: Decimal;

        match discount.discount_type.to_lowercase().as_str() {
            "percentage" | "percent" | "%" => {
                discount_amount = (net_amount * discount_value / dec!(100)).round_dp(2);
                breakdown.push(DiscountBreakdown {
                    description: format!("{}% discount", discount_value),
                    discount_type: "percentage".to_string(),
                    value: discount_value.to_string(),
                    amount: discount_amount.to_string(),
                });
            }
            "amount" | "fixed" => {
                discount_amount = discount_value.min(net_amount);
                breakdown.push(DiscountBreakdown {
                    description: "Fixed discount".to_string(),
                    discount_type: "amount".to_string(),
                    value: discount_value.to_string(),
                    amount: discount_amount.to_string(),
                });
            }
            _ => {
                discount_amount = Decimal::ZERO;
            }
        }

        total_discount += discount_amount;
        net_amount -= discount_amount;
    }

    // Ensure net amount is not negative
    net_amount = net_amount.max(Decimal::ZERO);

    // Calculate tax
    let tax_rate: Decimal = input.tax_rate.as_ref()
        .and_then(|r| r.parse().ok())
        .unwrap_or(Decimal::ZERO);

    let (tax_amount, total_amount) = if input.include_tax.unwrap_or(false) {
        // Price is tax-inclusive, extract tax
        let tax = net_amount - (net_amount / (dec!(1) + tax_rate / dec!(100)));
        (tax.round_dp(2), net_amount)
    } else {
        // Add tax to net amount
        let tax = (net_amount * tax_rate / dec!(100)).round_dp(2);
        (tax, net_amount + tax)
    };

    // Calculate effective values
    let discount_percentage = if gross_amount > Decimal::ZERO {
        (total_discount / gross_amount * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    let unit_price_after_discount = if quantity > Decimal::ZERO {
        (net_amount / quantity).round_dp(2)
    } else {
        Decimal::ZERO
    };

    let effective_rate = if quantity > Decimal::ZERO {
        (total_amount / quantity).round_dp(2)
    } else {
        Decimal::ZERO
    };

    PriceResult {
        base_price: base_price.round_dp(2).to_string(),
        quantity: quantity.to_string(),
        gross_amount: gross_amount.round_dp(2).to_string(),
        discount_amount: total_discount.round_dp(2).to_string(),
        discount_percentage: discount_percentage.to_string(),
        net_amount: net_amount.round_dp(2).to_string(),
        tax_amount: tax_amount.to_string(),
        total_amount: total_amount.round_dp(2).to_string(),
        unit_price_after_discount: unit_price_after_discount.to_string(),
        effective_rate: effective_rate.to_string(),
        breakdown,
    }
}

/// Calculate simple percentage discount
#[wasm_bindgen]
pub fn apply_percentage_discount(amount: &str, percentage: &str) -> String {
    let amount: Decimal = amount.parse().unwrap_or(Decimal::ZERO);
    let percentage: Decimal = percentage.parse().unwrap_or(Decimal::ZERO);

    let discount = (amount * percentage / dec!(100)).round_dp(2);
    (amount - discount).round_dp(2).to_string()
}

/// Calculate discount amount from percentage
#[wasm_bindgen]
pub fn calculate_discount_amount(amount: &str, percentage: &str) -> String {
    let amount: Decimal = amount.parse().unwrap_or(Decimal::ZERO);
    let percentage: Decimal = percentage.parse().unwrap_or(Decimal::ZERO);

    (amount * percentage / dec!(100)).round_dp(2).to_string()
}

/// Calculate percentage from discount amount
#[wasm_bindgen]
pub fn calculate_discount_percentage(amount: &str, discount: &str) -> String {
    let amount: Decimal = amount.parse().unwrap_or(Decimal::ZERO);
    let discount: Decimal = discount.parse().unwrap_or(Decimal::ZERO);

    if amount.is_zero() {
        return "0".to_string();
    }

    (discount / amount * dec!(100)).round_dp(2).to_string()
}

/// Calculate margin percentage
#[wasm_bindgen]
pub fn calculate_margin(cost: &str, selling_price: &str) -> JsValue {
    let cost: Decimal = cost.parse().unwrap_or(Decimal::ZERO);
    let selling_price: Decimal = selling_price.parse().unwrap_or(Decimal::ZERO);

    let profit = selling_price - cost;

    let margin_percentage = if selling_price > Decimal::ZERO {
        (profit / selling_price * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    let markup_percentage = if cost > Decimal::ZERO {
        (profit / cost * dec!(100)).round_dp(2)
    } else {
        Decimal::ZERO
    };

    let result = serde_json::json!({
        "cost": cost.round_dp(2).to_string(),
        "sellingPrice": selling_price.round_dp(2).to_string(),
        "profit": profit.round_dp(2).to_string(),
        "marginPercentage": margin_percentage.to_string(),
        "markupPercentage": markup_percentage.to_string(),
        "isProfitable": profit > Decimal::ZERO
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Calculate selling price from cost and margin
#[wasm_bindgen]
pub fn price_from_margin(cost: &str, margin_percentage: &str) -> String {
    let cost: Decimal = cost.parse().unwrap_or(Decimal::ZERO);
    let margin: Decimal = margin_percentage.parse().unwrap_or(Decimal::ZERO);

    if margin >= dec!(100) {
        return "0".to_string();
    }

    // Selling Price = Cost / (1 - Margin%)
    let selling_price = cost / (dec!(1) - margin / dec!(100));
    selling_price.round_dp(2).to_string()
}

/// Calculate selling price from cost and markup
#[wasm_bindgen]
pub fn price_from_markup(cost: &str, markup_percentage: &str) -> String {
    let cost: Decimal = cost.parse().unwrap_or(Decimal::ZERO);
    let markup: Decimal = markup_percentage.parse().unwrap_or(Decimal::ZERO);

    // Selling Price = Cost × (1 + Markup%)
    let selling_price = cost * (dec!(1) + markup / dec!(100));
    selling_price.round_dp(2).to_string()
}

/// Apply tiered pricing
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceTier {
    pub min_qty: String,
    pub max_qty: Option<String>,
    pub price: String,
}

/// Get price for quantity based on tiers
#[wasm_bindgen]
pub fn get_tiered_price(quantity: &str, tiers: JsValue) -> String {
    let quantity: Decimal = quantity.parse().unwrap_or(dec!(1));
    let tiers: Vec<PriceTier> = serde_wasm_bindgen::from_value(tiers).unwrap_or_default();

    for tier in &tiers {
        let min: Decimal = tier.min_qty.parse().unwrap_or(Decimal::ZERO);
        let max: Decimal = tier.max_qty.as_ref()
            .and_then(|m| m.parse().ok())
            .unwrap_or(Decimal::MAX);

        if quantity >= min && quantity <= max {
            return tier.price.clone();
        }
    }

    // Return first tier price as default
    tiers.first()
        .map(|t| t.price.clone())
        .unwrap_or_else(|| "0".to_string())
}

/// Calculate line total
#[wasm_bindgen]
pub fn calculate_line_total(
    unit_price: &str,
    quantity: &str,
    discount_percentage: &str,
) -> JsValue {
    let price: Decimal = unit_price.parse().unwrap_or(Decimal::ZERO);
    let qty: Decimal = quantity.parse().unwrap_or(dec!(1));
    let discount: Decimal = discount_percentage.parse().unwrap_or(Decimal::ZERO);

    let gross = price * qty;
    let discount_amount = (gross * discount / dec!(100)).round_dp(2);
    let net = gross - discount_amount;

    let result = serde_json::json!({
        "grossAmount": gross.round_dp(2).to_string(),
        "discountAmount": discount_amount.to_string(),
        "netAmount": net.round_dp(2).to_string(),
        "effectivePrice": (net / qty).round_dp(2).to_string()
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Round price to nearest standard value
#[wasm_bindgen]
pub fn round_price(price: &str, rounding: &str) -> String {
    let price: Decimal = price.parse().unwrap_or(Decimal::ZERO);

    let rounded = match rounding.to_lowercase().as_str() {
        "nearest5" => (price / dec!(5)).round() * dec!(5),
        "nearest10" => (price / dec!(10)).round() * dec!(10),
        "nearest50" => (price / dec!(50)).round() * dec!(50),
        "nearest100" => (price / dec!(100)).round() * dec!(100),
        "ceiling" => price.ceil(),
        "floor" => price.floor(),
        _ => price.round_dp(2),
    };

    rounded.to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_percentage_discount() {
        assert_eq!(apply_percentage_discount("1000", "10"), "900.00");
        assert_eq!(apply_percentage_discount("500", "25"), "375.00");
    }

    #[test]
    fn test_margin_calculation() {
        // Cost 80, Selling 100 => 20% margin, 25% markup
        let result = calculate_margin("80", "100");
        // Just verify it returns something valid
        assert!(!result.is_null());
    }

    #[test]
    fn test_price_from_margin() {
        // Cost 80, Margin 20% => Selling 100
        assert_eq!(price_from_margin("80", "20"), "100.00");
    }

    #[test]
    fn test_price_from_markup() {
        // Cost 80, Markup 25% => Selling 100
        assert_eq!(price_from_markup("80", "25"), "100.00");
    }

    #[test]
    fn test_round_price() {
        assert_eq!(round_price("123.45", "nearest5"), "125");
        assert_eq!(round_price("123.45", "nearest10"), "120");
        assert_eq!(round_price("123.45", "ceiling"), "124");
        assert_eq!(round_price("123.45", "floor"), "123");
    }
}
