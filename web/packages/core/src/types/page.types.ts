/**
 * Page Generic Types
 * Handles all page-level concerns: data fetching, routing params, filters, and actions
 */

import type { Component, Snippet } from 'svelte';
import type {
  BaseError,
  BreadcrumbItem,
  Action,
  BulkAction,
  PaginationState,
  SortConfig,
  SelectionState,
  DateRange,
  LoadingState,
  ColorVariant,
} from './common.types.js';

// ============================================================================
// PAGE METADATA
// ============================================================================

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

// ============================================================================
// PAGE GENERIC TYPE
// ============================================================================

/**
 * Generic Page Type
 * @template TData - The main data type for the page
 * @template TParams - Route parameters type
 * @template TFilters - Filter state type
 * @template TActions - Available actions type (string union)
 */
export interface Page<
  TData = unknown,
  TParams extends Record<string, string | number> = Record<string, string | number>,
  TFilters extends Record<string, unknown> = Record<string, unknown>,
  TActions extends string = string
> {
  // Metadata
  meta: PageMeta;

  // State
  data: TData | null;
  params: TParams;
  filters: TFilters;
  loadingState: LoadingState;
  error: PageError | null;

  // Actions
  actions: Record<TActions, Action<TData>>;
  bulkActions?: Record<string, BulkAction>;

  // Methods
  load: (params?: TParams) => Promise<void>;
  reload: () => Promise<void>;
  setFilters: (filters: Partial<TFilters>) => void;
  resetFilters: () => void;
  executeAction: (actionId: TActions, data?: TData) => Promise<void>;

  // Lifecycle hooks
  onBeforeLoad?: () => void | Promise<void>;
  onAfterLoad?: (data: TData) => void;
  onError?: (error: PageError) => void;
  onBeforeUnload?: () => void | boolean;
}

// ============================================================================
// PAGE VARIANTS
// ============================================================================

/**
 * List Page Type
 * For pages displaying a list/table of items
 */
export interface ListPage<
  TItem,
  TFilters extends Record<string, unknown> = Record<string, unknown>
> extends Page<TItem[], Record<string, never>, TFilters, 'create' | 'export' | 'import' | 'refresh'> {
  // List-specific state
  pagination: PaginationState;
  sorting: SortConfig;
  selection: SelectionState<TItem>;
  columns: ColumnConfig<TItem>[];
  searchQuery: string;

  // List-specific methods
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
export interface DetailPage<
  TEntity,
  TParams extends { id: string | number } = { id: string }
> extends Page<TEntity, TParams, Record<string, never>, 'edit' | 'delete' | 'duplicate' | 'print'> {
  // Detail-specific state
  isEditing: boolean;
  isDirty: boolean;
  originalEntity: TEntity | null;

  // Detail-specific methods
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
export interface FormPage<
  TValues extends Record<string, unknown>,
  TParams extends Record<string, string | number> = Record<string, string | number>
> extends Page<TValues, TParams, Record<string, never>, 'submit' | 'reset' | 'cancel'> {
  // Form-specific state
  mode: 'create' | 'edit' | 'view';
  initialValues: TValues;
  validation: FormValidationState<TValues>;

  // Form-specific methods
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
export interface DashboardPage<TWidgets extends string = string>
  extends Page<Record<TWidgets, unknown>, Record<string, never>, DateRangeFilter, 'refresh' | 'customize'> {
  // Dashboard-specific state
  widgets: WidgetConfig<TWidgets>[];
  dateRange: DateRange;
  activeWidgets: Set<TWidgets>;

  // Dashboard-specific methods
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
export interface ReportPage<
  TData,
  TParams extends Record<string, unknown> = Record<string, unknown>
> extends Page<TData, Record<string, never>, TParams, 'export' | 'schedule' | 'share'> {
  // Report-specific state
  chartType: ChartType;
  groupBy: string[];
  aggregations: Aggregation[];

  // Report-specific methods
  setChartType: (type: ChartType) => void;
  setGroupBy: (fields: string[]) => void;
  addAggregation: (aggregation: Aggregation) => void;
  removeAggregation: (id: string) => void;
  exportReport: (format: 'pdf' | 'xlsx' | 'csv') => Promise<void>;
  scheduleReport: (schedule: ReportSchedule) => Promise<void>;
}

// ============================================================================
// SUPPORTING TYPES
// ============================================================================

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
  position: { row: number; col: number };
  span: { rows: number; cols: number };
  refreshInterval?: number;
  visible?: boolean;
  config?: Record<string, unknown>;
}

/** Chart type */
export type ChartType =
  | 'line'
  | 'bar'
  | 'pie'
  | 'doughnut'
  | 'area'
  | 'scatter'
  | 'radar'
  | 'gauge'
  | 'funnel'
  | 'heatmap';

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

// ============================================================================
// PAGE SLOTS
// ============================================================================

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

// ============================================================================
// PAGE CONFIGURATION
// ============================================================================

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
