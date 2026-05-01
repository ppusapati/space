/**
 * Masters Service Factories
 * Typed ConnectRPC clients for items, parties, locations, UOM, CoA, tax codes
 */
import type { Client } from '@connectrpc/connect';
import { ItemService } from '@samavāya/proto/gen/business/masters/item/proto/item_pb.js';
import { EntityService } from '@samavāya/proto/gen/business/masters/party/proto/party_pb.js';
import { LocationService } from '@samavāya/proto/gen/business/masters/location/proto/location_pb.js';
import { UOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/uom_pb.js';
import { ChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/chartofaccounts_pb.js';
import { TaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/taxcode_pb.js';
import { AgricultureItemService } from '@samavāya/proto/gen/business/masters/item/proto/agriculture/item_agriculture_pb.js';
import { AgriculturePartyService } from '@samavāya/proto/gen/business/masters/party/proto/agriculture/party_agriculture_pb.js';
import { AgricultureLocationService } from '@samavāya/proto/gen/business/masters/location/proto/agriculture/location_agriculture_pb.js';
import { AgricultureUOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/agriculture/uom_agriculture_pb.js';
import { AgricultureChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/agriculture/chartofaccounts_agriculture_pb.js';
import { AgricultureTaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/agriculture/taxcode_agriculture_pb.js';
import { MfgVerticalItemService } from '@samavāya/proto/gen/business/masters/item/proto/mfgvertical/item_mfgvertical_pb.js';
import { MfgVerticalPartyService } from '@samavāya/proto/gen/business/masters/party/proto/mfgvertical/party_mfgvertical_pb.js';
import { MfgVerticalLocationService } from '@samavāya/proto/gen/business/masters/location/proto/mfgvertical/location_mfgvertical_pb.js';
import { MfgVerticalUOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/mfgvertical/uom_mfgvertical_pb.js';
import { MfgVerticalChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/mfgvertical/chartofaccounts_mfgvertical_pb.js';
import { MfgVerticalTaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/mfgvertical/taxcode_mfgvertical_pb.js';
import { solarItemService } from '@samavāya/proto/gen/business/masters/item/proto/solar/item_solar_pb.js';
import { solarPartyService } from '@samavāya/proto/gen/business/masters/party/proto/solar/party_solar_pb.js';
import { solarLocationService } from '@samavāya/proto/gen/business/masters/location/proto/solar/location_solar_pb.js';
import { solarUOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/solar/uom_solar_pb.js';
import { solarChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/solar/chartofaccounts_solar_pb.js';
import { solarTaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/solar/taxcode_solar_pb.js';
import { ConstructionMastersService } from '@samavāya/proto/gen/business/masters/masters/proto/construction/masters_construction_pb.js';
import { WaterMastersService } from '@samavāya/proto/gen/business/masters/masters/proto/water/masters_water_pb.js';
export { ItemService, EntityService, LocationService, UOMService, ChartOfAccountsService, TaxCodeService, AgricultureItemService, AgriculturePartyService, AgricultureLocationService, AgricultureUOMService, AgricultureChartOfAccountsService, AgricultureTaxCodeService, MfgVerticalItemService, MfgVerticalPartyService, MfgVerticalLocationService, MfgVerticalUOMService, MfgVerticalChartOfAccountsService, MfgVerticalTaxCodeService, solarItemService, solarPartyService, solarLocationService, solarUOMService, solarChartOfAccountsService, solarTaxCodeService, ConstructionMastersService, WaterMastersService, };
/** Typed client for ItemService (items, variants, categories, attributes, UOMs, prices) */
export declare function getItemService(): Client<typeof ItemService>;
/** Typed client for EntityService (parties — customers, vendors, contacts) */
export declare function getPartyService(): Client<typeof EntityService>;
/** Typed client for LocationService (warehouses, branches, zones) */
export declare function getLocationService(): Client<typeof LocationService>;
/** Typed client for UOMService (units of measure, conversions) */
export declare function getUOMService(): Client<typeof UOMService>;
/** Typed client for ChartOfAccountsService (GL accounts, account groups) */
export declare function getChartOfAccountsService(): Client<typeof ChartOfAccountsService>;
/** Typed client for TaxCodeService (tax codes, tax groups, exemptions) */
export declare function getTaxCodeService(): Client<typeof TaxCodeService>;
export declare function getAgricultureItemService(): Client<typeof AgricultureItemService>;
export declare function getAgriculturePartyService(): Client<typeof AgriculturePartyService>;
export declare function getAgricultureLocationService(): Client<typeof AgricultureLocationService>;
export declare function getAgricultureUOMService(): Client<typeof AgricultureUOMService>;
export declare function getAgricultureChartOfAccountsService(): Client<typeof AgricultureChartOfAccountsService>;
export declare function getAgricultureTaxCodeService(): Client<typeof AgricultureTaxCodeService>;
export declare function getMfgVerticalItemService(): Client<typeof MfgVerticalItemService>;
export declare function getMfgVerticalPartyService(): Client<typeof MfgVerticalPartyService>;
export declare function getMfgVerticalLocationService(): Client<typeof MfgVerticalLocationService>;
export declare function getMfgVerticalUOMService(): Client<typeof MfgVerticalUOMService>;
export declare function getMfgVerticalChartOfAccountsService(): Client<typeof MfgVerticalChartOfAccountsService>;
export declare function getMfgVerticalTaxCodeService(): Client<typeof MfgVerticalTaxCodeService>;
export declare function getSolarItemService(): Client<typeof solarItemService>;
export declare function getSolarPartyService(): Client<typeof solarPartyService>;
export declare function getSolarLocationService(): Client<typeof solarLocationService>;
export declare function getSolarUOMService(): Client<typeof solarUOMService>;
export declare function getSolarChartOfAccountsService(): Client<typeof solarChartOfAccountsService>;
export declare function getSolarTaxCodeService(): Client<typeof solarTaxCodeService>;
export declare function getConstructionMastersService(): Client<typeof ConstructionMastersService>;
export declare function getWaterMastersService(): Client<typeof WaterMastersService>;
//# sourceMappingURL=masters.d.ts.map