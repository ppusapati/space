import { Component, Snippet } from 'svelte';
import { BaseError, BreadcrumbItem, Action, BulkAction, PaginationState, SortConfig, SelectionState, DateRange, LoadingState } from './common.types.js';
/** Page metadata for SEO and navigation */
export interface PageMeta {
    title: string;
    description?: string;
    keywords?: string[];
    breadcrumbs?: BreadcrumbItem[];
    permissions?: string[];
    module?: string;
    icon?: string;
}
/** Page error */
export interface PageError extends BaseError {
    retryable?: boolean;
}
/**
 * Generic Page Type
 * @template TData - The main data type for the page
 * @template TParams - Route parameters type
 * @template TFilters - Filter state type
 * @template TActions - Available actions type (string union)
 */
export interface Page<TData = unknown, TParams extends Record<string, string | number> = Record<string, string | number>, TFilters extends Record<string, unknown> = Record<string, unknown>, TActions extends string = string> {
    meta: PageMeta;
    data: TData | null;
    params: TParams;
    filters: TFilters;
    loadingState: LoadingState;
    error: PageError | null;
    actions: Record<TActions, Action<TData>>;
    bulkActions?: Record<string, BulkAction>;
    load: (params?: TParams) => Promise<void>;
    reload: () => Promise<void>;
    setFilters: (filters: Partial<TFilters>) => void;
    resetFilters: () => void;
    executeAction: (actionId: TActions, data?: TData) => Promise<void>;
    onBeforeLoad?: () => void | Promise<void>;
    onAfterLoad?: (data: TData) => void;
    onError?: (error: PageError) => void;
    onBeforeUnload?: () => void | boolean;
}
/**
 * List Page Type
 * For pages displaying a list/table of items
 */
export interface ListPage<TItem, TFilters extends Record<string, unknown> = Record<string, unknown>> extends Page<TItem[], Record<string, never>, TFilters, 'create' | 'export' | 'import' | 'refresh'> {
    pagination: PaginationState;
    sorting: SortConfig;
    selection: SelectionState<TItem>;
    columns: ColumnConfig<TItem>[];
    searchQuery: string;
    setPage: (page: number) => void;
    setPageSize: (size: number) => void;
    setSort: (column: string, direction: 'asc' | 'desc' | null) => void;
    search: (query: string) => void;
    selectItem: (item: TItem) => void;
    deselectItem: (item: TItem) => void;
    selectAll: () => void;
    clearSelection: () => void;
    exportData: (format: 'csv' | 'xlsx' | 'pdf') => Promise<void>;
}
/**
 * Detail Page Type
 * For pages displaying a single entity
 */
export interface DetailPage<TEntity, TParams extends {
    id: string | number;
} = {
    id: string;
}> extends Page<TEntity, TParams, Record<string, never>, 'edit' | 'delete' | 'duplicate' | 'print'> {
    isEditing: boolean;
    isDirty: boolean;
    originalEntity: TEntity | null;
    startEdit: () => void;
    cancelEdit: () => void;
    save: () => Promise<void>;
    delete: () => Promise<void>;
    duplicate: () => Promise<void>;
    print: () => void;
}
/**
 * Form Page Type
 * For pages with form submission
 */
export interface FormPage<TValues extends Record<string, unknown>, TParams extends Record<string, string | number> = Record<string, string | number>> extends Page<TValues, TParams, Record<string, never>, 'submit' | 'reset' | 'cancel'> {
    mode: 'create' | 'edit' | 'view';
    initialValues: TValues;
    validation: FormValidationState<TValues>;
    setFieldValue: <K extends keyof TValues>(field: K, value: TValues[K]) => void;
    setFieldError: (field: keyof TValues, error: string | null) => void;
    setFieldTouched: (field: keyof TValues, touched?: boolean) => void;
    validate: () => Promise<boolean>;
    submit: () => Promise<void>;
    reset: () => void;
}
/**
 * Dashboard Page Type
 * For dashboard pages with widgets
 */
export interface DashboardPage<TWidgets extends string = string> extends Page<Record<TWidgets, unknown>, Record<string, never>, DateRangeFilter, 'refresh' | 'customize'> {
    widgets: WidgetConfig<TWidgets>[];
    dateRange: DateRange;
    activeWidgets: Set<TWidgets>;
    setDateRange: (range: DateRange) => void;
    refreshWidget: (widgetId: TWidgets) => Promise<void>;
    toggleWidget: (widgetId: TWidgets) => void;
    reorderWidgets: (order: TWidgets[]) => void;
    saveLayout: () => Promise<void>;
}
/**
 * Report Page Type
 * For report/analytics pages
 */
export interface ReportPage<TData, TParams extends Record<string, unknown> = Record<string, unknown>> extends Page<TData, Record<string, never>, TParams, 'export' | 'schedule' | 'share'> {
    chartType: ChartType;
    groupBy: string[];
    aggregations: Aggregation[];
    setChartType: (type: ChartType) => void;
    setGroupBy: (fields: string[]) => void;
    addAggregation: (aggregation: Aggregation) => void;
    removeAggregation: (id: string) => void;
    exportReport: (format: 'pdf' | 'xlsx' | 'csv') => Promise<void>;
    scheduleReport: (schedule: ReportSchedule) => Promise<void>;
}
/** Column configuration for list pages */
export interface ColumnConfig<TItem> {
    key: keyof TItem | string;
    header: string;
    width?: string;
    minWidth?: string;
    maxWidth?: string;
    align?: 'left' | 'center' | 'right';
    sortable?: boolean;
    filterable?: boolean;
    visible?: boolean;
    frozen?: 'left' | 'right';
    resizable?: boolean;
    format?: (value: unknown, row: TItem) => string;
    component?: Component;
    className?: string | ((row: TItem) => string);
}
/** Form validation state */
export interface FormValidationState<TValues> {
    isValid: boolean;
    isValidating: boolean;
    touched: Partial<Record<keyof TValues, boolean>>;
    dirty: Partial<Record<keyof TValues, boolean>>;
    errors: Partial<Record<keyof TValues, string>>;
    warnings: Partial<Record<keyof TValues, string>>;
}
/** Date range filter */
export interface DateRangeFilter extends Record<string, unknown> {
    dateRange: DateRange;
}
/** Widget configuration */
export interface WidgetConfig<TId extends string = string> {
    id: TId;
    title: string;
    type: 'chart' | 'metric' | 'table' | 'list' | 'calendar' | 'custom';
    size: 'sm' | 'md' | 'lg' | 'xl';
    position: {
        row: number;
        col: number;
    };
    span: {
        rows: number;
        cols: number;
    };
    refreshInterval?: number;
    visible?: boolean;
    config?: Record<string, unknown>;
}
/** Chart type */
export type ChartType = 'line' | 'bar' | 'pie' | 'doughnut' | 'area' | 'scatter' | 'radar' | 'gauge' | 'funnel' | 'heatmap';
/** Aggregation configuration */
export interface Aggregation {
    id: string;
    field: string;
    function: 'sum' | 'avg' | 'min' | 'max' | 'count' | 'distinct';
    label?: string;
}
/** Report schedule */
export interface ReportSchedule {
    frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly';
    dayOfWeek?: number;
    dayOfMonth?: number;
    time: string;
    recipients: string[];
    format: 'pdf' | 'xlsx' | 'csv';
}
/** Page slot configuration */
export interface PageSlots<TData = unknown> {
    header?: Snippet<[TData | null]>;
    toolbar?: Snippet<[TData | null]>;
    filters?: Snippet<[TData | null]>;
    content?: Snippet<[TData | null]>;
    footer?: Snippet<[TData | null]>;
    empty?: Snippet;
    loading?: Snippet;
    error?: Snippet<[PageError]>;
}
/** Page configuration */
export interface PageConfig {
    id: string;
    module: string;
    path: string;
    title: string;
    description?: string;
    icon?: string;
    permissions?: string[];
    layout?: 'default' | 'centered' | 'sidebar' | 'split' | 'full';
    maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
    padding?: 'none' | 'sm' | 'md' | 'lg';
    showBreadcrumbs?: boolean;
    showBackButton?: boolean;
}
//# sourceMappingURL=page.types.d.ts.map