/**
 * Navigation Store
 * Handles app navigation, menu state, breadcrumbs, and history
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const initialNavigationState = {
    currentModule: null,
    currentPath: '/',
    previousPath: null,
    modules: [],
    menuItems: [],
    breadcrumbs: [],
    isNavigating: false,
};
const initialSidebarState = {
    isCollapsed: false,
    isHovered: false,
    expandedGroups: new Set(),
    activeItemId: null,
    pinnedItems: [],
    recentItems: [],
};
const initialHistoryState = {
    entries: [],
    currentIndex: -1,
    maxEntries: 50,
};
// ============================================================================
// NAVIGATION STORE
// ============================================================================
function createNavigationStore() {
    const store = writable(initialNavigationState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const currentModule = derived(store, ($s) => $s.currentModule);
    const currentPath = derived(store, ($s) => $s.currentPath);
    const modules = derived(store, ($s) => $s.modules);
    const menuItems = derived(store, ($s) => $s.menuItems);
    const breadcrumbs = derived(store, ($s) => $s.breadcrumbs);
    const isNavigating = derived(store, ($s) => $s.isNavigating);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function setModules(modules) {
        update((s) => ({
            ...s,
            modules: modules.sort((a, b) => a.order - b.order),
        }));
    }
    function setCurrentModule(moduleId) {
        update((s) => ({ ...s, currentModule: moduleId }));
    }
    function setCurrentPath(path) {
        update((s) => ({
            ...s,
            previousPath: s.currentPath,
            currentPath: path,
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
            breadcrumbs: [...s.breadcrumbs, item],
        }));
    }
    function setNavigating(isNavigating) {
        update((s) => ({ ...s, isNavigating }));
    }
    function getModuleByPath(path) {
        const state = get(store);
        return state.modules.find((m) => path.startsWith(m.path));
    }
    function getVisibleModules(permissions = []) {
        const state = get(store);
        return state.modules.filter((m) => {
            if (!m.visible)
                return false;
            if (!m.permissions || m.permissions.length === 0)
                return true;
            return m.permissions.some((p) => permissions.includes(p));
        });
    }
    function reset() {
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
    const store = writable(initialSidebarState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const isCollapsed = derived(store, ($s) => $s.isCollapsed);
    const isHovered = derived(store, ($s) => $s.isHovered);
    const expandedGroups = derived(store, ($s) => $s.expandedGroups);
    const activeItemId = derived(store, ($s) => $s.activeItemId);
    const pinnedItems = derived(store, ($s) => $s.pinnedItems);
    const recentItems = derived(store, ($s) => $s.recentItems);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function setCollapsed(collapsed) {
        update((s) => ({ ...s, isCollapsed: collapsed }));
        localStorage.setItem('sidebar_collapsed', String(collapsed));
    }
    function toggleCollapsed() {
        update((s) => {
            const collapsed = !s.isCollapsed;
            localStorage.setItem('sidebar_collapsed', String(collapsed));
            return { ...s, isCollapsed: collapsed };
        });
    }
    function setHovered(hovered) {
        update((s) => ({ ...s, isHovered: hovered }));
    }
    function expandGroup(groupId) {
        update((s) => {
            const expandedGroups = new Set(s.expandedGroups);
            expandedGroups.add(groupId);
            return { ...s, expandedGroups };
        });
    }
    function collapseGroup(groupId) {
        update((s) => {
            const expandedGroups = new Set(s.expandedGroups);
            expandedGroups.delete(groupId);
            return { ...s, expandedGroups };
        });
    }
    function toggleGroup(groupId) {
        const state = get(store);
        if (state.expandedGroups.has(groupId)) {
            collapseGroup(groupId);
        }
        else {
            expandGroup(groupId);
        }
    }
    function setActiveItem(itemId) {
        update((s) => ({ ...s, activeItemId: itemId }));
    }
    function pinItem(itemId) {
        update((s) => {
            if (s.pinnedItems.includes(itemId))
                return s;
            const pinnedItems = [...s.pinnedItems, itemId];
            localStorage.setItem('sidebar_pinned', JSON.stringify(pinnedItems));
            return { ...s, pinnedItems };
        });
    }
    function unpinItem(itemId) {
        update((s) => {
            const pinnedItems = s.pinnedItems.filter((id) => id !== itemId);
            localStorage.setItem('sidebar_pinned', JSON.stringify(pinnedItems));
            return { ...s, pinnedItems };
        });
    }
    function addRecentItem(itemId) {
        update((s) => {
            const recentItems = [itemId, ...s.recentItems.filter((id) => id !== itemId)].slice(0, 10);
            localStorage.setItem('sidebar_recent', JSON.stringify(recentItems));
            return { ...s, recentItems };
        });
    }
    function clearRecentItems() {
        update((s) => ({ ...s, recentItems: [] }));
        localStorage.removeItem('sidebar_recent');
    }
    function loadState() {
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
    function reset() {
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
    const store = writable(initialHistoryState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const entries = derived(store, ($s) => $s.entries);
    const currentEntry = derived(store, ($s) => $s.entries[$s.currentIndex] ?? null);
    const canGoBack = derived(store, ($s) => $s.currentIndex > 0);
    const canGoForward = derived(store, ($s) => $s.currentIndex < $s.entries.length - 1);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function push(entry) {
        update((s) => {
            // Remove forward entries if we're not at the end
            const entries = s.entries.slice(0, s.currentIndex + 1);
            // Add new entry
            const newEntry = {
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
    function goBack() {
        const state = get(store);
        if (state.currentIndex <= 0)
            return null;
        update((s) => ({ ...s, currentIndex: s.currentIndex - 1 }));
        return get(store).entries[get(store).currentIndex] ?? null;
    }
    function goForward() {
        const state = get(store);
        if (state.currentIndex >= state.entries.length - 1)
            return null;
        update((s) => ({ ...s, currentIndex: s.currentIndex + 1 }));
        return get(store).entries[get(store).currentIndex] ?? null;
    }
    function goTo(index) {
        const state = get(store);
        if (index < 0 || index >= state.entries.length)
            return null;
        update((s) => ({ ...s, currentIndex: index }));
        return state.entries[index] ?? null;
    }
    function clear() {
        store.set(initialHistoryState);
    }
    function setMaxEntries(max) {
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
