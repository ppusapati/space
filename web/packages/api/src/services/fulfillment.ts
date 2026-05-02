/**
 * Fulfillment Service Factories
 * Typed ConnectRPC clients for fulfillment, returns and shipping.
 *
 * Phase F.7 (2026-04-22): All vertical sub-services retired and unified
 * via classregistry taxonomies on FulfillmentOrder (fulfillment),
 * ReturnRequest (returns) and Shipment (shipping). Callers pass
 * `class=<classname>` on Create/Update/List to retain vertical-specific
 * behaviour:
 *   - agri_bulk_fulfillment, construction_site_fulfillment,
 *     solar_plant_fulfillment, water_utility_fulfillment,
 *     manufacturing_interplant_fulfillment on FulfillmentOrder
 *   - equivalent classes on ReturnRequest (F.7.2) and Shipment (F.7.3)
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { FulfillmentService } from '@chetana/proto/gen/business/fulfillment/fulfillment/proto/fulfillment_pb.js';
import { ReturnsService } from '@chetana/proto/gen/business/fulfillment/returns/proto/returns_pb.js';
import { ShippingService } from '@chetana/proto/gen/business/fulfillment/shipping/proto/shipping_pb.js';

// Vertical-specific fulfillment/returns/shipping service descriptors
// retired in Phase F.7 — the classregistry taxonomy now carries the
// vertical semantics. See the module docstring above for class keys.

export { FulfillmentService, ReturnsService, ShippingService };

/** Typed client for FulfillmentService (fulfillment orders, waves, pick/pack/ship) */
export function getFulfillmentService(): Client<typeof FulfillmentService> {
  return getApiClient().getService(FulfillmentService);
}

/** Typed client for ReturnsService (return orders, RMAs) */
export function getReturnsService(): Client<typeof ReturnsService> {
  return getApiClient().getService(ReturnsService);
}

/** Typed client for ShippingService (shipments, tracking, freight) */
export function getShippingService(): Client<typeof ShippingService> {
  return getApiClient().getService(ShippingService);
}

// ─── Vertical factories retired in Phase F.7 ───────────────────────────────
// F.7.1 FulfillmentOrder:
//   getAgricultureFulfillmentService, getConstructionFulfillmentService,
//   getMfgVerticalFulfillmentService, getSolarFulfillmentService,
//   getWaterFulfillmentService
// F.7.2 ReturnRequest:
//   getAgricultureReturnsService, getConstructionReturnsService,
//   getMfgVerticalReturnsService, getSolarReturnsService,
//   getWaterReturnsService
// F.7.3 Shipment:
//   getAgricultureShippingService, getConstructionShippingService,
//   getMfgVerticalShippingService, getSolarShippingService,
//   getWaterShippingService
// All retired — callers use the base services above with `class=...`
// from config/class_registry/{fulfillment,returns,shipping}.yaml.
