/**
 * useList Composable
 * Creates a reactive list state with filtering, sorting, pagination, and selection
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type {
  List,
  ListState,
  ListConfig,
  ListError,
  PaginationState,
  SortConfig,
  SortDirection,
  SelectionState,
  FilterValue,
  ExportFormat,
  ExportOptions,
} from '../types/index.js';

// ============================================================================
// TYPES
// ============================================================================

export interface UseListOptions<TItem, TFilters extends Record<string, unknown>> {
  config: ListConfig<TItem>;
  initialItems?: TItem[];
  initialFilters?: TFilters;
  initialSort?: SortConfig;
  initialPageSize?: number;
  fetchData?: (params: FetchParams<TFilters>) => Promise<FetchResult<TItem>>;
  onError?: (error: ListError) => void;
}

export interface FetchParams<TFilters> {
  filters: TFilters;
  activeFilters: FilterValue[];
  searchQuery: string;
  sort: SortConfig;
  pagination: { page: number; pageSize: number };
}

export interface FetchResult<TItem> {
  items: TItem[];
  total: number;
}

export interface UseListReturn<TItem, TFilters extends Record<string, unknown>> {
  // Stores
  items: Writable<TItem[]>;
  filteredItems: Readable<TItem[]>;
  displayItems: Readable<TItem[]>;
  state: Writable<ListState>;
  filters: Writable<TFilters>;
  activeFilters: Writable<FilterValue[]>;
  searchQuery: Writable<string>;
  sort: Writable<SortConfig>;
  pagination: Writable<PaginationState>;
  selection: Writable<SelectionState<TItem>>;

  // Methods
  load: () => Promise<void>;
  reload: () => Promise<void>;
  refresh: () => Promise<void>;

  // Filter methods
  setFilter: <K extends keyof TFilters>(key: K, value: TFilters[K]) => void;
  setFilters: (filters: Partial<TFilters>) => void;
  clearFilter: (key: keyof TFilters) => void;
  clearAllFilters: () => void;
  addActiveFilter: (filter: FilterValue) => void;
  removeActiveFilter: (field: string) => void;

  // Search methods
  search: (query: string) => void;
  clearSearch: () => void;

  // Sort methods
  setSort: (column: string, direction: SortDirection) => void;
  toggleSort: (column: string) => void;
  clearSort: () => void;

  // Pagination methods
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  nextPage: () => void;
  prevPage: () => void;
  goToFirst: () => void;
  goToLast: () => void;

  // Selection methods
  selectItem: (item: TItem) => void;
  deselectItem: (item: TItem) => void;
  toggleItem: (item: TItem) => void;
  selectAll: () => void;
  deselectAll: () => void;
  selectRange: (startIndex: number, endIndex: number) => void;
  isSelected: (item: TItem) => boolean;

  // Export
  exportData: (format: ExportFormat, options?: ExportOptions) => Promise<void>;

  // Utility
  getItemKey: (item: TItem) => string | number;
}

// ============================================================================
// IMPLEMENTATION
// ============================================================================

export function useList<TItem, TFilters extends Record<string, unknown> = Record<string, unknown>>(
  options: UseListOptions<TItem, TFilters>
): UseListReturn<TItem, TFilters> {
  const { config, initialItems = [], initialFilters, initialSort, initialPageSize = 10, fetchData, onError } = options;

  // ============================================================================
  // STORES
  // ============================================================================

  const items = writable<TItem[]>(initialItems);

  const state = writable<ListState>({
    isLoading: false,
    isRefreshing: false,
    isEmpty: initialItems.length === 0,
    hasError: false,
  });

  const filters = writable<TFilters>((initialFilters ?? {}) as TFilters);
  const activeFilters = writable<FilterValue[]>([]);
  const searchQuery = writable<string>('');

  const sort = writable<SortConfig>(
    initialSort ?? config.defaultSort ?? { column: null, direction: null }
  );

  const pagination = writable<PaginationState>({
    page: 1,
    pageSize: initialPageSize,
    total: initialItems.length,
    totalPages: Math.ceil(initialItems.length / initialPageSize),
    hasNext: initialItems.length > initialPageSize,
    hasPrevious: false,
  });

  const selection = writable<SelectionState<TItem>>({
    selectedItems: [],
    selectedKeys: new Set(),
    isAllSelected: false,
    isIndeterminate: false,
  });

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const filteredItems = derived(
    [items, filters, activeFilters, searchQuery, sort],
    ([$items, $filters, $activeFilters, $searchQuery, $sort]) => {
      let result = [...$items];

      // Apply search
      if ($searchQuery && config.searchable && config.searchFields) {
        const query = $searchQuery.toLowerCase();
        result = result.filter((item) =>
          config.searchFields!.some((field) => {
            const value = item[field];
            return value != null && String(value).toLowerCase().includes(query);
          })
        );
      }

      // Apply active filters
      for (const filter of $activeFilters) {
        result = result.filter((item) => {
          const value = (item as Record<string, unknown>)[filter.field];
          return applyFilterOperator(value, filter.operator, filter.value, filter.secondValue);
        });
      }

      // Apply sorting
      if ($sort.column && $sort.direction) {
        result.sort((a, b) => {
          const aVal = (a as Record<string, unknown>)[$sort.column!];
          const bVal = (b as Record<string, unknown>)[$sort.column!];

          let comparison = 0;
          if (aVal == null && bVal == null) comparison = 0;
          else if (aVal == null) comparison = 1;
          else if (bVal == null) comparison = -1;
          else if (typeof aVal === 'string' && typeof bVal === 'string') {
            comparison = aVal.localeCompare(bVal);
          } else if (typeof aVal === 'number' && typeof bVal === 'number') {
            comparison = aVal - bVal;
          } else if (aVal instanceof Date && bVal instanceof Date) {
            comparison = aVal.getTime() - bVal.getTime();
          } else {
            comparison = String(aVal).localeCompare(String(bVal));
          }

          return $sort.direction === 'desc' ? -comparison : comparison;
        });
      }

      return result;
    }
  );

  const displayItems = derived([filteredItems, pagination], ([$filteredItems, $pagination]) => {
    if (!config.paginated) return $filteredItems;

    const start = ($pagination.page - 1) * $pagination.pageSize;
    const end = start + $pagination.pageSize;
    return $filteredItems.slice(start, end);
  });

  // ============================================================================
  // UTILITY FUNCTIONS
  // ============================================================================

  function getItemKey(item: TItem): string | number {
    if (typeof config.itemKey === 'function') {
      return config.itemKey(item);
    }
    return item[config.itemKey] as string | number;
  }

  function updatePagination(total: number) {
    pagination.update(($p) => {
      const totalPages = Math.ceil(total / $p.pageSize);
      return {
        ...$p,
        total,
        totalPages,
        hasNext: $p.page < totalPages,
        hasPrevious: $p.page > 1,
      };
    });
  }

  function applyFilterOperator(
    value: unknown,
    operator: string,
    filterValue: unknown,
    secondValue?: unknown
  ): boolean {
    switch (operator) {
      case 'equals':
        return value === filterValue;
      case 'notEquals':
        return value !== filterValue;
      case 'contains':
        return String(value).toLowerCase().includes(String(filterValue).toLowerCase());
      case 'notContains':
        return !String(value).toLowerCase().includes(String(filterValue).toLowerCase());
      case 'startsWith':
        return String(value).toLowerCase().startsWith(String(filterValue).toLowerCase());
      case 'endsWith':
        return String(value).toLowerCase().endsWith(String(filterValue).toLowerCase());
      case 'gt':
        return Number(value) > Number(filterValue);
      case 'gte':
        return Number(value) >= Number(filterValue);
      case 'lt':
        return Number(value) < Number(filterValue);
      case 'lte':
        return Number(value) <= Number(filterValue);
      case 'between':
        return Number(value) >= Number(filterValue) && Number(value) <= Number(secondValue);
      case 'in':
        return Array.isArray(filterValue) && filterValue.includes(value);
      case 'notIn':
        return Array.isArray(filterValue) && !filterValue.includes(value);
      case 'isEmpty':
        return value == null || value === '';
      case 'isNotEmpty':
        return value != null && value !== '';
      case 'isNull':
        return value == null;
      case 'isNotNull':
        return value != null;
      default:
        return true;
    }
  }

  // ============================================================================
  // DATA LOADING
  // ============================================================================

  async function load(): Promise<void> {
    if (!fetchData) return;

    state.update(($s) => ({ ...$s, isLoading: true, hasError: false, error: undefined }));

    try {
      const $filters = get(filters);
      const $activeFilters = get(activeFilters);
      const $searchQuery = get(searchQuery);
      const $sort = get(sort);
      const $pagination = get(pagination);

      const result = await fetchData({
        filters: $filters,
        activeFilters: $activeFilters,
        searchQuery: $searchQuery,
        sort: $sort,
        pagination: { page: $pagination.page, pageSize: $pagination.pageSize },
      });

      items.set(result.items);
      updatePagination(result.total);

      state.update(($s) => ({
        ...$s,
        isLoading: false,
        isEmpty: result.items.length === 0,
        lastUpdated: new Date(),
      }));
    } catch (error) {
      const listError: ListError = {
        code: 'LOAD_ERROR',
        message: error instanceof Error ? error.message : 'Failed to load data',
        retryable: true,
      };

      state.update(($s) => ({ ...$s, isLoading: false, hasError: true, error: listError }));
      onError?.(listError);
    }
  }

  async function reload(): Promise<void> {
    pagination.update(($p) => ({ ...$p, page: 1 }));
    await load();
  }

  async function refresh(): Promise<void> {
    state.update(($s) => ({ ...$s, isRefreshing: true }));
    await load();
    state.update(($s) => ({ ...$s, isRefreshing: false }));
  }

  // ============================================================================
  // FILTER METHODS
  // ============================================================================

  function setFilter<K extends keyof TFilters>(key: K, value: TFilters[K]): void {
    filters.update(($f) => ({ ...$f, [key]: value }));
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  function setFilters(newFilters: Partial<TFilters>): void {
    filters.update(($f) => ({ ...$f, ...newFilters }));
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  function clearFilter(key: keyof TFilters): void {
    filters.update(($f) => {
      const newFilters = { ...$f };
      delete newFilters[key];
      return newFilters;
    });
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  function clearAllFilters(): void {
    filters.set({} as TFilters);
    activeFilters.set([]);
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  function addActiveFilter(filter: FilterValue): void {
    activeFilters.update(($f) => {
      const existing = $f.findIndex((f) => f.field === filter.field);
      if (existing >= 0) {
        const newFilters = [...$f];
        newFilters[existing] = filter;
        return newFilters;
      }
      return [...$f, filter];
    });
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  function removeActiveFilter(field: string): void {
    activeFilters.update(($f) => $f.filter((f) => f.field !== field));
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  // ============================================================================
  // SEARCH METHODS
  // ============================================================================

  function search(query: string): void {
    searchQuery.set(query);
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  function clearSearch(): void {
    searchQuery.set('');
    pagination.update(($p) => ({ ...$p, page: 1 }));
  }

  // ============================================================================
  // SORT METHODS
  // ============================================================================

  function setSort(column: string, direction: SortDirection): void {
    sort.set({ column, direction });
  }

  function toggleSort(column: string): void {
    sort.update(($s) => {
      if ($s.column !== column) {
        return { column, direction: 'asc' };
      }
      if ($s.direction === 'asc') {
        return { column, direction: 'desc' };
      }
      return { column: null, direction: null };
    });
  }

  function clearSort(): void {
    sort.set({ column: null, direction: null });
  }

  // ============================================================================
  // PAGINATION METHODS
  // ============================================================================

  function setPage(page: number): void {
    pagination.update(($p) => ({
      ...$p,
      page: Math.max(1, Math.min(page, $p.totalPages)),
      hasNext: page < $p.totalPages,
      hasPrevious: page > 1,
    }));
  }

  function setPageSize(size: number): void {
    pagination.update(($p) => {
      const totalPages = Math.ceil($p.total / size);
      const page = Math.min($p.page, totalPages);
      return {
        ...$p,
        pageSize: size,
        page,
        totalPages,
        hasNext: page < totalPages,
        hasPrevious: page > 1,
      };
    });
  }

  function nextPage(): void {
    pagination.update(($p) => {
      if (!$p.hasNext) return $p;
      const page = $p.page + 1;
      return {
        ...$p,
        page,
        hasNext: page < $p.totalPages,
        hasPrevious: true,
      };
    });
  }

  function prevPage(): void {
    pagination.update(($p) => {
      if (!$p.hasPrevious) return $p;
      const page = $p.page - 1;
      return {
        ...$p,
        page,
        hasNext: true,
        hasPrevious: page > 1,
      };
    });
  }

  function goToFirst(): void {
    setPage(1);
  }

  function goToLast(): void {
    const $p = get(pagination);
    setPage($p.totalPages);
  }

  // ============================================================================
  // SELECTION METHODS
  // ============================================================================

  function selectItem(item: TItem): void {
    const key = getItemKey(item);
    selection.update(($s) => {
      if ($s.selectedKeys.has(key)) return $s;

      const selectedKeys = new Set($s.selectedKeys);
      selectedKeys.add(key);

      const selectedItems = config.multiSelect ? [...$s.selectedItems, item] : [item];

      const $items = get(items);
      const isAllSelected = selectedKeys.size === $items.length;

      return {
        selectedItems,
        selectedKeys,
        isAllSelected,
        isIndeterminate: !isAllSelected && selectedKeys.size > 0,
      };
    });
  }

  function deselectItem(item: TItem): void {
    const key = getItemKey(item);
    selection.update(($s) => {
      if (!$s.selectedKeys.has(key)) return $s;

      const selectedKeys = new Set($s.selectedKeys);
      selectedKeys.delete(key);

      const selectedItems = $s.selectedItems.filter((i) => getItemKey(i) !== key);

      return {
        selectedItems,
        selectedKeys,
        isAllSelected: false,
        isIndeterminate: selectedKeys.size > 0,
      };
    });
  }

  function toggleItem(item: TItem): void {
    const key = getItemKey(item);
    const $s = get(selection);
    if ($s.selectedKeys.has(key)) {
      deselectItem(item);
    } else {
      selectItem(item);
    }
  }

  function selectAll(): void {
    const $items = get(items);
    selection.set({
      selectedItems: [...$items],
      selectedKeys: new Set($items.map(getItemKey)),
      isAllSelected: true,
      isIndeterminate: false,
    });
  }

  function deselectAll(): void {
    selection.set({
      selectedItems: [],
      selectedKeys: new Set(),
      isAllSelected: false,
      isIndeterminate: false,
    });
  }

  function selectRange(startIndex: number, endIndex: number): void {
    const $items = get(items);
    const start = Math.min(startIndex, endIndex);
    const end = Math.max(startIndex, endIndex);
    const rangeItems = $items.slice(start, end + 1);

    selection.update(($s) => {
      const selectedKeys = new Set($s.selectedKeys);
      rangeItems.forEach((item) => selectedKeys.add(getItemKey(item)));

      const selectedItems = $items.filter((item) => selectedKeys.has(getItemKey(item)));

      return {
        selectedItems,
        selectedKeys,
        isAllSelected: selectedKeys.size === $items.length,
        isIndeterminate: selectedKeys.size > 0 && selectedKeys.size < $items.length,
      };
    });
  }

  function isSelected(item: TItem): boolean {
    const key = getItemKey(item);
    return get(selection).selectedKeys.has(key);
  }

  // ============================================================================
  // EXPORT
  // ============================================================================

  async function exportData(format: ExportFormat, options?: ExportOptions): Promise<void> {
    const $items = options?.includeSelection ? get(selection).selectedItems : get(filteredItems);

    // Get columns to export
    const columns = options?.columns
      ? config.columns.filter((c) => options.columns!.includes(String(c.key)))
      : config.columns.filter((c) => c.visible !== false);

    // Build data
    const data = $items.map((item) => {
      const row: Record<string, unknown> = {};
      for (const col of columns) {
        const value = (item as Record<string, unknown>)[col.key as string];
        row[col.header] = col.format ? col.format(value, item, 0) : value;
      }
      return row;
    });

    // Generate export based on format
    switch (format) {
      case 'csv':
        exportCSV(data, columns.map((c) => c.header), options);
        break;
      case 'json':
        exportJSON(data, options);
        break;
      default:
        console.warn(`Export format ${format} not implemented`);
    }
  }

  function exportCSV(
    data: Record<string, unknown>[],
    headers: string[],
    options?: ExportOptions
  ): void {
    const delimiter = options?.delimiter ?? ',';
    const quoteChar = options?.quoteChar ?? '"';

    const escapeValue = (value: unknown): string => {
      const str = value == null ? '' : String(value);
      if (str.includes(delimiter) || str.includes(quoteChar) || str.includes('\n')) {
        return `${quoteChar}${str.replace(new RegExp(quoteChar, 'g'), quoteChar + quoteChar)}${quoteChar}`;
      }
      return str;
    };

    const lines: string[] = [];

    if (options?.includeHeaders !== false) {
      lines.push(headers.map(escapeValue).join(delimiter));
    }

    for (const row of data) {
      lines.push(headers.map((h) => escapeValue(row[h])).join(delimiter));
    }

    const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8;' });
    downloadBlob(blob, `${options?.filename ?? 'export'}.csv`);
  }

  function exportJSON(data: Record<string, unknown>[], options?: ExportOptions): void {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    downloadBlob(blob, `${options?.filename ?? 'export'}.json`);
  }

  function downloadBlob(blob: Blob, filename: string): void {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    // Stores
    items,
    filteredItems,
    displayItems,
    state,
    filters,
    activeFilters,
    searchQuery,
    sort,
    pagination,
    selection,

    // Methods
    load,
    reload,
    refresh,

    // Filter methods
    setFilter,
    setFilters,
    clearFilter,
    clearAllFilters,
    addActiveFilter,
    removeActiveFilter,

    // Search methods
    search,
    clearSearch,

    // Sort methods
    setSort,
    toggleSort,
    clearSort,

    // Pagination methods
    setPage,
    setPageSize,
    nextPage,
    prevPage,
    goToFirst,
    goToLast,

    // Selection methods
    selectItem,
    deselectItem,
    toggleItem,
    selectAll,
    deselectAll,
    selectRange,
    isSelected,

    // Export
    exportData,

    // Utility
    getItemKey,
  };
}
