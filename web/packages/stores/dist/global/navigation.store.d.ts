/**
 * Navigation Store
 * Handles app navigation, menu state, breadcrumbs, and history
 */
import { type Readable } from 'svelte/store';
import type { MenuItem, BreadcrumbItem } from '@samavāya/core';
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
export declare const navigationStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<NavigationState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    currentModule: Readable<string | null>;
    currentPath: Readable<string>;
    modules: Readable<Module[]>;
    menuItems: Readable<MenuItem[]>;
    breadcrumbs: Readable<BreadcrumbItem[]>;
    isNavigating: Readable<boolean>;
    setModules: (modules: Module[]) => void;
    setCurrentModule: (moduleId: string | null) => void;
    setCurrentPath: (path: string) => void;
    setMenuItems: (items: MenuItem[]) => void;
    setBreadcrumbs: (items: BreadcrumbItem[]) => void;
    addBreadcrumb: (item: BreadcrumbItem) => void;
    setNavigating: (isNavigating: boolean) => void;
    getModuleByPath: (path: string) => Module | undefined;
    getVisibleModules: (permissions?: string[]) => Module[];
    reset: () => void;
};
export declare const sidebarStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<SidebarState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    isCollapsed: Readable<boolean>;
    isHovered: Readable<boolean>;
    expandedGroups: Readable<Set<string>>;
    activeItemId: Readable<string | null>;
    pinnedItems: Readable<string[]>;
    recentItems: Readable<string[]>;
    setCollapsed: (collapsed: boolean) => void;
    toggleCollapsed: () => void;
    setHovered: (hovered: boolean) => void;
    expandGroup: (groupId: string) => void;
    collapseGroup: (groupId: string) => void;
    toggleGroup: (groupId: string) => void;
    setActiveItem: (itemId: string | null) => void;
    pinItem: (itemId: string) => void;
    unpinItem: (itemId: string) => void;
    addRecentItem: (itemId: string) => void;
    clearRecentItems: () => void;
    loadState: () => void;
    reset: () => void;
};
export declare const historyStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<HistoryState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    entries: Readable<HistoryEntry[]>;
    currentEntry: Readable<HistoryEntry | null>;
    canGoBack: Readable<boolean>;
    canGoForward: Readable<boolean>;
    push: (entry: Omit<HistoryEntry, "timestamp">) => void;
    goBack: () => HistoryEntry | null;
    goForward: () => HistoryEntry | null;
    goTo: (index: number) => HistoryEntry | null;
    clear: () => void;
    setMaxEntries: (max: number) => void;
};
//# sourceMappingURL=navigation.store.d.ts.map