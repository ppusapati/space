/**
 * Purchase Service Factories
 * Typed ConnectRPC clients for procurement, PO, purchase invoice
 */
import { getApiClient } from '../client/client.js';
import { ProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/procurement_pb.js';
import { PurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/purchaseorder_pb.js';
import { PurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/purchaseinvoice_pb.js';
// Vertical-specific — Agriculture
import { AgricultureProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/agriculture/procurement_agriculture_pb.js';
import { AgriculturePurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/agriculture/purchaseinvoice_agriculture_pb.js';
import { AgriculturePurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/agriculture/purchaseorder_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/construction/procurement_construction_pb.js';
import { ConstructionPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/construction/purchaseorder_construction_pb.js';
// Vertical-specific — Construction Vertical
import { ConstructionVerticalPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/constructionvertical/purchaseorder_constructionvertical_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/mfgvertical/procurement_mfgvertical_pb.js';
import { MfgVerticalPurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/mfgvertical/purchaseinvoice_mfgvertical_pb.js';
import { MfgVerticalPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/mfgvertical/purchaseorder_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/solar/procurement_solar_pb.js';
import { solarPurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/solar/purchaseinvoice_solar_pb.js';
import { solarPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/solar/purchaseorder_solar_pb.js';
// Vertical-specific — Water
import { WaterProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/water/procurement_water_pb.js';
import { WaterPurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/water/purchaseinvoice_water_pb.js';
import { WaterPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/water/purchaseorder_water_pb.js';
// Vertical-specific — Work Vertical
import { WorkVerticalPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/workvertical/purchaseorder_workvertical_pb.js';
export { ProcurementService, PurchaseOrderService, PurchaseInvoiceService };
export function getProcurementService() {
    return getApiClient().getService(ProcurementService);
}
export function getPurchaseOrderService() {
    return getApiClient().getService(PurchaseOrderService);
}
export function getPurchaseInvoiceService() {
    return getApiClient().getService(PurchaseInvoiceService);
}
export { AgricultureProcurementService, ConstructionProcurementService, MfgVerticalProcurementService, solarProcurementService, WaterProcurementService, AgriculturePurchaseInvoiceService, MfgVerticalPurchaseInvoiceService, solarPurchaseInvoiceService, WaterPurchaseInvoiceService, AgriculturePurchaseOrderService, ConstructionPurchaseOrderService, ConstructionVerticalPurchaseOrderService, MfgVerticalPurchaseOrderService, solarPurchaseOrderService, WaterPurchaseOrderService, WorkVerticalPurchaseOrderService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureProcurementService() {
    return getApiClient().getService(AgricultureProcurementService);
}
export function getAgriculturePurchaseInvoiceService() {
    return getApiClient().getService(AgriculturePurchaseInvoiceService);
}
export function getAgriculturePurchaseOrderService() {
    return getApiClient().getService(AgriculturePurchaseOrderService);
}
// ─── Construction Vertical Factories ───
export function getConstructionProcurementService() {
    return getApiClient().getService(ConstructionProcurementService);
}
export function getConstructionPurchaseOrderService() {
    return getApiClient().getService(ConstructionPurchaseOrderService);
}
// ─── Construction Vertical Vertical Factories ───
export function getConstructionVerticalPurchaseOrderService() {
    return getApiClient().getService(ConstructionVerticalPurchaseOrderService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalProcurementService() {
    return getApiClient().getService(MfgVerticalProcurementService);
}
export function getMfgVerticalPurchaseInvoiceService() {
    return getApiClient().getService(MfgVerticalPurchaseInvoiceService);
}
export function getMfgVerticalPurchaseOrderService() {
    return getApiClient().getService(MfgVerticalPurchaseOrderService);
}
// ─── Solar Vertical Factories ───
export function getSolarProcurementService() {
    return getApiClient().getService(solarProcurementService);
}
export function getSolarPurchaseInvoiceService() {
    return getApiClient().getService(solarPurchaseInvoiceService);
}
export function getSolarPurchaseOrderService() {
    return getApiClient().getService(solarPurchaseOrderService);
}
// ─── Water Vertical Factories ───
export function getWaterProcurementService() {
    return getApiClient().getService(WaterProcurementService);
}
export function getWaterPurchaseInvoiceService() {
    return getApiClient().getService(WaterPurchaseInvoiceService);
}
export function getWaterPurchaseOrderService() {
    return getApiClient().getService(WaterPurchaseOrderService);
}
// ─── Work Vertical Vertical Factories ───
export function getWorkVerticalPurchaseOrderService() {
    return getApiClient().getService(WorkVerticalPurchaseOrderService);
}
//# sourceMappingURL=purchase.js.map