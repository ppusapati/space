/**
 * Asset Service Factories
 * Typed ConnectRPC clients for assets, depreciation, equipment, maintenance, vehicles
 */
import { getApiClient } from '../client/client.js';
import { AssetService } from '@samavāya/proto/gen/business/asset/asset/proto/asset_pb.js';
import { DepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/depreciation_pb.js';
import { EquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/equipment_pb.js';
import { MaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/maintenance_pb.js';
import { VehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/vehicle_pb.js';
// Vertical-specific — Agriculture
import { AgricultureAssetService } from '@samavāya/proto/gen/business/asset/asset/proto/agriculture/asset_agriculture_pb.js';
import { AgricultureDepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/agriculture/depreciation_agriculture_pb.js';
import { AgricultureEquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/agriculture/equipment_agriculture_pb.js';
import { AgricultureMaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/agriculture/maintenance_agriculture_pb.js';
import { AgricultureVehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/agriculture/vehicle_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionAssetService } from '@samavāya/proto/gen/business/asset/asset/proto/construction/asset_construction_pb.js';
import { ConstructionDepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/construction/depreciation_construction_pb.js';
import { ConstructionEquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/construction/equipment_construction_pb.js';
import { ConstructionMaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/construction/maintenance_construction_pb.js';
import { ConstructionVehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/construction/vehicle_construction_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalAssetService } from '@samavāya/proto/gen/business/asset/asset/proto/mfgvertical/asset_mfgvertical_pb.js';
import { MfgVerticalDepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/mfgvertical/depreciation_mfgvertical_pb.js';
import { MfgVerticalEquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/mfgvertical/equipment_mfgvertical_pb.js';
import { MfgVerticalMaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/mfgvertical/maintenance_mfgvertical_pb.js';
import { MfgVerticalVehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/mfgvertical/vehicle_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarAssetService } from '@samavāya/proto/gen/business/asset/asset/proto/solar/asset_solar_pb.js';
import { solarDepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/solar/depreciation_solar_pb.js';
import { solarEquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/solar/equipment_solar_pb.js';
import { solarMaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/solar/maintenance_solar_pb.js';
import { solarVehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/solar/vehicle_solar_pb.js';
// Vertical-specific — Water
import { WaterAssetService } from '@samavāya/proto/gen/business/asset/asset/proto/water/asset_water_pb.js';
import { WaterDepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/water/depreciation_water_pb.js';
import { WaterEquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/water/equipment_water_pb.js';
import { WaterMaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/water/maintenance_water_pb.js';
import { WaterVehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/water/vehicle_water_pb.js';
export { AssetService, DepreciationService, EquipmentService, MaintenanceService, VehicleService };
export function getAssetService() {
    return getApiClient().getService(AssetService);
}
export function getDepreciationService() {
    return getApiClient().getService(DepreciationService);
}
export function getEquipmentService() {
    return getApiClient().getService(EquipmentService);
}
export function getMaintenanceService() {
    return getApiClient().getService(MaintenanceService);
}
export function getVehicleService() {
    return getApiClient().getService(VehicleService);
}
export { AgricultureAssetService, ConstructionAssetService, MfgVerticalAssetService, solarAssetService, WaterAssetService, AgricultureDepreciationService, ConstructionDepreciationService, MfgVerticalDepreciationService, solarDepreciationService, WaterDepreciationService, AgricultureEquipmentService, ConstructionEquipmentService, MfgVerticalEquipmentService, solarEquipmentService, WaterEquipmentService, AgricultureMaintenanceService, ConstructionMaintenanceService, MfgVerticalMaintenanceService, solarMaintenanceService, WaterMaintenanceService, AgricultureVehicleService, ConstructionVehicleService, MfgVerticalVehicleService, solarVehicleService, WaterVehicleService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureAssetService() {
    return getApiClient().getService(AgricultureAssetService);
}
export function getAgricultureDepreciationService() {
    return getApiClient().getService(AgricultureDepreciationService);
}
export function getAgricultureEquipmentService() {
    return getApiClient().getService(AgricultureEquipmentService);
}
export function getAgricultureMaintenanceService() {
    return getApiClient().getService(AgricultureMaintenanceService);
}
export function getAgricultureVehicleService() {
    return getApiClient().getService(AgricultureVehicleService);
}
// ─── Construction Vertical Factories ───
export function getConstructionAssetService() {
    return getApiClient().getService(ConstructionAssetService);
}
export function getConstructionDepreciationService() {
    return getApiClient().getService(ConstructionDepreciationService);
}
export function getConstructionEquipmentService() {
    return getApiClient().getService(ConstructionEquipmentService);
}
export function getConstructionMaintenanceService() {
    return getApiClient().getService(ConstructionMaintenanceService);
}
export function getConstructionVehicleService() {
    return getApiClient().getService(ConstructionVehicleService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalAssetService() {
    return getApiClient().getService(MfgVerticalAssetService);
}
export function getMfgVerticalDepreciationService() {
    return getApiClient().getService(MfgVerticalDepreciationService);
}
export function getMfgVerticalEquipmentService() {
    return getApiClient().getService(MfgVerticalEquipmentService);
}
export function getMfgVerticalMaintenanceService() {
    return getApiClient().getService(MfgVerticalMaintenanceService);
}
export function getMfgVerticalVehicleService() {
    return getApiClient().getService(MfgVerticalVehicleService);
}
// ─── Solar Vertical Factories ───
export function getSolarAssetService() {
    return getApiClient().getService(solarAssetService);
}
export function getSolarDepreciationService() {
    return getApiClient().getService(solarDepreciationService);
}
export function getSolarEquipmentService() {
    return getApiClient().getService(solarEquipmentService);
}
export function getSolarMaintenanceService() {
    return getApiClient().getService(solarMaintenanceService);
}
export function getSolarVehicleService() {
    return getApiClient().getService(solarVehicleService);
}
// ─── Water Vertical Factories ───
export function getWaterAssetService() {
    return getApiClient().getService(WaterAssetService);
}
export function getWaterDepreciationService() {
    return getApiClient().getService(WaterDepreciationService);
}
export function getWaterEquipmentService() {
    return getApiClient().getService(WaterEquipmentService);
}
export function getWaterMaintenanceService() {
    return getApiClient().getService(WaterMaintenanceService);
}
export function getWaterVehicleService() {
    return getApiClient().getService(WaterVehicleService);
}
//# sourceMappingURL=asset.js.map