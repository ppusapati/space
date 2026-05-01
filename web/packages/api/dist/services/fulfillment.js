/**
 * Fulfillment Service Factories
 * Typed ConnectRPC clients for returns and shipping
 */
import { getApiClient } from '../client/client.js';
// Base service descriptors
import { ReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/returns_pb.js';
import { ShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/shipping_pb.js';
// Vertical-specific — Agriculture
import { AgricultureFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/agriculture/fulfillment_agriculture_pb.js';
import { AgricultureReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/agriculture/returns_agriculture_pb.js';
import { AgricultureShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/agriculture/shipping_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/construction/fulfillment_construction_pb.js';
import { ConstructionReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/construction/returns_construction_pb.js';
import { ConstructionShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/construction/shipping_construction_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/mfgvertical/fulfillment_mfgvertical_pb.js';
import { MfgVerticalReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/mfgvertical/returns_mfgvertical_pb.js';
import { MfgVerticalShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/mfgvertical/shipping_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/solar/fulfillment_solar_pb.js';
import { solarReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/solar/returns_solar_pb.js';
import { solarShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/solar/shipping_solar_pb.js';
// Vertical-specific — Water
import { WaterFulfillmentService } from '@samavāya/proto/gen/business/fulfillment/fulfillment/proto/water/fulfillment_water_pb.js';
import { WaterReturnsService } from '@samavāya/proto/gen/business/fulfillment/returns/proto/water/returns_water_pb.js';
import { WaterShippingService } from '@samavāya/proto/gen/business/fulfillment/shipping/proto/water/shipping_water_pb.js';
export { ReturnsService, ShippingService };
// ─── Returns ────────────────────────────────────────────────────────────────
/** Typed client for ReturnsService (return orders, RMAs) */
export function getReturnsService() {
    return getApiClient().getService(ReturnsService);
}
// ─── Shipping ───────────────────────────────────────────────────────────────
/** Typed client for ShippingService (shipments, tracking, freight) */
export function getShippingService() {
    return getApiClient().getService(ShippingService);
}
export { AgricultureFulfillmentService, ConstructionFulfillmentService, MfgVerticalFulfillmentService, solarFulfillmentService, WaterFulfillmentService, AgricultureReturnsService, ConstructionReturnsService, MfgVerticalReturnsService, solarReturnsService, WaterReturnsService, AgricultureShippingService, ConstructionShippingService, MfgVerticalShippingService, solarShippingService, WaterShippingService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureFulfillmentService() {
    return getApiClient().getService(AgricultureFulfillmentService);
}
export function getAgricultureReturnsService() {
    return getApiClient().getService(AgricultureReturnsService);
}
export function getAgricultureShippingService() {
    return getApiClient().getService(AgricultureShippingService);
}
// ─── Construction Vertical Factories ───
export function getConstructionFulfillmentService() {
    return getApiClient().getService(ConstructionFulfillmentService);
}
export function getConstructionReturnsService() {
    return getApiClient().getService(ConstructionReturnsService);
}
export function getConstructionShippingService() {
    return getApiClient().getService(ConstructionShippingService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalFulfillmentService() {
    return getApiClient().getService(MfgVerticalFulfillmentService);
}
export function getMfgVerticalReturnsService() {
    return getApiClient().getService(MfgVerticalReturnsService);
}
export function getMfgVerticalShippingService() {
    return getApiClient().getService(MfgVerticalShippingService);
}
// ─── Solar Vertical Factories ───
export function getSolarFulfillmentService() {
    return getApiClient().getService(solarFulfillmentService);
}
export function getSolarReturnsService() {
    return getApiClient().getService(solarReturnsService);
}
export function getSolarShippingService() {
    return getApiClient().getService(solarShippingService);
}
// ─── Water Vertical Factories ───
export function getWaterFulfillmentService() {
    return getApiClient().getService(WaterFulfillmentService);
}
export function getWaterReturnsService() {
    return getApiClient().getService(WaterReturnsService);
}
export function getWaterShippingService() {
    return getApiClient().getService(WaterShippingService);
}
//# sourceMappingURL=fulfillment.js.map