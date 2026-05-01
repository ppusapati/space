/**
 * Masters Service Factories
 * Typed ConnectRPC clients for items, parties, locations, UOM, CoA, tax codes
 */
import { getApiClient } from '../client/client.js';
// Base service descriptors
import { ItemService } from '@samavāya/proto/gen/business/masters/item/proto/item_pb.js';
import { EntityService } from '@samavāya/proto/gen/business/masters/party/proto/party_pb.js';
import { LocationService } from '@samavāya/proto/gen/business/masters/location/proto/location_pb.js';
import { UOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/uom_pb.js';
import { ChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/chartofaccounts_pb.js';
import { TaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/taxcode_pb.js';
// Vertical-specific service descriptors — Agriculture
import { AgricultureItemService } from '@samavāya/proto/gen/business/masters/item/proto/agriculture/item_agriculture_pb.js';
import { AgriculturePartyService } from '@samavāya/proto/gen/business/masters/party/proto/agriculture/party_agriculture_pb.js';
import { AgricultureLocationService } from '@samavāya/proto/gen/business/masters/location/proto/agriculture/location_agriculture_pb.js';
import { AgricultureUOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/agriculture/uom_agriculture_pb.js';
import { AgricultureChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/agriculture/chartofaccounts_agriculture_pb.js';
import { AgricultureTaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/agriculture/taxcode_agriculture_pb.js';
// Vertical-specific service descriptors — MfgVertical
import { MfgVerticalItemService } from '@samavāya/proto/gen/business/masters/item/proto/mfgvertical/item_mfgvertical_pb.js';
import { MfgVerticalPartyService } from '@samavāya/proto/gen/business/masters/party/proto/mfgvertical/party_mfgvertical_pb.js';
import { MfgVerticalLocationService } from '@samavāya/proto/gen/business/masters/location/proto/mfgvertical/location_mfgvertical_pb.js';
import { MfgVerticalUOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/mfgvertical/uom_mfgvertical_pb.js';
import { MfgVerticalChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/mfgvertical/chartofaccounts_mfgvertical_pb.js';
import { MfgVerticalTaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/mfgvertical/taxcode_mfgvertical_pb.js';
// Vertical-specific service descriptors — Solar
import { solarItemService } from '@samavāya/proto/gen/business/masters/item/proto/solar/item_solar_pb.js';
import { solarPartyService } from '@samavāya/proto/gen/business/masters/party/proto/solar/party_solar_pb.js';
import { solarLocationService } from '@samavāya/proto/gen/business/masters/location/proto/solar/location_solar_pb.js';
import { solarUOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/solar/uom_solar_pb.js';
import { solarChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/solar/chartofaccounts_solar_pb.js';
import { solarTaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/solar/taxcode_solar_pb.js';
// Vertical-specific service descriptors — Construction
import { ConstructionMastersService } from '@samavāya/proto/gen/business/masters/masters/proto/construction/masters_construction_pb.js';
// Vertical-specific service descriptors — Water
import { WaterMastersService } from '@samavāya/proto/gen/business/masters/masters/proto/water/masters_water_pb.js';
// Re-export service descriptors
export { ItemService, EntityService, LocationService, UOMService, ChartOfAccountsService, TaxCodeService, AgricultureItemService, AgriculturePartyService, AgricultureLocationService, AgricultureUOMService, AgricultureChartOfAccountsService, AgricultureTaxCodeService, MfgVerticalItemService, MfgVerticalPartyService, MfgVerticalLocationService, MfgVerticalUOMService, MfgVerticalChartOfAccountsService, MfgVerticalTaxCodeService, solarItemService, solarPartyService, solarLocationService, solarUOMService, solarChartOfAccountsService, solarTaxCodeService, ConstructionMastersService, WaterMastersService, };
// ─── Base Service Factories ──────────────────────────────────────────────────
/** Typed client for ItemService (items, variants, categories, attributes, UOMs, prices) */
export function getItemService() {
    return getApiClient().getService(ItemService);
}
/** Typed client for EntityService (parties — customers, vendors, contacts) */
export function getPartyService() {
    return getApiClient().getService(EntityService);
}
/** Typed client for LocationService (warehouses, branches, zones) */
export function getLocationService() {
    return getApiClient().getService(LocationService);
}
/** Typed client for UOMService (units of measure, conversions) */
export function getUOMService() {
    return getApiClient().getService(UOMService);
}
/** Typed client for ChartOfAccountsService (GL accounts, account groups) */
export function getChartOfAccountsService() {
    return getApiClient().getService(ChartOfAccountsService);
}
/** Typed client for TaxCodeService (tax codes, tax groups, exemptions) */
export function getTaxCodeService() {
    return getApiClient().getService(TaxCodeService);
}
// ─── Agriculture Vertical Factories ──────────────────────────────────────────
export function getAgricultureItemService() {
    return getApiClient().getService(AgricultureItemService);
}
export function getAgriculturePartyService() {
    return getApiClient().getService(AgriculturePartyService);
}
export function getAgricultureLocationService() {
    return getApiClient().getService(AgricultureLocationService);
}
export function getAgricultureUOMService() {
    return getApiClient().getService(AgricultureUOMService);
}
export function getAgricultureChartOfAccountsService() {
    return getApiClient().getService(AgricultureChartOfAccountsService);
}
export function getAgricultureTaxCodeService() {
    return getApiClient().getService(AgricultureTaxCodeService);
}
// ─── MfgVertical Factories ───────────────────────────────────────────────────
export function getMfgVerticalItemService() {
    return getApiClient().getService(MfgVerticalItemService);
}
export function getMfgVerticalPartyService() {
    return getApiClient().getService(MfgVerticalPartyService);
}
export function getMfgVerticalLocationService() {
    return getApiClient().getService(MfgVerticalLocationService);
}
export function getMfgVerticalUOMService() {
    return getApiClient().getService(MfgVerticalUOMService);
}
export function getMfgVerticalChartOfAccountsService() {
    return getApiClient().getService(MfgVerticalChartOfAccountsService);
}
export function getMfgVerticalTaxCodeService() {
    return getApiClient().getService(MfgVerticalTaxCodeService);
}
// ─── Solar Vertical Factories ───────────────────────────────────────────────
export function getSolarItemService() {
    return getApiClient().getService(solarItemService);
}
export function getSolarPartyService() {
    return getApiClient().getService(solarPartyService);
}
export function getSolarLocationService() {
    return getApiClient().getService(solarLocationService);
}
export function getSolarUOMService() {
    return getApiClient().getService(solarUOMService);
}
export function getSolarChartOfAccountsService() {
    return getApiClient().getService(solarChartOfAccountsService);
}
export function getSolarTaxCodeService() {
    return getApiClient().getService(solarTaxCodeService);
}
// ─── Construction Vertical Factories ────────────────────────────────────────
export function getConstructionMastersService() {
    return getApiClient().getService(ConstructionMastersService);
}
// ─── Water Vertical Factories ───────────────────────────────────────────────
export function getWaterMastersService() {
    return getApiClient().getService(WaterMastersService);
}
//# sourceMappingURL=masters.js.map