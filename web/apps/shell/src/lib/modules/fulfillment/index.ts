/**
 * Fulfillment domain module.
 *
 * Entities (verified live 2026-04-26):
 *   - fulfillment-orders: FulfillmentService.ListFulfillmentOrders. Response
 *     key `fulfillmentOrders`.
 *   - shipments: ShippingService.ListShipments. Response key `shipments`.
 *
 * Note: the agent recon initially recommended formId `delivery_confirmation`
 * for shipments but the live probe returned 404. The actual id (verified via
 * GetFormSchema) is `form_fulfillment_delivery_confirmation`.
 */
import type { DomainModule } from '../index.js';

export const fulfillment: DomainModule = {
  id: 'fulfillment',
  label: 'Fulfillment',
  entities: [
    {
      slug: 'fulfillment-orders',
      label: 'Fulfillment Orders',
      formId: 'form_fulfillment_shipment_creation',
      listEndpoint: '/fulfillment.fulfillment.api.v1.FulfillmentService/ListFulfillmentOrders',
      responseRowsKey: 'fulfillmentOrders',
      responseTotalKey: 'pagination.totalCount',
      columns: ['orderId', 'salesOrderId', 'status', 'promisedDate', 'priority'],
    },
    {
      slug: 'shipments',
      label: 'Shipments',
      formId: 'form_fulfillment_delivery_confirmation',
      listEndpoint: '/fulfillment.shipping.api.v1.ShippingService/ListShipments',
      responseRowsKey: 'shipments',
      responseTotalKey: 'pagination.totalCount',
      columns: ['shipmentId', 'carrier', 'status', 'shipDate', 'trackingNumber'],
    },
  ],
};
