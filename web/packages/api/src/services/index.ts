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

// Insights (BI, dashboards, reports, search) — legacy, kept for backward compat
export * from './insights.js';

// BI (canonical: core + verticals + query engine, connectors, NL, realtime, embed)
// NOTE: bi.ts re-exports the same core services as insights.ts for the
// legacy BIAnalyticsService / DashboardService / ReportService surfaces.
// Barrel re-export of the full file would cause duplicate-export errors.
// We selectively re-export only the new core/bi/* factories (BI-prefixed
// to avoid legacy collisions) so consumers can import them via the
// package root. Full legacy surface still reachable via explicit
// '@samavāya/api/services/bi' import.
export {
  BIDatasetService,
  BIQueryService,
  BIReportService,
  BIPresentationService,
  getBIDatasetService,
  getBIQueryService,
  getBIReportService,
  getBIPresentationService,
} from './bi.js';

// Extension (land module)
export * from './extension.js';

// FormService (form-first rendering: module discovery, schema, submission)
export {
  listModules,
  listForms,
  getFormSchema,
  submitForm,
  FormServiceError,
} from './formService.js';
export type {
  ModuleSummary,
  FormSummary,
  ProtoFormDefinition,
  SubmitFormResponse,
  SubmitValidationError,
  ListModulesResponse,
  ListFormsResponse,
  GetFormSchemaResponse,
} from './formService.js';

