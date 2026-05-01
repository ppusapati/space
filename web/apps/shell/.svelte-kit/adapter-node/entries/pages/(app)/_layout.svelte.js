import { s as store_get, b as attr_class, e as ensure_array_like, a as attr, u as unsubscribe_stores } from "../../../chunks/index2.js";
import "../../../chunks/ChartTypePicker.svelte_svelte_type_style_lang.js";
import "clsx";
import "echarts";
import { w as writable, d as derived, g as get } from "../../../chunks/index.js";
import { a as authStore } from "../../../chunks/auth.store.js";
import { t as themeStore } from "../../../chunks/theme.store.js";
import { e as escape_html } from "../../../chunks/escaping.js";
import { p as page } from "../../../chunks/stores.js";
const initialNavigationState = {
  currentModule: null,
  currentPath: "/",
  previousPath: null,
  modules: [],
  menuItems: [],
  breadcrumbs: [],
  isNavigating: false
};
const initialSidebarState = {
  isCollapsed: false,
  isHovered: false,
  expandedGroups: /* @__PURE__ */ new Set(),
  activeItemId: null,
  pinnedItems: [],
  recentItems: []
};
const initialHistoryState = {
  entries: [],
  currentIndex: -1,
  maxEntries: 50
};
function createNavigationStore() {
  const store = writable(initialNavigationState);
  const { subscribe, update } = store;
  const currentModule = derived(store, ($s) => $s.currentModule);
  const currentPath = derived(store, ($s) => $s.currentPath);
  const modules = derived(store, ($s) => $s.modules);
  const menuItems = derived(store, ($s) => $s.menuItems);
  const breadcrumbs = derived(store, ($s) => $s.breadcrumbs);
  const isNavigating = derived(store, ($s) => $s.isNavigating);
  function setModules(modules2) {
    update((s) => ({
      ...s,
      modules: modules2.sort((a, b) => a.order - b.order)
    }));
  }
  function setCurrentModule(moduleId) {
    update((s) => ({ ...s, currentModule: moduleId }));
  }
  function setCurrentPath(path) {
    update((s) => ({
      ...s,
      previousPath: s.currentPath,
      currentPath: path
    }));
  }
  function setMenuItems(items) {
    update((s) => ({ ...s, menuItems: items }));
  }
  function setBreadcrumbs(items) {
    update((s) => ({ ...s, breadcrumbs: items }));
  }
  function addBreadcrumb(item) {
    update((s) => ({
      ...s,
      breadcrumbs: [...s.breadcrumbs, item]
    }));
  }
  function setNavigating(isNavigating2) {
    update((s) => ({ ...s, isNavigating: isNavigating2 }));
  }
  function getModuleByPath(path) {
    const state = get(store);
    return state.modules.find((m) => path.startsWith(m.path));
  }
  function getVisibleModules(permissions = []) {
    const state = get(store);
    return state.modules.filter((m) => {
      if (!m.visible) return false;
      if (!m.permissions || m.permissions.length === 0) return true;
      return m.permissions.some((p) => permissions.includes(p));
    });
  }
  function reset() {
    store.set(initialNavigationState);
  }
  return {
    subscribe,
    // Derived stores
    currentModule,
    currentPath,
    modules,
    menuItems,
    breadcrumbs,
    isNavigating,
    // Actions
    setModules,
    setCurrentModule,
    setCurrentPath,
    setMenuItems,
    setBreadcrumbs,
    addBreadcrumb,
    setNavigating,
    getModuleByPath,
    getVisibleModules,
    reset
  };
}
function createSidebarStore() {
  const store = writable(initialSidebarState);
  const { subscribe, update } = store;
  const isCollapsed = derived(store, ($s) => $s.isCollapsed);
  const isHovered = derived(store, ($s) => $s.isHovered);
  const expandedGroups = derived(store, ($s) => $s.expandedGroups);
  const activeItemId = derived(store, ($s) => $s.activeItemId);
  const pinnedItems = derived(store, ($s) => $s.pinnedItems);
  const recentItems = derived(store, ($s) => $s.recentItems);
  function setCollapsed(collapsed) {
    update((s) => ({ ...s, isCollapsed: collapsed }));
    localStorage.setItem("sidebar_collapsed", String(collapsed));
  }
  function toggleCollapsed() {
    update((s) => {
      const collapsed = !s.isCollapsed;
      localStorage.setItem("sidebar_collapsed", String(collapsed));
      return { ...s, isCollapsed: collapsed };
    });
  }
  function setHovered(hovered) {
    update((s) => ({ ...s, isHovered: hovered }));
  }
  function expandGroup(groupId) {
    update((s) => {
      const expandedGroups2 = new Set(s.expandedGroups);
      expandedGroups2.add(groupId);
      return { ...s, expandedGroups: expandedGroups2 };
    });
  }
  function collapseGroup(groupId) {
    update((s) => {
      const expandedGroups2 = new Set(s.expandedGroups);
      expandedGroups2.delete(groupId);
      return { ...s, expandedGroups: expandedGroups2 };
    });
  }
  function toggleGroup(groupId) {
    const state = get(store);
    if (state.expandedGroups.has(groupId)) {
      collapseGroup(groupId);
    } else {
      expandGroup(groupId);
    }
  }
  function setActiveItem(itemId) {
    update((s) => ({ ...s, activeItemId: itemId }));
  }
  function pinItem(itemId) {
    update((s) => {
      if (s.pinnedItems.includes(itemId)) return s;
      const pinnedItems2 = [...s.pinnedItems, itemId];
      localStorage.setItem("sidebar_pinned", JSON.stringify(pinnedItems2));
      return { ...s, pinnedItems: pinnedItems2 };
    });
  }
  function unpinItem(itemId) {
    update((s) => {
      const pinnedItems2 = s.pinnedItems.filter((id) => id !== itemId);
      localStorage.setItem("sidebar_pinned", JSON.stringify(pinnedItems2));
      return { ...s, pinnedItems: pinnedItems2 };
    });
  }
  function addRecentItem(itemId) {
    update((s) => {
      const recentItems2 = [itemId, ...s.recentItems.filter((id) => id !== itemId)].slice(0, 10);
      localStorage.setItem("sidebar_recent", JSON.stringify(recentItems2));
      return { ...s, recentItems: recentItems2 };
    });
  }
  function clearRecentItems() {
    update((s) => ({ ...s, recentItems: [] }));
    localStorage.removeItem("sidebar_recent");
  }
  function loadState() {
    const collapsed = localStorage.getItem("sidebar_collapsed");
    const pinned = localStorage.getItem("sidebar_pinned");
    const recent = localStorage.getItem("sidebar_recent");
    update((s) => ({
      ...s,
      isCollapsed: collapsed === "true",
      pinnedItems: pinned ? JSON.parse(pinned) : [],
      recentItems: recent ? JSON.parse(recent) : []
    }));
  }
  function reset() {
    localStorage.removeItem("sidebar_collapsed");
    localStorage.removeItem("sidebar_pinned");
    localStorage.removeItem("sidebar_recent");
    store.set(initialSidebarState);
  }
  return {
    subscribe,
    // Derived stores
    isCollapsed,
    isHovered,
    expandedGroups,
    activeItemId,
    pinnedItems,
    recentItems,
    // Actions
    setCollapsed,
    toggleCollapsed,
    setHovered,
    expandGroup,
    collapseGroup,
    toggleGroup,
    setActiveItem,
    pinItem,
    unpinItem,
    addRecentItem,
    clearRecentItems,
    loadState,
    reset
  };
}
function createHistoryStore() {
  const store = writable(initialHistoryState);
  const { subscribe, update } = store;
  const entries = derived(store, ($s) => $s.entries);
  const currentEntry = derived(
    store,
    ($s) => $s.entries[$s.currentIndex] ?? null
  );
  const canGoBack = derived(store, ($s) => $s.currentIndex > 0);
  const canGoForward = derived(
    store,
    ($s) => $s.currentIndex < $s.entries.length - 1
  );
  function push(entry) {
    update((s) => {
      const entries2 = s.entries.slice(0, s.currentIndex + 1);
      const newEntry = {
        ...entry,
        timestamp: /* @__PURE__ */ new Date()
      };
      entries2.push(newEntry);
      while (entries2.length > s.maxEntries) {
        entries2.shift();
      }
      return {
        ...s,
        entries: entries2,
        currentIndex: entries2.length - 1
      };
    });
  }
  function goBack() {
    const state = get(store);
    if (state.currentIndex <= 0) return null;
    update((s) => ({ ...s, currentIndex: s.currentIndex - 1 }));
    return get(store).entries[get(store).currentIndex] ?? null;
  }
  function goForward() {
    const state = get(store);
    if (state.currentIndex >= state.entries.length - 1) return null;
    update((s) => ({ ...s, currentIndex: s.currentIndex + 1 }));
    return get(store).entries[get(store).currentIndex] ?? null;
  }
  function goTo(index) {
    const state = get(store);
    if (index < 0 || index >= state.entries.length) return null;
    update((s) => ({ ...s, currentIndex: index }));
    return state.entries[index] ?? null;
  }
  function clear() {
    store.set(initialHistoryState);
  }
  function setMaxEntries(max) {
    update((s) => {
      const entries2 = s.entries.slice(-max);
      return {
        ...s,
        maxEntries: max,
        entries: entries2,
        currentIndex: Math.min(s.currentIndex, entries2.length - 1)
      };
    });
  }
  return {
    subscribe,
    // Derived stores
    entries,
    currentEntry,
    canGoBack,
    canGoForward,
    // Actions
    push,
    goBack,
    goForward,
    goTo,
    clear,
    setMaxEntries
  };
}
createNavigationStore();
const sidebarStore = createSidebarStore();
createHistoryStore();
const __vite_import_meta_env__ = {};
const FORM_SERVICE_BASE = "/platform.formservice.api.v1.FormService";
function getBaseUrl() {
  if (typeof import.meta !== "undefined" && __vite_import_meta_env__?.VITE_API_URL) {
    return void 0;
  }
  return "http://localhost:8130";
}
async function rpcCall(method, request) {
  const url = `${getBaseUrl()}${FORM_SERVICE_BASE}/${method}`;
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(request)
  });
  if (!response.ok) {
    throw new Error(`FormService.${method}: ${response.status} ${response.statusText}`);
  }
  return response.json();
}
function createModuleStore() {
  let state = $state({
    modules: [],
    selectedModuleId: null,
    forms: [],
    isLoadingModules: false,
    isLoadingForms: false,
    moduleError: null,
    formError: null,
    isApiDriven: false
  });
  async function loadModules() {
    state.isLoadingModules = true;
    state.moduleError = null;
    try {
      const response = await rpcCall("ListModules", { context: {} });
      state.modules = response.modules ?? [];
      state.isApiDriven = state.modules.length > 0;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load modules";
      state.moduleError = message;
      state.isApiDriven = false;
      console.warn("[moduleStore] API unavailable, UI should fall back to static registry:", message);
    } finally {
      state.isLoadingModules = false;
    }
  }
  async function selectModule(moduleId) {
    if (state.selectedModuleId === moduleId && state.forms.length > 0) {
      return;
    }
    state.selectedModuleId = moduleId;
    state.forms = [];
    state.isLoadingForms = true;
    state.formError = null;
    try {
      const response = await rpcCall("ListForms", { context: {}, moduleId });
      state.forms = response.forms ?? [];
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load forms";
      state.formError = message;
      console.warn(`[moduleStore] Failed to load forms for ${moduleId}:`, message);
    } finally {
      state.isLoadingForms = false;
    }
  }
  function clearSelection() {
    state.selectedModuleId = null;
    state.forms = [];
    state.formError = null;
  }
  async function refresh() {
    await loadModules();
    if (state.selectedModuleId) {
      const current = state.selectedModuleId;
      state.selectedModuleId = null;
      await selectModule(current);
    }
  }
  return {
    /** Read the full state (reactive via $state) */
    get state() {
      return state;
    },
    /** Reactive getters for individual state properties */
    get modules() {
      return state.modules;
    },
    get selectedModuleId() {
      return state.selectedModuleId;
    },
    get forms() {
      return state.forms;
    },
    get isLoadingModules() {
      return state.isLoadingModules;
    },
    get isLoadingForms() {
      return state.isLoadingForms;
    },
    get moduleError() {
      return state.moduleError;
    },
    get formError() {
      return state.formError;
    },
    get isApiDriven() {
      return state.isApiDriven;
    },
    get isLoading() {
      return state.isLoadingModules || state.isLoadingForms;
    },
    /** Actions */
    loadModules,
    selectModule,
    clearSelection,
    refresh
  };
}
const moduleStore = createModuleStore();
const MODULE_REGISTRY = [
  {
    id: "dashboard",
    label: "Dashboard",
    path: "/",
    icon: "M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z",
    order: 0
  },
  {
    id: "identity",
    label: "Identity",
    path: "/identity",
    icon: "M12 2a5 5 0 015 5v1A5 5 0 017 8V7a5 5 0 015-5zM20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2",
    order: 1,
    sections: [
      {
        title: "Access Management",
        items: [
          { label: "Users", path: "/identity/users" },
          { label: "Roles", path: "/identity/roles" },
          { label: "Sessions", path: "/identity/sessions" }
        ]
      }
    ]
  },
  {
    id: "masters",
    label: "Masters",
    path: "/masters",
    icon: "M20 7H4a2 2 0 00-2 2v10a2 2 0 002 2h16a2 2 0 002-2V9a2 2 0 00-2-2zM16 21V5a2 2 0 00-2-2h-4a2 2 0 00-2 2v16",
    order: 2,
    sections: [
      {
        title: "Master Data",
        items: [
          { label: "Items", path: "/masters/items" },
          { label: "Parties", path: "/masters/parties" },
          { label: "Locations", path: "/masters/locations" },
          { label: "Chart of Accounts", path: "/masters/chart-of-accounts" },
          { label: "UOM", path: "/masters/uom" },
          { label: "Tax Codes", path: "/masters/tax-codes" }
        ]
      }
    ]
  },
  {
    id: "finance",
    label: "Finance",
    path: "/finance",
    icon: "M12 2v20M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6",
    order: 3,
    sections: [
      {
        title: "General Ledger",
        items: [
          { label: "Chart of Accounts", path: "/finance/gl/accounts" },
          { label: "Journal Entries", path: "/finance/gl/journal-entries" },
          { label: "Fiscal Periods", path: "/finance/gl/fiscal-periods" }
        ]
      },
      {
        title: "Accounts Payable",
        items: [
          { label: "Bills", path: "/finance/ap/bills" },
          { label: "Payments", path: "/finance/ap/payments" },
          { label: "AP Aging", path: "/finance/ap/aging" }
        ]
      },
      {
        title: "Accounts Receivable",
        items: [
          { label: "Invoices", path: "/finance/ar/invoices" },
          { label: "Receipts", path: "/finance/ar/receipts" },
          { label: "AR Aging", path: "/finance/ar/aging" }
        ]
      },
      {
        title: "Reports",
        items: [
          { label: "Trial Balance", path: "/finance/reports/trial-balance" },
          { label: "Balance Sheet", path: "/finance/reports/balance-sheet" },
          { label: "Income Statement", path: "/finance/reports/income-statement" },
          { label: "Cash Flow", path: "/finance/reports/cash-flow" }
        ]
      }
    ]
  },
  {
    id: "sales",
    label: "Sales",
    path: "/sales",
    icon: "M3 3v18h18M18.7 8l-5.1 5.2-2.8-2.7L7 14.3",
    order: 4,
    sections: [
      {
        title: "Sales",
        items: [
          { label: "CRM", path: "/sales/crm" },
          { label: "Sales Orders", path: "/sales/orders" },
          { label: "Invoices", path: "/sales/invoices" },
          { label: "Pricing", path: "/sales/pricing" },
          { label: "Dealers", path: "/sales/dealers" },
          { label: "Commissions", path: "/sales/commissions" }
        ]
      }
    ]
  },
  {
    id: "purchase",
    label: "Purchase",
    path: "/purchase",
    icon: "M9 21a1 1 0 100-2 1 1 0 000 2zM20 21a1 1 0 100-2 1 1 0 000 2zM1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6",
    order: 5,
    sections: [
      {
        title: "Procurement",
        items: [
          { label: "Requisitions", path: "/purchase/requisitions" },
          { label: "Purchase Orders", path: "/purchase/orders" },
          { label: "Invoices", path: "/purchase/invoices" },
          { label: "Vendors", path: "/purchase/vendors" }
        ]
      }
    ]
  },
  {
    id: "inventory",
    label: "Inventory",
    path: "/inventory",
    icon: "M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16zM3.27 6.96L12 12.01l8.73-5.05M12 22.08V12",
    order: 6,
    sections: [
      {
        title: "Inventory",
        items: [
          { label: "Stock", path: "/inventory/stock" },
          { label: "Lot & Serial", path: "/inventory/lot-serial" },
          { label: "Quality", path: "/inventory/quality" },
          { label: "Warehouse", path: "/inventory/warehouse" },
          { label: "Transfers", path: "/inventory/transfers" }
        ]
      }
    ]
  },
  {
    id: "hr",
    label: "HR",
    path: "/hr",
    icon: "M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75",
    order: 7,
    sections: [
      {
        title: "HR Management",
        items: [
          { label: "Employees", path: "/hr/employees" },
          { label: "Payroll", path: "/hr/payroll" },
          { label: "Leave", path: "/hr/leave" },
          { label: "Attendance", path: "/hr/attendance" },
          { label: "Recruitment", path: "/hr/recruitment" },
          { label: "Training", path: "/hr/training" }
        ]
      }
    ]
  },
  {
    id: "manufacturing",
    label: "Manufacturing",
    path: "/manufacturing",
    icon: "M2 20h20M6 20V4l6 4V4l6 4v12",
    order: 8,
    sections: [
      {
        title: "Production",
        items: [
          { label: "BOM", path: "/manufacturing/bom" },
          { label: "Production Orders", path: "/manufacturing/production" },
          { label: "Job Cards", path: "/manufacturing/job-cards" },
          { label: "Routing", path: "/manufacturing/routing" },
          { label: "Shop Floor", path: "/manufacturing/shop-floor" }
        ]
      }
    ]
  },
  {
    id: "projects",
    label: "Projects",
    path: "/projects",
    icon: "M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z",
    order: 9,
    sections: [
      {
        title: "Project Management",
        items: [
          { label: "Projects", path: "/projects/list" },
          { label: "Tasks", path: "/projects/tasks" },
          { label: "BOQ", path: "/projects/boq" },
          { label: "Timesheets", path: "/projects/timesheets" },
          { label: "Billing", path: "/projects/billing" }
        ]
      }
    ]
  },
  {
    id: "asset",
    label: "Assets",
    path: "/asset",
    icon: "M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z",
    order: 10,
    sections: [
      {
        title: "Asset Management",
        items: [
          { label: "Assets", path: "/asset/list" },
          { label: "Depreciation", path: "/asset/depreciation" },
          { label: "Maintenance", path: "/asset/maintenance" },
          { label: "Vehicles", path: "/asset/vehicles" }
        ]
      }
    ]
  },
  {
    id: "fulfillment",
    label: "Fulfillment",
    path: "/fulfillment",
    icon: "M16 16l3-8 3 8c-1.05.63-2.26 1-3.5 1s-2.45-.37-3.5-1zM2 16l3-8 3 8c-1.05.63-2.26 1-3.5 1S2.45 16.63 2 16zM7 21h10M12 3v18",
    order: 11,
    sections: [
      {
        title: "Fulfillment",
        items: [
          { label: "Shipping", path: "/fulfillment/shipping" },
          { label: "Returns", path: "/fulfillment/returns" }
        ]
      }
    ]
  },
  {
    id: "insights",
    label: "Insights",
    path: "/insights",
    icon: "M18 20V10M12 20V4M6 20v-6",
    order: 12,
    sections: [
      {
        title: "Analytics",
        items: [
          { label: "Dashboards", path: "/insights/dashboards" },
          { label: "Reports", path: "/insights/reports" },
          { label: "BI Analytics", path: "/insights/bi" }
        ]
      }
    ]
  },
  {
    id: "workflow",
    label: "Workflow",
    path: "/workflow",
    icon: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2M12 12h4M12 16h4M8 12h.01M8 16h.01",
    order: 13,
    sections: [
      {
        title: "Workflow Engine",
        items: [
          { label: "Approvals", path: "/workflow/approvals" },
          { label: "Form Builder", path: "/workflow/form-builder" },
          { label: "Escalations", path: "/workflow/escalations" },
          { label: "Workflows", path: "/workflow/workflows" }
        ]
      }
    ]
  },
  {
    id: "budget",
    label: "Budget",
    path: "/budget",
    icon: "M9 7h6M9 11h6M9 15h4M5 3h14a2 2 0 012 2v14a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2z",
    order: 14,
    sections: [
      {
        title: "Budget Management",
        items: [
          { label: "Budgets", path: "/budget/budgets" },
          { label: "Variance", path: "/budget/variance" },
          { label: "Capex", path: "/budget/capex" },
          { label: "Forecasting", path: "/budget/forecasting" }
        ]
      }
    ]
  },
  {
    id: "banking",
    label: "Banking",
    path: "/banking",
    icon: "M3 21h18M3 10h18M5 6l7-3 7 3M4 10v11M20 10v11M8 14v3M12 14v3M16 14v3",
    order: 15,
    sections: [
      {
        title: "Statutory & Banking",
        items: [
          { label: "Bank Accounts", path: "/banking/accounts" },
          { label: "GST", path: "/banking/gst" },
          { label: "TDS", path: "/banking/tds" },
          { label: "E-Invoice", path: "/banking/e-invoice" },
          { label: "E-Way Bill", path: "/banking/e-way-bill" }
        ]
      }
    ]
  },
  {
    id: "notifications",
    label: "Notifications",
    path: "/notifications",
    icon: "M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0",
    order: 16,
    sections: [
      {
        title: "Notification Center",
        items: [
          { label: "Notifications", path: "/notifications/list" },
          { label: "Templates", path: "/notifications/templates" }
        ]
      }
    ]
  },
  {
    id: "audit",
    label: "Audit",
    path: "/audit",
    icon: "M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z",
    order: 17,
    sections: [
      {
        title: "Audit & Compliance",
        items: [
          { label: "Audit Log", path: "/audit/log" },
          { label: "Changelog", path: "/audit/changelog" },
          { label: "Compliance", path: "/audit/compliance" },
          { label: "GDPR", path: "/audit/gdpr" },
          { label: "Retention", path: "/audit/retention" }
        ]
      }
    ]
  },
  {
    id: "platform",
    label: "Platform",
    path: "/platform",
    icon: "M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4",
    order: 18,
    sections: [
      {
        title: "Platform Services",
        items: [
          { label: "Scheduler", path: "/platform/scheduler" },
          { label: "File Storage", path: "/platform/file-storage" },
          { label: "Integrations", path: "/platform/integrations" },
          { label: "SLA", path: "/platform/sla" },
          { label: "Webhooks", path: "/platform/webhooks" },
          { label: "System Settings", path: "/platform/settings" }
        ]
      }
    ]
  },
  {
    id: "communication",
    label: "Communication",
    path: "/communication",
    icon: "M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z",
    order: 19,
    sections: [
      {
        title: "Communication",
        items: [
          { label: "Chat", path: "/communication/chat" },
          { label: "Currency", path: "/communication/currency" },
          { label: "Localization", path: "/communication/i18n" }
        ]
      }
    ]
  },
  {
    id: "data",
    label: "Data",
    path: "/data",
    icon: "M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4M4 12c0 2.21 3.582 4 8 4s8-1.79 8-4",
    order: 20,
    sections: [
      {
        title: "Data Management",
        items: [
          { label: "Import/Export", path: "/data/bridge" },
          { label: "Archive", path: "/data/archive" },
          { label: "Backup & DR", path: "/data/backup" }
        ]
      }
    ]
  },
  {
    id: "land",
    label: "Land Acquisition",
    path: "/land",
    icon: "M1 22l5-10 5 10M1 22h10M8 8l4-6 4 6M8 8h8M18 22l3-6 3 6M18 22h6M2 17h7M15 13h5",
    order: 21,
    sections: [
      {
        title: "Land Parcels",
        items: [
          { label: "Land Parcel", path: "/land/land-parcel" },
          { label: "GIS & Spatial", path: "/land/gis-spatial" },
          { label: "Field Operations", path: "/land/field-operations" }
        ]
      },
      {
        title: "Legal & Compliance",
        items: [
          { label: "Compliance", path: "/land/compliance" },
          { label: "Due Diligence", path: "/land/due-diligence" },
          { label: "Legal Cases", path: "/land/legal-case" }
        ]
      },
      {
        title: "Transactions",
        items: [
          { label: "Negotiation", path: "/land/negotiation" },
          { label: "Stakeholders", path: "/land/stakeholder" },
          { label: "Land Finance", path: "/land/land-finance" }
        ]
      },
      {
        title: "Analysis & Leasing",
        items: [
          { label: "Risk Scoring", path: "/land/risk-scoring" },
          { label: "Land Insights", path: "/land/land-insights" },
          { label: "Govt Lease", path: "/land/govt-lease" },
          { label: "Grid Interconnection", path: "/land/grid-interconnection" },
          { label: "Right of Way", path: "/land/right-of-way" },
          { label: "Renewable Energy Finance", path: "/land/renewable-energy-finance" }
        ]
      }
    ]
  }
];
function getEnabledModules(enabledIds) {
  if (!enabledIds) return MODULE_REGISTRY;
  const idSet = new Set(enabledIds);
  return MODULE_REGISTRY.filter((m) => m.id === "dashboard" || idSet.has(m.id));
}
function ErpShell($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let {
      activeModule,
      currentPath = "",
      enabledModules,
      pageTitle,
      brandName = "samavāya",
      children
    } = $$props;
    const isCollapsed = store_get($$store_subs ??= {}, "$sidebarStore", sidebarStore).isCollapsed;
    const userName = store_get($$store_subs ??= {}, "$authStore", authStore).user?.name || store_get($$store_subs ??= {}, "$authStore", authStore).user?.displayName || "User";
    const userInitial = userName.charAt(0).toUpperCase();
    const staticModules = getEnabledModules(enabledModules);
    const modules = () => {
      if (!moduleStore.isApiDriven) return staticModules;
      const apiMap = new Map(moduleStore.modules.map((m) => [m.moduleId, m]));
      return staticModules.map((mod) => {
        const apiMod = apiMap.get(mod.id);
        if (apiMod) {
          return { ...mod, label: apiMod.label || mod.label };
        }
        return mod;
      });
    };
    const currentModuleDef = MODULE_REGISTRY.find((m) => m.id === activeModule);
    const staticSections = currentModuleDef?.sections ?? [];
    const apiFormItems = () => {
      if (!moduleStore.isApiDriven || moduleStore.forms.length === 0) return null;
      if (moduleStore.selectedModuleId !== activeModule) return null;
      return {
        title: "Forms",
        items: moduleStore.forms.map((f) => ({ label: f.title, path: `/forms/${f.formId}` }))
      };
    };
    const sections = () => {
      const base = [...staticSections];
      const apiSection = apiFormItems();
      if (apiSection && apiSection.items.length > 0) {
        base.push(apiSection);
      }
      return base;
    };
    const derivedTitle = () => {
      if (pageTitle) return pageTitle;
      if (!currentPath) return currentModuleDef?.label ?? "Dashboard";
      for (const section of sections()) {
        for (const item of section.items) {
          if (currentPath.startsWith(item.path)) return item.label;
        }
      }
      if (currentPath.startsWith("/forms/")) {
        const formId = currentPath.split("/forms/")[1]?.split("/")[0];
        const form = moduleStore.forms.find((f) => f.formId === formId);
        if (form) return form.title;
      }
      return currentModuleDef?.label ?? "Dashboard";
    };
    function isModuleActive(mod) {
      if (mod.id === "dashboard") return currentPath === "/" || currentPath === "/dashboard";
      return mod.id === activeModule;
    }
    function isNavItemActive(path) {
      return currentPath.startsWith(path);
    }
    $$renderer2.push(`<div${attr_class("erp-layout svelte-1dknexg", void 0, { "sidebar-collapsed": isCollapsed })}><aside class="module-sidebar svelte-1dknexg"><div class="module-sidebar-header svelte-1dknexg"><span class="brand-icon svelte-1dknexg">S</span> `);
    if (!isCollapsed) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<span class="brand-name svelte-1dknexg">${escape_html(brandName)}</span>`);
    } else {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--></div> <nav class="module-nav svelte-1dknexg"><!--[-->`);
    const each_array = ensure_array_like(modules());
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let mod = each_array[$$index];
      $$renderer2.push(`<a${attr("href", mod.path)}${attr_class("module-item svelte-1dknexg", void 0, { "active": isModuleActive(mod) })}${attr("title", isCollapsed ? mod.label : void 0)}><svg class="module-icon svelte-1dknexg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path${attr("d", mod.icon)}></path></svg> `);
      if (!isCollapsed) {
        $$renderer2.push("<!--[-->");
        $$renderer2.push(`<span class="module-label svelte-1dknexg">${escape_html(mod.label)}</span>`);
      } else {
        $$renderer2.push("<!--[!-->");
      }
      $$renderer2.push(`<!--]--></a>`);
    }
    $$renderer2.push(`<!--]--></nav> <div class="module-sidebar-footer svelte-1dknexg"><button class="sidebar-action svelte-1dknexg" title="Toggle sidebar"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="svelte-1dknexg">`);
    if (isCollapsed) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<path d="M9 18l6-6-6-6"></path>`);
    } else {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`<path d="M15 18l-6-6 6-6"></path>`);
    }
    $$renderer2.push(`<!--]--></svg></button> <button class="sidebar-action svelte-1dknexg" title="Toggle theme"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="svelte-1dknexg">`);
    if (store_get($$store_subs ??= {}, "$themeStore", themeStore).mode === "dark") {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>`);
    } else {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`<path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"></path>`);
    }
    $$renderer2.push(`<!--]--></svg></button></div></aside> `);
    if (sections().length > 0 && !isCollapsed) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<aside class="sub-nav svelte-1dknexg"><div class="sub-nav-header svelte-1dknexg"><h2 class="sub-nav-title svelte-1dknexg">${escape_html(currentModuleDef?.label)}</h2></div> <nav class="sub-nav-content svelte-1dknexg"><!--[-->`);
      const each_array_1 = ensure_array_like(sections());
      for (let $$index_2 = 0, $$length = each_array_1.length; $$index_2 < $$length; $$index_2++) {
        let section = each_array_1[$$index_2];
        $$renderer2.push(`<div class="sub-nav-section svelte-1dknexg"><h3 class="sub-nav-section-title svelte-1dknexg">${escape_html(section.title)}</h3> <ul class="sub-nav-list svelte-1dknexg"><!--[-->`);
        const each_array_2 = ensure_array_like(section.items);
        for (let $$index_1 = 0, $$length2 = each_array_2.length; $$index_1 < $$length2; $$index_1++) {
          let item = each_array_2[$$index_1];
          $$renderer2.push(`<li><a${attr("href", item.path)}${attr_class("sub-nav-link svelte-1dknexg", void 0, { "active": isNavItemActive(item.path) })}>${escape_html(item.label)}</a></li>`);
        }
        $$renderer2.push(`<!--]--></ul></div>`);
      }
      $$renderer2.push(`<!--]--></nav></aside>`);
    } else {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--> <div class="main-wrapper svelte-1dknexg"><header class="app-header svelte-1dknexg"><div class="header-left"><h1 class="page-title svelte-1dknexg">${escape_html(derivedTitle())}</h1></div> <div class="header-right svelte-1dknexg"><div class="user-menu svelte-1dknexg"><span class="user-name svelte-1dknexg">${escape_html(userName)}</span> <div class="user-avatar svelte-1dknexg">${escape_html(userInitial)}</div></div></div></header> <main class="main-content svelte-1dknexg">`);
    children($$renderer2);
    $$renderer2.push(`<!----></main></div></div>`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function _layout($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let { children } = $$props;
    const MODULE_IDS = [
      "identity",
      "masters",
      "finance",
      "sales",
      "purchase",
      "inventory",
      "hr",
      "manufacturing",
      "projects",
      "asset",
      "fulfillment",
      "insights",
      "workflow",
      "budget",
      "banking",
      "notifications",
      "audit",
      "platform",
      "communication",
      "data",
      "land"
    ];
    const activeModule = () => {
      const path = store_get($$store_subs ??= {}, "$page", page).url.pathname;
      const segment = path.split("/")[1] ?? "";
      return MODULE_IDS.includes(segment) ? segment : "dashboard";
    };
    ErpShell($$renderer2, {
      activeModule: activeModule(),
      currentPath: store_get($$store_subs ??= {}, "$page", page).url.pathname,
      children: ($$renderer3) => {
        children($$renderer3);
        $$renderer3.push(`<!---->`);
      }
    });
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
export {
  _layout as default
};
