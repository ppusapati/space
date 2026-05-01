/**
 * Asset Service Factories
 * Typed ConnectRPC clients for assets, depreciation, equipment, maintenance, vehicles
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { AssetService } from '@samavāya/proto/gen/business/asset/asset/proto/asset_pb.js';
import { DepreciationService } from '@samavāya/proto/gen/business/asset/depreciation/proto/depreciation_pb.js';
import { EquipmentService } from '@samavāya/proto/gen/business/asset/equipment/proto/equipment_pb.js';
import { MaintenanceService } from '@samavāya/proto/gen/business/asset/maintenance/proto/maintenance_pb.js';
import { VehicleService } from '@samavāya/proto/gen/business/asset/vehicle/proto/vehicle_pb.js';

// Vertical-specific — Agriculture
// AgricultureAssetService retired — Phase F.2.3 consolidation. Use
// AssetService with class="farmland" / "orchard" and the attributes
// in config/class_registry/asset.yaml.
// AgricultureDepreciationService retired — Phase F.2.5 consolidation. Use
// DepreciationService with class="agri_bearer_plant" / "incometax_block"
// / "indas_book" per config/class_registry/depreciation.yaml.
// AgricultureEquipmentService retired — Phase F.2.2 consolidation. Use
// EquipmentService with class="agriculture_sprayer" / "grain_silo" /
// "cold_storage" / "water_pump" and the attributes in
// config/class_registry/equipment.yaml.
// AgricultureMaintenanceService retired — Phase F.2.4 consolidation. Use
// MaintenanceService with class="seasonal_service" / "irrigation_check" /
// "preventive_service" per config/class_registry/maintenance.yaml.
// AgricultureVehicleService retired — Phase F.2.1 consolidation. Use
// VehicleService with class="farm_tractor" / "farm_harvester" /
// "cold_chain_reefer" and the attributes in
// config/class_registry/vehicle.yaml.
// Vertical-specific — Construction
// ConstructionAssetService retired — Phase F.2.3 consolidation. Use
// AssetService with class="office_building" / "warehouse" /
// "construction_site".
// ConstructionDepreciationService retired — Phase F.2.5 consolidation. Use
// DepreciationService with class="realestate_structure" / "indas_book"
// / "companies_act_book".
// ConstructionEquipmentService retired — Phase F.2.2 consolidation. Use
// EquipmentService with class="construction_equipment".
// ConstructionMaintenanceService retired — Phase F.2.4 consolidation. Use
// MaintenanceService with class="site_safety_walk" / "corrective_repair"
// / "statutory_inspection".
// ConstructionVehicleService retired — Phase F.2.1 consolidation. Use
// VehicleService with class="construction_dumper" / "construction_mixer".
// Vertical-specific — MfgVertical (Manufacturing)
// MfgVerticalAssetService retired — Phase F.2.3 consolidation. Use
// AssetService with class="factory_land" / "factory_building".
// MfgVerticalDepreciationService retired — Phase F.2.5 consolidation. Use
// DepreciationService with class="indas_book" / "incometax_block" /
// "ifrs_book" / "us_macrs_tax".
// MfgVerticalEquipmentService retired — Phase F.2.2 consolidation. Use
// EquipmentService with class="cnc_machine" / "tooling_set" /
// "lab_instrument".
// MfgVerticalMaintenanceService retired — Phase F.2.4 consolidation. Use
// MaintenanceService with class="cnc_calibration" / "tooling_change" /
// "predictive_service" / "preventive_service".
// MfgVerticalVehicleService retired — Phase F.2.1 consolidation. Use
// VehicleService with class="forklift" / "material_transport_truck".
// Vertical-specific — Solar
// solarAssetService retired — Phase F.2.3 consolidation. Use
// AssetService with class="solar_farm".
// solarDepreciationService retired — Phase F.2.5 consolidation. Use
// DepreciationService with class="solar_accelerated" /
// "incometax_block" / "indas_book".
// solarEquipmentService retired — Phase F.2.2 consolidation. Use
// EquipmentService with class="solar_testing_rig".
// solarMaintenanceService retired — Phase F.2.4 consolidation. Use
// MaintenanceService with class="panel_cleaning" / "string_test" /
// "predictive_service".
// solarVehicleService retired — Phase F.2.1 consolidation. Use
// VehicleService with class="forklift" or "material_transport_truck"
// (solar verticals share these classes with manufacturing).
// Vertical-specific — Water
// WaterAssetService retired — Phase F.2.3 consolidation. Use
// AssetService with class="water_treatment_plant".
// WaterDepreciationService retired — Phase F.2.5 consolidation. Use
// DepreciationService with class="indas_book" / "incometax_block".
// WaterEquipmentService retired — Phase F.2.2 consolidation. Use
// EquipmentService with class="water_pump" (shared across water +
// agriculture domains per class_registry).
// WaterMaintenanceService retired — Phase F.2.4 consolidation. Use
// MaintenanceService with class="pump_ppm" / "effluent_test" /
// "statutory_inspection".
// WaterVehicleService retired — Phase F.2.1 consolidation. Use
// VehicleService with class="water_tanker".

export { AssetService, DepreciationService, EquipmentService, MaintenanceService, VehicleService };

export function getAssetService(): Client<typeof AssetService> {
  return getApiClient().getService(AssetService);
}

export function getDepreciationService(): Client<typeof DepreciationService> {
  return getApiClient().getService(DepreciationService);
}

export function getEquipmentService(): Client<typeof EquipmentService> {
  return getApiClient().getService(EquipmentService);
}

export function getMaintenanceService(): Client<typeof MaintenanceService> {
  return getApiClient().getService(MaintenanceService);
}

export function getVehicleService(): Client<typeof VehicleService> {
  return getApiClient().getService(VehicleService);
}


export {
  // Depreciation verticals retired — Phase F.2.5 consolidation. Use the single
  // DepreciationService above with `class` set per config/class_registry/depreciation.yaml.
  // Vehicle verticals retired — Phase F.2.1 consolidation. Use the single
  // VehicleService above with `class` set per config/class_registry/vehicle.yaml.
  // Equipment verticals retired — Phase F.2.2 consolidation. Use the single
  // EquipmentService above with `class` set per config/class_registry/equipment.yaml.
  // Asset verticals retired — Phase F.2.3 consolidation. Use the single
  // AssetService above with `class` set per config/class_registry/asset.yaml.
};

// ─── Agriculture Vertical Factories ───

// getAgricultureAssetService retired — Phase F.2.3 consolidation.
// Use getAssetService() with class="farmland" / "orchard".

// getAgricultureDepreciationService retired — Phase F.2.5 consolidation.
// Use getDepreciationService() with class="agri_bearer_plant" /
// "incometax_block" / "indas_book".

// getAgricultureEquipmentService retired — Phase F.2.2 consolidation.
// Use getEquipmentService() with class="agriculture_sprayer" /
// "grain_silo" / "cold_storage" / "water_pump".

// getAgricultureMaintenanceService retired — Phase F.2.4 consolidation.
// Use getMaintenanceService() with class="seasonal_service" /
// "irrigation_check" / "preventive_service".

// getAgricultureVehicleService retired — Phase F.2.1 consolidation.
// Use getVehicleService() with class="farm_tractor" / "farm_harvester"
// / "cold_chain_reefer".

// ─── Construction Vertical Factories ───

// getConstructionAssetService retired — Phase F.2.3 consolidation.
// Use getAssetService() with class="office_building" / "warehouse" /
// "construction_site".

// getConstructionDepreciationService retired — Phase F.2.5 consolidation.
// Use getDepreciationService() with class="realestate_structure" /
// "indas_book" / "companies_act_book".

// getConstructionEquipmentService retired — Phase F.2.2 consolidation.
// Use getEquipmentService() with class="construction_equipment".

// getConstructionMaintenanceService retired — Phase F.2.4 consolidation.
// Use getMaintenanceService() with class="site_safety_walk" /
// "corrective_repair" / "statutory_inspection".

// getConstructionVehicleService retired — Phase F.2.1 consolidation.
// Use getVehicleService() with class="construction_dumper" or
// "construction_mixer".

// ─── MfgVertical (Manufacturing) Vertical Factories ───

// getMfgVerticalAssetService retired — Phase F.2.3 consolidation.
// Use getAssetService() with class="factory_land" / "factory_building".

// getMfgVerticalDepreciationService retired — Phase F.2.5 consolidation.
// Use getDepreciationService() with class="indas_book" /
// "incometax_block" / "ifrs_book" / "us_macrs_tax".

// getMfgVerticalEquipmentService retired — Phase F.2.2 consolidation.
// Use getEquipmentService() with class="cnc_machine" / "tooling_set" /
// "lab_instrument".

// getMfgVerticalMaintenanceService retired — Phase F.2.4 consolidation.
// Use getMaintenanceService() with class="cnc_calibration" /
// "tooling_change" / "predictive_service" / "preventive_service".

// getMfgVerticalVehicleService retired — Phase F.2.1 consolidation.
// Use getVehicleService() with class="forklift" or
// "material_transport_truck".

// ─── Solar Vertical Factories ───

// getSolarAssetService retired — Phase F.2.3 consolidation.
// Use getAssetService() with class="solar_farm".

// getSolarDepreciationService retired — Phase F.2.5 consolidation.
// Use getDepreciationService() with class="solar_accelerated" /
// "incometax_block" / "indas_book".

// getSolarEquipmentService retired — Phase F.2.2 consolidation.
// Use getEquipmentService() with class="solar_testing_rig".

// getSolarMaintenanceService retired — Phase F.2.4 consolidation.
// Use getMaintenanceService() with class="panel_cleaning" /
// "string_test" / "predictive_service".

// getSolarVehicleService retired — Phase F.2.1 consolidation.
// Use getVehicleService() with class="forklift" or
// "material_transport_truck" (shared with manufacturing).

// ─── Water Vertical Factories ───

// getWaterAssetService retired — Phase F.2.3 consolidation.
// Use getAssetService() with class="water_treatment_plant".

// getWaterDepreciationService retired — Phase F.2.5 consolidation.
// Use getDepreciationService() with class="indas_book" /
// "incometax_block".

// getWaterEquipmentService retired — Phase F.2.2 consolidation.
// Use getEquipmentService() with class="water_pump".

// getWaterMaintenanceService retired — Phase F.2.4 consolidation.
// Use getMaintenanceService() with class="pump_ppm" /
// "effluent_test" / "statutory_inspection".

// getWaterVehicleService retired — Phase F.2.1 consolidation.
// Use getVehicleService() with class="water_tanker".

