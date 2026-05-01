/**
 * Navigation Store
 * Handles app navigation, menu state, breadcrumbs, and history
 */

import { writable, derived, get, type Readable } from 'svelte/store';
import type { MenuItem, BreadcrumbItem } from '@samavāya/core';

// ============================================================================
// TYPES
// ============================================================================

export interface Module {
  id: string;
  name: string;
  displayName: string;
  icon: string;
  path: string;
  description?: string;
  order: number;
  visible: boolean;
  permissions?: string[];
  children?: ModuleItem[];
}

export interface ModuleItem {
  id: string;
  label: string;
  path: string;
  icon?: string;
  badge?: string | number;
  permissions?: string[];
  children?: ModuleItem[];
}

export interface NavigationState {
  currentModule: string | null;
  currentPath: string;
  previousPath: string | null;
  modules: Module[];
  menuItems: MenuItem[];
  breadcrumbs: BreadcrumbItem[];
  isNavigating: boolean;
}

export interface SidebarState {
  isCollapsed: boolean;
  isHovered: boolean;
  expandedGroups: Set<string>;
  activeItemId: string | null;
  pinnedItems: string[];
  recentItems: string[];
}

export interface HistoryEntry {
  path: string;
  title: string;
  timestamp: Date;
  module?: string;
}

export interface HistoryState {
  entries: HistoryEntry[];
  currentIndex: number;
  maxEntries: number;
}

// ============================================================================
// INITIAL STATE
// ============================================================================

const initialNavigationState: NavigationState = {
  currentModule: null,
  currentPath: '/',
  previousPath: null,
  modules: [],
  menuItems: [],
  breadcrumbs: [],
  isNavigating: false,
};

const initialSidebarState: SidebarState = {
  isCollapsed: false,
  isHovered: false,
  expandedGroups: new Set(),
  activeItemId: null,
  pinnedItems: [],
  recentItems: [],
};

const initialHistoryState: HistoryState = {
  entries: [],
  currentIndex: -1,
  maxEntries: 50,
};

// ============================================================================
// NAVIGATION STORE
// ============================================================================

function createNavigationStore() {
  const store = writable<NavigationState>(initialNavigationState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const currentModule: Readable<string | null> = derived(store, ($s) => $s.currentModule);
  const currentPath: Readable<string> = derived(store, ($s) => $s.currentPath);
  const modules: Readable<Module[]> = derived(store, ($s) => $s.modules);
  const menuItems: Readable<MenuItem[]> = derived(store, ($s) => $s.menuItems);
  const breadcrumbs: Readable<BreadcrumbItem[]> = derived(store, ($s) => $s.breadcrumbs);
  const isNavigating: Readable<boolean> = derived(store, ($s) => $s.isNavigating);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function setModules(modules: Module[]): void {
    update((s) => ({
      ...s,
      modules: modules.sort((a, b) => a.order - b.order),
    }));
  }

  function setCurrentModule(moduleId: string | null): void {
    update((s) => ({ ...s, currentModule: moduleId }));
  }

  function setCurrentPath(path: string): void {
    update((s) => ({
      ...s,
      previousPath: s.currentPath,
      currentPath: path,
    }));
  }

  function setMenuItems(items: MenuItem[]): void {
    update((s) => ({ ...s, menuItems: items }));
  }

  function setBreadcrumbs(items: BreadcrumbItem[]): void {
    update((s) => ({ ...s, breadcrumbs: items }));
  }

  function addBreadcrumb(item: BreadcrumbItem): void {
    update((s) => ({
      ...s,
      breadcrumbs: [...s.breadcrumbs, item],
    }));
  }

  function setNavigating(isNavigating: boolean): void {
    update((s) => ({ ...s, isNavigating }));
  }

  function getModuleByPath(path: string): Module | undefined {
    const state = get(store);
    return state.modules.find((m) => path.startsWith(m.path));
  }

  function getVisibleModules(permissions: string[] = []): Module[] {
    const state = get(store);
    return state.modules.filter((m) => {
      if (!m.visible) return false;
      if (!m.permissions || m.permissions.length === 0) return true;
      return m.permissions.some((p) => permissions.includes(p));
    });
  }

  function reset(): void {
    store.set(initialNavigationState);
  }

  // ============================================================================
  // RETURN
  // ============================================================================

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
    reset,
  };
}

// ============================================================================
// SIDEBAR STORE
// ============================================================================

function createSidebarStore() {
  const store = writable<SidebarState>(initialSidebarState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const isCollapsed: Readable<boolean> = derived(store, ($s) => $s.isCollapsed);
  const isHovered: Readable<boolean> = derived(store, ($s) => $s.isHovered);
  const expandedGroups: Readable<Set<string>> = derived(store, ($s) => $s.expandedGroups);
  const activeItemId: Readable<string | null> = derived(store, ($s) => $s.activeItemId);
  const pinnedItems: Readable<string[]> = derived(store, ($s) => $s.pinnedItems);
  const recentItems: Readable<string[]> = derived(store, ($s) => $s.recentItems);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function setCollapsed(collapsed: boolean): void {
    update((s) => ({ ...s, isCollapsed: collapsed }));
    localStorage.setItem('sidebar_collapsed', String(collapsed));
  }

  function toggleCollapsed(): void {
    update((s) => {
      const collapsed = !s.isCollapsed;
      localStorage.setItem('sidebar_collapsed', String(collapsed));
      return { ...s, isCollapsed: collapsed };
    });
  }

  function setHovered(hovered: boolean): void {
    update((s) => ({ ...s, isHovered: hovered }));
  }

  function expandGroup(groupId: string): void {
    update((s) => {
      const expandedGroups = new Set(s.expandedGroups);
      expandedGroups.add(groupId);
      return { ...s, expandedGroups };
    });
  }

  function collapseGroup(groupId: string): void {
    update((s) => {
      const expandedGroups = new Set(s.expandedGroups);
      expandedGroups.delete(groupId);
      return { ...s, expandedGroups };
    });
  }

  function toggleGroup(groupId: string): void {
    const state = get(store);
    if (state.expandedGroups.has(groupId)) {
      collapseGroup(groupId);
    } else {
      expandGroup(groupId);
    }
  }

  function setActiveItem(itemId: string | null): void {
    update((s) => ({ ...s, activeItemId: itemId }));
  }

  function pinItem(itemId: string): void {
    update((s) => {
      if (s.pinnedItems.includes(itemId)) return s;
      const pinnedItems = [...s.pinnedItems, itemId];
      localStorage.setItem('sidebar_pinned', JSON.stringify(pinnedItems));
      return { ...s, pinnedItems };
    });
  }

  function unpinItem(itemId: string): void {
    update((s) => {
      const pinnedItems = s.pinnedItems.filter((id) => id !== itemId);
      localStorage.setItem('sidebar_pinned', JSON.stringify(pinnedItems));
      return { ...s, pinnedItems };
    });
  }

  function addRecentItem(itemId: string): void {
    update((s) => {
      const recentItems = [itemId, ...s.recentItems.filter((id) => id !== itemId)].slice(0, 10);
      localStorage.setItem('sidebar_recent', JSON.stringify(recentItems));
      return { ...s, recentItems };
    });
  }

  function clearRecentItems(): void {
    update((s) => ({ ...s, recentItems: [] }));
    localStorage.removeItem('sidebar_recent');
  }

  function loadState(): void {
    const collapsed = localStorage.getItem('sidebar_collapsed');
    const pinned = localStorage.getItem('sidebar_pinned');
    const recent = localStorage.getItem('sidebar_recent');

    update((s) => ({
      ...s,
      isCollapsed: collapsed === 'true',
      pinnedItems: pinned ? JSON.parse(pinned) : [],
      recentItems: recent ? JSON.parse(recent) : [],
    }));
  }

  function reset(): void {
    localStorage.removeItem('sidebar_collapsed');
    localStorage.removeItem('sidebar_pinned');
    localStorage.removeItem('sidebar_recent');
    store.set(initialSidebarState);
  }

  // ============================================================================
  // RETURN
  // ============================================================================

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
    reset,
  };
}

// ============================================================================
// HISTORY STORE
// ============================================================================

function createHistoryStore() {
  const store = writable<HistoryState>(initialHistoryState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const entries: Readable<HistoryEntry[]> = derived(store, ($s) => $s.entries);
  const currentEntry: Readable<HistoryEntry | null> = derived(
    store,
    ($s) => $s.entries[$s.currentIndex] ?? null
  );
  const canGoBack: Readable<boolean> = derived(store, ($s) => $s.currentIndex > 0);
  const canGoForward: Readable<boolean> = derived(
    store,
    ($s) => $s.currentIndex < $s.entries.length - 1
  );

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function push(entry: Omit<HistoryEntry, 'timestamp'>): void {
    update((s) => {
      // Remove forward entries if we're not at the end
      const entries = s.entries.slice(0, s.currentIndex + 1);

      // Add new entry
      const newEntry: HistoryEntry = {
        ...entry,
        timestamp: new Date(),
      };
      entries.push(newEntry);

      // Limit entries
      while (entries.length > s.maxEntries) {
        entries.shift();
      }

      return {
        ...s,
        entries,
        currentIndex: entries.length - 1,
      };
    });
  }

  function goBack(): HistoryEntry | null {
    const state = get(store);
    if (state.currentIndex <= 0) return null;

    update((s) => ({ ...s, currentIndex: s.currentIndex - 1 }));
    return get(store).entries[get(store).currentIndex] ?? null;
  }

  function goForward(): HistoryEntry | null {
    const state = get(store);
    if (state.currentIndex >= state.entries.length - 1) return null;

    update((s) => ({ ...s, currentIndex: s.currentIndex + 1 }));
    return get(store).entries[get(store).currentIndex] ?? null;
  }

  function goTo(index: number): HistoryEntry | null {
    const state = get(store);
    if (index < 0 || index >= state.entries.length) return null;

    update((s) => ({ ...s, currentIndex: index }));
    return state.entries[index] ?? null;
  }

  function clear(): void {
    store.set(initialHistoryState);
  }

  function setMaxEntries(max: number): void {
    update((s) => {
      const entries = s.entries.slice(-max);
      return {
        ...s,
        maxEntries: max,
        entries,
        currentIndex: Math.min(s.currentIndex, entries.length - 1),
      };
    });
  }

  // ============================================================================
  // RETURN
  // ============================================================================

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
    setMaxEntries,
  };
}

// ============================================================================
// EXPORT
// ============================================================================

export const navigationStore = createNavigationStore();
export const sidebarStore = createSidebarStore();
export const historyStore = createHistoryStore();
