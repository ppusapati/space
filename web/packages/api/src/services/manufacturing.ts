/**
 * Manufacturing Service Factories
 * Typed ConnectRPC clients for BOM, routing, production, job cards, etc.
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { BOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/bom_pb.js';
import { RoutingService } from '@samavāya/proto/gen/business/manufacturing/routing/proto/routing_pb.js';
import { ProductionOrderService } from '@samavāya/proto/gen/business/manufacturing/productionorder/proto/productionorder_pb.js';
import { JobCardService } from '@samavāya/proto/gen/business/manufacturing/jobcard/proto/jobcard_pb.js';
import { ShopFloorService } from '@samavāya/proto/gen/business/manufacturing/shopfloor/proto/shopfloor_pb.js';
import { WorkCenterService } from '@samavāya/proto/gen/business/manufacturing/workcenter/proto/workcenter_pb.js';
import { SubcontractingService } from '@samavāya/proto/gen/business/manufacturing/subcontracting/proto/subcontracting_pb.js';
import { MfgQualityService } from '@samavāya/proto/gen/business/manufacturing/mfgquality/proto/mfgquality_pb.js';
import { PlanningService as MfgPlanningService } from '@samavāya/proto/gen/business/manufacturing/planning/proto/planning_pb.js';

// Vertical-specific — Agriculture
// AgricultureBOMService retired — Phase F.3.1 consolidation. Use
// BOMService with class="crop_input_kit" / "food_recipe" per
// config/class_registry/bom.yaml.
// AgricultureJobCardService retired — Phase F.3.4 consolidation. Use
// JobCardService with class="agri_harvest_task" per
// config/class_registry/jobcard.yaml.
// AgricultureMfgQualityService retired — Phase F.3.5 consolidation. Use
// MfgQualityService with class="agri_produce_grading" / "haccp_food" per
// config/class_registry/mfgquality.yaml.
// AgriculturePlanningService retired — Phase F.3.6 consolidation. Use
// PlanningService with class="agri_seasonal_plan" per
// config/class_registry/planning.yaml.
// AgricultureProductionOrderService retired — Phase F.3.3 consolidation.
// Use ProductionOrderService with class="agri_batch_processing" per
// config/class_registry/productionorder.yaml.
// AgricultureRoutingService retired — Phase F.3.2 consolidation. Use
// RoutingService with class="agri_crop_processing" per
// config/class_registry/routing.yaml.
// AgricultureShopFloorService retired — Phase F.3.7 consolidation. Use
// ShopFloorService with class="agri_processing_station" per
// config/class_registry/shopfloor.yaml.
// AgricultureSubcontractingService retired — Phase F.3.8 consolidation.
// Use SubcontractingService with class="agri_contract_farming" per
// config/class_registry/subcontracting.yaml.
// AgricultureWorkCenterService retired — Phase F.1 consolidation. Use
// WorkCenterService with class="agriculture_processing" and the
// attributes defined in config/class_registry/workcenter.yaml.
// Vertical-specific — Construction
// ConstructionJobCardService retired — Phase F.3.4 consolidation. Use
// JobCardService with class="construction_daily_progress".
// ConstructionMfgPlanningService retired — Phase F.3.6 consolidation.
// Use PlanningService with class="construction_schedule" /
// "project_planning".
// ConstructionProductionOrderService retired — Phase F.3.3 consolidation.
// Use ProductionOrderService with class="construction_work_package"
// / "engineer_to_order".
// ConstructionRoutingService retired — Phase F.3.2 consolidation. Use
// RoutingService with class="construction_work_package" /
// "project_based".
// ConstructionShopFloorService retired — Phase F.3.7 consolidation. Use
// ShopFloorService with class="construction_site_station".
// ConstructionSubcontractingService retired — Phase F.3.8 consolidation.
// Use SubcontractingService with class="epc_contract" /
// "labor_contracting".
// Vertical-specific — Construction Vertical
// ConstructionVerticalBOMService retired — Phase F.3.1 consolidation.
// Use BOMService with class="construction_boq".
// Vertical-specific — MfgVertical (Manufacturing)
// MfgVerticalBOMService retired — Phase F.3.1 consolidation. Use
// BOMService with class="discrete_assembly" / "electronics_pcb" /
// "process_batch" / "pharma_batch_record" / "textile_cut_sew".
// MfgVerticalJobCardService retired — Phase F.3.4 consolidation. Use
// JobCardService with class="production_operation" / "changeover_setup"
// / "quality_inspection" / "rework_operation" / "first_article" /
// "training".
// MfgVerticalMfgQualityService retired — Phase F.3.5 consolidation. Use
// MfgQualityService with class="iso9001_general" / "iatf16949_automotive"
// / "as9100_aerospace" / "iso13485_medical" / "gmp_pharma_batch".
// MfgVerticalPlanningService retired — Phase F.3.6 consolidation. Use
// PlanningService with class="mps_discrete" / "mrp_standard" /
// "aps_finite_capacity" / "kanban_pull_signal" / "s_op" / "capacity_rcp".
// MfgVerticalProductionOrderService retired — Phase F.3.3 consolidation.
// Use ProductionOrderService with class="make_to_stock" / "make_to_order"
// / "engineer_to_order" / "rework_order" / "prototype_order" /
// "process_batch" / "kanban_replenishment".
// MfgVerticalRoutingService retired — Phase F.3.2 consolidation. Use
// RoutingService with class="discrete_assembly_line" /
// "cellular_manufacturing" / "process_flow" / "batch_process".
// MfgVerticalShopFloorService retired — Phase F.3.7 consolidation. Use
// ShopFloorService with class="manual_assembly" / "semi_automated_cell" /
// "fully_automated_line" / "cnc_machining" / "injection_molding" /
// "pharma_cleanroom" / "food_processing_line" / "textile_sewing_line".
// MfgVerticalSubcontractingService retired — Phase F.3.8 consolidation.
// Use SubcontractingService with class="job_work" / "toll_manufacturing"
// / "cmt_textile" / "assembly_outsourcing".
// MfgVerticalWorkCenterService retired — Phase F.1 consolidation.
// Use WorkCenterService with class="mfg_discrete" or "mfg_batch".
// Vertical-specific — Solar
// solarBOMService retired — Phase F.3.1 consolidation. Use BOMService
// with class="solar_module_assembly".
// solarJobCardService retired — Phase F.3.4 consolidation. Use
// JobCardService with class="solar_module_step".
// solarMfgQualityService retired — Phase F.3.5 consolidation. Use
// MfgQualityService with class="iec61215_solar".
// solarPlanningService retired — Phase F.3.6 consolidation. Use
// PlanningService with class="solar_generation_plan".
// solarProductionOrderService retired — Phase F.3.3 consolidation.
// Use ProductionOrderService with class="solar_module_build".
// solarRoutingService retired — Phase F.3.2 consolidation. Use
// RoutingService with class="solar_module_process".
// solarShopFloorService retired — Phase F.3.7 consolidation. Use
// ShopFloorService with class="solar_module_line".
// solarSubcontractingService retired — Phase F.3.8 consolidation.
// Use SubcontractingService with class="solar_epc".
// solarWorkCenterService retired — Phase F.1 consolidation.
// Use WorkCenterService with class="solar_mfg_line".
// Vertical-specific — Water
// WaterBOMService retired — Phase F.3.1 consolidation. Use BOMService
// with class="water_treatment_assembly".
// WaterJobCardService retired — Phase F.3.4 consolidation. Use
// JobCardService with class="water_operation_run".
// WaterMfgPlanningService retired — Phase F.3.6 consolidation. Use
// PlanningService with class="water_distribution_plan".
// WaterProductionOrderService retired — Phase F.3.3 consolidation.
// Use ProductionOrderService with class="water_treatment_run".
// WaterRoutingService retired — Phase F.3.2 consolidation. Use
// RoutingService with class="water_treatment_process".
// WaterShopFloorService retired — Phase F.3.7 consolidation. Use
// ShopFloorService with class="water_treatment_station".
// WaterSubcontractingService retired — Phase F.3.8 consolidation.
// Use SubcontractingService with class="water_om_contract".
// WaterWorkCenterService retired — Phase F.1 (2026-04-20) consolidated
// the five workcenter verticals into the single WorkCenterService in
// manufacturing/workcenter/proto/workcenter.proto. Water treatment is
// now a class (`water_treatment`) defined in
// config/class_registry/workcenter.yaml; the frontend renders it via
// ListWorkCenterClasses + GetWorkCenterClassSchema.
// Vertical-specific — Work Vertical
// WorkVerticalBOMService retired — Phase F.3.1 consolidation. Use
// BOMService with class="construction_boq" (work vertical is covered
// by construction_boq in the BOM registry).

export {
  BOMService, RoutingService, ProductionOrderService, JobCardService,
  ShopFloorService, WorkCenterService, SubcontractingService,
  MfgQualityService, MfgPlanningService,
};

export function getBOMService(): Client<typeof BOMService> {
  return getApiClient().getService(BOMService);
}

export function getRoutingService(): Client<typeof RoutingService> {
  return getApiClient().getService(RoutingService);
}

export function getProductionOrderService(): Client<typeof ProductionOrderService> {
  return getApiClient().getService(ProductionOrderService);
}

export function getJobCardService(): Client<typeof JobCardService> {
  return getApiClient().getService(JobCardService);
}

export function getShopFloorService(): Client<typeof ShopFloorService> {
  return getApiClient().getService(ShopFloorService);
}

export function getWorkCenterService(): Client<typeof WorkCenterService> {
  return getApiClient().getService(WorkCenterService);
}

export function getSubcontractingService(): Client<typeof SubcontractingService> {
  return getApiClient().getService(SubcontractingService);
}

export function getMfgQualityService(): Client<typeof MfgQualityService> {
  return getApiClient().getService(MfgQualityService);
}

export function getMfgPlanningService(): Client<typeof MfgPlanningService> {
  return getApiClient().getService(MfgPlanningService);
}


export {
  // BOM verticals retired — Phase F.3.1 consolidation. Use the single
  // BOMService above with `class` set per config/class_registry/bom.yaml.
  // JobCard verticals retired — Phase F.3.4 consolidation. Use the single
  // JobCardService above with `class` set per config/class_registry/jobcard.yaml.
  // MfgQuality verticals retired — Phase F.3.5 consolidation. Use the single
  // MfgQualityService above with `class` set per config/class_registry/mfgquality.yaml.
  // Planning verticals retired — Phase F.3.6 consolidation. Use the single
  // PlanningService above with `class` set per config/class_registry/planning.yaml.
  // ProductionOrder verticals retired — Phase F.3.3 consolidation. Use the single
  // ProductionOrderService above with `class` set per config/class_registry/productionorder.yaml.
  // Routing verticals retired — Phase F.3.2 consolidation. Use the single
  // RoutingService above with `class` set per config/class_registry/routing.yaml.
  // ShopFloor verticals retired — Phase F.3.7 consolidation. Use the single
  // ShopFloorService above with `class` set per config/class_registry/shopfloor.yaml.
  // Subcontracting verticals retired — Phase F.3.8 consolidation. Use the single
  // SubcontractingService above with `class` set per config/class_registry/subcontracting.yaml.
  // WorkCenter verticals retired — Phase F.1 consolidation. Use the
  // single WorkCenterService above with `class` set per
  // config/class_registry/workcenter.yaml.
};

// ─── Agriculture Vertical Factories ───

// getAgricultureBOMService retired — Phase F.3.1 consolidation.
// Use getBOMService() with class="crop_input_kit" / "food_recipe".

// getAgricultureJobCardService retired — Phase F.3.4 consolidation.
// Use getJobCardService() with class="agri_harvest_task".

// getAgricultureMfgQualityService retired — Phase F.3.5 consolidation.
// Use getMfgQualityService() with class="agri_produce_grading" /
// "haccp_food".

// getAgricultureMfgPlanningService retired — Phase F.3.6 consolidation.
// Use getMfgPlanningService() with class="agri_seasonal_plan".

// getAgricultureProductionOrderService retired — Phase F.3.3 consolidation.
// Use getProductionOrderService() with class="agri_batch_processing".

// getAgricultureRoutingService retired — Phase F.3.2 consolidation.
// Use getRoutingService() with class="agri_crop_processing".

// getAgricultureShopFloorService retired — Phase F.3.7 consolidation.
// Use getShopFloorService() with class="agri_processing_station".

// getAgricultureSubcontractingService retired — Phase F.3.8 consolidation.
// Use getSubcontractingService() with class="agri_contract_farming".

// getAgricultureWorkCenterService retired — Phase F.1 consolidation.
// Use getWorkCenterService() with class="agriculture_processing".

// ─── Construction Vertical Factories ───

// getConstructionJobCardService retired — Phase F.3.4 consolidation.
// Use getJobCardService() with class="construction_daily_progress".

// getConstructionMfgPlanningService retired — Phase F.3.6 consolidation.
// Use getMfgPlanningService() with class="construction_schedule" /
// "project_planning".

// getConstructionProductionOrderService retired — Phase F.3.3 consolidation.
// Use getProductionOrderService() with class="construction_work_package"
// / "engineer_to_order".

// getConstructionRoutingService retired — Phase F.3.2 consolidation.
// Use getRoutingService() with class="construction_work_package" /
// "project_based".

// getConstructionShopFloorService retired — Phase F.3.7 consolidation.
// Use getShopFloorService() with class="construction_site_station".

// getConstructionSubcontractingService retired — Phase F.3.8 consolidation.
// Use getSubcontractingService() with class="epc_contract" /
// "labor_contracting".

// ─── Construction Vertical Vertical Factories ───

// getConstructionVerticalBOMService retired — Phase F.3.1 consolidation.
// Use getBOMService() with class="construction_boq".

// ─── MfgVertical (Manufacturing) Vertical Factories ───

// getMfgVerticalBOMService retired — Phase F.3.1 consolidation.
// Use getBOMService() with class="discrete_assembly" / "electronics_pcb"
// / "process_batch" / "pharma_batch_record" / "textile_cut_sew".

// getMfgVerticalJobCardService retired — Phase F.3.4 consolidation.
// Use getJobCardService() with class="production_operation" /
// "changeover_setup" / "quality_inspection" / "rework_operation" /
// "first_article" / "training".

// getMfgVerticalMfgQualityService retired — Phase F.3.5 consolidation.
// Use getMfgQualityService() with class="iso9001_general" /
// "iatf16949_automotive" / "as9100_aerospace" / "iso13485_medical" /
// "gmp_pharma_batch".

// getMfgVerticalMfgPlanningService retired — Phase F.3.6 consolidation.
// Use getMfgPlanningService() with class="mps_discrete" / "mrp_standard"
// / "aps_finite_capacity" / "kanban_pull_signal" / "s_op" / "capacity_rcp".

// getMfgVerticalProductionOrderService retired — Phase F.3.3 consolidation.
// Use getProductionOrderService() with class="make_to_stock" /
// "make_to_order" / "engineer_to_order" / "rework_order" /
// "prototype_order" / "process_batch" / "kanban_replenishment".

// getMfgVerticalRoutingService retired — Phase F.3.2 consolidation.
// Use getRoutingService() with class="discrete_assembly_line" /
// "cellular_manufacturing" / "process_flow" / "batch_process".

// getMfgVerticalShopFloorService retired — Phase F.3.7 consolidation.
// Use getShopFloorService() with class="manual_assembly" /
// "semi_automated_cell" / "fully_automated_line" / "cnc_machining" /
// "injection_molding" / "pharma_cleanroom" / "food_processing_line" /
// "textile_sewing_line".

// getMfgVerticalSubcontractingService retired — Phase F.3.8 consolidation.
// Use getSubcontractingService() with class="job_work" /
// "toll_manufacturing" / "cmt_textile" / "assembly_outsourcing".

// getMfgVerticalWorkCenterService retired — Phase F.1 consolidation.
// Use getWorkCenterService() with class="mfg_discrete" or "mfg_batch".

// ─── Solar Vertical Factories ───

// getSolarBOMService retired — Phase F.3.1 consolidation.
// Use getBOMService() with class="solar_module_assembly".

// getSolarJobCardService retired — Phase F.3.4 consolidation.
// Use getJobCardService() with class="solar_module_step".

// getSolarMfgQualityService retired — Phase F.3.5 consolidation.
// Use getMfgQualityService() with class="iec61215_solar".

// getSolarMfgPlanningService retired — Phase F.3.6 consolidation.
// Use getMfgPlanningService() with class="solar_generation_plan".

// getSolarProductionOrderService retired — Phase F.3.3 consolidation.
// Use getProductionOrderService() with class="solar_module_build".

// getSolarRoutingService retired — Phase F.3.2 consolidation.
// Use getRoutingService() with class="solar_module_process".

// getSolarShopFloorService retired — Phase F.3.7 consolidation.
// Use getShopFloorService() with class="solar_module_line".

// getSolarSubcontractingService retired — Phase F.3.8 consolidation.
// Use getSubcontractingService() with class="solar_epc".

// getSolarWorkCenterService retired — Phase F.1 consolidation.
// Use getWorkCenterService() with class="solar_mfg_line".

// ─── Water Vertical Factories ───

// getWaterBOMService retired — Phase F.3.1 consolidation.
// Use getBOMService() with class="water_treatment_assembly".

// getWaterJobCardService retired — Phase F.3.4 consolidation.
// Use getJobCardService() with class="water_operation_run".

// getWaterMfgPlanningService retired — Phase F.3.6 consolidation.
// Use getMfgPlanningService() with class="water_distribution_plan".

// getWaterProductionOrderService retired — Phase F.3.3 consolidation.
// Use getProductionOrderService() with class="water_treatment_run".

// getWaterRoutingService retired — Phase F.3.2 consolidation.
// Use getRoutingService() with class="water_treatment_process".

// getWaterShopFloorService retired — Phase F.3.7 consolidation.
// Use getShopFloorService() with class="water_treatment_station".

// getWaterSubcontractingService retired — Phase F.3.8 consolidation.
// Use getSubcontractingService() with class="water_om_contract".

// getWaterWorkCenterService retired — Phase F.1 consolidation. Use
// getWorkCenterService() with class="water_treatment" and the
// attributes declared in config/class_registry/workcenter.yaml.

// ─── Work Vertical Vertical Factories ───

// getWorkVerticalBOMService retired — Phase F.3.1 consolidation.
// Use getBOMService() with class="construction_boq".

