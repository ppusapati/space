/**
 * useSelection Composable
 * Creates a reactive selection state for managing selected items
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type { SelectionState } from '../types/index.js';

// ============================================================================
// TYPES
// ============================================================================

export interface UseSelectionOptions<T> {
  initialSelection?: T[];
  getKey?: (item: T) => string | number;
  multiSelect?: boolean;
  maxSelection?: number;
  onChange?: (selection: SelectionState<T>) => void;
}

export interface UseSelectionReturn<T> {
  // State
  selection: Readable<SelectionState<T>>;
  selectedItems: Readable<T[]>;
  selectedKeys: Readable<Set<string | number>>;
  isAllSelected: Readable<boolean>;
  isIndeterminate: Readable<boolean>;
  selectedCount: Readable<number>;

  // Methods
  select: (item: T) => void;
  deselect: (item: T) => void;
  toggle: (item: T) => void;
  selectAll: (items: T[]) => void;
  deselectAll: () => void;
  selectRange: (items: T[], startIndex: number, endIndex: number) => void;
  setSelection: (items: T[]) => void;
  isSelected: (item: T) => boolean;
  getSelectedKeys: () => Set<string | number>;
  getSelectedItems: () => T[];
  reset: () => void;
}

// ============================================================================
// IMPLEMENTATION
// ============================================================================

export function useSelection<T>(options: UseSelectionOptions<T> = {}): UseSelectionReturn<T> {
  const {
    initialSelection = [],
    getKey = (item: T) => {
      if (typeof item === 'object' && item !== null && 'id' in item) {
        return (item as { id: string | number }).id;
      }
      return JSON.stringify(item);
    },
    multiSelect = true,
    maxSelection,
    onChange,
  } = options;

  // ============================================================================
  // STORES
  // ============================================================================

  const selectedItemsStore = writable<T[]>(initialSelection);
  const selectedKeysStore = writable<Set<string | number>>(
    new Set(initialSelection.map(getKey))
  );
  const allItemsStore = writable<T[]>([]);

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const selectedItems = { subscribe: selectedItemsStore.subscribe } as Readable<T[]>;
  const selectedKeys = { subscribe: selectedKeysStore.subscribe } as Readable<Set<string | number>>;

  const selectedCount = derived(selectedItemsStore, ($items) => $items.length);

  const isAllSelected = derived(
    [selectedItemsStore, allItemsStore],
    ([$selected, $all]) => $all.length > 0 && $selected.length === $all.length
  );

  const isIndeterminate = derived(
    [selectedItemsStore, allItemsStore],
    ([$selected, $all]) => $selected.length > 0 && $selected.length < $all.length
  );

  const selection = derived(
    [selectedItemsStore, selectedKeysStore, isAllSelected, isIndeterminate],
    ([$items, $keys, $isAll, $isIndeterminate]) => ({
      selectedItems: $items,
      selectedKeys: $keys,
      isAllSelected: $isAll,
      isIndeterminate: $isIndeterminate,
    })
  );

  // ============================================================================
  // NOTIFY CHANGES
  // ============================================================================

  let isInitial = true;
  selection.subscribe(($selection) => {
    if (!isInitial) {
      onChange?.($selection);
    }
    isInitial = false;
  });

  // ============================================================================
  // METHODS
  // ============================================================================

  function select(item: T): void {
    const key = getKey(item);
    const $keys = get(selectedKeysStore);

    // Already selected
    if ($keys.has(key)) return;

    // Check max selection
    if (maxSelection && $keys.size >= maxSelection) return;

    if (multiSelect) {
      selectedKeysStore.update(($k) => {
        const newKeys = new Set($k);
        newKeys.add(key);
        return newKeys;
      });
      selectedItemsStore.update(($items) => [...$items, item]);
    } else {
      // Single select - replace
      selectedKeysStore.set(new Set([key]));
      selectedItemsStore.set([item]);
    }
  }

  function deselect(item: T): void {
    const key = getKey(item);

    selectedKeysStore.update(($k) => {
      const newKeys = new Set($k);
      newKeys.delete(key);
      return newKeys;
    });

    selectedItemsStore.update(($items) => $items.filter((i) => getKey(i) !== key));
  }

  function toggle(item: T): void {
    const key = getKey(item);
    const $keys = get(selectedKeysStore);

    if ($keys.has(key)) {
      deselect(item);
    } else {
      select(item);
    }
  }

  function selectAll(items: T[]): void {
    if (!multiSelect) return;

    allItemsStore.set(items);

    let itemsToSelect = items;
    if (maxSelection && items.length > maxSelection) {
      itemsToSelect = items.slice(0, maxSelection);
    }

    selectedItemsStore.set([...itemsToSelect]);
    selectedKeysStore.set(new Set(itemsToSelect.map(getKey)));
  }

  function deselectAll(): void {
    selectedItemsStore.set([]);
    selectedKeysStore.set(new Set());
  }

  function selectRange(items: T[], startIndex: number, endIndex: number): void {
    if (!multiSelect) return;

    const start = Math.min(startIndex, endIndex);
    const end = Math.max(startIndex, endIndex);
    const rangeItems = items.slice(start, end + 1);

    let newItems: T[];
    if (maxSelection) {
      const $items = get(selectedItemsStore);
      const available = maxSelection - $items.length;
      newItems = [...$items, ...rangeItems.slice(0, available)];
    } else {
      const $items = get(selectedItemsStore);
      const $keys = get(selectedKeysStore);
      newItems = [...$items];

      for (const item of rangeItems) {
        const key = getKey(item);
        if (!$keys.has(key)) {
          newItems.push(item);
        }
      }
    }

    selectedItemsStore.set(newItems);
    selectedKeysStore.set(new Set(newItems.map(getKey)));
  }

  function setSelection(items: T[]): void {
    let itemsToSet = items;
    if (maxSelection && items.length > maxSelection) {
      itemsToSet = items.slice(0, maxSelection);
    }
    if (!multiSelect && itemsToSet.length > 1) {
      const firstItem = itemsToSet[0];
      if (firstItem !== undefined) {
        itemsToSet = [firstItem];
      }
    }

    selectedItemsStore.set([...itemsToSet]);
    selectedKeysStore.set(new Set(itemsToSet.map(getKey)));
  }

  function isSelected(item: T): boolean {
    const key = getKey(item);
    return get(selectedKeysStore).has(key);
  }

  function getSelectedKeys(): Set<string | number> {
    return new Set(get(selectedKeysStore));
  }

  function getSelectedItems(): T[] {
    return [...get(selectedItemsStore)];
  }

  function reset(): void {
    selectedItemsStore.set([...initialSelection]);
    selectedKeysStore.set(new Set(initialSelection.map(getKey)));
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    // State
    selection,
    selectedItems,
    selectedKeys,
    isAllSelected,
    isIndeterminate,
    selectedCount,

    // Methods
    select,
    deselect,
    toggle,
    selectAll,
    deselectAll,
    selectRange,
    setSelection,
    isSelected,
    getSelectedKeys,
    getSelectedItems,
    reset,
  };
}

// ============================================================================
// ROW SELECTION HELPER
// ============================================================================

export interface UseRowSelectionOptions<T> {
  getRowKey: (row: T) => string | number;
  multiSelect?: boolean;
  onChange?: (selectedRows: T[]) => void;
}

export function useRowSelection<T>(options: UseRowSelectionOptions<T>) {
  const { getRowKey, multiSelect = true, onChange } = options;

  return useSelection<T>({
    getKey: getRowKey,
    multiSelect,
    onChange: (selection) => onChange?.(selection.selectedItems),
  });
}

// ============================================================================
// CHECKBOX GROUP HELPER
// ============================================================================

export interface UseCheckboxGroupOptions<T> {
  options: T[];
  initialSelected?: T[];
  getValue?: (option: T) => string | number;
  maxSelected?: number;
  onChange?: (selected: T[]) => void;
}

export function useCheckboxGroup<T>(options: UseCheckboxGroupOptions<T>) {
  const { options: allOptions, initialSelected = [], getValue, maxSelected, onChange } = options;

  const selection = useSelection<T>({
    initialSelection: initialSelected,
    getKey: getValue ?? ((item) => {
      if (typeof item === 'object' && item !== null && 'value' in item) {
        return (item as { value: string | number }).value;
      }
      return JSON.stringify(item);
    }),
    multiSelect: true,
    maxSelection: maxSelected,
    onChange: (state) => onChange?.(state.selectedItems),
  });

  return {
    ...selection,
    options: allOptions,
    toggleAll: () => {
      if (get(selection.isAllSelected)) {
        selection.deselectAll();
      } else {
        selection.selectAll(allOptions);
      }
    },
  };
}
