/**
 * UI Store
 * Handles UI state like modals, drawers, loading, and command palette
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const initialModalState = {
    stack: [],
};
const initialDrawerState = {
    current: null,
    stack: [],
};
const initialLoadingState = {
    global: false,
    items: new Map(),
};
const initialCommandPaletteState = {
    isOpen: false,
    query: '',
    groups: [],
    selectedIndex: 0,
    mode: 'search',
};
// ============================================================================
// MODAL STORE
// ============================================================================
function createModalStore() {
    const store = writable(initialModalState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const stack = derived(store, ($s) => $s.stack);
    const current = derived(store, ($s) => $s.stack[$s.stack.length - 1] ?? null);
    const isAnyOpen = derived(store, ($s) => $s.stack.length > 0);
    const count = derived(store, ($s) => $s.stack.length);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function open(modal) {
        return new Promise((resolve, reject) => {
            const id = `modal-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
            const item = {
                ...modal,
                id,
                resolve: resolve,
                reject,
            };
            update((s) => ({
                ...s,
                stack: [...s.stack, item],
            }));
        });
    }
    function close(id, result) {
        update((s) => {
            if (!id) {
                // Close top modal
                const modal = s.stack[s.stack.length - 1];
                modal?.resolve?.(result);
                return { ...s, stack: s.stack.slice(0, -1) };
            }
            const modal = s.stack.find((m) => m.id === id);
            modal?.resolve?.(result);
            return { ...s, stack: s.stack.filter((m) => m.id !== id) };
        });
    }
    function closeAll() {
        const state = get(store);
        state.stack.forEach((modal) => modal.resolve?.(undefined));
        update((s) => ({ ...s, stack: [] }));
    }
    function reject(id, reason) {
        update((s) => {
            if (!id) {
                const modal = s.stack[s.stack.length - 1];
                modal?.reject?.(reason);
                return { ...s, stack: s.stack.slice(0, -1) };
            }
            const modal = s.stack.find((m) => m.id === id);
            modal?.reject?.(reason);
            return { ...s, stack: s.stack.filter((m) => m.id !== id) };
        });
    }
    function updateProps(id, props) {
        update((s) => ({
            ...s,
            stack: s.stack.map((m) => m.id === id ? { ...m, props: { ...m.props, ...props } } : m),
        }));
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        stack,
        current,
        isAnyOpen,
        count,
        // Actions
        open,
        close,
        closeAll,
        reject,
        updateProps,
    };
}
// ============================================================================
// DRAWER STORE
// ============================================================================
function createDrawerStore() {
    const store = writable(initialDrawerState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const current = derived(store, ($s) => $s.current);
    const isOpen = derived(store, ($s) => $s.current !== null);
    const stack = derived(store, ($s) => $s.stack);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function open(drawer) {
        const id = `drawer-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        const item = { ...drawer, id };
        update((s) => ({
            ...s,
            current: item,
            stack: [...s.stack, item],
        }));
        return id;
    }
    function close(id) {
        update((s) => {
            if (!id || s.current?.id === id) {
                // Close current
                const newStack = s.stack.slice(0, -1);
                return {
                    ...s,
                    current: newStack[newStack.length - 1] ?? null,
                    stack: newStack,
                };
            }
            const newStack = s.stack.filter((d) => d.id !== id);
            return {
                ...s,
                current: s.current?.id === id ? (newStack[newStack.length - 1] ?? null) : s.current,
                stack: newStack,
            };
        });
    }
    function closeAll() {
        update((s) => ({ ...s, current: null, stack: [] }));
    }
    function updateProps(id, props) {
        update((s) => ({
            ...s,
            current: s.current?.id === id ? { ...s.current, props: { ...s.current.props, ...props } } : s.current,
            stack: s.stack.map((d) => d.id === id ? { ...d, props: { ...d.props, ...props } } : d),
        }));
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        current,
        isOpen,
        stack,
        // Actions
        open,
        close,
        closeAll,
        updateProps,
    };
}
// ============================================================================
// LOADING STORE
// ============================================================================
function createLoadingStore() {
    const store = writable(initialLoadingState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const isGlobalLoading = derived(store, ($s) => $s.global);
    const globalMessage = derived(store, ($s) => $s.globalMessage);
    const activeCount = derived(store, ($s) => $s.items.size);
    const isAnyLoading = derived(store, ($s) => $s.global || $s.items.size > 0);
    const progress = derived(store, ($s) => $s.progress);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function setGlobal(loading, message) {
        update((s) => ({ ...s, global: loading, globalMessage: message }));
    }
    function start(id, options) {
        update((s) => {
            const items = new Map(s.items);
            items.set(id, { id, ...options });
            return { ...s, items };
        });
    }
    function stop(id) {
        update((s) => {
            const items = new Map(s.items);
            items.delete(id);
            return { ...s, items };
        });
    }
    function stopAll() {
        update((s) => ({ ...s, items: new Map(), global: false }));
    }
    function updateProgress(id, progress, message) {
        update((s) => {
            const items = new Map(s.items);
            const item = items.get(id);
            if (item) {
                items.set(id, { ...item, progress, message });
            }
            return { ...s, items };
        });
    }
    function setProgress(value, message) {
        update((s) => ({ ...s, progress: { value, message } }));
    }
    function clearProgress() {
        update((s) => ({ ...s, progress: undefined }));
    }
    function isLoading(id) {
        return get(store).items.has(id);
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        isGlobalLoading,
        globalMessage,
        activeCount,
        isAnyLoading,
        progress,
        // Actions
        setGlobal,
        start,
        stop,
        stopAll,
        updateProgress,
        setProgress,
        clearProgress,
        isLoading,
    };
}
// ============================================================================
// COMMAND PALETTE STORE
// ============================================================================
function createCommandPaletteStore() {
    const store = writable(initialCommandPaletteState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const isOpen = derived(store, ($s) => $s.isOpen);
    const query = derived(store, ($s) => $s.query);
    const groups = derived(store, ($s) => $s.groups);
    const selectedIndex = derived(store, ($s) => $s.selectedIndex);
    const filteredGroups = derived(store, ($s) => {
        if (!$s.query)
            return $s.groups;
        const q = $s.query.toLowerCase();
        return $s.groups
            .map((group) => ({
            ...group,
            items: group.items.filter((item) => item.label.toLowerCase().includes(q) ||
                item.description?.toLowerCase().includes(q) ||
                item.keywords?.some((k) => k.toLowerCase().includes(q))),
        }))
            .filter((group) => group.items.length > 0);
    });
    const flatItems = derived(filteredGroups, ($groups) => $groups.flatMap((g) => g.items));
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function open(mode = 'search') {
        update((s) => ({ ...s, isOpen: true, mode, query: '', selectedIndex: 0 }));
    }
    function close() {
        update((s) => ({ ...s, isOpen: false, query: '', selectedIndex: 0 }));
    }
    function toggle() {
        const state = get(store);
        if (state.isOpen) {
            close();
        }
        else {
            open();
        }
    }
    function setQuery(query) {
        update((s) => ({ ...s, query, selectedIndex: 0 }));
    }
    function registerGroup(group) {
        update((s) => {
            const existingIndex = s.groups.findIndex((g) => g.id === group.id);
            if (existingIndex >= 0) {
                const groups = [...s.groups];
                groups[existingIndex] = group;
                return { ...s, groups };
            }
            return {
                ...s,
                groups: [...s.groups, group].sort((a, b) => (a.priority ?? 0) - (b.priority ?? 0)),
            };
        });
    }
    function unregisterGroup(groupId) {
        update((s) => ({
            ...s,
            groups: s.groups.filter((g) => g.id !== groupId),
        }));
    }
    function registerCommand(groupId, command) {
        update((s) => ({
            ...s,
            groups: s.groups.map((g) => g.id === groupId ? { ...g, items: [...g.items, command] } : g),
        }));
    }
    function unregisterCommand(commandId) {
        update((s) => ({
            ...s,
            groups: s.groups.map((g) => ({
                ...g,
                items: g.items.filter((c) => c.id !== commandId),
            })),
        }));
    }
    function selectNext() {
        const items = get(flatItems);
        update((s) => ({
            ...s,
            selectedIndex: Math.min(s.selectedIndex + 1, items.length - 1),
        }));
    }
    function selectPrev() {
        update((s) => ({
            ...s,
            selectedIndex: Math.max(s.selectedIndex - 1, 0),
        }));
    }
    function selectIndex(index) {
        const items = get(flatItems);
        update((s) => ({
            ...s,
            selectedIndex: Math.max(0, Math.min(index, items.length - 1)),
        }));
    }
    async function executeSelected() {
        const state = get(store);
        const items = get(flatItems);
        const item = items[state.selectedIndex];
        if (item && !item.disabled) {
            close();
            await item.action();
        }
    }
    async function execute(commandId) {
        const items = get(flatItems);
        const item = items.find((i) => i.id === commandId);
        if (item && !item.disabled) {
            close();
            await item.action();
        }
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        isOpen,
        query,
        groups,
        selectedIndex,
        filteredGroups,
        flatItems,
        // Actions
        open,
        close,
        toggle,
        setQuery,
        registerGroup,
        unregisterGroup,
        registerCommand,
        unregisterCommand,
        selectNext,
        selectPrev,
        selectIndex,
        executeSelected,
        execute,
    };
}
// ============================================================================
// EXPORT
// ============================================================================
export const modalStore = createModalStore();
export const drawerStore = createDrawerStore();
export const loadingStore = createLoadingStore();
export const commandPaletteStore = createCommandPaletteStore();
