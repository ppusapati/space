/**
 * Entity Store Factory
 * Creates a reusable store for CRUD operations on entities
 */
import { type Writable, type Readable } from 'svelte/store';
import type { BaseEntity, PaginationState, SortConfig, SelectionState, LoadingState } from '@samavāya/core';
export interface EntityStoreConfig<TEntity extends BaseEntity, TFilters = Record<string, unknown>> {
    name: string;
    idField?: keyof TEntity;
    defaultPageSize?: number;
    defaultSort?: SortConfig;
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
    items: TEntity[];
    loadingState: LoadingState;
    error: EntityError | null;
    pagination: PaginationState;
    sort: SortConfig;
    filters: TFilters;
    search: string;
    selection: SelectionState<TEntity>;
    current: TEntity | null;
    currentLoading: boolean;
    currentError: EntityError | null;
    lastFetched: Date | null;
}
export interface EntityError {
    code: string;
    message: string;
    details?: Record<string, unknown>;
}
export interface EntityStoreReturn<TEntity extends BaseEntity, TFilters = Record<string, unknown>> {
    subscribe: Writable<EntityState<TEntity, TFilters>>['subscribe'];
    items: Readable<TEntity[]>;
    isLoading: Readable<boolean>;
    isEmpty: Readable<boolean>;
    hasError: Readable<boolean>;
    pagination: Readable<PaginationState>;
    sort: Readable<SortConfig>;
    filters: Readable<TFilters>;
    selection: Readable<SelectionState<TEntity>>;
    current: Readable<TEntity | null>;
    fetchList: () => Promise<void>;
    refresh: () => Promise<void>;
    setPage: (page: number) => void;
    setPageSize: (size: number) => void;
    setSort: (column: string, direction: 'asc' | 'desc' | null) => void;
    setFilters: (filters: Partial<TFilters>) => void;
    resetFilters: () => void;
    setSearch: (query: string) => void;
    select: (entity: TEntity) => void;
    deselect: (entity: TEntity) => void;
    toggleSelect: (entity: TEntity) => void;
    selectAll: () => void;
    deselectAll: () => void;
    isSelected: (entity: TEntity) => boolean;
    fetchOne: (id: string) => Promise<TEntity | null>;
    create: (data: Partial<TEntity>) => Promise<TEntity | null>;
    update: (id: string, data: Partial<TEntity>) => Promise<TEntity | null>;
    remove: (id: string) => Promise<boolean>;
    bulkRemove: (ids: string[]) => Promise<boolean>;
    getById: (id: string) => TEntity | undefined;
    setCurrent: (entity: TEntity | null) => void;
    reset: () => void;
}
export declare function createEntityStore<TEntity extends BaseEntity, TFilters extends Record<string, unknown> = Record<string, unknown>>(config: EntityStoreConfig<TEntity, TFilters>): EntityStoreReturn<TEntity, TFilters>;
//# sourceMappingURL=createEntityStore.d.ts.map