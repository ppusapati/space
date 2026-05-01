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
// Identity (auth, users, tenants, roles, access)
export * from './identity.js';
// Masters (items, parties, locations, UOM, CoA, tax codes)
export * from './masters.js';
// Finance (ledger, AR, AP, journal, billing, tax, reconciliation)
export * from './finance.js';
// Sales (CRM, orders, invoices, pricing, territory, dealer)
export * from './sales.js';
// Purchase (procurement, PO, purchase invoice)
export * from './purchase.js';
// Inventory (core, lot-serial, quality, planning, WMS)
export * from './inventory.js';
// HR (employee, payroll, leave, attendance, recruitment, training)
export * from './hr.js';
// Manufacturing (BOM, routing, production, job cards, shop floor)
export * from './manufacturing.js';
// Projects (project, task, BOQ, billing, costing, timesheet)
export * from './projects.js';
// Asset (assets, depreciation, equipment, maintenance, vehicles)
export * from './asset.js';
// Workflow (workflow, approval, escalation)
export * from './workflow.js';
// Fulfillment (returns, shipping)
export * from './fulfillment.js';
// Platform & Infrastructure (notifications, banking, budget, audit, data, etc.)
export * from './platform.js';
// Insights (BI, dashboards, reports, search)
export * from './insights.js';
// Extension (land module)
export * from './extension.js';
//# sourceMappingURL=index.js.map