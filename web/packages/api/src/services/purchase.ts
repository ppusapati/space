/**
 * Purchase Service Factories
 * Typed ConnectRPC clients for procurement, PO, purchase invoice
 *
 * Phase F.9 (2026-04-22): All vertical sub-services retired and unified
 * via classregistry taxonomies on RFQ/ProcurementRequest (procurement),
 * PurchaseOrder (purchaseorder) and PurchaseInvoice (purchaseinvoice).
 * Callers pass `class=<classname>` on Create/Update/List to retain
 * vertical-specific behaviour:
 *   - agri_input_po, construction_site_po, solar_plant_po,
 *     water_utility_po, manufacturing_plant_po on PurchaseOrder
 *   - equivalent classes on ProcurementRequest and PurchaseInvoice
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { ProcurementService } from '@chetana/proto/gen/business/purchase/procurement/proto/procurement_pb.js';
import { PurchaseOrderService } from '@chetana/proto/gen/business/purchase/purchaseorder/proto/purchaseorder_pb.js';
import { PurchaseInvoiceService } from '@chetana/proto/gen/business/purchase/purchaseinvoice/proto/purchaseinvoice_pb.js';

// Vertical-specific procurement/PO/PurchaseInvoice service descriptors
// retired in Phase F.9 — the classregistry taxonomy now carries the
// vertical semantics. See the module docstring above for class keys.

export { ProcurementService, PurchaseOrderService, PurchaseInvoiceService };

export function getProcurementService(): Client<typeof ProcurementService> {
  return getApiClient().getService(ProcurementService);
}

export function getPurchaseOrderService(): Client<typeof PurchaseOrderService> {
  return getApiClient().getService(PurchaseOrderService);
}

export function getPurchaseInvoiceService(): Client<typeof PurchaseInvoiceService> {
  return getApiClient().getService(PurchaseInvoiceService);
}

// ─── Vertical factories retired in Phase F.9 ─────────────────────────────
// getAgricultureProcurementService,   getAgriculturePurchaseInvoiceService,
// getAgriculturePurchaseOrderService, getConstructionProcurementService,
// getConstructionPurchaseOrderService, getConstructionVerticalPurchaseOrderService,
// getMfgVerticalProcurementService,   getMfgVerticalPurchaseInvoiceService,
// getMfgVerticalPurchaseOrderService, getSolarProcurementService,
// getSolarPurchaseInvoiceService,     getSolarPurchaseOrderService,
// getWaterProcurementService,         getWaterPurchaseInvoiceService,
// getWaterPurchaseOrderService,       getWorkVerticalPurchaseOrderService
// are all retired — callers use the base services above with the
// appropriate `class=...` parameter from config/class_registry/*.yaml.
