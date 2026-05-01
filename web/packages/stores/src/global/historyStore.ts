/**
 * History Store (Undo/Redo)
 *
 * Generic undo/redo functionality for any state:
 * - Configurable history depth
 * - Batch operations
 * - State snapshots
 * - Action descriptions for UI
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export interface HistoryEntry<T> {
  state: T;
  timestamp: Date;
  description?: string;
}

export interface HistoryConfig {
  /** Maximum history entries (default: 50) */
  maxHistory: number;
  /** Whether to track redo stack */
  enableRedo: boolean;
}

export interface HistoryState<T> {
  current: T;
  undoStack: HistoryEntry<T>[];
  redoStack: HistoryEntry<T>[];
  isBatching: boolean;
  batchDescription?: string;
}

// ============================================================================
// Factory
// ============================================================================

const DEFAULT_CONFIG: HistoryConfig = {
  maxHistory: 50,
  enableRedo: true,
};

/**
 * Create a history-enabled store for any state type
 */
export function createHistoryStore<T>(initialState: T, config?: Partial<HistoryConfig>) {
  const cfg: HistoryConfig = { ...DEFAULT_CONFIG, ...config };

  const { subscribe, set, update } = writable<HistoryState<T>>({
    current: initialState,
    undoStack: [],
    redoStack: [],
    isBatching: false,
  });

  let batchedState: T | null = null;

  function pushToUndo(entry: HistoryEntry<T>) {
    update(s => {
      const undoStack = [...s.undoStack, entry];
      if (undoStack.length > cfg.maxHistory) undoStack.shift();
      return { ...s, undoStack, redoStack: cfg.enableRedo ? [] : s.redoStack };
    });
  }

  return {
    subscribe,

    /** Get current state */
    get current(): T {
      return get({ subscribe }).current;
    },

    /** Push new state with optional description */
    push(newState: T, description?: string) {
      const state = get({ subscribe });

      if (state.isBatching) {
        batchedState = newState;
        return;
      }

      pushToUndo({ state: state.current, timestamp: new Date(), description });
      update(s => ({ ...s, current: newState }));
    },

    /** Update state using a function */
    modify(fn: (current: T) => T, description?: string) {
      const state = get({ subscribe });
      const newState = fn(state.current);
      this.push(newState, description);
    },

    /** Undo last change */
    undo(): boolean {
      const state = get({ subscribe });
      if (state.undoStack.length === 0) return false;

      const lastEntry = state.undoStack[state.undoStack.length - 1]!;
      const undoStack = state.undoStack.slice(0, -1);
      const redoStack = cfg.enableRedo
        ? [...state.redoStack, { state: state.current, timestamp: new Date() }]
        : state.redoStack;

      set({ ...state, current: lastEntry.state, undoStack, redoStack, isBatching: false });
      return true;
    },

    /** Redo last undone change */
    redo(): boolean {
      if (!cfg.enableRedo) return false;
      const state = get({ subscribe });
      if (state.redoStack.length === 0) return false;

      const lastEntry = state.redoStack[state.redoStack.length - 1]!;
      const redoStack = state.redoStack.slice(0, -1);
      const undoStack = [...state.undoStack, { state: state.current, timestamp: new Date() }];

      set({ ...state, current: lastEntry.state, undoStack, redoStack, isBatching: false });
      return true;
    },

    /** Start batch operation (multiple changes as one undo step) */
    startBatch(description?: string) {
      const state = get({ subscribe });
      batchedState = state.current;
      update(s => ({ ...s, isBatching: true, batchDescription: description }));
    },

    /** End batch operation */
    endBatch() {
      const state = get({ subscribe });
      if (!state.isBatching) return;

      if (batchedState !== null) {
        const originalState = state.undoStack.length > 0
          ? state.undoStack[state.undoStack.length - 1]!.state
          : state.current;

        // Only push if state actually changed
        if (JSON.stringify(originalState) !== JSON.stringify(batchedState)) {
          pushToUndo({ state: originalState, timestamp: new Date(), description: state.batchDescription });
          update(s => ({ ...s, current: batchedState!, isBatching: false, batchDescription: undefined }));
        } else {
          update(s => ({ ...s, isBatching: false, batchDescription: undefined }));
        }
      } else {
        update(s => ({ ...s, isBatching: false, batchDescription: undefined }));
      }
      batchedState = null;
    },

    /** Cancel batch operation */
    cancelBatch() {
      batchedState = null;
      update(s => ({ ...s, isBatching: false, batchDescription: undefined }));
    },

    /** Clear all history */
    clearHistory() {
      update(s => ({ ...s, undoStack: [], redoStack: [] }));
    },

    /** Reset to initial state and clear history */
    reset() {
      set({ current: initialState, undoStack: [], redoStack: [], isBatching: false });
    },

    /** Go to specific point in history */
    goTo(index: number): boolean {
      const state = get({ subscribe });
      if (index < 0 || index >= state.undoStack.length) return false;

      const targetEntry = state.undoStack[index]!;
      const undoStack = state.undoStack.slice(0, index);
      const redoEntries = state.undoStack.slice(index + 1).map(e => e);
      const redoStack = cfg.enableRedo ? [...redoEntries.reverse(), { state: state.current, timestamp: new Date() }] : [];

      set({ ...state, current: targetEntry.state, undoStack, redoStack: redoStack as HistoryEntry<T>[], isBatching: false });
      return true;
    },

    /** Check if can undo */
    get canUndo(): boolean {
      return get({ subscribe }).undoStack.length > 0;
    },

    /** Check if can redo */
    get canRedo(): boolean {
      return cfg.enableRedo && get({ subscribe }).redoStack.length > 0;
    },

    /** Get undo stack descriptions */
    get undoDescriptions(): string[] {
      return get({ subscribe }).undoStack.map(e => e.description || 'Change').reverse();
    },

    /** Get redo stack descriptions */
    get redoDescriptions(): string[] {
      return get({ subscribe }).redoStack.map(e => e.description || 'Change').reverse();
    },
  };
}

// ============================================================================
// Derived Helpers
// ============================================================================

/** Create derived stores for a history store */
export function deriveHistoryState<T>(historyStore: ReturnType<typeof createHistoryStore<T>>) {
  return {
    current: derived(historyStore, $h => $h.current),
    canUndo: derived(historyStore, $h => $h.undoStack.length > 0),
    canRedo: derived(historyStore, $h => $h.redoStack.length > 0),
    undoCount: derived(historyStore, $h => $h.undoStack.length),
    redoCount: derived(historyStore, $h => $h.redoStack.length),
    isBatching: derived(historyStore, $h => $h.isBatching),
  };
}

// ============================================================================
// useHistory Composable (Svelte 5 Runes)
// ============================================================================

/**
 * Svelte 5 composable for undo/redo functionality
 * Usage:
 * ```svelte
 * <script>
 *   const history = useHistory({ name: '', age: 0 });
 *
 *   function updateName(name: string) {
 *     history.push({ ...history.current, name }, 'Update name');
 *   }
 * </script>
 *
 * <button onclick={() => history.undo()} disabled={!history.canUndo}>Undo</button>
 * <button onclick={() => history.redo()} disabled={!history.canRedo}>Redo</button>
 * ```
 */
export function useHistory<T>(initialState: T, config?: Partial<HistoryConfig>) {
  const store = createHistoryStore(initialState, config);

  // Return an object that exposes the store's methods and reactive getters
  return {
    subscribe: store.subscribe,
    push: store.push.bind(store),
    modify: store.modify.bind(store),
    undo: store.undo.bind(store),
    redo: store.redo.bind(store),
    startBatch: store.startBatch.bind(store),
    endBatch: store.endBatch.bind(store),
    cancelBatch: store.cancelBatch.bind(store),
    clearHistory: store.clearHistory.bind(store),
    reset: store.reset.bind(store),
    goTo: store.goTo.bind(store),
    get current() { return store.current; },
    get canUndo() { return store.canUndo; },
    get canRedo() { return store.canRedo; },
    get undoDescriptions() { return store.undoDescriptions; },
    get redoDescriptions() { return store.redoDescriptions; },
  };
}
