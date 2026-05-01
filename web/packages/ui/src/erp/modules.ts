/**
 * ERP Module Registry
 *
 * Defines all available modules with their routes, icons, and metadata.
 * The sidebar reads this registry + user permissions to show available modules.
 */

export interface ModuleDef {
  id: string;
  label: string;
  path: string;
  /** SVG path data for the module icon (24x24 viewBox, stroke-based) */
  icon: string;
  order: number;
  /** Sub-navigation sections shown when this module is active */
  sections?: ModuleSection[];
}

export interface ModuleSection {
  title: string;
  items: ModuleSectionItem[];
}

export interface ModuleSectionItem {
  label: string;
  path: string;
  icon?: string;
}

/**
 * All ERP modules. Sidebar filters this based on which modules are deployed
 * and which the user has permissions for.
 */
export const MODULE_REGISTRY: ModuleDef[] = [
  {
    id: 'dashboard',
    label: 'Dashboard',
    path: '/',
    icon: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
    order: 0,
  },
  {
    id: 'identity',
    label: 'Identity',
    path: '/identity',
    icon: 'M12 2a5 5 0 015 5v1A5 5 0 017 8V7a5 5 0 015-5zM20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2',
    order: 1,
    sections: [
      {
        title: 'Tenancy & Users',
        items: [
          { label: 'Tenants', path: '/identity/tenants' },
          { label: 'Entities', path: '/identity/entities' },
          { label: 'Users', path: '/identity/users' },
        ],
      },
      {
        title: 'Access Control',
        items: [
          { label: 'Roles', path: '/identity/roles' },
          { label: 'Permissions', path: '/identity/permissions' },
          { label: 'Resources', path: '/identity/resources' },
          { label: 'ACLs', path: '/identity/acls' },
          { label: 'Attributes', path: '/identity/attributes' },
          { label: 'Policies', path: '/identity/policies' },
          { label: 'Delegations', path: '/identity/delegations' },
          { label: 'User Roles (hierarchy)', path: '/identity/user-roles' },
          { label: 'Permission Definitions', path: '/identity/permission-defs' },
        ],
      },
      {
        title: 'Policy Decision Point',
        items: [
          { label: 'Evaluation Strategies', path: '/identity/evaluation-strategies' },
          { label: 'Custom Rules', path: '/identity/custom-rules' },
          { label: 'Test Cases', path: '/identity/test-cases' },
          { label: 'Attribute Providers', path: '/identity/attribute-providers' },
        ],
      },
    ],
  },
  {
    id: 'masters',
    label: 'Masters',
    path: '/masters',
    icon: 'M20 7H4a2 2 0 00-2 2v10a2 2 0 002 2h16a2 2 0 002-2V9a2 2 0 00-2-2zM16 21V5a2 2 0 00-2-2h-4a2 2 0 00-2 2v16',
    order: 2,
    sections: [
      {
        title: 'Master Data',
        items: [
          { label: 'Items', path: '/masters/items' },
          { label: 'Item Categories', path: '/masters/categories' },
          { label: 'Units of Measure', path: '/masters/uoms' },
          { label: 'Parties', path: '/masters/parties' },
          { label: 'Locations', path: '/masters/locations' },
          { label: 'Tax Codes', path: '/masters/tax-codes' },
          { label: 'Chart of Accounts', path: '/masters/accounts' },
        ],
      },
    ],
  },
  {
    id: 'finance',
    label: 'Finance',
    path: '/finance',
    icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6',
    order: 3,
    sections: [
      {
        title: 'General Ledger',
        items: [
          { label: 'Chart of Accounts', path: '/finance/gl/accounts' },
          { label: 'Journal Entries', path: '/finance/journal-entries' },
          { label: 'Fiscal Periods', path: '/finance/gl/fiscal-periods' },
        ],
      },
      {
        title: 'Accounts Payable',
        items: [
          { label: 'Bills', path: '/finance/ap/bills' },
          { label: 'Payments', path: '/finance/ap/payments' },
          { label: 'AP Aging', path: '/finance/ap/aging' },
        ],
      },
      {
        title: 'Accounts Receivable',
        items: [
          { label: 'Invoices', path: '/finance/ar/invoices' },
          { label: 'Receipts', path: '/finance/ar/receipts' },
          { label: 'AR Aging', path: '/finance/ar/aging' },
        ],
      },
      {
        title: 'Reports',
        items: [
          { label: 'Trial Balance', path: '/finance/reports/trial-balance' },
          { label: 'Balance Sheet', path: '/finance/reports/balance-sheet' },
          { label: 'Income Statement', path: '/finance/reports/income-statement' },
          { label: 'Cash Flow', path: '/finance/reports/cash-flow' },
        ],
      },
    ],
  },
  {
    id: 'sales',
    label: 'Sales',
    path: '/sales',
    icon: 'M3 3v18h18M18.7 8l-5.1 5.2-2.8-2.7L7 14.3',
    order: 4,
    sections: [
      {
        title: 'Sales',
        items: [
          { label: 'CRM', path: '/sales/crm' },
          { label: 'Sales Orders', path: '/sales/orders' },
          { label: 'Invoices', path: '/sales/invoices' },
          { label: 'Pricing', path: '/sales/pricing' },
          { label: 'Dealers', path: '/sales/dealers' },
          { label: 'Commissions', path: '/sales/commissions' },
        ],
      },
    ],
  },
  {
    id: 'purchase',
    label: 'Purchase',
    path: '/purchase',
    icon: 'M9 21a1 1 0 100-2 1 1 0 000 2zM20 21a1 1 0 100-2 1 1 0 000 2zM1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6',
    order: 5,
    sections: [
      {
        title: 'Procurement',
        items: [
          { label: 'Requisitions', path: '/purchase/requisitions' },
          { label: 'Purchase Orders', path: '/purchase/purchase-orders' },
          { label: 'Invoices', path: '/purchase/invoices' },
          { label: 'Vendors', path: '/purchase/vendors' },
        ],
      },
    ],
  },
  {
    id: 'inventory',
    label: 'Inventory',
    path: '/inventory',
    icon: 'M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16zM3.27 6.96L12 12.01l8.73-5.05M12 22.08V12',
    order: 6,
    sections: [
      {
        title: 'Inventory',
        items: [
          { label: 'Stock Transfers', path: '/inventory/stock-transfers' },
          { label: 'Cycle Counts', path: '/inventory/cycle-counts' },
          { label: 'Stock', path: '/inventory/stock' },
          { label: 'Lot & Serial', path: '/inventory/lot-serial' },
          { label: 'Quality', path: '/inventory/quality' },
          { label: 'Warehouse', path: '/inventory/warehouse' },
        ],
      },
    ],
  },
  {
    id: 'hr',
    label: 'HR',
    path: '/hr',
    icon: 'M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75',
    order: 7,
    sections: [
      {
        title: 'HR Management',
        items: [
          { label: 'Employees', path: '/hr/employees' },
          { label: 'Payroll', path: '/hr/payroll' },
          { label: 'Leave', path: '/hr/leave' },
          { label: 'Attendance', path: '/hr/attendance' },
          { label: 'Recruitment', path: '/hr/recruitment' },
          { label: 'Training', path: '/hr/training' },
        ],
      },
    ],
  },
  {
    id: 'manufacturing',
    label: 'Manufacturing',
    path: '/manufacturing',
    icon: 'M2 20h20M6 20V4l6 4V4l6 4v12',
    order: 8,
    sections: [
      {
        title: 'Production',
        items: [
          { label: 'BOM', path: '/manufacturing/boms' },
          { label: 'Production Orders', path: '/manufacturing/production-orders' },
          { label: 'Work Centers', path: '/manufacturing/work-centers' },
          { label: 'Job Cards', path: '/manufacturing/job-cards' },
          { label: 'Routing', path: '/manufacturing/routing' },
          { label: 'Shop Floor', path: '/manufacturing/shop-floor' },
        ],
      },
    ],
  },
  {
    id: 'projects',
    label: 'Projects',
    path: '/projects',
    icon: 'M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z',
    order: 9,
    sections: [
      {
        title: 'Project Management',
        items: [
          { label: 'Projects', path: '/projects/projects' },
          { label: 'Tasks', path: '/projects/tasks' },
          { label: 'BOQ', path: '/projects/boq' },
          { label: 'Timesheets', path: '/projects/timesheets' },
          { label: 'Billing', path: '/projects/billing' },
        ],
      },
    ],
  },
  {
    id: 'asset',
    label: 'Assets',
    path: '/asset',
    icon: 'M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z',
    order: 10,
    sections: [
      {
        title: 'Asset Register',
        items: [
          { label: 'Assets', path: '/asset/assets' },
          { label: 'Categories', path: '/asset/categories' },
          { label: 'Asset Classes', path: '/asset/asset-classes' },
          { label: 'Transfers', path: '/asset/transfers' },
          { label: 'Disposals', path: '/asset/disposals' },
          { label: 'Revaluations', path: '/asset/revaluations' },
        ],
      },
      {
        title: 'Depreciation',
        items: [
          { label: 'Setups', path: '/asset/depreciation-setups' },
          { label: 'Schedules', path: '/asset/depreciation-schedules' },
          { label: 'Runs', path: '/asset/depreciation-runs' },
          { label: 'Book-Tax Differences', path: '/asset/book-tax-differences' },
          { label: 'Depreciation Classes', path: '/asset/depreciation-classes' },
        ],
      },
      {
        title: 'Equipment',
        items: [
          { label: 'Equipment', path: '/asset/equipment' },
          { label: 'Equipment Categories', path: '/asset/equipment-categories' },
          { label: 'Certifications', path: '/asset/certifications' },
          { label: 'Equipment Classes', path: '/asset/equipment-classes' },
        ],
      },
      {
        title: 'Maintenance',
        items: [
          { label: 'Maintenance Requests', path: '/asset/maintenance-requests' },
          { label: 'Work Orders', path: '/asset/work-orders' },
          { label: 'PM Schedules', path: '/asset/pm-schedules' },
          { label: 'Spare Parts', path: '/asset/spare-parts' },
          { label: 'Checklist Templates', path: '/asset/checklist-templates' },
          { label: 'Maintenance Classes', path: '/asset/maintenance-classes' },
        ],
      },
      {
        title: 'Fleet',
        items: [
          { label: 'Vehicles', path: '/asset/vehicles' },
          { label: 'Trips', path: '/asset/trips' },
          { label: 'Fuel Entries', path: '/asset/fuel-entries' },
          { label: 'Driver Assignments', path: '/asset/driver-assignments' },
          { label: 'Vehicle Classes', path: '/asset/vehicle-classes' },
        ],
      },
    ],
  },
  {
    id: 'fulfillment',
    label: 'Fulfillment',
    path: '/fulfillment',
    icon: 'M16 16l3-8 3 8c-1.05.63-2.26 1-3.5 1s-2.45-.37-3.5-1zM2 16l3-8 3 8c-1.05.63-2.26 1-3.5 1S2.45 16.63 2 16zM7 21h10M12 3v18',
    order: 11,
    sections: [
      {
        title: 'Fulfillment',
        items: [
          { label: 'Fulfillment Orders', path: '/fulfillment/fulfillment-orders' },
          { label: 'Shipments', path: '/fulfillment/shipments' },
          { label: 'Returns', path: '/fulfillment/returns' },
        ],
      },
    ],
  },
  {
    id: 'insights',
    label: 'Insights',
    path: '/insights',
    icon: 'M18 20V10M12 20V4M6 20v-6',
    order: 12,
    sections: [
      {
        title: 'Analytics',
        items: [
          { label: 'Dashboards', path: '/insights/dashboards' },
          { label: 'Reports', path: '/insights/reports' },
          { label: 'BI Analytics', path: '/insights/bi' },
        ],
      },
    ],
  },
  {
    id: 'workflow',
    label: 'Workflow',
    path: '/workflow',
    icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2M12 12h4M12 16h4M8 12h.01M8 16h.01',
    order: 13,
    sections: [
      {
        title: 'Process Engine',
        items: [
          { label: 'Workflow Definitions', path: '/workflow/workflow-definitions' },
          { label: 'Workflow Executions', path: '/workflow/workflow-executions' },
          { label: 'Workflow Tasks', path: '/workflow/workflow-tasks' },
        ],
      },
      {
        title: 'Approvals',
        items: [
          { label: 'Approval Instances', path: '/workflow/approval-instances' },
          { label: 'Approval Delegates', path: '/workflow/approval-delegates' },
          { label: 'Department Levels', path: '/workflow/department-levels' },
        ],
      },
      {
        title: 'Escalation',
        items: [
          { label: 'Escalation Rules', path: '/workflow/escalation-rules' },
          { label: 'Escalation Chains', path: '/workflow/escalation-chains' },
          { label: 'Escalation Triggers', path: '/workflow/escalation-triggers' },
        ],
      },
    ],
  },
  {
    id: 'budget',
    label: 'Budget',
    path: '/budget',
    icon: 'M9 7h6M9 11h6M9 15h4M5 3h14a2 2 0 012 2v14a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2z',
    order: 14,
    sections: [
      {
        title: 'Budget Management',
        items: [
          { label: 'Budgets', path: '/budget/budgets' },
          // /budget/budgets is wired live via lib/modules/budget/index.ts
          // (BudgetService.ListBudgets returns 200 under JWT). The other
          // entries below remain aspirational until their entities are added.
          { label: 'Variance', path: '/budget/variance' },
          { label: 'Capex', path: '/budget/capex' },
          { label: 'Forecasting', path: '/budget/forecasting' },
        ],
      },
    ],
  },
  {
    id: 'banking',
    label: 'Banking',
    path: '/banking',
    icon: 'M3 21h18M3 10h18M5 6l7-3 7 3M4 10v11M20 10v11M8 14v3M12 14v3M16 14v3',
    order: 15,
    sections: [
      {
        title: 'Statutory & Banking',
        items: [
          { label: 'Bank Accounts', path: '/banking/bank-accounts' },
          { label: 'GST', path: '/banking/gst' },
          { label: 'TDS', path: '/banking/tds' },
          { label: 'E-Invoice', path: '/banking/e-invoice' },
          { label: 'E-Way Bill', path: '/banking/e-way-bill' },
        ],
      },
    ],
  },
  {
    id: 'notifications',
    label: 'Notifications',
    path: '/notifications',
    icon: 'M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0',
    order: 16,
    sections: [
      {
        title: 'Notification Center',
        items: [
          { label: 'Notifications', path: '/notifications/notifications' },
          { label: 'Notification Templates', path: '/notifications/notification-templates' },
        ],
      },
      {
        title: 'Template Authoring',
        items: [
          { label: 'Templates', path: '/notifications/templates' },
        ],
      },
    ],
  },
  {
    id: 'audit',
    label: 'Audit',
    path: '/audit',
    icon: 'M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z',
    order: 17,
    sections: [
      {
        title: 'Audit & Compliance',
        items: [
          { label: 'Audit Log', path: '/audit/log' },
          { label: 'Compliance Rules', path: '/audit/compliance-rules' },
          { label: 'Changelog', path: '/audit/changelog' },
          { label: 'Compliance', path: '/audit/compliance' },
          { label: 'GDPR', path: '/audit/gdpr' },
          { label: 'Retention', path: '/audit/retention' },
        ],
      },
    ],
  },
  {
    id: 'platform',
    label: 'Platform',
    path: '/platform',
    icon: 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4',
    order: 18,
    sections: [
      {
        title: 'API Gateway',
        items: [
          { label: 'API Routes', path: '/platform/api-routes' },
          { label: 'API Keys', path: '/platform/api-keys' },
          { label: 'Rate Limits', path: '/platform/rate-limits' },
          { label: 'Circuit Breakers', path: '/platform/circuit-breakers' },
          { label: 'API Versions', path: '/platform/api-versions' },
        ],
      },
      {
        title: 'Storage & Files',
        items: [
          { label: 'Files', path: '/platform/files' },
          { label: 'Folders', path: '/platform/folders' },
          { label: 'QR Codes', path: '/platform/qr-codes' },
          { label: 'Label Templates', path: '/platform/label-templates' },
        ],
      },
      {
        title: 'Jobs & Scheduling',
        items: [
          { label: 'Scheduled Jobs', path: '/platform/scheduled-jobs' },
          { label: 'Batch Definitions', path: '/platform/batches' },
          { label: 'Batch Schedules', path: '/platform/batch-schedules' },
          { label: 'Queues', path: '/platform/queues' },
        ],
      },
      {
        title: 'Integration & Print',
        items: [
          { label: 'Integrations', path: '/platform/integrations' },
          { label: 'Printers', path: '/platform/printers' },
          { label: 'Print Jobs', path: '/platform/print-jobs' },
          { label: 'Print Templates', path: '/platform/print-templates' },
          { label: 'Webhooks', path: '/platform/webhooks' },
        ],
      },
      {
        title: 'Reliability & Config',
        items: [
          { label: 'SLA Calendars', path: '/platform/sla-calendars' },
          { label: 'SLA Rules', path: '/platform/sla-rules' },
          { label: 'System Settings', path: '/platform/settings' },
          { label: 'Feature Flags', path: '/platform/feature-flags' },
        ],
      },
    ],
  },
  {
    id: 'communication',
    label: 'Communication',
    path: '/communication',
    icon: 'M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z',
    order: 19,
    sections: [
      {
        title: 'Chat',
        items: [
          { label: 'Conversations', path: '/communication/conversations' },
        ],
      },
      {
        title: 'Currency',
        items: [
          { label: 'Currencies', path: '/communication/currencies' },
          { label: 'Conversion Logs', path: '/communication/conversion-logs' },
          { label: 'Rate Alerts', path: '/communication/rate-alerts' },
        ],
      },
    ],
  },
  {
    id: 'data',
    label: 'Data',
    path: '/data',
    icon: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4M4 12c0 2.21 3.582 4 8 4s8-1.79 8-4',
    order: 20,
    sections: [
      {
        title: 'Data Management',
        items: [
          { label: 'Backup Policies', path: '/data/backup-policies' },
          { label: 'Import/Export', path: '/data/bridge' },
          { label: 'Archive', path: '/data/archive' },
          { label: 'Backup & DR', path: '/data/backup' },
        ],
      },
    ],
  },
  {
    id: 'land',
    label: 'Land Acquisition',
    path: '/land',
    icon: 'M1 22l5-10 5 10M1 22h10M8 8l4-6 4 6M8 8h8M18 22l3-6 3 6M18 22h6M2 17h7M15 13h5',
    order: 21,
    sections: [
      {
        title: 'Land Parcels',
        items: [
          { label: 'Land Parcel', path: '/land/land-parcel' },
          { label: 'GIS & Spatial', path: '/land/gis-spatial' },
          { label: 'Field Operations', path: '/land/field-operations' },
        ],
      },
      {
        title: 'Legal & Compliance',
        items: [
          { label: 'Compliance', path: '/land/compliance' },
          { label: 'Due Diligence', path: '/land/due-diligence' },
          { label: 'Legal Cases', path: '/land/legal-case' },
        ],
      },
      {
        title: 'Transactions',
        items: [
          { label: 'Negotiation', path: '/land/negotiation' },
          { label: 'Stakeholders', path: '/land/stakeholder' },
          { label: 'Land Finance', path: '/land/land-finance' },
        ],
      },
      {
        title: 'Analysis & Leasing',
        items: [
          { label: 'Risk Scoring', path: '/land/risk-scoring' },
          { label: 'Land Insights', path: '/land/land-insights' },
          { label: 'Govt Lease', path: '/land/govt-lease' },
          { label: 'Grid Interconnection', path: '/land/grid-interconnection' },
          { label: 'Right of Way', path: '/land/right-of-way' },
          { label: 'Renewable Energy Finance', path: '/land/renewable-energy-finance' },
        ],
      },
    ],
  },
];

/** Get a module definition by ID */
export function getModule(id: string): ModuleDef | undefined {
  return MODULE_REGISTRY.find((m) => m.id === id);
}

/** Get modules filtered by a list of enabled module IDs */
export function getEnabledModules(enabledIds?: string[]): ModuleDef[] {
  if (!enabledIds) return MODULE_REGISTRY;
  const idSet = new Set(enabledIds);
  return MODULE_REGISTRY.filter((m) => m.id === 'dashboard' || idSet.has(m.id));
}
