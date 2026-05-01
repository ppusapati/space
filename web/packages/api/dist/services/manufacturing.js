/**
 * Manufacturing Service Factories
 * Typed ConnectRPC clients for BOM, routing, production, job cards, etc.
 */
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
import { AgricultureBOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/agriculture/bom_agri_pb.js';
import { AgricultureJobCardService } from '@samavāya/proto/gen/business/manufacturing/jobcard/proto/agriculture/jobcard_agri_pb.js';
import { AgricultureMfgQualityService } from '@samavāya/proto/gen/business/manufacturing/mfgquality/proto/agriculture/mfgquality_agri_pb.js';
import { AgriculturePlanningService as AgricultureMfgPlanningService } from '@samavāya/proto/gen/business/manufacturing/planning/proto/agriculture/planning_agri_pb.js';
import { AgricultureProductionOrderService } from '@samavāya/proto/gen/business/manufacturing/productionorder/proto/agriculture/productionorder_agri_pb.js';
import { AgricultureRoutingService } from '@samavāya/proto/gen/business/manufacturing/routing/proto/agriculture/routing_agri_pb.js';
import { AgricultureShopFloorService } from '@samavāya/proto/gen/business/manufacturing/shopfloor/proto/agriculture/shopfloor_agri_pb.js';
import { AgricultureSubcontractingService } from '@samavāya/proto/gen/business/manufacturing/subcontracting/proto/agriculture/subcontracting_agri_pb.js';
import { AgricultureWorkCenterService } from '@samavāya/proto/gen/business/manufacturing/workcenter/proto/agriculture/workcenter_agri_pb.js';
// Vertical-specific — Construction
import { ConstructionJobCardService } from '@samavāya/proto/gen/business/manufacturing/jobcard/proto/construction/jobcard_construction_pb.js';
import { ConstructionMfgPlanningService } from '@samavāya/proto/gen/business/manufacturing/planning/proto/construction/planning_construction_pb.js';
import { ConstructionProductionOrderService } from '@samavāya/proto/gen/business/manufacturing/productionorder/proto/construction/productionorder_construction_pb.js';
import { ConstructionRoutingService } from '@samavāya/proto/gen/business/manufacturing/routing/proto/construction/routing_construction_pb.js';
import { ConstructionShopFloorService } from '@samavāya/proto/gen/business/manufacturing/shopfloor/proto/construction/shopfloor_construction_pb.js';
import { ConstructionSubcontractingService } from '@samavāya/proto/gen/business/manufacturing/subcontracting/proto/construction/subcontracting_construction_pb.js';
// Vertical-specific — Construction Vertical
import { ConstructionVerticalBOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/constructionvertical/bom_constructionvertical_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalBOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/mfgvertical/bom_mfgvertical_pb.js';
import { MfgVerticalJobCardService } from '@samavāya/proto/gen/business/manufacturing/jobcard/proto/mfgvertical/jobcard_mfgvertical_pb.js';
import { MfgVerticalMfgQualityService } from '@samavāya/proto/gen/business/manufacturing/mfgquality/proto/mfgvertical/mfgquality_mfgvertical_pb.js';
import { MfgVerticalPlanningService as MfgVerticalMfgPlanningService } from '@samavāya/proto/gen/business/manufacturing/planning/proto/mfgvertical/planning_mfgvertical_pb.js';
import { MfgVerticalProductionOrderService } from '@samavāya/proto/gen/business/manufacturing/productionorder/proto/mfgvertical/productionorder_mfgvertical_pb.js';
import { MfgVerticalRoutingService } from '@samavāya/proto/gen/business/manufacturing/routing/proto/mfgvertical/routing_mfgvertical_pb.js';
import { MfgVerticalShopFloorService } from '@samavāya/proto/gen/business/manufacturing/shopfloor/proto/mfgvertical/shopfloor_mfgvertical_pb.js';
import { MfgVerticalSubcontractingService } from '@samavāya/proto/gen/business/manufacturing/subcontracting/proto/mfgvertical/subcontracting_mfgvertical_pb.js';
import { MfgVerticalWorkCenterService } from '@samavāya/proto/gen/business/manufacturing/workcenter/proto/mfgvertical/workcenter_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarBOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/solar/bom_solar_pb.js';
import { solarJobCardService } from '@samavāya/proto/gen/business/manufacturing/jobcard/proto/solar/jobcard_solar_pb.js';
import { solarMfgQualityService } from '@samavāya/proto/gen/business/manufacturing/mfgquality/proto/solar/mfgquality_solar_pb.js';
import { solarPlanningService as solarMfgPlanningService } from '@samavāya/proto/gen/business/manufacturing/planning/proto/solar/planning_solar_pb.js';
import { solarProductionOrderService } from '@samavāya/proto/gen/business/manufacturing/productionorder/proto/solar/productionorder_solar_pb.js';
import { solarRoutingService } from '@samavāya/proto/gen/business/manufacturing/routing/proto/solar/routing_solar_pb.js';
import { solarShopFloorService } from '@samavāya/proto/gen/business/manufacturing/shopfloor/proto/solar/shopfloor_solar_pb.js';
import { solarSubcontractingService } from '@samavāya/proto/gen/business/manufacturing/subcontracting/proto/solar/subcontracting_solar_pb.js';
import { solarWorkCenterService } from '@samavāya/proto/gen/business/manufacturing/workcenter/proto/solar/workcenter_solar_pb.js';
// Vertical-specific — Water
import { WaterBOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/water/bom_water_pb.js';
import { WaterJobCardService } from '@samavāya/proto/gen/business/manufacturing/jobcard/proto/water/jobcard_water_pb.js';
import { WaterMfgPlanningService } from '@samavāya/proto/gen/business/manufacturing/planning/proto/water/planning_water_pb.js';
import { WaterProductionOrderService } from '@samavāya/proto/gen/business/manufacturing/productionorder/proto/water/productionorder_water_pb.js';
import { WaterRoutingService } from '@samavāya/proto/gen/business/manufacturing/routing/proto/water/routing_water_pb.js';
import { WaterShopFloorService } from '@samavāya/proto/gen/business/manufacturing/shopfloor/proto/water/shopfloor_water_pb.js';
import { WaterSubcontractingService } from '@samavāya/proto/gen/business/manufacturing/subcontracting/proto/water/subcontracting_water_pb.js';
import { WaterWorkCenterService } from '@samavāya/proto/gen/business/manufacturing/workcenter/proto/water/workcenter_water_pb.js';
// Vertical-specific — Work Vertical
import { WorkVerticalBOMService } from '@samavāya/proto/gen/business/manufacturing/bom/proto/workvertical/bom_workvertical_pb.js';
export { BOMService, RoutingService, ProductionOrderService, JobCardService, ShopFloorService, WorkCenterService, SubcontractingService, MfgQualityService, MfgPlanningService, };
export function getBOMService() {
    return getApiClient().getService(BOMService);
}
export function getRoutingService() {
    return getApiClient().getService(RoutingService);
}
export function getProductionOrderService() {
    return getApiClient().getService(ProductionOrderService);
}
export function getJobCardService() {
    return getApiClient().getService(JobCardService);
}
export function getShopFloorService() {
    return getApiClient().getService(ShopFloorService);
}
export function getWorkCenterService() {
    return getApiClient().getService(WorkCenterService);
}
export function getSubcontractingService() {
    return getApiClient().getService(SubcontractingService);
}
export function getMfgQualityService() {
    return getApiClient().getService(MfgQualityService);
}
export function getMfgPlanningService() {
    return getApiClient().getService(MfgPlanningService);
}
export { AgricultureBOMService, ConstructionVerticalBOMService, MfgVerticalBOMService, solarBOMService, WaterBOMService, WorkVerticalBOMService, AgricultureJobCardService, ConstructionJobCardService, MfgVerticalJobCardService, solarJobCardService, WaterJobCardService, AgricultureMfgQualityService, MfgVerticalMfgQualityService, solarMfgQualityService, AgricultureMfgPlanningService, ConstructionMfgPlanningService, MfgVerticalMfgPlanningService, solarMfgPlanningService, WaterMfgPlanningService, AgricultureProductionOrderService, ConstructionProductionOrderService, MfgVerticalProductionOrderService, solarProductionOrderService, WaterProductionOrderService, AgricultureRoutingService, ConstructionRoutingService, MfgVerticalRoutingService, solarRoutingService, WaterRoutingService, AgricultureShopFloorService, ConstructionShopFloorService, MfgVerticalShopFloorService, solarShopFloorService, WaterShopFloorService, AgricultureSubcontractingService, ConstructionSubcontractingService, MfgVerticalSubcontractingService, solarSubcontractingService, WaterSubcontractingService, AgricultureWorkCenterService, MfgVerticalWorkCenterService, solarWorkCenterService, WaterWorkCenterService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureBOMService() {
    return getApiClient().getService(AgricultureBOMService);
}
export function getAgricultureJobCardService() {
    return getApiClient().getService(AgricultureJobCardService);
}
export function getAgricultureMfgQualityService() {
    return getApiClient().getService(AgricultureMfgQualityService);
}
export function getAgricultureMfgPlanningService() {
    return getApiClient().getService(AgricultureMfgPlanningService);
}
export function getAgricultureProductionOrderService() {
    return getApiClient().getService(AgricultureProductionOrderService);
}
export function getAgricultureRoutingService() {
    return getApiClient().getService(AgricultureRoutingService);
}
export function getAgricultureShopFloorService() {
    return getApiClient().getService(AgricultureShopFloorService);
}
export function getAgricultureSubcontractingService() {
    return getApiClient().getService(AgricultureSubcontractingService);
}
export function getAgricultureWorkCenterService() {
    return getApiClient().getService(AgricultureWorkCenterService);
}
// ─── Construction Vertical Factories ───
export function getConstructionJobCardService() {
    return getApiClient().getService(ConstructionJobCardService);
}
export function getConstructionMfgPlanningService() {
    return getApiClient().getService(ConstructionMfgPlanningService);
}
export function getConstructionProductionOrderService() {
    return getApiClient().getService(ConstructionProductionOrderService);
}
export function getConstructionRoutingService() {
    return getApiClient().getService(ConstructionRoutingService);
}
export function getConstructionShopFloorService() {
    return getApiClient().getService(ConstructionShopFloorService);
}
export function getConstructionSubcontractingService() {
    return getApiClient().getService(ConstructionSubcontractingService);
}
// ─── Construction Vertical Vertical Factories ───
export function getConstructionVerticalBOMService() {
    return getApiClient().getService(ConstructionVerticalBOMService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalBOMService() {
    return getApiClient().getService(MfgVerticalBOMService);
}
export function getMfgVerticalJobCardService() {
    return getApiClient().getService(MfgVerticalJobCardService);
}
export function getMfgVerticalMfgQualityService() {
    return getApiClient().getService(MfgVerticalMfgQualityService);
}
export function getMfgVerticalMfgPlanningService() {
    return getApiClient().getService(MfgVerticalMfgPlanningService);
}
export function getMfgVerticalProductionOrderService() {
    return getApiClient().getService(MfgVerticalProductionOrderService);
}
export function getMfgVerticalRoutingService() {
    return getApiClient().getService(MfgVerticalRoutingService);
}
export function getMfgVerticalShopFloorService() {
    return getApiClient().getService(MfgVerticalShopFloorService);
}
export function getMfgVerticalSubcontractingService() {
    return getApiClient().getService(MfgVerticalSubcontractingService);
}
export function getMfgVerticalWorkCenterService() {
    return getApiClient().getService(MfgVerticalWorkCenterService);
}
// ─── Solar Vertical Factories ───
export function getSolarBOMService() {
    return getApiClient().getService(solarBOMService);
}
export function getSolarJobCardService() {
    return getApiClient().getService(solarJobCardService);
}
export function getSolarMfgQualityService() {
    return getApiClient().getService(solarMfgQualityService);
}
export function getSolarMfgPlanningService() {
    return getApiClient().getService(solarMfgPlanningService);
}
export function getSolarProductionOrderService() {
    return getApiClient().getService(solarProductionOrderService);
}
export function getSolarRoutingService() {
    return getApiClient().getService(solarRoutingService);
}
export function getSolarShopFloorService() {
    return getApiClient().getService(solarShopFloorService);
}
export function getSolarSubcontractingService() {
    return getApiClient().getService(solarSubcontractingService);
}
export function getSolarWorkCenterService() {
    return getApiClient().getService(solarWorkCenterService);
}
// ─── Water Vertical Factories ───
export function getWaterBOMService() {
    return getApiClient().getService(WaterBOMService);
}
export function getWaterJobCardService() {
    return getApiClient().getService(WaterJobCardService);
}
export function getWaterMfgPlanningService() {
    return getApiClient().getService(WaterMfgPlanningService);
}
export function getWaterProductionOrderService() {
    return getApiClient().getService(WaterProductionOrderService);
}
export function getWaterRoutingService() {
    return getApiClient().getService(WaterRoutingService);
}
export function getWaterShopFloorService() {
    return getApiClient().getService(WaterShopFloorService);
}
export function getWaterSubcontractingService() {
    return getApiClient().getService(WaterSubcontractingService);
}
export function getWaterWorkCenterService() {
    return getApiClient().getService(WaterWorkCenterService);
}
// ─── Work Vertical Vertical Factories ───
export function getWorkVerticalBOMService() {
    return getApiClient().getService(WorkVerticalBOMService);
}
//# sourceMappingURL=manufacturing.js.map