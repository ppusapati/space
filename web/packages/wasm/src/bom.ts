/**
 * Samavaya BOM - TypeScript Bindings
 * Bill of Materials calculations using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';
import type {
  BomItem,
  BomExplosionResult,
  BomCostResult,
  WhereUsedResult,
  MrpResult,
  ProductionSchedule,
} from './types';

// Type for the raw WASM module
interface BomWasm {
  explode_bom: (productId: string, bomData: unknown, quantity: string, level?: number) => unknown;
  explode_bom_single_level: (productId: string, bomData: unknown, quantity: string) => unknown;
  validate_bom_circular: (productId: string, bomData: unknown) => unknown;
  get_bom_tree: (productId: string, bomData: unknown) => unknown;
  calculate_cost_rollup: (productId: string, bomData: unknown, itemCosts: unknown) => unknown;
  calculate_bom_cost: (productId: string, bomData: unknown, quantity: string, itemCosts: unknown) => unknown;
  compare_bom_costs: (productId: string, bomVersions: unknown, itemCosts: unknown) => unknown;
  make_vs_buy_analysis: (productId: string, bomData: unknown, makeItemCosts: unknown, buyItemCosts: unknown) => unknown;
  where_used_single_level: (componentId: string, bomData: unknown) => unknown;
  where_used_all_levels: (componentId: string, bomData: unknown) => unknown;
  analyze_impact: (componentId: string, bomData: unknown, impactType: string) => unknown;
  find_common_components: (productIds: unknown, bomData: unknown) => unknown;
  run_mrp: (requirements: unknown, bomData: unknown, inventory: unknown, inTransit: unknown) => unknown;
  calculate_reorder_point: (avgDailyUsage: string, leadTimeDays: number, safetyStockDays: number) => unknown;
  calculate_eoq: (annualDemand: string, orderingCost: string, holdingCostPerUnit: string) => unknown;
  calculate_safety_stock: (avgDailyUsage: string, usageStdDev: string, leadTimeDays: number, serviceLevel: string) => unknown;
  generate_production_schedule: (mrpResult: unknown, productionCapacity: unknown) => unknown;
}

let wasmModule: BomWasm | null = null;

/**
 * Initialize the BOM module
 */
async function ensureLoaded(): Promise<BomWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<BomWasm>('bom' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// BOM Explosion Functions
// ============================================================================

/**
 * Explode BOM to all levels
 * @param productId - Product ID to explode
 * @param bomData - BOM structure data
 * @param quantity - Quantity to explode for
 * @param maxLevel - Maximum level to explode (optional)
 * @returns Multi-level BOM explosion
 */
export async function explodeBom(
  productId: string,
  bomData: Record<string, BomItem[]>,
  quantity: string,
  maxLevel?: number
): Promise<BomExplosionResult> {
  const wasm = await ensureLoaded();
  return wasm.explode_bom(productId, bomData, quantity, maxLevel) as BomExplosionResult;
}

/**
 * Explode BOM to single level only
 * @param productId - Product ID to explode
 * @param bomData - BOM structure data
 * @param quantity - Quantity to explode for
 * @returns Single-level BOM explosion
 */
export async function explodeBomSingleLevel(
  productId: string,
  bomData: Record<string, BomItem[]>,
  quantity: string
): Promise<BomExplosionResult> {
  const wasm = await ensureLoaded();
  return wasm.explode_bom_single_level(productId, bomData, quantity) as BomExplosionResult;
}

/**
 * Validate BOM for circular references
 * @param productId - Product ID to validate
 * @param bomData - BOM structure data
 * @returns Validation result with circular path if found
 */
export async function validateBomCircular(
  productId: string,
  bomData: Record<string, BomItem[]>
): Promise<{ isValid: boolean; circularPath?: string[] }> {
  const wasm = await ensureLoaded();
  return wasm.validate_bom_circular(productId, bomData) as { isValid: boolean; circularPath?: string[] };
}

/**
 * Get BOM as a tree structure
 * @param productId - Product ID
 * @param bomData - BOM structure data
 * @returns Tree representation of BOM
 */
export async function getBomTree(
  productId: string,
  bomData: Record<string, BomItem[]>
): Promise<unknown> {
  const wasm = await ensureLoaded();
  return wasm.get_bom_tree(productId, bomData);
}

// ============================================================================
// Costing Functions
// ============================================================================

/**
 * Calculate cost rollup for a product
 * @param productId - Product ID
 * @param bomData - BOM structure data
 * @param itemCosts - Cost per item
 * @returns Cost rollup with all levels
 */
export async function calculateCostRollup(
  productId: string,
  bomData: Record<string, BomItem[]>,
  itemCosts: Record<string, string>
): Promise<BomCostResult> {
  const wasm = await ensureLoaded();
  return wasm.calculate_cost_rollup(productId, bomData, itemCosts) as BomCostResult;
}

/**
 * Calculate total BOM cost for a quantity
 * @param productId - Product ID
 * @param bomData - BOM structure data
 * @param quantity - Quantity to calculate for
 * @param itemCosts - Cost per item
 * @returns Total cost calculation
 */
export async function calculateBomCost(
  productId: string,
  bomData: Record<string, BomItem[]>,
  quantity: string,
  itemCosts: Record<string, string>
): Promise<{ unitCost: string; totalCost: string; breakdown: BomCostResult }> {
  const wasm = await ensureLoaded();
  return wasm.calculate_bom_cost(productId, bomData, quantity, itemCosts) as {
    unitCost: string;
    totalCost: string;
    breakdown: BomCostResult;
  };
}

/**
 * Compare costs across BOM versions
 * @param productId - Product ID
 * @param bomVersions - Different BOM versions
 * @param itemCosts - Cost per item
 * @returns Cost comparison
 */
export async function compareBomCosts(
  productId: string,
  bomVersions: Record<string, Record<string, BomItem[]>>,
  itemCosts: Record<string, string>
): Promise<Array<{ version: string; cost: string; difference: string }>> {
  const wasm = await ensureLoaded();
  return wasm.compare_bom_costs(productId, bomVersions, itemCosts) as Array<{
    version: string;
    cost: string;
    difference: string;
  }>;
}

/**
 * Make vs Buy analysis
 * @param productId - Product ID
 * @param bomData - BOM structure data
 * @param makeItemCosts - Costs if making in-house
 * @param buyItemCosts - Costs if buying
 * @returns Analysis recommendation
 */
export async function makeVsBuyAnalysis(
  productId: string,
  bomData: Record<string, BomItem[]>,
  makeItemCosts: Record<string, string>,
  buyItemCosts: Record<string, string>
): Promise<{ makeCost: string; buyCost: string; recommendation: 'make' | 'buy'; savings: string }> {
  const wasm = await ensureLoaded();
  return wasm.make_vs_buy_analysis(productId, bomData, makeItemCosts, buyItemCosts) as {
    makeCost: string;
    buyCost: string;
    recommendation: 'make' | 'buy';
    savings: string;
  };
}

// ============================================================================
// Where-Used Functions
// ============================================================================

/**
 * Find where a component is used (single level)
 * @param componentId - Component ID
 * @param bomData - BOM structure data
 * @returns Products using this component
 */
export async function whereUsedSingleLevel(
  componentId: string,
  bomData: Record<string, BomItem[]>
): Promise<WhereUsedResult[]> {
  const wasm = await ensureLoaded();
  return wasm.where_used_single_level(componentId, bomData) as WhereUsedResult[];
}

/**
 * Find where a component is used (all levels)
 * @param componentId - Component ID
 * @param bomData - BOM structure data
 * @returns All products using this component
 */
export async function whereUsedAllLevels(
  componentId: string,
  bomData: Record<string, BomItem[]>
): Promise<WhereUsedResult[]> {
  const wasm = await ensureLoaded();
  return wasm.where_used_all_levels(componentId, bomData) as WhereUsedResult[];
}

/**
 * Analyze impact of component change
 * @param componentId - Component ID
 * @param bomData - BOM structure data
 * @param impactType - Type of impact (cost, availability, quality)
 * @returns Impact analysis
 */
export async function analyzeImpact(
  componentId: string,
  bomData: Record<string, BomItem[]>,
  impactType: 'cost' | 'availability' | 'quality'
): Promise<{ affectedProducts: string[]; impactLevel: string; recommendations: string[] }> {
  const wasm = await ensureLoaded();
  return wasm.analyze_impact(componentId, bomData, impactType) as {
    affectedProducts: string[];
    impactLevel: string;
    recommendations: string[];
  };
}

/**
 * Find common components across products
 * @param productIds - Array of product IDs
 * @param bomData - BOM structure data
 * @returns Common components
 */
export async function findCommonComponents(
  productIds: string[],
  bomData: Record<string, BomItem[]>
): Promise<Array<{ componentId: string; usedIn: string[]; totalQuantity: string }>> {
  const wasm = await ensureLoaded();
  return wasm.find_common_components(productIds, bomData) as Array<{
    componentId: string;
    usedIn: string[];
    totalQuantity: string;
  }>;
}

// ============================================================================
// MRP Functions
// ============================================================================

/**
 * Run Material Requirements Planning
 * @param requirements - Demand requirements
 * @param bomData - BOM structure data
 * @param inventory - Current inventory
 * @param inTransit - In-transit quantities
 * @returns MRP calculation
 */
export async function runMrp(
  requirements: Array<{ productId: string; quantity: string; dueDate: string }>,
  bomData: Record<string, BomItem[]>,
  inventory: Record<string, string>,
  inTransit: Record<string, string>
): Promise<MrpResult> {
  const wasm = await ensureLoaded();
  return wasm.run_mrp(requirements, bomData, inventory, inTransit) as MrpResult;
}

/**
 * Calculate reorder point for an item
 * @param avgDailyUsage - Average daily usage
 * @param leadTimeDays - Lead time in days
 * @param safetyStockDays - Safety stock in days
 * @returns Reorder point
 */
export async function calculateReorderPoint(
  avgDailyUsage: string,
  leadTimeDays: number,
  safetyStockDays: number
): Promise<{ reorderPoint: string; safetyStock: string }> {
  const wasm = await ensureLoaded();
  return wasm.calculate_reorder_point(avgDailyUsage, leadTimeDays, safetyStockDays) as {
    reorderPoint: string;
    safetyStock: string;
  };
}

/**
 * Calculate Economic Order Quantity
 * @param annualDemand - Annual demand
 * @param orderingCost - Cost per order
 * @param holdingCostPerUnit - Holding cost per unit per year
 * @returns EOQ and related metrics
 */
export async function calculateEoq(
  annualDemand: string,
  orderingCost: string,
  holdingCostPerUnit: string
): Promise<{ eoq: string; annualOrderingCost: string; annualHoldingCost: string; totalCost: string }> {
  const wasm = await ensureLoaded();
  return wasm.calculate_eoq(annualDemand, orderingCost, holdingCostPerUnit) as {
    eoq: string;
    annualOrderingCost: string;
    annualHoldingCost: string;
    totalCost: string;
  };
}

/**
 * Calculate safety stock
 * @param avgDailyUsage - Average daily usage
 * @param usageStdDev - Standard deviation of usage
 * @param leadTimeDays - Lead time in days
 * @param serviceLevel - Target service level (e.g., "0.95")
 * @returns Safety stock quantity
 */
export async function calculateSafetyStock(
  avgDailyUsage: string,
  usageStdDev: string,
  leadTimeDays: number,
  serviceLevel: string
): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.calculate_safety_stock(avgDailyUsage, usageStdDev, leadTimeDays, serviceLevel) as string;
}

/**
 * Generate production schedule from MRP
 * @param mrpResult - MRP calculation result
 * @param productionCapacity - Production capacity per day
 * @returns Production schedule
 */
export async function generateProductionSchedule(
  mrpResult: MrpResult,
  productionCapacity: Record<string, string>
): Promise<ProductionSchedule> {
  const wasm = await ensureLoaded();
  return wasm.generate_production_schedule(mrpResult, productionCapacity) as ProductionSchedule;
}
