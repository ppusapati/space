/**
 * List Generic Types
 * Comprehensive list handling with filtering, sorting, pagination, and selection
 */

import type { Component, Snippet } from 'svelte';
import type {
  BaseProps,
  LoadingState,
  PaginationState,
  SortConfig,
  SortDirection,
  SelectionState,
  FilterValue,
  FilterOperator,
  Action,
  BulkAction,
  ColorVariant,
  Size,
} from './common.types.js';

// ============================================================================
// LIST GENERIC TYPE
// ============================================================================

/**
 * Generic List Type
 * @template TItem - The item type in the list
 * @template TFilters - Filter state type
 */
export interface List<
  TItem,
  TFilters extends Record<string, unknown> = Record<string, unknown>
> {
  // Data
  items: TItem[];
  filteredItems: TItem[];
  displayItems: TItem[]; // After pagination

  // State
  state: ListState;
  loadingState: LoadingState;

  // Filters
  filters: TFilters;
  activeFilters: FilterValue[];
  searchQuery: string;

  // Sorting
  sort: SortConfig;

  // Pagination
  pagination: PaginationState;

  // Selection
  selection: SelectionState<TItem>;

  // Configuration
  config: ListConfig<TItem>;

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

  // Bulk operations
  executeBulkAction: (actionId: string) => Promise<void>;

  // Export
  exportData: (format: ExportFormat, options?: ExportOptions) => Promise<void>;

  // Events
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

  // Item identification
  itemKey: keyof TItem | ((item: TItem) => string | number);

  // Columns
  columns: ListColumn<TItem>[];

  // Features
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

  // Actions
  rowActions?: Action<TItem>[];
  bulkActions?: BulkAction<TItem>[];

  // Display
  density?: 'compact' | 'default' | 'comfortable';
  striped?: boolean;
  bordered?: boolean;
  hoverable?: boolean;
  stickyHeader?: boolean;

  // Virtual scrolling
  virtualScroll?: boolean;
  rowHeight?: number;
  overscan?: number;

  // Empty state
  emptyTitle?: string;
  emptyDescription?: string;
  emptyIcon?: string;
  emptyAction?: Action<void>;

  // Row customization
  rowClass?: string | ((item: TItem, index: number) => string);
  rowStyle?: string | ((item: TItem, index: number) => string);
  isRowDisabled?: (item: TItem) => boolean;
  isRowExpandable?: (item: TItem) => boolean;
}

// ============================================================================
// COLUMN TYPES
// ============================================================================

/** List column configuration */
export interface ListColumn<TItem> {
  key: keyof TItem | string;
  header: string;
  description?: string;

  // Sizing
  width?: string;
  minWidth?: string;
  maxWidth?: string;
  flex?: number;

  // Alignment
  align?: 'left' | 'center' | 'right';
  headerAlign?: 'left' | 'center' | 'right';

  // Features
  sortable?: boolean;
  sortKey?: string; // API field name for sorting
  filterable?: boolean;
  filterType?: FilterType;
  searchable?: boolean;
  resizable?: boolean;
  visible?: boolean;
  frozen?: 'left' | 'right';
  hidden?: boolean;

  // Rendering
  format?: (value: unknown, row: TItem, index: number) => string;
  render?: (value: unknown, row: TItem, index: number) => string;
  component?: Component;
  snippet?: Snippet<[{ value: unknown; row: TItem; index: number }]>;

  // Styling
  cellClass?: string | ((value: unknown, row: TItem) => string);
  headerClass?: string;

  // Aggregation (for footer)
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

// ============================================================================
// FILTER TYPES
// ============================================================================

/** Filter type */
export type FilterType =
  | 'text'
  | 'number'
  | 'date'
  | 'dateRange'
  | 'select'
  | 'multiSelect'
  | 'boolean'
  | 'range'
  | 'custom';

/** Filter definition */
export interface FilterDefinition<TItem> {
  key: string;
  field: keyof TItem | string;
  label: string;
  type: FilterType;
  operators?: FilterOperator[];
  defaultOperator?: FilterOperator;

  // For select/multiSelect
  options?: FilterOption[];
  loadOptions?: (query: string) => Promise<FilterOption[]>;

  // For number/range
  min?: number;
  max?: number;
  step?: number;

  // For date/dateRange
  minDate?: Date;
  maxDate?: Date;
  dateFormat?: string;

  // Custom filter
  component?: Component;
  customFilter?: (item: TItem, value: unknown) => boolean;

  // Display
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

// ============================================================================
// EXPORT TYPES
// ============================================================================

/** Export format */
export type ExportFormat = 'csv' | 'xlsx' | 'pdf' | 'json';

/** Export options */
export interface ExportOptions {
  filename?: string;
  columns?: string[]; // Which columns to export
  includeHeaders?: boolean;
  includeSelection?: boolean; // Only export selected items
  encoding?: string;

  // CSV options
  delimiter?: string;
  quoteChar?: string;

  // PDF options
  orientation?: 'portrait' | 'landscape';
  pageSize?: 'A4' | 'Letter' | 'Legal';
  title?: string;
  subtitle?: string;

  // XLSX options
  sheetName?: string;
}

// ============================================================================
// LIST VARIANTS
// ============================================================================

/**
 * Data Grid List
 * Full-featured data grid with all features
 */
export interface DataGridList<TItem, TFilters extends Record<string, unknown>>
  extends List<TItem, TFilters> {
  // Additional grid features
  columnGroups?: ColumnGroup<TItem>[];
  frozenColumns: { left: number; right: number };
  expandedRows: Set<string | number>;
  editingCell: { rowKey: string | number; column: string } | null;

  // Column management
  reorderColumns: (fromIndex: number, toIndex: number) => void;
  resizeColumn: (column: string, width: number) => void;
  toggleColumnVisibility: (column: string) => void;
  freezeColumn: (column: string, side: 'left' | 'right' | null) => void;
  resetColumns: () => void;

  // Row expansion
  expandRow: (rowKey: string | number) => void;
  collapseRow: (rowKey: string | number) => void;
  toggleRow: (rowKey: string | number) => void;
  expandAll: () => void;
  collapseAll: () => void;

  // Cell editing
  startEdit: (rowKey: string | number, column: string) => void;
  cancelEdit: () => void;
  saveEdit: (value: unknown) => Promise<void>;
}

/**
 * Simple List
 * Basic list without grid features
 */
export interface SimpleList<TItem> extends Omit<List<TItem>, 'columns' | 'sort'> {
  // Simplified configuration
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
export interface GroupedList<TItem, TFilters extends Record<string, unknown>>
  extends List<TItem, TFilters> {
  // Grouping
  groupBy: keyof TItem | ((item: TItem) => string);
  groups: ListGroup<TItem>[];
  collapsedGroups: Set<string>;
  groupSort?: 'asc' | 'desc' | ((a: string, b: string) => number);

  // Group methods
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
export interface TreeList<TItem extends TreeListItem>
  extends Omit<List<TItem>, 'pagination'> {
  // Tree state
  expandedKeys: Set<string | number>;
  loadedKeys: Set<string | number>;
  loadingKeys: Set<string | number>;

  // Tree config
  parentKey: keyof TItem;
  childrenKey: keyof TItem;
  lazyLoad?: boolean;
  loadChildren?: (parent: TItem) => Promise<TItem[]>;

  // Tree methods
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
  // Virtual scroll state
  scrollOffset: number;
  visibleRange: { start: number; end: number };
  totalHeight: number;

  // Virtual scroll config
  rowHeight: number | ((index: number) => number);
  overscan: number;
  bufferSize: number;

  // Virtual scroll methods
  scrollToIndex: (index: number, align?: 'start' | 'center' | 'end') => void;
  scrollToOffset: (offset: number) => void;
  measureRow: (index: number) => number;
}

// ============================================================================
// LIST SLOTS
// ============================================================================

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

// ============================================================================
// LIST EVENTS
// ============================================================================

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
