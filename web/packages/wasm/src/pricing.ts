/**
 * Samavaya Pricing - TypeScript Bindings
 * Price calculations, discounts, margins using WASM
 */

import { loadWasmModule } from './loader';
import type {
  Discount,
  PriceInput,
  PriceResult,
  MarginResult,
  PriceTier,
  LineTotalResult,
} from './types';

// Type for the raw WASM module
interface PricingWasm {
  calculate_price: (input: unknown) => unknown;
  apply_percentage_discount: (amount: string, percentage: string) => string;
  calculate_discount_amount: (amount: string, percentage: string) => string;
  calculate_discount_percentage: (amount: string, discount: string) => string;
  calculate_margin: (cost: string, sellingPrice: string) => unknown;
  price_from_margin: (cost: string, marginPercentage: string) => string;
  price_from_markup: (cost: string, markupPercentage: string) => string;
  get_tiered_price: (quantity: string, tiers: unknown) => string;
  calculate_line_total: (unitPrice: string, quantity: string, discountPercentage: string) => unknown;
  round_price: (price: string, rounding: string) => string;
}

let wasmModule: PricingWasm | null = null;

/**
 * Initialize the pricing module
 */
async function ensureLoaded(): Promise<PricingWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<PricingWasm>('pricing');
  }
  return wasmModule;
}

// ============================================================================
// Price Calculation Functions
// ============================================================================

/**
 * Calculate price with discounts and tax
 * @param input - Price calculation input
 * @returns Detailed price breakdown
 */
export async function calculatePrice(input: PriceInput): Promise<PriceResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_price(input) as PriceResult;
}

/**
 * Calculate line total for an invoice line
 * @param unitPrice - Unit price
 * @param quantity - Quantity
 * @param discountPercentage - Discount percentage
 * @returns Line total breakdown
 */
export async function calculateLineTotal(
  unitPrice: string,
  quantity: string,
  discountPercentage = '0'
): Promise<LineTotalResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_line_total(unitPrice, quantity, discountPercentage) as LineTotalResult;
}

// ============================================================================
// Discount Functions
// ============================================================================

/**
 * Apply a percentage discount to an amount
 * @param amount - Original amount
 * @param percentage - Discount percentage
 * @returns Discounted amount
 */
export async function applyPercentageDiscount(
  amount: string,
  percentage: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.apply_percentage_discount(amount, percentage);
}

/**
 * Calculate discount amount from percentage
 * @param amount - Base amount
 * @param percentage - Discount percentage
 * @returns Discount amount
 */
export async function calculateDiscountAmount(
  amount: string,
  percentage: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.calculate_discount_amount(amount, percentage);
}

/**
 * Calculate discount percentage from discount amount
 * @param amount - Base amount
 * @param discount - Discount amount
 * @returns Discount percentage
 */
export async function calculateDiscountPercentage(
  amount: string,
  discount: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.calculate_discount_percentage(amount, discount);
}

// ============================================================================
// Margin & Markup Functions
// ============================================================================

/**
 * Calculate margin and markup percentages
 * @param cost - Cost price
 * @param sellingPrice - Selling price
 * @returns Margin calculation result
 */
export async function calculateMargin(
  cost: string,
  sellingPrice: string
): Promise<MarginResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_margin(cost, sellingPrice) as MarginResult;
}

/**
 * Calculate selling price from cost and margin percentage
 * @param cost - Cost price
 * @param marginPercentage - Desired margin percentage
 * @returns Selling price
 */
export async function priceFromMargin(
  cost: string,
  marginPercentage: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.price_from_margin(cost, marginPercentage);
}

/**
 * Calculate selling price from cost and markup percentage
 * @param cost - Cost price
 * @param markupPercentage - Desired markup percentage
 * @returns Selling price
 */
export async function priceFromMarkup(
  cost: string,
  markupPercentage: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.price_from_markup(cost, markupPercentage);
}

// ============================================================================
// Tiered Pricing Functions
// ============================================================================

/**
 * Get price for a quantity based on price tiers
 * @param quantity - Order quantity
 * @param tiers - Array of price tiers
 * @returns Price for the quantity
 */
export async function getTieredPrice(
  quantity: string,
  tiers: PriceTier[]
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.get_tiered_price(quantity, tiers);
}

// ============================================================================
// Price Rounding Functions
// ============================================================================

/**
 * Rounding strategies
 */
export type RoundingStrategy =
  | 'nearest5'
  | 'nearest10'
  | 'nearest50'
  | 'nearest100'
  | 'ceiling'
  | 'floor'
  | 'standard';

/**
 * Round price to a standard value
 * @param price - Price to round
 * @param rounding - Rounding strategy
 * @returns Rounded price
 */
export async function roundPrice(
  price: string,
  rounding: RoundingStrategy = 'standard'
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.round_price(price, rounding);
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Create a simple price input for quick calculations
 * @param basePrice - Base unit price
 * @param quantity - Quantity
 * @param discountPercent - Discount percentage (optional)
 * @param taxRate - Tax rate percentage (optional)
 * @returns PriceInput object
 */
export function createPriceInput(
  basePrice: string,
  quantity: string,
  discountPercent?: string,
  taxRate?: string
): PriceInput {
  const discounts: Discount[] = [];

  if (discountPercent && parseFloat(discountPercent) > 0) {
    discounts.push({
      discountType: 'percentage',
      value: discountPercent,
    });
  }

  return {
    basePrice,
    quantity,
    discounts,
    taxRate,
    includeTax: false,
  };
}

/**
 * Create tiered pricing configuration
 * @param tiers - Array of [minQty, price] or [minQty, maxQty, price] tuples
 * @returns Array of PriceTier objects
 */
export function createPriceTiers(
  tiers: Array<[string, string] | [string, string, string]>
): PriceTier[] {
  return tiers.map((tier, index) => {
    if (tier.length === 2) {
      const nextTier = tiers[index + 1];
      return {
        minQty: tier[0],
        maxQty: nextTier ? String(parseFloat(nextTier[0]) - 1) : undefined,
        price: tier[1],
      };
    }
    return {
      minQty: tier[0],
      maxQty: tier[1] || undefined,
      price: tier[2],
    };
  });
}
