/**
 * Service Factories — Typed ConnectRPC clients for all backend modules
 *
 * Usage:
 *   import { getItemService } from '@samavāya/api/services';
 *   const items = getItemService();
 *   const response = await items.listItems({ ... });
 *
 * Or import from specific module:
 *   import { getItemService } from '@samavāya/api/services/masters';
 */
export * from './identity.js';
export * from './masters.js';
export * from './finance.js';
export * from './sales.js';
export * from './purchase.js';
export * from './inventory.js';
export * from './hr.js';
export * from './manufacturing.js';
export * from './projects.js';
export * from './asset.js';
export * from './workflow.js';
export * from './fulfillment.js';
export * from './platform.js';
export * from './insights.js';
export * from './extension.js';
//# sourceMappingURL=index.d.ts.map