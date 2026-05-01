/**
 * Entity Store Factory
 * Creates a reusable store for CRUD operations on entities
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type {
  BaseEntity,
  PaginationState,
  SortConfig,
  SelectionState,
  LoadingState,
} from '@samavāya/core';

// ============================================================================
// TYPES
// ============================================================================

export interface EntityStoreConfig<TEntity extends BaseEntity, TFilters = Record<string, unknown>> {
  name: string;
  idField?: keyof TEntity;
  defaultPageSize?: number;
  defaultSort?: SortConfig;

  // API handlers
  fetchList?: (params: FetchListParams<TFilters>) => Promise<FetchListResult<TEntity>>;
  fetchOne?: (id: string) => Promise<TEntity>;
  create?: (data: Partial<TEntity>) => Promise<TEntity>;
  update?: (id: string, data: Partial<TEntity>) => Promise<TEntity>;
  remove?: (id: string) => Promise<void>;
  bulkRemove?: (ids: string[]) => Promise<void>;
}

export interface FetchListParams<TFilters> {
  page: number;
  pageSize: number;
  sort: SortConfig;
  filters: TFilters;
  search?: string;
}

export interface FetchListResult<TEntity> {
  items: TEntity[];
  total: number;
}

export interface EntityState<TEntity extends BaseEntity, TFilters = Record<string, unknown>> {
  // List state
  items: TEntity[];
  loadingState: LoadingState;
  error: EntityError | null;

  // Pagination
  pagination: PaginationState;

  // Sorting
  sort: SortConfig;

  // Filtering
  filters: TFilters;
  search: string;

  // Selection
  selection: SelectionState<TEntity>;

  // Single entity
  current: TEntity | null;
  currentLoading: boolean;
  currentError: EntityError | null;

  // Timestamps
  lastFetched: Date | null;
}

export interface EntityError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface EntityStoreReturn<TEntity extends BaseEntity, TFilters = Record<string, unknown>> {
  // State
  subscribe: Writable<EntityState<TEntity, TFilters>>['subscribe'];

  // Derived stores
  items: Readable<TEntity[]>;
  isLoading: Readable<boolean>;
  isEmpty: Readable<boolean>;
  hasError: Readable<boolean>;
  pagination: Readable<PaginationState>;
  sort: Readable<SortConfig>;
  filters: Readable<TFilters>;
  selection: Readable<SelectionState<TEntity>>;
  current: Readable<TEntity | null>;

  // List actions
  fetchList: () => Promise<void>;
  refresh: () => Promise<void>;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setSort: (column: string, direction: 'asc' | 'desc' | null) => void;
  setFilters: (filters: Partial<TFilters>) => void;
  resetFilters: () => void;
  setSearch: (query: string) => void;

  // Selection actions
  select: (entity: TEntity) => void;
  deselect: (entity: TEntity) => void;
  toggleSelect: (entity: TEntity) => void;
  selectAll: () => void;
  deselectAll: () => void;
  isSelected: (entity: TEntity) => boolean;

  // CRUD actions
  fetchOne: (id: string) => Promise<TEntity | null>;
  create: (data: Partial<TEntity>) => Promise<TEntity | null>;
  update: (id: string, data: Partial<TEntity>) => Promise<TEntity | null>;
  remove: (id: string) => Promise<boolean>;
  bulkRemove: (ids: string[]) => Promise<boolean>;

  // Utility
  getById: (id: string) => TEntity | undefined;
  setCurrent: (entity: TEntity | null) => void;
  reset: () => void;
}

// ============================================================================
// FACTORY
// ============================================================================

export function createEntityStore<
  TEntity extends BaseEntity,
  TFilters extends Record<string, unknown> = Record<string, unknown>
>(config: EntityStoreConfig<TEntity, TFilters>): EntityStoreReturn<TEntity, TFilters> {
  const {
    name,
    idField = 'id' as keyof TEntity,
    defaultPageSize = 10,
    defaultSort = { column: null, direction: null },
    fetchList: fetchListFn,
    fetchOne: fetchOneFn,
    create: createFn,
    update: updateFn,
    remove: removeFn,
    bulkRemove: bulkRemoveFn,
  } = config;

  // ============================================================================
  // INITIAL STATE
  // ============================================================================

  const initialState: EntityState<TEntity, TFilters> = {
    items: [],
    loadingState: 'idle',
    error: null,
    pagination: {
      page: 1,
      pageSize: defaultPageSize,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrevious: false,
    },
    sort: defaultSort,
    filters: {} as TFilters,
    search: '',
    selection: {
      selectedItems: [],
      selectedKeys: new Set(),
      isAllSelected: false,
      isIndeterminate: false,
    },
    current: null,
    currentLoading: false,
    currentError: null,
    lastFetched: null,
  };

  // ============================================================================
  // STORE
  // ============================================================================

  const store = writable<EntityState<TEntity, TFilters>>(initialState);
  const { subscribe, set, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const items: Readable<TEntity[]> = derived(store, ($s) => $s.items);
  const isLoading: Readable<boolean> = derived(store, ($s) => $s.loadingState === 'loading');
  const isEmpty: Readable<boolean> = derived(store, ($s) => $s.items.length === 0 && $s.loadingState !== 'loading');
  const hasError: Readable<boolean> = derived(store, ($s) => $s.error !== null);
  const pagination: Readable<PaginationState> = derived(store, ($s) => $s.pagination);
  const sort: Readable<SortConfig> = derived(store, ($s) => $s.sort);
  const filters: Readable<TFilters> = derived(store, ($s) => $s.filters);
  const selection: Readable<SelectionState<TEntity>> = derived(store, ($s) => $s.selection);
  const current: Readable<TEntity | null> = derived(store, ($s) => $s.current);

  // ============================================================================
  // HELPERS
  // ============================================================================

  function getEntityId(entity: TEntity): string {
    return String(entity[idField]);
  }

  function updatePagination(total: number): void {
    update((s) => {
      const totalPages = Math.ceil(total / s.pagination.pageSize);
      return {
        ...s,
        pagination: {
          ...s.pagination,
          total,
          totalPages,
          hasNext: s.pagination.page < totalPages,
          hasPrevious: s.pagination.page > 1,
        },
      };
    });
  }

  // ============================================================================
  // LIST ACTIONS
  // ============================================================================

  async function fetchList(): Promise<void> {
    if (!fetchListFn) {
      console.warn(`[${name}Store] fetchList not configured`);
      return;
    }

    update((s) => ({ ...s, loadingState: 'loading', error: null }));

    try {
      const state = get(store);
      const result = await fetchListFn({
        page: state.pagination.page,
        pageSize: state.pagination.pageSize,
        sort: state.sort,
        filters: state.filters,
        search: state.search,
      });

      update((s) => ({
        ...s,
        items: result.items,
        loadingState: 'success',
        lastFetched: new Date(),
      }));
      updatePagination(result.total);
    } catch (error) {
      update((s) => ({
        ...s,
        loadingState: 'error',
        error: {
          code: 'FETCH_LIST_ERROR',
          message: error instanceof Error ? error.message : 'Failed to fetch list',
        },
      }));
    }
  }

  async function refresh(): Promise<void> {
    update((s) => ({ ...s, loadingState: 'refreshing' }));
    await fetchList();
  }

  function setPage(page: number): void {
    update((s) => ({
      ...s,
      pagination: { ...s.pagination, page },
    }));
    fetchList();
  }

  function setPageSize(size: number): void {
    update((s) => ({
      ...s,
      pagination: { ...s.pagination, pageSize: size, page: 1 },
    }));
    fetchList();
  }

  function setSort(column: string, direction: 'asc' | 'desc' | null): void {
    update((s) => ({
      ...s,
      sort: { column, direction },
      pagination: { ...s.pagination, page: 1 },
    }));
    fetchList();
  }

  function setFilters(newFilters: Partial<TFilters>): void {
    update((s) => ({
      ...s,
      filters: { ...s.filters, ...newFilters },
      pagination: { ...s.pagination, page: 1 },
    }));
    fetchList();
  }

  function resetFilters(): void {
    update((s) => ({
      ...s,
      filters: {} as TFilters,
      search: '',
      pagination: { ...s.pagination, page: 1 },
    }));
    fetchList();
  }

  function setSearch(query: string): void {
    update((s) => ({
      ...s,
      search: query,
      pagination: { ...s.pagination, page: 1 },
    }));
    fetchList();
  }

  // ============================================================================
  // SELECTION ACTIONS
  // ============================================================================

  function select(entity: TEntity): void {
    const id = getEntityId(entity);
    update((s) => {
      if (s.selection.selectedKeys.has(id)) return s;

      const selectedKeys = new Set(s.selection.selectedKeys);
      selectedKeys.add(id);
      const selectedItems = [...s.selection.selectedItems, entity];

      return {
        ...s,
        selection: {
          selectedItems,
          selectedKeys,
          isAllSelected: selectedKeys.size === s.items.length,
          isIndeterminate: selectedKeys.size > 0 && selectedKeys.size < s.items.length,
        },
      };
    });
  }

  function deselect(entity: TEntity): void {
    const id = getEntityId(entity);
    update((s) => {
      if (!s.selection.selectedKeys.has(id)) return s;

      const selectedKeys = new Set(s.selection.selectedKeys);
      selectedKeys.delete(id);
      const selectedItems = s.selection.selectedItems.filter((e) => getEntityId(e) !== id);

      return {
        ...s,
        selection: {
          selectedItems,
          selectedKeys,
          isAllSelected: false,
          isIndeterminate: selectedKeys.size > 0,
        },
      };
    });
  }

  function toggleSelect(entity: TEntity): void {
    const id = getEntityId(entity);
    const state = get(store);
    if (state.selection.selectedKeys.has(id)) {
      deselect(entity);
    } else {
      select(entity);
    }
  }

  function selectAll(): void {
    const state = get(store);
    update((s) => ({
      ...s,
      selection: {
        selectedItems: [...s.items],
        selectedKeys: new Set(s.items.map(getEntityId)),
        isAllSelected: true,
        isIndeterminate: false,
      },
    }));
  }

  function deselectAll(): void {
    update((s) => ({
      ...s,
      selection: {
        selectedItems: [],
        selectedKeys: new Set(),
        isAllSelected: false,
        isIndeterminate: false,
      },
    }));
  }

  function isSelected(entity: TEntity): boolean {
    const state = get(store);
    return state.selection.selectedKeys.has(getEntityId(entity));
  }

  // ============================================================================
  // CRUD ACTIONS
  // ============================================================================

  async function fetchOne(id: string): Promise<TEntity | null> {
    if (!fetchOneFn) {
      console.warn(`[${name}Store] fetchOne not configured`);
      return null;
    }

    update((s) => ({ ...s, currentLoading: true, currentError: null }));

    try {
      const entity = await fetchOneFn(id);
      update((s) => ({ ...s, current: entity, currentLoading: false }));
      return entity;
    } catch (error) {
      update((s) => ({
        ...s,
        currentLoading: false,
        currentError: {
          code: 'FETCH_ONE_ERROR',
          message: error instanceof Error ? error.message : 'Failed to fetch entity',
        },
      }));
      return null;
    }
  }

  async function create(data: Partial<TEntity>): Promise<TEntity | null> {
    if (!createFn) {
      console.warn(`[${name}Store] create not configured`);
      return null;
    }

    try {
      const entity = await createFn(data);
      update((s) => ({
        ...s,
        items: [entity, ...s.items],
        pagination: { ...s.pagination, total: s.pagination.total + 1 },
      }));
      return entity;
    } catch (error) {
      update((s) => ({
        ...s,
        error: {
          code: 'CREATE_ERROR',
          message: error instanceof Error ? error.message : 'Failed to create entity',
        },
      }));
      return null;
    }
  }

  async function updateEntity(id: string, data: Partial<TEntity>): Promise<TEntity | null> {
    if (!updateFn) {
      console.warn(`[${name}Store] update not configured`);
      return null;
    }

    try {
      const entity = await updateFn(id, data);
      update((s) => ({
        ...s,
        items: s.items.map((e) => (getEntityId(e) === id ? entity : e)),
        current: s.current && getEntityId(s.current) === id ? entity : s.current,
      }));
      return entity;
    } catch (error) {
      update((s) => ({
        ...s,
        error: {
          code: 'UPDATE_ERROR',
          message: error instanceof Error ? error.message : 'Failed to update entity',
        },
      }));
      return null;
    }
  }

  async function remove(id: string): Promise<boolean> {
    if (!removeFn) {
      console.warn(`[${name}Store] remove not configured`);
      return false;
    }

    try {
      await removeFn(id);
      update((s) => ({
        ...s,
        items: s.items.filter((e) => getEntityId(e) !== id),
        pagination: { ...s.pagination, total: s.pagination.total - 1 },
        current: s.current && getEntityId(s.current) === id ? null : s.current,
      }));
      return true;
    } catch (error) {
      update((s) => ({
        ...s,
        error: {
          code: 'REMOVE_ERROR',
          message: error instanceof Error ? error.message : 'Failed to remove entity',
        },
      }));
      return false;
    }
  }

  async function bulkRemove(ids: string[]): Promise<boolean> {
    if (!bulkRemoveFn) {
      console.warn(`[${name}Store] bulkRemove not configured`);
      return false;
    }

    try {
      await bulkRemoveFn(ids);
      const idSet = new Set(ids);
      update((s) => ({
        ...s,
        items: s.items.filter((e) => !idSet.has(getEntityId(e))),
        pagination: { ...s.pagination, total: s.pagination.total - ids.length },
        selection: {
          selectedItems: [],
          selectedKeys: new Set(),
          isAllSelected: false,
          isIndeterminate: false,
        },
      }));
      return true;
    } catch (error) {
      update((s) => ({
        ...s,
        error: {
          code: 'BULK_REMOVE_ERROR',
          message: error instanceof Error ? error.message : 'Failed to remove entities',
        },
      }));
      return false;
    }
  }

  // ============================================================================
  // UTILITY
  // ============================================================================

  function getById(id: string): TEntity | undefined {
    const state = get(store);
    return state.items.find((e) => getEntityId(e) === id);
  }

  function setCurrent(entity: TEntity | null): void {
    update((s) => ({ ...s, current: entity }));
  }

  function reset(): void {
    set(initialState);
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    subscribe,
    // Derived stores
    items,
    isLoading,
    isEmpty,
    hasError,
    pagination,
    sort,
    filters,
    selection,
    current,
    // List actions
    fetchList,
    refresh,
    setPage,
    setPageSize,
    setSort,
    setFilters,
    resetFilters,
    setSearch,
    // Selection actions
    select,
    deselect,
    toggleSelect,
    selectAll,
    deselectAll,
    isSelected,
    // CRUD actions
    fetchOne,
    create,
    update: updateEntity,
    remove,
    bulkRemove,
    // Utility
    getById,
    setCurrent,
    reset,
  };
}
