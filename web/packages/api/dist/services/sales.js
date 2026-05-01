/**
 * Sales Service Factories
 * Typed ConnectRPC clients for CRM, orders, invoices, pricing, territory, etc.
 */
import { getApiClient } from '../client/client.js';
import { CRMService } from '@samavāya/proto/gen/business/sales/crm/proto/crm_pb.js';
import { SalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/salesorder_pb.js';
import { SalesInvoiceService } from '@samavāya/proto/gen/business/sales/salesinvoice/proto/salesinvoice_pb.js';
import { PricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/pricing_pb.js';
import { TerritoryService } from '@samavāya/proto/gen/business/sales/territory/proto/territory_pb.js';
import { CommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/commission_pb.js';
import { DealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/dealer_pb.js';
import { FieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/fieldsales_pb.js';
import { RoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/routeplanning_pb.js';
import { SalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/salesanalytics_pb.js';
// Vertical-specific — Agriculture
import { AgricultureCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/agriculture/commission_agriculture_pb.js';
import { AgricultureCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/agriculture/crm_agriculture_pb.js';
import { AgricultureDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/agriculture/dealer_agriculture_pb.js';
import { AgricultureFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/agriculture/fieldsales_agriculture_pb.js';
import { AgriculturePricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/agriculture/pricing_agriculture_pb.js';
import { AgricultureRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/agriculture/routeplanning_agriculture_pb.js';
import { AgricultureSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/agriculture/salesanalytics_agriculture_pb.js';
import { AgricultureSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/agriculture/salesorder_agriculture_pb.js';
import { AgricultureTerritoryService } from '@samavāya/proto/gen/business/sales/territory/proto/agriculture/territory_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/construction/commission_construction_pb.js';
import { ConstructionCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/construction/crm_construction_pb.js';
import { ConstructionDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/construction/dealer_construction_pb.js';
import { ConstructionFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/construction/fieldsales_construction_pb.js';
import { ConstructionPricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/construction/pricing_construction_pb.js';
import { ConstructionRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/construction/routeplanning_construction_pb.js';
import { ConstructionSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/construction/salesanalytics_construction_pb.js';
import { ConstructionSalesInvoiceService } from '@samavāya/proto/gen/business/sales/salesinvoice/proto/construction/salesinvoice_construction_pb.js';
import { ConstructionTerritoryService } from '@samavāya/proto/gen/business/sales/territory/proto/construction/territory_construction_pb.js';
// Vertical-specific — Construction Vertical
import { ConstructionVerticalSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/constructionvertical/salesorder_constructionvertical_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/mfgvertical/commission_mfgvertical_pb.js';
import { MfgVerticalCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/mfgvertical/crm_mfgvertical_pb.js';
import { MfgVerticalDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/mfgvertical/dealer_mfgvertical_pb.js';
import { MfgVerticalFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/mfgvertical/fieldsales_mfgvertical_pb.js';
import { MfgVerticalPricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/mfgvertical/pricing_mfgvertical_pb.js';
import { MfgVerticalRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/mfgvertical/routeplanning_mfgvertical_pb.js';
import { MfgVerticalSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/mfgvertical/salesanalytics_mfgvertical_pb.js';
import { MfgVerticalSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/mfgvertical/salesorder_mfgvertical_pb.js';
import { MfgVerticalTerritoryService } from '@samavāya/proto/gen/business/sales/territory/proto/mfgvertical/territory_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/solar/commission_solar_pb.js';
import { solarCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/solar/crm_solar_pb.js';
import { solarDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/solar/dealer_solar_pb.js';
import { solarFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/solar/fieldsales_solar_pb.js';
import { solarPricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/solar/pricing_solar_pb.js';
import { solarRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/solar/routeplanning_solar_pb.js';
import { solarSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/solar/salesanalytics_solar_pb.js';
import { solarSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/solar/salesorder_solar_pb.js';
import { solarTerritoryService } from '@samavāya/proto/gen/business/sales/territory/proto/solar/territory_solar_pb.js';
// Vertical-specific — Water
import { WaterCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/water/commission_water_pb.js';
import { WaterCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/water/crm_water_pb.js';
import { WaterDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/water/dealer_water_pb.js';
import { WaterFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/water/fieldsales_water_pb.js';
import { WaterPricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/water/pricing_water_pb.js';
import { WaterRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/water/routeplanning_water_pb.js';
import { WaterSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/water/salesanalytics_water_pb.js';
import { WaterSalesInvoiceService } from '@samavāya/proto/gen/business/sales/salesinvoice/proto/water/salesinvoice_water_pb.js';
import { WaterSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/water/salesorder_water_pb.js';
import { WaterTerritoryService } from '@samavāya/proto/gen/business/sales/territory/proto/water/territory_water_pb.js';
// Vertical-specific — Work Vertical
import { WorkVerticalSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/workvertical/salesorder_workvertical_pb.js';
export { CRMService, SalesOrderService, SalesInvoiceService, PricingService, TerritoryService, CommissionService, DealerService, FieldSalesService, RoutePlanningService, SalesAnalyticsService, };
export function getCRMService() {
    return getApiClient().getService(CRMService);
}
export function getSalesOrderService() {
    return getApiClient().getService(SalesOrderService);
}
export function getSalesInvoiceService() {
    return getApiClient().getService(SalesInvoiceService);
}
export function getPricingService() {
    return getApiClient().getService(PricingService);
}
export function getTerritoryService() {
    return getApiClient().getService(TerritoryService);
}
export function getCommissionService() {
    return getApiClient().getService(CommissionService);
}
export function getDealerService() {
    return getApiClient().getService(DealerService);
}
export function getFieldSalesService() {
    return getApiClient().getService(FieldSalesService);
}
export function getRoutePlanningService() {
    return getApiClient().getService(RoutePlanningService);
}
export function getSalesAnalyticsService() {
    return getApiClient().getService(SalesAnalyticsService);
}
export { AgricultureCommissionService, ConstructionCommissionService, MfgVerticalCommissionService, solarCommissionService, WaterCommissionService, AgricultureCRMService, ConstructionCRMService, MfgVerticalCRMService, solarCRMService, WaterCRMService, AgricultureDealerService, ConstructionDealerService, MfgVerticalDealerService, solarDealerService, WaterDealerService, AgricultureFieldSalesService, ConstructionFieldSalesService, MfgVerticalFieldSalesService, solarFieldSalesService, WaterFieldSalesService, AgriculturePricingService, ConstructionPricingService, MfgVerticalPricingService, solarPricingService, WaterPricingService, AgricultureRoutePlanningService, ConstructionRoutePlanningService, MfgVerticalRoutePlanningService, solarRoutePlanningService, WaterRoutePlanningService, AgricultureSalesAnalyticsService, ConstructionSalesAnalyticsService, MfgVerticalSalesAnalyticsService, solarSalesAnalyticsService, WaterSalesAnalyticsService, ConstructionSalesInvoiceService, WaterSalesInvoiceService, AgricultureSalesOrderService, ConstructionVerticalSalesOrderService, MfgVerticalSalesOrderService, solarSalesOrderService, WaterSalesOrderService, WorkVerticalSalesOrderService, AgricultureTerritoryService, ConstructionTerritoryService, MfgVerticalTerritoryService, solarTerritoryService, WaterTerritoryService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureCommissionService() {
    return getApiClient().getService(AgricultureCommissionService);
}
export function getAgricultureCRMService() {
    return getApiClient().getService(AgricultureCRMService);
}
export function getAgricultureDealerService() {
    return getApiClient().getService(AgricultureDealerService);
}
export function getAgricultureFieldSalesService() {
    return getApiClient().getService(AgricultureFieldSalesService);
}
export function getAgriculturePricingService() {
    return getApiClient().getService(AgriculturePricingService);
}
export function getAgricultureRoutePlanningService() {
    return getApiClient().getService(AgricultureRoutePlanningService);
}
export function getAgricultureSalesAnalyticsService() {
    return getApiClient().getService(AgricultureSalesAnalyticsService);
}
export function getAgricultureSalesOrderService() {
    return getApiClient().getService(AgricultureSalesOrderService);
}
export function getAgricultureTerritoryService() {
    return getApiClient().getService(AgricultureTerritoryService);
}
// ─── Construction Vertical Factories ───
export function getConstructionCommissionService() {
    return getApiClient().getService(ConstructionCommissionService);
}
export function getConstructionCRMService() {
    return getApiClient().getService(ConstructionCRMService);
}
export function getConstructionDealerService() {
    return getApiClient().getService(ConstructionDealerService);
}
export function getConstructionFieldSalesService() {
    return getApiClient().getService(ConstructionFieldSalesService);
}
export function getConstructionPricingService() {
    return getApiClient().getService(ConstructionPricingService);
}
export function getConstructionRoutePlanningService() {
    return getApiClient().getService(ConstructionRoutePlanningService);
}
export function getConstructionSalesAnalyticsService() {
    return getApiClient().getService(ConstructionSalesAnalyticsService);
}
export function getConstructionSalesInvoiceService() {
    return getApiClient().getService(ConstructionSalesInvoiceService);
}
export function getConstructionTerritoryService() {
    return getApiClient().getService(ConstructionTerritoryService);
}
// ─── Construction Vertical Vertical Factories ───
export function getConstructionVerticalSalesOrderService() {
    return getApiClient().getService(ConstructionVerticalSalesOrderService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalCommissionService() {
    return getApiClient().getService(MfgVerticalCommissionService);
}
export function getMfgVerticalCRMService() {
    return getApiClient().getService(MfgVerticalCRMService);
}
export function getMfgVerticalDealerService() {
    return getApiClient().getService(MfgVerticalDealerService);
}
export function getMfgVerticalFieldSalesService() {
    return getApiClient().getService(MfgVerticalFieldSalesService);
}
export function getMfgVerticalPricingService() {
    return getApiClient().getService(MfgVerticalPricingService);
}
export function getMfgVerticalRoutePlanningService() {
    return getApiClient().getService(MfgVerticalRoutePlanningService);
}
export function getMfgVerticalSalesAnalyticsService() {
    return getApiClient().getService(MfgVerticalSalesAnalyticsService);
}
export function getMfgVerticalSalesOrderService() {
    return getApiClient().getService(MfgVerticalSalesOrderService);
}
export function getMfgVerticalTerritoryService() {
    return getApiClient().getService(MfgVerticalTerritoryService);
}
// ─── Solar Vertical Factories ───
export function getSolarCommissionService() {
    return getApiClient().getService(solarCommissionService);
}
export function getSolarCRMService() {
    return getApiClient().getService(solarCRMService);
}
export function getSolarDealerService() {
    return getApiClient().getService(solarDealerService);
}
export function getSolarFieldSalesService() {
    return getApiClient().getService(solarFieldSalesService);
}
export function getSolarPricingService() {
    return getApiClient().getService(solarPricingService);
}
export function getSolarRoutePlanningService() {
    return getApiClient().getService(solarRoutePlanningService);
}
export function getSolarSalesAnalyticsService() {
    return getApiClient().getService(solarSalesAnalyticsService);
}
export function getSolarSalesOrderService() {
    return getApiClient().getService(solarSalesOrderService);
}
export function getSolarTerritoryService() {
    return getApiClient().getService(solarTerritoryService);
}
// ─── Water Vertical Factories ───
export function getWaterCommissionService() {
    return getApiClient().getService(WaterCommissionService);
}
export function getWaterCRMService() {
    return getApiClient().getService(WaterCRMService);
}
export function getWaterDealerService() {
    return getApiClient().getService(WaterDealerService);
}
export function getWaterFieldSalesService() {
    return getApiClient().getService(WaterFieldSalesService);
}
export function getWaterPricingService() {
    return getApiClient().getService(WaterPricingService);
}
export function getWaterRoutePlanningService() {
    return getApiClient().getService(WaterRoutePlanningService);
}
export function getWaterSalesAnalyticsService() {
    return getApiClient().getService(WaterSalesAnalyticsService);
}
export function getWaterSalesInvoiceService() {
    return getApiClient().getService(WaterSalesInvoiceService);
}
export function getWaterSalesOrderService() {
    return getApiClient().getService(WaterSalesOrderService);
}
export function getWaterTerritoryService() {
    return getApiClient().getService(WaterTerritoryService);
}
// ─── Work Vertical Vertical Factories ───
export function getWorkVerticalSalesOrderService() {
    return getApiClient().getService(WorkVerticalSalesOrderService);
}
//# sourceMappingURL=sales.js.map