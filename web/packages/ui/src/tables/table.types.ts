/**
 * Table component types
 */

import type { Size, SortDirection, FilterOperator, FilterValue, PaginationState } from '../types';

/** Column definition */
export interface TableColumn<T = Record<string, unknown>> {
  /** Unique column key (corresponds to data field) */
  key: string;
  /** Column header text */
  header: string;
  /** Column width */
  width?: string;
  /** Minimum width */
  minWidth?: string;
  /** Maximum width */
  maxWidth?: string;
  /** Text alignment */
  align?: 'left' | 'center' | 'right';
  /** Whether column is sortable */
  sortable?: boolean;
  /** Whether column is filterable */
  filterable?: boolean;
  /** Filter type for this column */
  filterType?: 'text' | 'number' | 'date' | 'select' | 'boolean';
  /** Filter options (for select type) */
  filterOptions?: { value: string; label: string }[];
  /** Whether column is visible */
  visible?: boolean;
  /** Whether column is resizable */
  resizable?: boolean;
  /** Whether column is sticky */
  sticky?: 'left' | 'right';
  /** Custom cell renderer key */
  cellRenderer?: string;
  /** Custom header renderer key */
  headerRenderer?: string;
  /** Value formatter function */
  format?: (value: unknown, row: T) => string;
  /** CSS class for cells */
  cellClass?: string;
  /** CSS class for header */
  headerClass?: string;
}

/** Table row */
export interface TableRow<T = Record<string, unknown>> {
  id: string | number;
  data: T;
  selected?: boolean;
  expanded?: boolean;
  children?: TableRow<T>[];
}

/** Sort state */
export interface SortState {
  column: string | null;
  direction: SortDirection;
}

/** Table props interface */
export interface TableProps<T = Record<string, unknown>> {
  /** Column definitions */
  columns: TableColumn<T>[];
  /** Table data */
  data: T[];
  /** Unique row identifier key */
  rowKey?: string;
  /** Table size variant */
  size?: Size;
  /** Whether table has striped rows */
  striped?: boolean;
  /** Whether table has hover effect */
  hoverable?: boolean;
  /** Whether table has borders */
  bordered?: boolean;
  /** Whether to show row selection */
  selectable?: boolean;
  /** Selection mode */
  selectionMode?: 'single' | 'multiple';
  /** Selected row keys */
  selectedKeys?: (string | number)[];
  /** Whether table is sortable */
  sortable?: boolean;
  /** Current sort state */
  sortState?: SortState;
  /** Loading state */
  loading?: boolean;
  /** Empty state message */
  emptyMessage?: string;
  /** Sticky header */
  stickyHeader?: boolean;
  /** Maximum height (enables scroll) */
  maxHeight?: string;
  /** Full width */
  fullWidth?: boolean;
}

/** Column resize state */
export interface ColumnWidthState {
  [columnKey: string]: number;
}

/** Column order state */
export type ColumnOrderState = string[];

/** DataGrid specific props */
export interface DataGridProps<T = Record<string, unknown>> extends TableProps<T> {
  /** Enable filtering */
  filterable?: boolean;
  /** Current filters */
  filters?: FilterValue[];
  /** Enable global search */
  searchable?: boolean;
  /** Current search query */
  searchQuery?: string;
  /** Enable pagination */
  paginated?: boolean;
  /** Pagination state */
  pagination?: PaginationState;
  /** Available page sizes */
  pageSizes?: number[];
  /** Enable export */
  exportable?: boolean;
  /** Export formats */
  exportFormats?: ('csv' | 'xlsx' | 'pdf')[];
  /** Export filename (without extension) */
  exportFilename?: string;
  /** Enable column visibility toggle */
  columnToggle?: boolean;
  /** Enable row expansion */
  expandable?: boolean;
  /** Toolbar position */
  toolbarPosition?: 'top' | 'bottom' | 'both';
  /** Enable column resizing */
  columnResizable?: boolean;
  /** Column widths state */
  columnWidths?: ColumnWidthState;
  /** Enable column reordering */
  columnReorderable?: boolean;
  /** Column order (array of column keys) */
  columnOrder?: ColumnOrderState;
}

/** Table classes */
export const tableClasses = {
  wrapper: 'w-full overflow-auto',
  table: 'w-full border-collapse',
  thead: 'bg-neutral-50',
  theadSticky: 'sticky top-0 z-10 bg-neutral-50',
  tr: 'border-b border-neutral-200',
  trHover: 'hover:bg-neutral-50',
  trStriped: 'even:bg-neutral-25',
  trSelected: 'bg-brand-primary-50',
  th: 'px-4 py-3 text-left text-xs font-semibold text-neutral-600 uppercase tracking-wider',
  thSortable: 'cursor-pointer select-none hover:bg-neutral-100',
  td: 'px-4 py-3 text-sm text-neutral-900',
  tdBordered: 'border border-neutral-200',
  checkbox: 'w-4 h-4',
  sortIcon: 'ml-1 inline-flex',
  loading: 'absolute inset-0 bg-neutral-white/75 flex items-center justify-center z-20',
  empty: 'py-12 text-center text-neutral-500',
};

/** Table size classes */
export const tableSizeClasses: Record<Size, { th: string; td: string }> = {
  xs: { th: 'px-2 py-1 text-xs', td: 'px-2 py-1 text-xs' },
  sm: { th: 'px-3 py-2 text-xs', td: 'px-3 py-2 text-sm' },
  md: { th: 'px-4 py-3 text-xs', td: 'px-4 py-3 text-sm' },
  lg: { th: 'px-5 py-4 text-sm', td: 'px-5 py-4 text-base' },
  xl: { th: 'px-6 py-5 text-sm', td: 'px-6 py-5 text-lg' },
};

/** DataGrid toolbar classes */
export const toolbarClasses = {
  container: 'flex flex-wrap items-center justify-between gap-4 py-3 px-4 bg-neutral-50 border-b border-neutral-200',
  left: 'flex items-center gap-3',
  right: 'flex items-center gap-3',
  search: 'relative',
  searchInput: 'pl-9 pr-4 py-2 text-sm border border-neutral-300 rounded-md focus:outline-none focus:ring-2 focus:ring-brand-primary-500 focus:border-brand-primary-500',
  searchIcon: 'absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-neutral-400',
  button: 'px-3 py-2 text-sm font-medium rounded-md transition-colors',
  buttonPrimary: 'bg-brand-primary-500 text-neutral-white hover:bg-brand-primary-600',
  buttonSecondary: 'bg-neutral-white border border-neutral-300 text-neutral-700 hover:bg-neutral-50',
  filterBadge: 'ml-1 px-1.5 py-0.5 text-xs bg-brand-primary-100 text-brand-primary-700 rounded-full',
};

/** Filter panel classes */
export const filterPanelClasses = {
  container: 'p-4 bg-neutral-white border-b border-neutral-200',
  row: 'flex flex-wrap items-center gap-3 mb-2',
  select: 'px-3 py-2 text-sm border border-neutral-300 rounded-md focus:outline-none focus:ring-2 focus:ring-brand-primary-500',
  input: 'px-3 py-2 text-sm border border-neutral-300 rounded-md focus:outline-none focus:ring-2 focus:ring-brand-primary-500',
  removeBtn: 'p-1 text-neutral-400 hover:text-semantic-error-500 rounded',
  addBtn: 'text-sm text-brand-primary-600 hover:text-brand-primary-700 font-medium',
};

/** Pagination classes */
export const paginationClasses = {
  container: 'flex flex-wrap items-center justify-between gap-4 py-3 px-4 bg-neutral-50 border-t border-neutral-200',
  info: 'text-sm text-neutral-600',
  controls: 'flex items-center gap-2',
  button: 'p-2 rounded-md border border-neutral-300 text-neutral-600 hover:bg-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed',
  pageButton: 'px-3 py-1 text-sm rounded-md',
  pageButtonActive: 'bg-brand-primary-500 text-neutral-white',
  pageButtonInactive: 'hover:bg-neutral-100 text-neutral-700',
  pageSizeSelect: 'px-2 py-1 text-sm border border-neutral-300 rounded-md',
};

/** Export dropdown classes */
export const exportDropdownClasses = {
  container: 'relative',
  button: 'flex items-center gap-2 px-3 py-2 text-sm font-medium bg-neutral-white border border-neutral-300 text-neutral-700 rounded-md hover:bg-neutral-50',
  dropdown: 'absolute right-0 mt-1 w-40 bg-neutral-white border border-neutral-200 rounded-md shadow-lg z-dropdown',
  item: 'block w-full px-4 py-2 text-sm text-left text-neutral-700 hover:bg-neutral-50',
};
