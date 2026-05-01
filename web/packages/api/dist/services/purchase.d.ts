/**
 * Purchase Service Factories
 * Typed ConnectRPC clients for procurement, PO, purchase invoice
 */
import type { Client } from '@connectrpc/connect';
import { ProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/procurement_pb.js';
import { PurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/purchaseorder_pb.js';
import { PurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/purchaseinvoice_pb.js';
import { AgricultureProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/agriculture/procurement_agriculture_pb.js';
import { AgriculturePurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/agriculture/purchaseinvoice_agriculture_pb.js';
import { AgriculturePurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/agriculture/purchaseorder_agriculture_pb.js';
import { ConstructionProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/construction/procurement_construction_pb.js';
import { ConstructionPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/construction/purchaseorder_construction_pb.js';
import { ConstructionVerticalPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/constructionvertical/purchaseorder_constructionvertical_pb.js';
import { MfgVerticalProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/mfgvertical/procurement_mfgvertical_pb.js';
import { MfgVerticalPurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/mfgvertical/purchaseinvoice_mfgvertical_pb.js';
import { MfgVerticalPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/mfgvertical/purchaseorder_mfgvertical_pb.js';
import { solarProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/solar/procurement_solar_pb.js';
import { solarPurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/solar/purchaseinvoice_solar_pb.js';
import { solarPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/solar/purchaseorder_solar_pb.js';
import { WaterProcurementService } from '@samavāya/proto/gen/business/purchase/procurement/proto/water/procurement_water_pb.js';
import { WaterPurchaseInvoiceService } from '@samavāya/proto/gen/business/purchase/purchaseinvoice/proto/water/purchaseinvoice_water_pb.js';
import { WaterPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/water/purchaseorder_water_pb.js';
import { WorkVerticalPurchaseOrderService } from '@samavāya/proto/gen/business/purchase/purchaseorder/proto/workvertical/purchaseorder_workvertical_pb.js';
export { ProcurementService, PurchaseOrderService, PurchaseInvoiceService };
export declare function getProcurementService(): Client<typeof ProcurementService>;
export declare function getPurchaseOrderService(): Client<typeof PurchaseOrderService>;
export declare function getPurchaseInvoiceService(): Client<typeof PurchaseInvoiceService>;
export { AgricultureProcurementService, ConstructionProcurementService, MfgVerticalProcurementService, solarProcurementService, WaterProcurementService, AgriculturePurchaseInvoiceService, MfgVerticalPurchaseInvoiceService, solarPurchaseInvoiceService, WaterPurchaseInvoiceService, AgriculturePurchaseOrderService, ConstructionPurchaseOrderService, ConstructionVerticalPurchaseOrderService, MfgVerticalPurchaseOrderService, solarPurchaseOrderService, WaterPurchaseOrderService, WorkVerticalPurchaseOrderService, };
export declare function getAgricultureProcurementService(): Client<typeof AgricultureProcurementService>;
export declare function getAgriculturePurchaseInvoiceService(): Client<typeof AgriculturePurchaseInvoiceService>;
export declare function getAgriculturePurchaseOrderService(): Client<typeof AgriculturePurchaseOrderService>;
export declare function getConstructionProcurementService(): Client<typeof ConstructionProcurementService>;
export declare function getConstructionPurchaseOrderService(): Client<typeof ConstructionPurchaseOrderService>;
export declare function getConstructionVerticalPurchaseOrderService(): Client<typeof ConstructionVerticalPurchaseOrderService>;
export declare function getMfgVerticalProcurementService(): Client<typeof MfgVerticalProcurementService>;
export declare function getMfgVerticalPurchaseInvoiceService(): Client<typeof MfgVerticalPurchaseInvoiceService>;
export declare function getMfgVerticalPurchaseOrderService(): Client<typeof MfgVerticalPurchaseOrderService>;
export declare function getSolarProcurementService(): Client<typeof solarProcurementService>;
export declare function getSolarPurchaseInvoiceService(): Client<typeof solarPurchaseInvoiceService>;
export declare function getSolarPurchaseOrderService(): Client<typeof solarPurchaseOrderService>;
export declare function getWaterProcurementService(): Client<typeof WaterProcurementService>;
export declare function getWaterPurchaseInvoiceService(): Client<typeof WaterPurchaseInvoiceService>;
export declare function getWaterPurchaseOrderService(): Client<typeof WaterPurchaseOrderService>;
export declare function getWorkVerticalPurchaseOrderService(): Client<typeof WorkVerticalPurchaseOrderService>;
//# sourceMappingURL=purchase.d.ts.map