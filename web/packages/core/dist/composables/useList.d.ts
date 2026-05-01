import { Writable, Readable } from 'svelte/store';
import { ListState, ListConfig, ListError, PaginationState, SortConfig, SortDirection, SelectionState, FilterValue, ExportFormat, ExportOptions } from '../types/index.js';
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
    pagination: {
        page: number;
        pageSize: number;
    };
}
export interface FetchResult<TItem> {
    items: TItem[];
    total: number;
}
export interface UseListReturn<TItem, TFilters extends Record<string, unknown>> {
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
    load: () => Promise<void>;
    reload: () => Promise<void>;
    refresh: () => Promise<void>;
    setFilter: <K extends keyof TFilters>(key: K, value: TFilters[K]) => void;
    setFilters: (filters: Partial<TFilters>) => void;
    clearFilter: (key: keyof TFilters) => void;
    clearAllFilters: () => void;
    addActiveFilter: (filter: FilterValue) => void;
    removeActiveFilter: (field: string) => void;
    search: (query: string) => void;
    clearSearch: () => void;
    setSort: (column: string, direction: SortDirection) => void;
    toggleSort: (column: string) => void;
    clearSort: () => void;
    setPage: (page: number) => void;
    setPageSize: (size: number) => void;
    nextPage: () => void;
    prevPage: () => void;
    goToFirst: () => void;
    goToLast: () => void;
    selectItem: (item: TItem) => void;
    deselectItem: (item: TItem) => void;
    toggleItem: (item: TItem) => void;
    selectAll: () => void;
    deselectAll: () => void;
    selectRange: (startIndex: number, endIndex: number) => void;
    isSelected: (item: TItem) => boolean;
    exportData: (format: ExportFormat, options?: ExportOptions) => Promise<void>;
    getItemKey: (item: TItem) => string | number;
}
export declare function useList<TItem, TFilters extends Record<string, unknown> = Record<string, unknown>>(options: UseListOptions<TItem, TFilters>): UseListReturn<TItem, TFilters>;
//# sourceMappingURL=useList.d.ts.map