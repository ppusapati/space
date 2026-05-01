/**
 * Inventory Service Factories
 * Typed ConnectRPC clients for inventory core, lot-serial, quality, planning, WMS
 */
import { getApiClient } from '../client/client.js';
import { InventoryCoreService } from '@samavāya/proto/gen/business/inventory/core/proto/inventorycore_pb.js';
import { BarcodeService } from '@samavāya/proto/gen/business/inventory/barcode/proto/barcode_pb.js';
import { CycleCountService } from '@samavāya/proto/gen/business/inventory/cycle-count/proto/cyclecount_pb.js';
import { LotSerialService } from '@samavāya/proto/gen/business/inventory/lot-serial/proto/lotserial_pb.js';
import { PlanningService } from '@samavāya/proto/gen/business/inventory/planning/proto/planning_pb.js';
import { QualityService } from '@samavāya/proto/gen/business/inventory/quality/proto/quality_pb.js';
import { StockTransferService } from '@samavāya/proto/gen/business/inventory/stock-transfer/proto/stocktransfer_pb.js';
import { WMSService } from '@samavāya/proto/gen/business/inventory/wms/proto/wms_pb.js';
// Vertical-specific — Agriculture
import { AgricultureInventoryCoreService } from '@samavāya/proto/gen/business/inventory/core/proto/agriculture/inventorycore_agri_pb.js';
import { AgricultureLotSerialService } from '@samavāya/proto/gen/business/inventory/lot-serial/proto/agriculture/lotserial_agri_pb.js';
import { AgriculturePlanningService } from '@samavāya/proto/gen/business/inventory/planning/proto/agriculture/planning_agri_pb.js';
import { AgricultureQualityService } from '@samavāya/proto/gen/business/inventory/quality/proto/agriculture/quality_agri_pb.js';
import { AgricultureWMSService } from '@samavāya/proto/gen/business/inventory/wms/proto/agriculture/wms_agri_pb.js';
// Vertical-specific — Construction
import { ConstructionLotSerialService } from '@samavāya/proto/gen/business/inventory/lot-serial/proto/construction/lot-serial_construction_pb.js';
import { ConstructionInventoryPlanningService } from '@samavāya/proto/gen/business/inventory/planning/proto/construction/planning_construction_pb.js';
import { ConstructionInventoryQualityService } from '@samavāya/proto/gen/business/inventory/quality/proto/construction/quality_construction_pb.js';
// Vertical-specific — Construction Vertical
import { ConstructionVerticalInventoryService } from '@samavāya/proto/gen/business/inventory/core/proto/constructionvertical/inventory_constructionvertical_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalInventoryCoreService } from '@samavāya/proto/gen/business/inventory/core/proto/mfgvertical/inventorycore_mfgvertical_pb.js';
import { MfgVerticalLotSerialService } from '@samavāya/proto/gen/business/inventory/lot-serial/proto/mfgvertical/lotserial_mfgvertical_pb.js';
import { MfgVerticalPlanningService } from '@samavāya/proto/gen/business/inventory/planning/proto/mfgvertical/planning_mfgvertical_pb.js';
import { MfgVerticalQualityService } from '@samavāya/proto/gen/business/inventory/quality/proto/mfgvertical/quality_mfgvertical_pb.js';
import { MfgVerticalWMSService } from '@samavāya/proto/gen/business/inventory/wms/proto/mfgvertical/wms_mfgvertical_pb.js';
// Vertical-specific — Solar
import { SolarInventoryCoreService } from '@samavāya/proto/gen/business/inventory/core/proto/solar/inventorycore_solar_pb.js';
import { SolarLotSerialService } from '@samavāya/proto/gen/business/inventory/lot-serial/proto/solar/lotserial_solar_pb.js';
import { SolarPlanningService } from '@samavāya/proto/gen/business/inventory/planning/proto/solar/planning_solar_pb.js';
import { SolarQualityService } from '@samavāya/proto/gen/business/inventory/quality/proto/solar/quality_solar_pb.js';
import { SolarWMSService } from '@samavāya/proto/gen/business/inventory/wms/proto/solar/wms_solar_pb.js';
// Vertical-specific — Water
import { WaterInventoryService } from '@samavāya/proto/gen/business/inventory/core/proto/water/core_water_pb.js';
import { WaterLotSerialService } from '@samavāya/proto/gen/business/inventory/lot-serial/proto/water/lot-serial_water_pb.js';
import { WaterInventoryPlanningService } from '@samavāya/proto/gen/business/inventory/planning/proto/water/planning_water_pb.js';
import { WaterInventoryQualityService } from '@samavāya/proto/gen/business/inventory/quality/proto/water/quality_water_pb.js';
import { WaterWMSService } from '@samavāya/proto/gen/business/inventory/wms/proto/water/wms_water_pb.js';
// Vertical-specific — Work Vertical
import { WorkVerticalInventoryService } from '@samavāya/proto/gen/business/inventory/core/proto/workvertical/inventory_workvertical_pb.js';
export { InventoryCoreService, BarcodeService, CycleCountService, LotSerialService, PlanningService, QualityService, StockTransferService, WMSService, };
export function getInventoryService() {
    return getApiClient().getService(InventoryCoreService);
}
export function getBarcodeService() {
    return getApiClient().getService(BarcodeService);
}
export function getCycleCountService() {
    return getApiClient().getService(CycleCountService);
}
export function getLotSerialService() {
    return getApiClient().getService(LotSerialService);
}
export function getInventoryPlanningService() {
    return getApiClient().getService(PlanningService);
}
export function getQualityService() {
    return getApiClient().getService(QualityService);
}
export function getStockTransferService() {
    return getApiClient().getService(StockTransferService);
}
export function getWMSService() {
    return getApiClient().getService(WMSService);
}
export { AgricultureInventoryCoreService, ConstructionVerticalInventoryService, MfgVerticalInventoryCoreService, SolarInventoryCoreService, WaterInventoryService, WorkVerticalInventoryService, AgricultureLotSerialService, ConstructionLotSerialService, MfgVerticalLotSerialService, SolarLotSerialService, WaterLotSerialService, AgriculturePlanningService, ConstructionInventoryPlanningService, MfgVerticalPlanningService, SolarPlanningService, WaterInventoryPlanningService, AgricultureQualityService, ConstructionInventoryQualityService, MfgVerticalQualityService, SolarQualityService, WaterInventoryQualityService, AgricultureWMSService, MfgVerticalWMSService, SolarWMSService, WaterWMSService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureInventoryCoreService() {
    return getApiClient().getService(AgricultureInventoryCoreService);
}
export function getAgricultureLotSerialService() {
    return getApiClient().getService(AgricultureLotSerialService);
}
export function getAgriculturePlanningService() {
    return getApiClient().getService(AgriculturePlanningService);
}
export function getAgricultureQualityService() {
    return getApiClient().getService(AgricultureQualityService);
}
export function getAgricultureWMSService() {
    return getApiClient().getService(AgricultureWMSService);
}
// ─── Construction Vertical Factories ───
export function getConstructionLotSerialService() {
    return getApiClient().getService(ConstructionLotSerialService);
}
export function getConstructionInventoryPlanningService() {
    return getApiClient().getService(ConstructionInventoryPlanningService);
}
export function getConstructionInventoryQualityService() {
    return getApiClient().getService(ConstructionInventoryQualityService);
}
// ─── Construction Vertical Vertical Factories ───
export function getConstructionVerticalInventoryService() {
    return getApiClient().getService(ConstructionVerticalInventoryService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalInventoryCoreService() {
    return getApiClient().getService(MfgVerticalInventoryCoreService);
}
export function getMfgVerticalLotSerialService() {
    return getApiClient().getService(MfgVerticalLotSerialService);
}
export function getMfgVerticalPlanningService() {
    return getApiClient().getService(MfgVerticalPlanningService);
}
export function getMfgVerticalQualityService() {
    return getApiClient().getService(MfgVerticalQualityService);
}
export function getMfgVerticalWMSService() {
    return getApiClient().getService(MfgVerticalWMSService);
}
// ─── Solar Vertical Factories ───
export function getSolarInventoryCoreService() {
    return getApiClient().getService(SolarInventoryCoreService);
}
export function getSolarLotSerialService() {
    return getApiClient().getService(SolarLotSerialService);
}
export function getSolarPlanningService() {
    return getApiClient().getService(SolarPlanningService);
}
export function getSolarQualityService() {
    return getApiClient().getService(SolarQualityService);
}
export function getSolarWMSService() {
    return getApiClient().getService(SolarWMSService);
}
// ─── Water Vertical Factories ───
export function getWaterInventoryService() {
    return getApiClient().getService(WaterInventoryService);
}
export function getWaterLotSerialService() {
    return getApiClient().getService(WaterLotSerialService);
}
export function getWaterInventoryPlanningService() {
    return getApiClient().getService(WaterInventoryPlanningService);
}
export function getWaterInventoryQualityService() {
    return getApiClient().getService(WaterInventoryQualityService);
}
export function getWaterWMSService() {
    return getApiClient().getService(WaterWMSService);
}
// ─── Work Vertical Vertical Factories ───
export function getWorkVerticalInventoryService() {
    return getApiClient().getService(WorkVerticalInventoryService);
}
//# sourceMappingURL=inventory.js.map