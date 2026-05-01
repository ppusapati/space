/**
 * Inventory domain module.
 *
 * Entities (verified live 2026-04-26 against the JWT monolith):
 *   - stock-transfers: StockTransferService.ListTransferOrders. Module path
 *     `inventory.stocktransfer` (no hyphen in proto package name) even though
 *     the directory is `business/inventory/stock-transfer`. Proto declares
 *     `repeated TransferOrder orders` — so the JSON key is `orders`, NOT
 *     `transferOrders` (the agent recon's first guess; corrected after
 *     reading the proto).
 *   - cycle-counts: CyclecountService.ListCountPlans (proto package name is
 *     `inventory.cyclecount`, no hyphen). Proto declares
 *     `repeated CountPlan plans` — so the JSON key is `plans`, NOT
 *     `countPlans`. Same correction.
 */
import type { DomainModule } from '../index.js';

export const inventory: DomainModule = {
  id: 'inventory',
  label: 'Inventory',
  entities: [
    {
      slug: 'stock-transfers',
      label: 'Stock Transfers',
      formId: 'form_inventory_transfer',
      listEndpoint: '/inventory.stocktransfer.api.v1.StockTransferService/ListTransferOrders',
      responseRowsKey: 'orders',
      responseTotalKey: 'pagination.totalCount',
      columns: ['transferOrderId', 'fromWarehouse', 'toWarehouse', 'status', 'transferDate'],
    },
    {
      slug: 'cycle-counts',
      label: 'Cycle Counts',
      formId: 'form_inventory_cycle_count',
      listEndpoint: '/inventory.cyclecount.api.v1.CycleCountService/ListCountPlans',
      responseRowsKey: 'plans',
      responseTotalKey: 'pagination.totalCount',
      columns: ['planId', 'warehouseId', 'status', 'scheduledDate', 'actualDate'],
    },
  ],
};
