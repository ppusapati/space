/**
 * Fulfillment Service Factories
 * Typed ConnectRPC clients for returns and shipping
 */
import type { Client } from '@connectrpc/connect';
import { ReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/returns_pb.js';
import { ShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/shipping_pb.js';
import { AgricultureFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/agriculture/fulfillment_agriculture_pb.js';
import { AgricultureReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/agriculture/returns_agriculture_pb.js';
import { AgricultureShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/agriculture/shipping_agriculture_pb.js';
import { ConstructionFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/construction/fulfillment_construction_pb.js';
import { ConstructionReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/construction/returns_construction_pb.js';
import { ConstructionShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/construction/shipping_construction_pb.js';
import { MfgVerticalFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/mfgvertical/fulfillment_mfgvertical_pb.js';
import { MfgVerticalReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/mfgvertical/returns_mfgvertical_pb.js';
import { MfgVerticalShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/mfgvertical/shipping_mfgvertical_pb.js';
import { solarFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/solar/fulfillment_solar_pb.js';
import { solarReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/solar/returns_solar_pb.js';
import { solarShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/solar/shipping_solar_pb.js';
import { WaterFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/water/fulfillment_water_pb.js';
import { WaterReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/water/returns_water_pb.js';
import { WaterShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/water/shipping_water_pb.js';
export { ReturnsService, ShippingService };
/** Typed client for ReturnsService (return orders, RMAs) */
export declare function getReturnsService(): Client<typeof ReturnsService>;
/** Typed client for ShippingService (shipments, tracking, freight) */
export declare function getShippingService(): Client<typeof ShippingService>;
export { AgricultureFulfillmentService, ConstructionFulfillmentService, MfgVerticalFulfillmentService, solarFulfillmentService, WaterFulfillmentService, AgricultureReturnsService, ConstructionReturnsService, MfgVerticalReturnsService, solarReturnsService, WaterReturnsService, AgricultureShippingService, ConstructionShippingService, MfgVerticalShippingService, solarShippingService, WaterShippingService, };
export declare function getAgricultureFulfillmentService(): Client<typeof AgricultureFulfillmentService>;
export declare function getAgricultureReturnsService(): Client<typeof AgricultureReturnsService>;
export declare function getAgricultureShippingService(): Client<typeof AgricultureShippingService>;
export declare function getConstructionFulfillmentService(): Client<typeof ConstructionFulfillmentService>;
export declare function getConstructionReturnsService(): Client<typeof ConstructionReturnsService>;
export declare function getConstructionShippingService(): Client<typeof ConstructionShippingService>;
export declare function getMfgVerticalFulfillmentService(): Client<typeof MfgVerticalFulfillmentService>;
export declare function getMfgVerticalReturnsService(): Client<typeof MfgVerticalReturnsService>;
export declare function getMfgVerticalShippingService(): Client<typeof MfgVerticalShippingService>;
export declare function getSolarFulfillmentService(): Client<typeof solarFulfillmentService>;
export declare function getSolarReturnsService(): Client<typeof solarReturnsService>;
export declare function getSolarShippingService(): Client<typeof solarShippingService>;
export declare function getWaterFulfillmentService(): Client<typeof WaterFulfillmentService>;
export declare function getWaterReturnsService(): Client<typeof WaterReturnsService>;
export declare function getWaterShippingService(): Client<typeof WaterShippingService>;
//# sourceMappingURL=fulfillment.d.ts.map