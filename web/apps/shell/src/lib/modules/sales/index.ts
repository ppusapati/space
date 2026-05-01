/**
 * Sales domain module.
 *
 * Entities:
 *   - orders: sales order entry. Verified live in Phase 2 (form's
 *     rpcEndpoint /sales.salesorder.api.v1.SalesOrderService/CreateSalesOrder
 *     is mounted; Create returns VALIDATION_ERROR for missing required
 *     fields, proving the inner handler is reached). List endpoint
 *     /sales.salesorder.api.v1.SalesOrderService/ListSalesOrders
 *     returns 200 with empty list under fresh tenant — empty-state UI
 *     is the right answer.
 */
import type { DomainModule } from '../index.js';

export const sales: DomainModule = {
  id: 'sales',
  label: 'Sales',
  entities: [
    {
      slug: 'orders',
      label: 'Sales Orders',
      formId: 'form_sales_sales_order',
      listEndpoint: '/sales.salesorder.api.v1.SalesOrderService/ListSalesOrders',
      responseRowsKey: 'orders',
      responseTotalKey: 'totalCount',
      columns: ['orderNumber', 'customerId', 'orderDate', 'status', 'totalAmount'],
    },
  ],
};
