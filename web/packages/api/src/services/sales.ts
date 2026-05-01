/**
 * Sales Service Factories
 * Typed ConnectRPC clients for CRM, orders, invoices, pricing, territory, etc.
 */

import type { Client } from '@connectrpc/connect';
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
// Phase F.8.10 (2026-04-21): AgricultureTerritoryService retired — use
// TerritoryService + class="agri_mandi_catchment" via config/class_registry/territory.yaml.
// Vertical-specific — Construction
import { ConstructionCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/construction/commission_construction_pb.js';
import { ConstructionCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/construction/crm_construction_pb.js';
import { ConstructionDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/construction/dealer_construction_pb.js';
import { ConstructionFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/construction/fieldsales_construction_pb.js';
import { ConstructionPricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/construction/pricing_construction_pb.js';
import { ConstructionRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/construction/routeplanning_construction_pb.js';
import { ConstructionSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/construction/salesanalytics_construction_pb.js';
import { ConstructionSalesInvoiceService } from '@samavāya/proto/gen/business/sales/salesinvoice/proto/construction/salesinvoice_construction_pb.js';
// Phase F.8.10 (2026-04-21): ConstructionTerritoryService retired — use
// TerritoryService + class="construction_project_zone" via config/class_registry/territory.yaml.
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
// Phase F.8.10 (2026-04-21): MfgVerticalTerritoryService retired — use
// TerritoryService + class="mfg_plant_service_area" via config/class_registry/territory.yaml.
// Vertical-specific — Solar
import { solarCommissionService } from '@samavāya/proto/gen/business/sales/commission/proto/solar/commission_solar_pb.js';
import { solarCRMService } from '@samavāya/proto/gen/business/sales/crm/proto/solar/crm_solar_pb.js';
import { solarDealerService } from '@samavāya/proto/gen/business/sales/dealer/proto/solar/dealer_solar_pb.js';
import { solarFieldSalesService } from '@samavāya/proto/gen/business/sales/fieldsales/proto/solar/fieldsales_solar_pb.js';
import { solarPricingService } from '@samavāya/proto/gen/business/sales/pricing/proto/solar/pricing_solar_pb.js';
import { solarRoutePlanningService } from '@samavāya/proto/gen/business/sales/routeplanning/proto/solar/routeplanning_solar_pb.js';
import { solarSalesAnalyticsService } from '@samavāya/proto/gen/business/sales/salesanalytics/proto/solar/salesanalytics_solar_pb.js';
import { solarSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/solar/salesorder_solar_pb.js';
// Phase F.8.10 (2026-04-21): solarTerritoryService retired — use
// TerritoryService + class="solar_discom_area" via config/class_registry/territory.yaml.
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
// Phase F.8.10 (2026-04-21): WaterTerritoryService retired — use
// TerritoryService + class="water_utility_zone" via config/class_registry/territory.yaml.
// Vertical-specific — Work Vertical
import { WorkVerticalSalesOrderService } from '@samavāya/proto/gen/business/sales/salesorder/proto/workvertical/salesorder_workvertical_pb.js';

export {
  CRMService, SalesOrderService, SalesInvoiceService, PricingService,
  TerritoryService, CommissionService, DealerService, FieldSalesService,
  RoutePlanningService, SalesAnalyticsService,
};

export function getCRMService(): Client<typeof CRMService> {
  return getApiClient().getService(CRMService);
}

export function getSalesOrderService(): Client<typeof SalesOrderService> {
  return getApiClient().getService(SalesOrderService);
}

export function getSalesInvoiceService(): Client<typeof SalesInvoiceService> {
  return getApiClient().getService(SalesInvoiceService);
}

export function getPricingService(): Client<typeof PricingService> {
  return getApiClient().getService(PricingService);
}

export function getTerritoryService(): Client<typeof TerritoryService> {
  return getApiClient().getService(TerritoryService);
}

export function getCommissionService(): Client<typeof CommissionService> {
  return getApiClient().getService(CommissionService);
}

export function getDealerService(): Client<typeof DealerService> {
  return getApiClient().getService(DealerService);
}

export function getFieldSalesService(): Client<typeof FieldSalesService> {
  return getApiClient().getService(FieldSalesService);
}

export function getRoutePlanningService(): Client<typeof RoutePlanningService> {
  return getApiClient().getService(RoutePlanningService);
}

export function getSalesAnalyticsService(): Client<typeof SalesAnalyticsService> {
  return getApiClient().getService(SalesAnalyticsService);
}


export {
  AgricultureCommissionService, ConstructionCommissionService, MfgVerticalCommissionService, solarCommissionService,
  WaterCommissionService, AgricultureCRMService, ConstructionCRMService, MfgVerticalCRMService,
  solarCRMService, WaterCRMService, AgricultureDealerService, ConstructionDealerService,
  MfgVerticalDealerService, solarDealerService, WaterDealerService, AgricultureFieldSalesService,
  ConstructionFieldSalesService, MfgVerticalFieldSalesService, solarFieldSalesService, WaterFieldSalesService,
  AgriculturePricingService, ConstructionPricingService, MfgVerticalPricingService, solarPricingService,
  WaterPricingService, AgricultureRoutePlanningService, ConstructionRoutePlanningService, MfgVerticalRoutePlanningService,
  solarRoutePlanningService, WaterRoutePlanningService, AgricultureSalesAnalyticsService, ConstructionSalesAnalyticsService,
  MfgVerticalSalesAnalyticsService, solarSalesAnalyticsService, WaterSalesAnalyticsService, ConstructionSalesInvoiceService,
  WaterSalesInvoiceService, AgricultureSalesOrderService, ConstructionVerticalSalesOrderService, MfgVerticalSalesOrderService,
  solarSalesOrderService, WaterSalesOrderService, WorkVerticalSalesOrderService,
  // Phase F.8.10 (2026-04-21): Agriculture/Construction/MfgVertical/solar/Water TerritoryService
  // retired — consolidated into TerritoryService with classregistry-driven verticals.
};

// ─── Agriculture Vertical Factories ───

export function getAgricultureCommissionService(): Client<typeof AgricultureCommissionService> {
  return getApiClient().getService(AgricultureCommissionService);
}

export function getAgricultureCRMService(): Client<typeof AgricultureCRMService> {
  return getApiClient().getService(AgricultureCRMService);
}

export function getAgricultureDealerService(): Client<typeof AgricultureDealerService> {
  return getApiClient().getService(AgricultureDealerService);
}

export function getAgricultureFieldSalesService(): Client<typeof AgricultureFieldSalesService> {
  return getApiClient().getService(AgricultureFieldSalesService);
}

export function getAgriculturePricingService(): Client<typeof AgriculturePricingService> {
  return getApiClient().getService(AgriculturePricingService);
}

export function getAgricultureRoutePlanningService(): Client<typeof AgricultureRoutePlanningService> {
  return getApiClient().getService(AgricultureRoutePlanningService);
}

export function getAgricultureSalesAnalyticsService(): Client<typeof AgricultureSalesAnalyticsService> {
  return getApiClient().getService(AgricultureSalesAnalyticsService);
}

export function getAgricultureSalesOrderService(): Client<typeof AgricultureSalesOrderService> {
  return getApiClient().getService(AgricultureSalesOrderService);
}

// Phase F.8.10 (2026-04-21): getAgricultureTerritoryService retired —
// call getTerritoryService() + class="agri_mandi_catchment" instead.

// ─── Construction Vertical Factories ───

export function getConstructionCommissionService(): Client<typeof ConstructionCommissionService> {
  return getApiClient().getService(ConstructionCommissionService);
}

export function getConstructionCRMService(): Client<typeof ConstructionCRMService> {
  return getApiClient().getService(ConstructionCRMService);
}

export function getConstructionDealerService(): Client<typeof ConstructionDealerService> {
  return getApiClient().getService(ConstructionDealerService);
}

export function getConstructionFieldSalesService(): Client<typeof ConstructionFieldSalesService> {
  return getApiClient().getService(ConstructionFieldSalesService);
}

export function getConstructionPricingService(): Client<typeof ConstructionPricingService> {
  return getApiClient().getService(ConstructionPricingService);
}

export function getConstructionRoutePlanningService(): Client<typeof ConstructionRoutePlanningService> {
  return getApiClient().getService(ConstructionRoutePlanningService);
}

export function getConstructionSalesAnalyticsService(): Client<typeof ConstructionSalesAnalyticsService> {
  return getApiClient().getService(ConstructionSalesAnalyticsService);
}

export function getConstructionSalesInvoiceService(): Client<typeof ConstructionSalesInvoiceService> {
  return getApiClient().getService(ConstructionSalesInvoiceService);
}

// Phase F.8.10 (2026-04-21): getConstructionTerritoryService retired —
// call getTerritoryService() + class="construction_project_zone" instead.

// ─── Construction Vertical Vertical Factories ───

export function getConstructionVerticalSalesOrderService(): Client<typeof ConstructionVerticalSalesOrderService> {
  return getApiClient().getService(ConstructionVerticalSalesOrderService);
}

// ─── MfgVertical (Manufacturing) Vertical Factories ───

export function getMfgVerticalCommissionService(): Client<typeof MfgVerticalCommissionService> {
  return getApiClient().getService(MfgVerticalCommissionService);
}

export function getMfgVerticalCRMService(): Client<typeof MfgVerticalCRMService> {
  return getApiClient().getService(MfgVerticalCRMService);
}

export function getMfgVerticalDealerService(): Client<typeof MfgVerticalDealerService> {
  return getApiClient().getService(MfgVerticalDealerService);
}

export function getMfgVerticalFieldSalesService(): Client<typeof MfgVerticalFieldSalesService> {
  return getApiClient().getService(MfgVerticalFieldSalesService);
}

export function getMfgVerticalPricingService(): Client<typeof MfgVerticalPricingService> {
  return getApiClient().getService(MfgVerticalPricingService);
}

export function getMfgVerticalRoutePlanningService(): Client<typeof MfgVerticalRoutePlanningService> {
  return getApiClient().getService(MfgVerticalRoutePlanningService);
}

export function getMfgVerticalSalesAnalyticsService(): Client<typeof MfgVerticalSalesAnalyticsService> {
  return getApiClient().getService(MfgVerticalSalesAnalyticsService);
}

export function getMfgVerticalSalesOrderService(): Client<typeof MfgVerticalSalesOrderService> {
  return getApiClient().getService(MfgVerticalSalesOrderService);
}

// Phase F.8.10 (2026-04-21): getMfgVerticalTerritoryService retired —
// call getTerritoryService() + class="mfg_plant_service_area" instead.

// ─── Solar Vertical Factories ───

export function getSolarCommissionService(): Client<typeof solarCommissionService> {
  return getApiClient().getService(solarCommissionService);
}

export function getSolarCRMService(): Client<typeof solarCRMService> {
  return getApiClient().getService(solarCRMService);
}

export function getSolarDealerService(): Client<typeof solarDealerService> {
  return getApiClient().getService(solarDealerService);
}

export function getSolarFieldSalesService(): Client<typeof solarFieldSalesService> {
  return getApiClient().getService(solarFieldSalesService);
}

export function getSolarPricingService(): Client<typeof solarPricingService> {
  return getApiClient().getService(solarPricingService);
}

export function getSolarRoutePlanningService(): Client<typeof solarRoutePlanningService> {
  return getApiClient().getService(solarRoutePlanningService);
}

export function getSolarSalesAnalyticsService(): Client<typeof solarSalesAnalyticsService> {
  return getApiClient().getService(solarSalesAnalyticsService);
}

export function getSolarSalesOrderService(): Client<typeof solarSalesOrderService> {
  return getApiClient().getService(solarSalesOrderService);
}

// Phase F.8.10 (2026-04-21): getSolarTerritoryService retired —
// call getTerritoryService() + class="solar_discom_area" instead.

// ─── Water Vertical Factories ───

export function getWaterCommissionService(): Client<typeof WaterCommissionService> {
  return getApiClient().getService(WaterCommissionService);
}

export function getWaterCRMService(): Client<typeof WaterCRMService> {
  return getApiClient().getService(WaterCRMService);
}

export function getWaterDealerService(): Client<typeof WaterDealerService> {
  return getApiClient().getService(WaterDealerService);
}

export function getWaterFieldSalesService(): Client<typeof WaterFieldSalesService> {
  return getApiClient().getService(WaterFieldSalesService);
}

export function getWaterPricingService(): Client<typeof WaterPricingService> {
  return getApiClient().getService(WaterPricingService);
}

export function getWaterRoutePlanningService(): Client<typeof WaterRoutePlanningService> {
  return getApiClient().getService(WaterRoutePlanningService);
}

export function getWaterSalesAnalyticsService(): Client<typeof WaterSalesAnalyticsService> {
  return getApiClient().getService(WaterSalesAnalyticsService);
}

export function getWaterSalesInvoiceService(): Client<typeof WaterSalesInvoiceService> {
  return getApiClient().getService(WaterSalesInvoiceService);
}

export function getWaterSalesOrderService(): Client<typeof WaterSalesOrderService> {
  return getApiClient().getService(WaterSalesOrderService);
}

// Phase F.8.10 (2026-04-21): getWaterTerritoryService retired —
// call getTerritoryService() + class="water_utility_zone" instead.

// ─── Work Vertical Vertical Factories ───

export function getWorkVerticalSalesOrderService(): Client<typeof WorkVerticalSalesOrderService> {
  return getApiClient().getService(WorkVerticalSalesOrderService);
}

