/**
 * Inventory Service Factories
 * Typed ConnectRPC clients for inventory core, lot-serial, quality, planning, WMS
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { InventoryCoreService } from '@chetana/proto/gen/business/inventory/core/proto/inventorycore_pb.js';
import { BarcodeService } from '@chetana/proto/gen/business/inventory/barcode/proto/barcode_pb.js';
import { CycleCountService } from '@chetana/proto/gen/business/inventory/cycle-count/proto/cyclecount_pb.js';
import { LotSerialService } from '@chetana/proto/gen/business/inventory/lot-serial/proto/lotserial_pb.js';
import { PlanningService } from '@chetana/proto/gen/business/inventory/planning/proto/planning_pb.js';
import { QualityService } from '@chetana/proto/gen/business/inventory/quality/proto/quality_pb.js';
import { StockTransferService } from '@chetana/proto/gen/business/inventory/stock-transfer/proto/stocktransfer_pb.js';
import { WMSService } from '@chetana/proto/gen/business/inventory/wms/proto/wms_pb.js';

// Vertical-specific — Agriculture
// AgricultureInventoryCoreService retired in Phase F.6.1 — callers use
// getInventoryService() with class="agri_cold_store" or
// "agri_warehouse_receipt_store" via the unified Warehouse classregistry.
// AgricultureLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="agri_harvest_lot".
// AgriculturePlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="agri_seasonal_plan".
// AgricultureQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="agri_produce_inspection".
// AgricultureWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="agri_bulk_despatch_wave".
// Vertical-specific — Construction
// ConstructionLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="construction_material_lot".
// ConstructionInventoryPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="construction_project_plan".
// ConstructionInventoryQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="construction_material_inspection".
// Vertical-specific — Construction Vertical
// ConstructionVerticalInventoryService retired in Phase F.6.1 — callers use
// getInventoryService() with class="construction_site_store".
// Vertical-specific — MfgVertical (Manufacturing)
// MfgVerticalInventoryCoreService retired in Phase F.6.1 — callers use
// getInventoryService() with class="manufacturing_plant_store".
// MfgVerticalLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="manufacturing_wip_lot".
// MfgVerticalPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="manufacturing_plant_plan".
// MfgVerticalQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="manufacturing_shopfloor_inspection".
// MfgVerticalWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="manufacturing_line_feed_wave".
// Vertical-specific — Solar
// SolarInventoryCoreService retired in Phase F.6.1 — callers use
// getInventoryService() with class="solar_plant_store".
// SolarLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="solar_module_lot".
// SolarPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="solar_om_plan".
// SolarQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="solar_module_flash_test".
// SolarWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="solar_plant_despatch_wave".
// Vertical-specific — Water
// WaterInventoryService retired in Phase F.6.1 — callers use
// getInventoryService() with class="water_utility_store".
// WaterLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="water_chemical_lot".
// WaterInventoryPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="water_utility_programme_plan".
// WaterInventoryQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="water_lab_testing".
// WaterWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="water_utility_despatch_wave".
// Vertical-specific — Work Vertical
// WorkVerticalInventoryService retired in Phase F.6.1 — callers use
// getInventoryService() with a construction/real-estate warehouse class.

export {
  InventoryCoreService, BarcodeService, CycleCountService, LotSerialService,
  PlanningService, QualityService, StockTransferService, WMSService,
};

export function getInventoryService(): Client<typeof InventoryCoreService> {
  return getApiClient().getService(InventoryCoreService);
}

export function getBarcodeService(): Client<typeof BarcodeService> {
  return getApiClient().getService(BarcodeService);
}

export function getCycleCountService(): Client<typeof CycleCountService> {
  return getApiClient().getService(CycleCountService);
}

export function getLotSerialService(): Client<typeof LotSerialService> {
  return getApiClient().getService(LotSerialService);
}

export function getInventoryPlanningService(): Client<typeof PlanningService> {
  return getApiClient().getService(PlanningService);
}

export function getQualityService(): Client<typeof QualityService> {
  return getApiClient().getService(QualityService);
}

export function getStockTransferService(): Client<typeof StockTransferService> {
  return getApiClient().getService(StockTransferService);
}

export function getWMSService(): Client<typeof WMSService> {
  return getApiClient().getService(WMSService);
}


// All inventory vertical services retired in Phase F.6.1–F.6.5:
// - F.6.1 InventoryCore verticals → Warehouse classregistry
// - F.6.2 LotSerial verticals → Lot classregistry
// - F.6.3 Planning verticals → ReplenishmentPlan classregistry
// - F.6.4 Quality verticals → InspectionPlan classregistry
// - F.6.5 WMS verticals → Wave classregistry

// ─── Agriculture Vertical Factories ───

// getAgricultureInventoryCoreService retired in Phase F.6.1 — callers use
// getInventoryService() with class="agri_cold_store" or "agri_warehouse_receipt_store".
// getAgricultureLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="agri_harvest_lot".
// getAgriculturePlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="agri_seasonal_plan".

// getAgricultureQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="agri_produce_inspection".

// getAgricultureWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="agri_bulk_despatch_wave".

// ─── Construction Vertical Factories ───

// getConstructionLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="construction_material_lot".
// getConstructionInventoryPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="construction_project_plan".

// getConstructionInventoryQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="construction_material_inspection".

// ─── Construction Vertical Vertical Factories ───

// getConstructionVerticalInventoryService retired in Phase F.6.1 — callers use
// getInventoryService() with class="construction_site_store".

// ─── MfgVertical (Manufacturing) Vertical Factories ───

// getMfgVerticalInventoryCoreService retired in Phase F.6.1 — callers use
// getInventoryService() with class="manufacturing_plant_store".
// getMfgVerticalLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="manufacturing_wip_lot".
// getMfgVerticalPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="manufacturing_plant_plan".

// getMfgVerticalQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="manufacturing_shopfloor_inspection".

// getMfgVerticalWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="manufacturing_line_feed_wave".

// ─── Solar Vertical Factories ───

// getSolarInventoryCoreService retired in Phase F.6.1 — callers use
// getInventoryService() with class="solar_plant_store".
// getSolarLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="solar_module_lot".
// getSolarPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="solar_om_plan".

// getSolarQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="solar_module_flash_test".

// getSolarWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="solar_plant_despatch_wave".

// ─── Water Vertical Factories ───

// getWaterInventoryService retired in Phase F.6.1 — callers use
// getInventoryService() with class="water_utility_store".
// getWaterLotSerialService retired in Phase F.6.2 — callers use
// getLotSerialService() with class="water_chemical_lot".
// getWaterInventoryPlanningService retired in Phase F.6.3 — callers use
// getInventoryPlanningService() with class="water_utility_programme_plan".

// getWaterInventoryQualityService retired in Phase F.6.4 — callers use
// getQualityService() with class="water_lab_testing".

// getWaterWMSService retired in Phase F.6.5 — callers use
// getWMSService() with class="water_utility_despatch_wave".

// ─── Work Vertical Vertical Factories ───

// getWorkVerticalInventoryService retired in Phase F.6.1 — callers use
// getInventoryService() with a construction/real-estate warehouse class.

