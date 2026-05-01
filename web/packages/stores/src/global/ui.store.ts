/**
 * UI Store
 * Handles UI state like modals, drawers, loading, and command palette
 */

import { writable, derived, get, type Readable } from 'svelte/store';
import type { Component, Snippet } from 'svelte';

// ============================================================================
// TYPES
// ============================================================================

export interface ModalItem {
  id: string;
  component?: Component;
  snippet?: Snippet;
  props?: Record<string, unknown>;
  config: {
    title?: string;
    size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
    closable?: boolean;
    closeOnEscape?: boolean;
    closeOnOverlay?: boolean;
    preventClose?: boolean;
  };
  resolve?: (value: unknown) => void;
  reject?: (reason?: unknown) => void;
}

export interface DrawerItem {
  id: string;
  component?: Component;
  snippet?: Snippet;
  props?: Record<string, unknown>;
  config: {
    title?: string;
    position?: 'left' | 'right' | 'top' | 'bottom';
    size?: 'sm' | 'md' | 'lg' | 'xl';
    closable?: boolean;
    closeOnEscape?: boolean;
    closeOnOverlay?: boolean;
    overlay?: boolean;
  };
}

export interface LoadingItem {
  id: string;
  message?: string;
  progress?: number;
  cancelable?: boolean;
  onCancel?: () => void;
}

export interface CommandItem {
  id: string;
  label: string;
  description?: string;
  icon?: string;
  shortcut?: string;
  group?: string;
  keywords?: string[];
  disabled?: boolean;
  action: () => void | Promise<void>;
}

export interface CommandGroup {
  id: string;
  label: string;
  priority?: number;
  items: CommandItem[];
}

export interface ModalState {
  stack: ModalItem[];
}

export interface DrawerState {
  current: DrawerItem | null;
  stack: DrawerItem[];
}

export interface LoadingState {
  global: boolean;
  globalMessage?: string;
  items: Map<string, LoadingItem>;
  progress?: {
    value: number;
    message?: string;
  };
}

export interface CommandPaletteState {
  isOpen: boolean;
  query: string;
  groups: CommandGroup[];
  selectedIndex: number;
  mode: 'search' | 'command';
}

// ============================================================================
// INITIAL STATE
// ============================================================================

const initialModalState: ModalState = {
  stack: [],
};

const initialDrawerState: DrawerState = {
  current: null,
  stack: [],
};

const initialLoadingState: LoadingState = {
  global: false,
  items: new Map(),
};

const initialCommandPaletteState: CommandPaletteState = {
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
  const store = writable<ModalState>(initialModalState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const stack: Readable<ModalItem[]> = derived(store, ($s) => $s.stack);
  const current: Readable<ModalItem | null> = derived(
    store,
    ($s) => $s.stack[$s.stack.length - 1] ?? null
  );
  const isAnyOpen: Readable<boolean> = derived(store, ($s) => $s.stack.length > 0);
  const count: Readable<number> = derived(store, ($s) => $s.stack.length);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function open<T = unknown>(modal: Omit<ModalItem, 'id' | 'resolve' | 'reject'>): Promise<T> {
    return new Promise((resolve, reject) => {
      const id = `modal-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      const item: ModalItem = {
        ...modal,
        id,
        resolve: resolve as (value: unknown) => void,
        reject,
      };

      update((s) => ({
        ...s,
        stack: [...s.stack, item],
      }));
    });
  }

  function close(id?: string, result?: unknown): void {
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

  function closeAll(): void {
    const state = get(store);
    state.stack.forEach((modal) => modal.resolve?.(undefined));
    update((s) => ({ ...s, stack: [] }));
  }

  function reject(id?: string, reason?: unknown): void {
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

  function updateProps(id: string, props: Record<string, unknown>): void {
    update((s) => ({
      ...s,
      stack: s.stack.map((m) =>
        m.id === id ? { ...m, props: { ...m.props, ...props } } : m
      ),
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
  const store = writable<DrawerState>(initialDrawerState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const current: Readable<DrawerItem | null> = derived(store, ($s) => $s.current);
  const isOpen: Readable<boolean> = derived(store, ($s) => $s.current !== null);
  const stack: Readable<DrawerItem[]> = derived(store, ($s) => $s.stack);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function open(drawer: Omit<DrawerItem, 'id'>): string {
    const id = `drawer-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const item: DrawerItem = { ...drawer, id };

    update((s) => ({
      ...s,
      current: item,
      stack: [...s.stack, item],
    }));

    return id;
  }

  function close(id?: string): void {
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

  function closeAll(): void {
    update((s) => ({ ...s, current: null, stack: [] }));
  }

  function updateProps(id: string, props: Record<string, unknown>): void {
    update((s) => ({
      ...s,
      current: s.current?.id === id ? { ...s.current, props: { ...s.current.props, ...props } } : s.current,
      stack: s.stack.map((d) =>
        d.id === id ? { ...d, props: { ...d.props, ...props } } : d
      ),
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
  const store = writable<LoadingState>(initialLoadingState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const isGlobalLoading: Readable<boolean> = derived(store, ($s) => $s.global);
  const globalMessage: Readable<string | undefined> = derived(store, ($s) => $s.globalMessage);
  const activeCount: Readable<number> = derived(store, ($s) => $s.items.size);
  const isAnyLoading: Readable<boolean> = derived(store, ($s) => $s.global || $s.items.size > 0);
  const progress: Readable<{ value: number; message?: string } | undefined> = derived(
    store,
    ($s) => $s.progress
  );

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function setGlobal(loading: boolean, message?: string): void {
    update((s) => ({ ...s, global: loading, globalMessage: message }));
  }

  function start(id: string, options?: Omit<LoadingItem, 'id'>): void {
    update((s) => {
      const items = new Map(s.items);
      items.set(id, { id, ...options });
      return { ...s, items };
    });
  }

  function stop(id: string): void {
    update((s) => {
      const items = new Map(s.items);
      items.delete(id);
      return { ...s, items };
    });
  }

  function stopAll(): void {
    update((s) => ({ ...s, items: new Map(), global: false }));
  }

  function updateProgress(id: string, progress: number, message?: string): void {
    update((s) => {
      const items = new Map(s.items);
      const item = items.get(id);
      if (item) {
        items.set(id, { ...item, progress, message });
      }
      return { ...s, items };
    });
  }

  function setProgress(value: number, message?: string): void {
    update((s) => ({ ...s, progress: { value, message } }));
  }

  function clearProgress(): void {
    update((s) => ({ ...s, progress: undefined }));
  }

  function isLoading(id: string): boolean {
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
  const store = writable<CommandPaletteState>(initialCommandPaletteState);
  const { subscribe, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const isOpen: Readable<boolean> = derived(store, ($s) => $s.isOpen);
  const query: Readable<string> = derived(store, ($s) => $s.query);
  const groups: Readable<CommandGroup[]> = derived(store, ($s) => $s.groups);
  const selectedIndex: Readable<number> = derived(store, ($s) => $s.selectedIndex);

  const filteredGroups: Readable<CommandGroup[]> = derived(store, ($s) => {
    if (!$s.query) return $s.groups;

    const q = $s.query.toLowerCase();
    return $s.groups
      .map((group) => ({
        ...group,
        items: group.items.filter(
          (item) =>
            item.label.toLowerCase().includes(q) ||
            item.description?.toLowerCase().includes(q) ||
            item.keywords?.some((k) => k.toLowerCase().includes(q))
        ),
      }))
      .filter((group) => group.items.length > 0);
  });

  const flatItems: Readable<CommandItem[]> = derived(filteredGroups, ($groups) =>
    $groups.flatMap((g) => g.items)
  );

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function open(mode: 'search' | 'command' = 'search'): void {
    update((s) => ({ ...s, isOpen: true, mode, query: '', selectedIndex: 0 }));
  }

  function close(): void {
    update((s) => ({ ...s, isOpen: false, query: '', selectedIndex: 0 }));
  }

  function toggle(): void {
    const state = get(store);
    if (state.isOpen) {
      close();
    } else {
      open();
    }
  }

  function setQuery(query: string): void {
    update((s) => ({ ...s, query, selectedIndex: 0 }));
  }

  function registerGroup(group: CommandGroup): void {
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

  function unregisterGroup(groupId: string): void {
    update((s) => ({
      ...s,
      groups: s.groups.filter((g) => g.id !== groupId),
    }));
  }

  function registerCommand(groupId: string, command: CommandItem): void {
    update((s) => ({
      ...s,
      groups: s.groups.map((g) =>
        g.id === groupId ? { ...g, items: [...g.items, command] } : g
      ),
    }));
  }

  function unregisterCommand(commandId: string): void {
    update((s) => ({
      ...s,
      groups: s.groups.map((g) => ({
        ...g,
        items: g.items.filter((c) => c.id !== commandId),
      })),
    }));
  }

  function selectNext(): void {
    const items = get(flatItems);
    update((s) => ({
      ...s,
      selectedIndex: Math.min(s.selectedIndex + 1, items.length - 1),
    }));
  }

  function selectPrev(): void {
    update((s) => ({
      ...s,
      selectedIndex: Math.max(s.selectedIndex - 1, 0),
    }));
  }

  function selectIndex(index: number): void {
    const items = get(flatItems);
    update((s) => ({
      ...s,
      selectedIndex: Math.max(0, Math.min(index, items.length - 1)),
    }));
  }

  async function executeSelected(): Promise<void> {
    const state = get(store);
    const items = get(flatItems);
    const item = items[state.selectedIndex];

    if (item && !item.disabled) {
      close();
      await item.action();
    }
  }

  async function execute(commandId: string): Promise<void> {
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
