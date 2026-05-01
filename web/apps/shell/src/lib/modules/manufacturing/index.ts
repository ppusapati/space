/**
 * Manufacturing domain module.
 *
 * Entities (verified live 2026-04-26):
 *   - boms: BOMService.ListBOMs. Response `{meta, boms, pagination}`. Note
 *     the rows key is just `boms` (proto declares `repeated BOM boms`).
 *   - production-orders: ProductionOrderService.ListProductionOrders.
 *     Response uses `orders` as the rows key (NOT `productionOrders`),
 *     matching the proto's `repeated ProductionOrder orders`.
 *   - work-centers: WorkCenterService.ListWorkCenters. Rows key `workCenters`.
 *
 * All three use the canonical packages.api.v1.pagination.Pagination message,
 * so totalCount is always nested under `pagination.totalCount`.
 */
import type { DomainModule } from '../index.js';

export const manufacturing: DomainModule = {
  id: 'manufacturing',
  label: 'Manufacturing',
  entities: [
    {
      slug: 'boms',
      label: 'Bills of Materials',
      formId: 'form_manufacturing_bom',
      listEndpoint: '/manufacturing.bom.api.v1.BOMService/ListBOMs',
      responseRowsKey: 'boms',
      responseTotalKey: 'pagination.totalCount',
      columns: ['bomCode', 'itemName', 'status', 'currentRevision', 'componentCount'],
    },
    {
      slug: 'production-orders',
      label: 'Production Orders',
      formId: 'production_order_management',
      listEndpoint: '/manufacturing.productionorder.api.v1.ProductionOrderService/ListProductionOrders',
      responseRowsKey: 'orders',
      responseTotalKey: 'pagination.totalCount',
      columns: ['orderNumber', 'itemCode', 'status', 'plannedQuantity', 'workCenterCode'],
    },
    {
      slug: 'work-centers',
      label: 'Work Centers',
      formId: 'form_manufacturing_equipment_master',
      listEndpoint: '/manufacturing.workcenter.api.v1.WorkCenterService/ListWorkCenters',
      responseRowsKey: 'workCenters',
      responseTotalKey: 'pagination.totalCount',
      // ListWorkCenters handler shipped 2026-04-26 (item 32 follow-up).
      // Pre-fix the entire WorkCenterHandler embedded
      // UnimplementedWorkCenterServiceHandler with no overrides — every
      // RPC returned 501. Now ListWorkCenters returns 200; other RPCs
      // (Create/Update/Delete/Activate/...) remain 501 until the team's
      // separate proto-field-mapping pass lands.
      columns: ['workCenterId', 'workCenterCode', 'workCenterName', 'workCenterType', 'status', 'department'],
    },
  ],
};
