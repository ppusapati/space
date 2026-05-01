/**
 * Masters Service Factories
 *
 * Typed ConnectRPC clients for items, parties, locations, UOM, CoA, tax codes.
 *
 * Phase F.11 (2026-04-22) consolidation retired the per-vertical factories —
 * Agriculture*, MfgVertical*, Solar*, Construction*, Water* — into the generic
 * services keyed by `class` from the classregistry (see
 * `backend/config/class_registry/{item,uom,party,location,taxcode,chartofaccounts}.yaml`).
 * Vertical-specific behaviour is now selected by passing `class=agri_produce`,
 * `class=construction_material`, `class=solar_component`, `class=water_chemical`,
 * `class=mfg_wip`, etc. on the generic ItemService / EntityService / LocationService
 * / UOMService / ChartOfAccountsService / TaxCodeService RPCs.
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

// Base service descriptors
import { ItemService } from '@samavāya/proto/gen/business/masters/item/proto/item_pb.js';
import { EntityService } from '@samavāya/proto/gen/business/masters/party/proto/party_pb.js';
import { LocationService } from '@samavāya/proto/gen/business/masters/location/proto/location_pb.js';
import { UOMService } from '@samavāya/proto/gen/business/masters/UOM/proto/uom_pb.js';
import { ChartOfAccountsService } from '@samavāya/proto/gen/business/masters/chartofaccounts/proto/chartofaccounts_pb.js';
import { TaxCodeService } from '@samavāya/proto/gen/business/masters/taxcode/proto/taxcode_pb.js';

// Phase F.11.1 — Agriculture*Item/Party/Location/UOM/CoA/TaxCode retired.
//   Replaced by the generic services above with `class=agri_*` filters
//   (e.g. agri_produce, agri_input_po).
//
// Phase F.11.2 — MfgVertical*Item/Party/Location/UOM/CoA/TaxCode retired.
//   Replaced by the generic services above with `class=mfg_wip` filters and
//   manufacturing-oriented plant PO classes.
//
// Phase F.11.3 — solar*Item/Party/Location/UOM/CoA/TaxCode retired.
//   Replaced by the generic services above with `class=solar_*` filters
//   (e.g. solar_component, solar_plant_location).
//
// Phase F.11.7 — ConstructionMastersService / WaterMastersService retired.
//   The masters/masters meta-service is no longer a distinct RPC; use the
//   generic services above with `class=construction_material`,
//   `class=construction_site`, `class=water_chemical`,
//   `class=water_plant_location`, etc.

// Re-export service descriptors
export {
  ItemService, EntityService, LocationService, UOMService, ChartOfAccountsService, TaxCodeService,
};

// ─── Base Service Factories ──────────────────────────────────────────────────

/** Typed client for ItemService (items, variants, categories, attributes, UOMs, prices) */
export function getItemService(): Client<typeof ItemService> {
  return getApiClient().getService(ItemService);
}

/** Typed client for EntityService (parties — customers, vendors, contacts) */
export function getPartyService(): Client<typeof EntityService> {
  return getApiClient().getService(EntityService);
}

/** Typed client for LocationService (warehouses, branches, zones) */
export function getLocationService(): Client<typeof LocationService> {
  return getApiClient().getService(LocationService);
}

/** Typed client for UOMService (units of measure, conversions) */
export function getUOMService(): Client<typeof UOMService> {
  return getApiClient().getService(UOMService);
}

/** Typed client for ChartOfAccountsService (GL accounts, account groups) */
export function getChartOfAccountsService(): Client<typeof ChartOfAccountsService> {
  return getApiClient().getService(ChartOfAccountsService);
}

/** Typed client for TaxCodeService (tax codes, tax groups, exemptions) */
export function getTaxCodeService(): Client<typeof TaxCodeService> {
  return getApiClient().getService(TaxCodeService);
}

// ─── Retired Vertical Factories (Phase F.11) ────────────────────────────────
//
// getAgricultureItemService, getAgriculturePartyService, getAgricultureLocationService,
// getAgricultureUOMService, getAgricultureChartOfAccountsService, getAgricultureTaxCodeService,
// getMfgVerticalItemService, getMfgVerticalPartyService, getMfgVerticalLocationService,
// getMfgVerticalUOMService, getMfgVerticalChartOfAccountsService, getMfgVerticalTaxCodeService,
// getSolarItemService, getSolarPartyService, getSolarLocationService,
// getSolarUOMService, getSolarChartOfAccountsService, getSolarTaxCodeService,
// getConstructionMastersService, getWaterMastersService
//
// All retired. Use the generic getItemService / getPartyService / getLocationService /
// getUOMService / getChartOfAccountsService / getTaxCodeService factories above and pass
// the appropriate `class` from config/class_registry/*.yaml on Create / Update / List calls.
