/**
 * Purchase domain module.
 *
 * Entities (verified live 2026-04-26 against the JWT monolith on :9090):
 *   - purchase-orders: PurchaseOrderService.ListPurchaseOrders. Response shape
 *     `{base, purchaseOrders: [...], pagination: {totalCount: N}}` —
 *     responseTotalKey is dotted because the count lives on the canonical
 *     packages.api.v1.pagination.Pagination message.
 *   - invoices: PurchaseInvoiceService.ListInvoices. Response key `invoices`,
 *     same nested pagination.
 *   - requisitions: ProcurementService.ListRequisitions (procurement service,
 *     not requisition — requisition is one of its endpoints). Response key
 *     `requisitions`.
 *
 * Field naming: ConnectRPC serializes proto snake_case as camelCase JSON.
 * Columns below use the camelCase form clients actually receive.
 */
import type { DomainModule } from '../index.js';

export const purchase: DomainModule = {
  id: 'purchase',
  label: 'Purchase',
  entities: [
    {
      slug: 'purchase-orders',
      label: 'Purchase Orders',
      formId: 'form_purchase_purchase_order',
      listEndpoint: '/purchase.purchaseorder.api.v1.PurchaseOrderService/ListPurchaseOrders',
      responseRowsKey: 'purchaseOrders',
      responseTotalKey: 'pagination.totalCount',
      columns: ['poNumber', 'vendorName', 'status', 'totalAmount', 'poDate'],
    },
    {
      slug: 'invoices',
      label: 'Vendor Invoices',
      formId: 'form_purchase_invoice',
      listEndpoint: '/purchase.purchaseinvoice.api.v1.PurchaseInvoiceService/ListInvoices',
      responseRowsKey: 'invoices',
      responseTotalKey: 'pagination.totalCount',
      columns: ['invoiceNumber', 'vendorName', 'status', 'totalAmount', 'invoiceDate'],
    },
    {
      slug: 'requisitions',
      label: 'Requisitions',
      formId: 'form_purchase_requisition',
      listEndpoint: '/purchase.procurement.api.v1.ProcurementService/ListRequisitions',
      responseRowsKey: 'requisitions',
      responseTotalKey: 'pagination.totalCount',
      columns: ['requisitionNumber', 'status', 'department', 'requestedDate'],
    },
  ],
};
