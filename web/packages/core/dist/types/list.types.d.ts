import { Component, Snippet } from 'svelte';
import { LoadingState, PaginationState, SortConfig, SortDirection, SelectionState, FilterValue, FilterOperator, Action, BulkAction, ColorVariant, Size } from './common.types.js';
/**
 * Generic List Type
 * @template TItem - The item type in the list
 * @template TFilters - Filter state type
 */
export interface List<TItem, TFilters extends Record<string, unknown> = Record<string, unknown>> {
    items: TItem[];
    filteredItems: TItem[];
    displayItems: TItem[];
    state: ListState;
    loadingState: LoadingState;
    filters: TFilters;
    activeFilters: FilterValue[];
    searchQuery: string;
    sort: SortConfig;
    pagination: PaginationState;
    selection: SelectionState<TItem>;
    config: ListConfig<TItem>;
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
    executeBulkAction: (actionId: string) => Promise<void>;
    exportData: (format: ExportFormat, options?: ExportOptions) => Promise<void>;
    onItemClick?: (item: TItem) => void;
    onItemDoubleClick?: (item: TItem) => void;
    onSelectionChange?: (selection: SelectionState<TItem>) => void;
    onFiltersChange?: (filters: TFilters) => void;
    onSortChange?: (sort: SortConfig) => void;
    onPageChange?: (page: number) => void;
}
/** List state */
export interface ListState {
    isLoading: boolean;
    isRefreshing: boolean;
    isEmpty: boolean;
    hasError: boolean;
    error?: ListError;
    lastUpdated?: Date;
}
/** List error */
export interface ListError {
    code: string;
    message: string;
    retryable?: boolean;
}
/** List configuration */
export interface ListConfig<TItem> {
    id: string;
    title?: string;
    description?: string;
    itemKey: keyof TItem | ((item: TItem) => string | number);
    columns: ListColumn<TItem>[];
    searchable?: boolean;
    searchFields?: (keyof TItem)[];
    searchPlaceholder?: string;
    filterable?: boolean;
    filterDefinitions?: FilterDefinition<TItem>[];
    sortable?: boolean;
    defaultSort?: SortConfig;
    paginated?: boolean;
    defaultPageSize?: number;
    pageSizeOptions?: number[];
    selectable?: boolean;
    multiSelect?: boolean;
    selectOnClick?: boolean;
    rowActions?: Action<TItem>[];
    bulkActions?: BulkAction<TItem>[];
    density?: 'compact' | 'default' | 'comfortable';
    striped?: boolean;
    bordered?: boolean;
    hoverable?: boolean;
    stickyHeader?: boolean;
    virtualScroll?: boolean;
    rowHeight?: number;
    overscan?: number;
    emptyTitle?: string;
    emptyDescription?: string;
    emptyIcon?: string;
    emptyAction?: Action<void>;
    rowClass?: string | ((item: TItem, index: number) => string);
    rowStyle?: string | ((item: TItem, index: number) => string);
    isRowDisabled?: (item: TItem) => boolean;
    isRowExpandable?: (item: TItem) => boolean;
}
/** List column configuration */
export interface ListColumn<TItem> {
    key: keyof TItem | string;
    header: string;
    description?: string;
    width?: string;
    minWidth?: string;
    maxWidth?: string;
    flex?: number;
    align?: 'left' | 'center' | 'right';
    headerAlign?: 'left' | 'center' | 'right';
    sortable?: boolean;
    sortKey?: string;
    filterable?: boolean;
    filterType?: FilterType;
    searchable?: boolean;
    resizable?: boolean;
    visible?: boolean;
    frozen?: 'left' | 'right';
    hidden?: boolean;
    format?: (value: unknown, row: TItem, index: number) => string;
    render?: (value: unknown, row: TItem, index: number) => string;
    component?: Component;
    snippet?: Snippet<[{
        value: unknown;
        row: TItem;
        index: number;
    }]>;
    cellClass?: string | ((value: unknown, row: TItem) => string);
    headerClass?: string;
    aggregate?: 'sum' | 'avg' | 'min' | 'max' | 'count' | ((items: TItem[]) => unknown);
    aggregateFormat?: (value: unknown) => string;
}
/** Column group */
export interface ColumnGroup<TItem> {
    id: string;
    header: string;
    columns: ListColumn<TItem>[];
    collapsible?: boolean;
    collapsed?: boolean;
}
/** Filter type */
export type FilterType = 'text' | 'number' | 'date' | 'dateRange' | 'select' | 'multiSelect' | 'boolean' | 'range' | 'custom';
/** Filter definition */
export interface FilterDefinition<TItem> {
    key: string;
    field: keyof TItem | string;
    label: string;
    type: FilterType;
    operators?: FilterOperator[];
    defaultOperator?: FilterOperator;
    options?: FilterOption[];
    loadOptions?: (query: string) => Promise<FilterOption[]>;
    min?: number;
    max?: number;
    step?: number;
    minDate?: Date;
    maxDate?: Date;
    dateFormat?: string;
    component?: Component;
    customFilter?: (item: TItem, value: unknown) => boolean;
    placeholder?: string;
    helperText?: string;
    defaultValue?: unknown;
    clearable?: boolean;
    collapsible?: boolean;
    collapsed?: boolean;
}
/** Filter option */
export interface FilterOption {
    label: string;
    value: unknown;
    icon?: string;
    color?: ColorVariant;
    count?: number;
    disabled?: boolean;
}
/** Active filter chip */
export interface ActiveFilterChip {
    field: string;
    label: string;
    operator: FilterOperator;
    operatorLabel: string;
    value: unknown;
    displayValue: string;
}
/** Export format */
export type ExportFormat = 'csv' | 'xlsx' | 'pdf' | 'json';
/** Export options */
export interface ExportOptions {
    filename?: string;
    columns?: string[];
    includeHeaders?: boolean;
    includeSelection?: boolean;
    encoding?: string;
    delimiter?: string;
    quoteChar?: string;
    orientation?: 'portrait' | 'landscape';
    pageSize?: 'A4' | 'Letter' | 'Legal';
    title?: string;
    subtitle?: string;
    sheetName?: string;
}
/**
 * Data Grid List
 * Full-featured data grid with all features
 */
export interface DataGridList<TItem, TFilters extends Record<string, unknown>> extends List<TItem, TFilters> {
    columnGroups?: ColumnGroup<TItem>[];
    frozenColumns: {
        left: number;
        right: number;
    };
    expandedRows: Set<string | number>;
    editingCell: {
        rowKey: string | number;
        column: string;
    } | null;
    reorderColumns: (fromIndex: number, toIndex: number) => void;
    resizeColumn: (column: string, width: number) => void;
    toggleColumnVisibility: (column: string) => void;
    freezeColumn: (column: string, side: 'left' | 'right' | null) => void;
    resetColumns: () => void;
    expandRow: (rowKey: string | number) => void;
    collapseRow: (rowKey: string | number) => void;
    toggleRow: (rowKey: string | number) => void;
    expandAll: () => void;
    collapseAll: () => void;
    startEdit: (rowKey: string | number, column: string) => void;
    cancelEdit: () => void;
    saveEdit: (value: unknown) => Promise<void>;
}
/**
 * Simple List
 * Basic list without grid features
 */
export interface SimpleList<TItem> extends Omit<List<TItem>, 'columns' | 'sort'> {
    itemTemplate?: Snippet<[TItem, number]>;
    itemComponent?: Component;
    orientation?: 'vertical' | 'horizontal';
    gap?: Size;
    showDividers?: boolean;
}
/**
 * Grouped List
 * List with item grouping
 */
export interface GroupedList<TItem, TFilters extends Record<string, unknown>> extends List<TItem, TFilters> {
    groupBy: keyof TItem | ((item: TItem) => string);
    groups: ListGroup<TItem>[];
    collapsedGroups: Set<string>;
    groupSort?: 'asc' | 'desc' | ((a: string, b: string) => number);
    toggleGroup: (groupKey: string) => void;
    expandAllGroups: () => void;
    collapseAllGroups: () => void;
    selectGroup: (groupKey: string) => void;
    deselectGroup: (groupKey: string) => void;
}
/** List group */
export interface ListGroup<TItem> {
    key: string;
    label: string;
    items: TItem[];
    count: number;
    collapsed: boolean;
    aggregates?: Record<string, unknown>;
}
/**
 * Tree List
 * Hierarchical list with parent-child relationships
 */
export interface TreeList<TItem extends TreeListItem> extends Omit<List<TItem>, 'pagination'> {
    expandedKeys: Set<string | number>;
    loadedKeys: Set<string | number>;
    loadingKeys: Set<string | number>;
    parentKey: keyof TItem;
    childrenKey: keyof TItem;
    lazyLoad?: boolean;
    loadChildren?: (parent: TItem) => Promise<TItem[]>;
    expand: (key: string | number) => void;
    collapse: (key: string | number) => void;
    toggle: (key: string | number) => void;
    expandAll: () => void;
    collapseAll: () => void;
    loadNode: (key: string | number) => Promise<void>;
}
/** Tree list item base */
export interface TreeListItem {
    id: string | number;
    parentId?: string | number | null;
    children?: TreeListItem[];
    isLeaf?: boolean;
    level?: number;
}
/**
 * Virtual List
 * Optimized for large datasets with virtual scrolling
 */
export interface VirtualList<TItem> extends Omit<List<TItem>, 'pagination'> {
    scrollOffset: number;
    visibleRange: {
        start: number;
        end: number;
    };
    totalHeight: number;
    rowHeight: number | ((index: number) => number);
    overscan: number;
    bufferSize: number;
    scrollToIndex: (index: number, align?: 'start' | 'center' | 'end') => void;
    scrollToOffset: (offset: number) => void;
    measureRow: (index: number) => number;
}
/** List slots */
export interface ListSlots<TItem> {
    header?: Snippet<[List<TItem>]>;
    toolbar?: Snippet<[List<TItem>]>;
    filters?: Snippet<[List<TItem>]>;
    beforeList?: Snippet<[List<TItem>]>;
    afterList?: Snippet<[List<TItem>]>;
    footer?: Snippet<[List<TItem>]>;
    empty?: Snippet;
    loading?: Snippet;
    error?: Snippet<[ListError]>;
    row?: Snippet<[TItem, number]>;
    expandedRow?: Snippet<[TItem, number]>;
    groupHeader?: Snippet<[ListGroup<TItem>]>;
    pagination?: Snippet<[PaginationState]>;
}
/** List events */
export interface ListEvents<TItem> {
    onLoad?: () => void;
    onLoadComplete?: (items: TItem[]) => void;
    onLoadError?: (error: ListError) => void;
    onItemClick?: (item: TItem, index: number) => void;
    onItemDoubleClick?: (item: TItem, index: number) => void;
    onItemContextMenu?: (item: TItem, index: number, event: MouseEvent) => void;
    onSelectionChange?: (selection: SelectionState<TItem>) => void;
    onFilterChange?: (filters: Record<string, unknown>) => void;
    onSortChange?: (sort: SortConfig) => void;
    onPageChange?: (pagination: PaginationState) => void;
    onColumnReorder?: (columns: string[]) => void;
    onColumnResize?: (column: string, width: number) => void;
    onRowExpand?: (item: TItem) => void;
    onRowCollapse?: (item: TItem) => void;
    onExport?: (format: ExportFormat) => void;
}
//# sourceMappingURL=list.types.d.ts.map