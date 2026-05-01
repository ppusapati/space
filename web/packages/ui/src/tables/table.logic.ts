/**
 * Table business logic - sorting, filtering, pagination, search
 */

import type { FilterOperator, FilterValue, SortDirection, PaginationState } from '../types';
import type { TableColumn, SortState } from './table.types';

/**
 * Sort data by column
 */
export function sortData<T extends Record<string, unknown>>(
  data: T[],
  sortState: SortState,
  columns: TableColumn<T>[]
): T[] {
  if (!sortState.column || !sortState.direction) {
    return data;
  }

  const column = columns.find(c => c.key === sortState.column);
  if (!column) return data;

  return [...data].sort((a, b) => {
    const aValue = getNestedValue(a, sortState.column!);
    const bValue = getNestedValue(b, sortState.column!);

    let comparison = 0;

    if (aValue === null || aValue === undefined) comparison = 1;
    else if (bValue === null || bValue === undefined) comparison = -1;
    else if (typeof aValue === 'string' && typeof bValue === 'string') {
      comparison = aValue.localeCompare(bValue);
    } else if (aValue instanceof Date && bValue instanceof Date) {
      comparison = aValue.getTime() - bValue.getTime();
    } else if (typeof aValue === 'number' && typeof bValue === 'number') {
      comparison = aValue - bValue;
    } else {
      comparison = String(aValue).localeCompare(String(bValue));
    }

    return sortState.direction === 'desc' ? -comparison : comparison;
  });
}

/**
 * Toggle sort direction
 */
export function toggleSort(
  currentState: SortState,
  columnKey: string
): SortState {
  if (currentState.column !== columnKey) {
    return { column: columnKey, direction: 'asc' };
  }

  if (currentState.direction === 'asc') {
    return { column: columnKey, direction: 'desc' };
  }

  return { column: null, direction: null };
}

/**
 * Filter data by filter values
 */
export function filterData<T extends Record<string, unknown>>(
  data: T[],
  filters: FilterValue[]
): T[] {
  if (!filters.length) return data;

  return data.filter(row => {
    return filters.every(filter => {
      const value = getNestedValue(row, filter.column);
      return applyFilter(value, filter.operator, filter.value, filter.secondValue);
    });
  });
}

/**
 * Apply single filter
 */
function applyFilter(
  value: unknown,
  operator: FilterOperator,
  filterValue: unknown,
  secondValue?: unknown
): boolean {
  // Handle null/undefined
  if (value === null || value === undefined) {
    if (operator === 'isEmpty') return true;
    if (operator === 'isNotEmpty') return false;
    return false;
  }

  const strValue = String(value).toLowerCase();
  const strFilterValue = String(filterValue).toLowerCase();

  switch (operator) {
    case 'equals':
      return strValue === strFilterValue;
    case 'notEquals':
      return strValue !== strFilterValue;
    case 'contains':
      return strValue.includes(strFilterValue);
    case 'startsWith':
      return strValue.startsWith(strFilterValue);
    case 'endsWith':
      return strValue.endsWith(strFilterValue);
    case 'greaterThan':
      return Number(value) > Number(filterValue);
    case 'lessThan':
      return Number(value) < Number(filterValue);
    case 'greaterThanOrEqual':
      return Number(value) >= Number(filterValue);
    case 'lessThanOrEqual':
      return Number(value) <= Number(filterValue);
    case 'between':
      return Number(value) >= Number(filterValue) && Number(value) <= Number(secondValue);
    case 'isEmpty':
      return value === '' || value === null || value === undefined;
    case 'isNotEmpty':
      return value !== '' && value !== null && value !== undefined;
    default:
      return true;
  }
}

/**
 * Search data across all columns
 */
export function searchData<T extends Record<string, unknown>>(
  data: T[],
  query: string,
  columns: TableColumn<T>[]
): T[] {
  if (!query.trim()) return data;

  const lowerQuery = query.toLowerCase();
  const searchableColumns = columns.filter(c => c.visible !== false);

  return data.filter(row => {
    return searchableColumns.some(column => {
      const value = getNestedValue(row, column.key);
      if (value === null || value === undefined) return false;

      const strValue = column.format
        ? column.format(value, row)
        : String(value);

      return strValue.toLowerCase().includes(lowerQuery);
    });
  });
}

/**
 * Paginate data
 */
export function paginateData<T>(
  data: T[],
  pagination: PaginationState
): T[] {
  const start = (pagination.page - 1) * pagination.pageSize;
  const end = start + pagination.pageSize;
  return data.slice(start, end);
}

/**
 * Calculate pagination info
 */
export function calculatePagination(
  totalItems: number,
  page: number,
  pageSize: number
): PaginationState & { totalPages: number; startItem: number; endItem: number } {
  const totalPages = Math.ceil(totalItems / pageSize);
  const startItem = totalItems === 0 ? 0 : (page - 1) * pageSize + 1;
  const endItem = Math.min(page * pageSize, totalItems);

  return {
    page,
    pageSize,
    total: totalItems,
    totalPages,
    startItem,
    endItem,
  };
}

/**
 * Get page numbers for pagination UI
 */
export function getPageNumbers(
  currentPage: number,
  totalPages: number,
  maxVisible: number = 7
): (number | 'ellipsis')[] {
  if (totalPages <= maxVisible) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }

  const pages: (number | 'ellipsis')[] = [];
  const sidePages = Math.floor((maxVisible - 3) / 2);

  // Always show first page
  pages.push(1);

  // Calculate range around current page
  let start = Math.max(2, currentPage - sidePages);
  let end = Math.min(totalPages - 1, currentPage + sidePages);

  // Adjust if at start or end
  if (currentPage <= sidePages + 2) {
    end = maxVisible - 2;
  } else if (currentPage >= totalPages - sidePages - 1) {
    start = totalPages - maxVisible + 3;
  }

  // Add ellipsis before range if needed
  if (start > 2) {
    pages.push('ellipsis');
  }

  // Add page range
  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  // Add ellipsis after range if needed
  if (end < totalPages - 1) {
    pages.push('ellipsis');
  }

  // Always show last page
  pages.push(totalPages);

  return pages;
}

/**
 * Get nested value from object using dot notation
 */
export function getNestedValue<T extends Record<string, unknown>>(
  obj: T,
  path: string
): unknown {
  return path.split('.').reduce((acc: unknown, part: string) => {
    if (acc && typeof acc === 'object' && part in acc) {
      return (acc as Record<string, unknown>)[part];
    }
    return undefined;
  }, obj);
}

/**
 * Process all table transformations
 */
export function processTableData<T extends Record<string, unknown>>(
  data: T[],
  columns: TableColumn<T>[],
  options: {
    searchQuery?: string;
    filters?: FilterValue[];
    sortState?: SortState;
    pagination?: PaginationState;
  }
): { data: T[]; total: number } {
  let result = [...data];

  // Apply search
  if (options.searchQuery) {
    result = searchData(result, options.searchQuery, columns);
  }

  // Apply filters
  if (options.filters?.length) {
    result = filterData(result, options.filters);
  }

  // Get total before pagination
  const total = result.length;

  // Apply sort
  if (options.sortState?.column) {
    result = sortData(result, options.sortState, columns);
  }

  // Apply pagination
  if (options.pagination) {
    result = paginateData(result, options.pagination);
  }

  return { data: result, total };
}

/**
 * Get filter operators for column type
 */
export function getFilterOperators(
  filterType: 'text' | 'number' | 'date' | 'select' | 'boolean'
): { value: FilterOperator; label: string }[] {
  const commonOperators = [
    { value: 'isEmpty' as FilterOperator, label: 'Is empty' },
    { value: 'isNotEmpty' as FilterOperator, label: 'Is not empty' },
  ];

  switch (filterType) {
    case 'text':
      return [
        { value: 'contains', label: 'Contains' },
        { value: 'equals', label: 'Equals' },
        { value: 'notEquals', label: 'Not equals' },
        { value: 'startsWith', label: 'Starts with' },
        { value: 'endsWith', label: 'Ends with' },
        ...commonOperators,
      ];
    case 'number':
      return [
        { value: 'equals', label: 'Equals' },
        { value: 'notEquals', label: 'Not equals' },
        { value: 'greaterThan', label: 'Greater than' },
        { value: 'lessThan', label: 'Less than' },
        { value: 'greaterThanOrEqual', label: 'Greater or equal' },
        { value: 'lessThanOrEqual', label: 'Less or equal' },
        { value: 'between', label: 'Between' },
        ...commonOperators,
      ];
    case 'date':
      return [
        { value: 'equals', label: 'Equals' },
        { value: 'notEquals', label: 'Not equals' },
        { value: 'greaterThan', label: 'After' },
        { value: 'lessThan', label: 'Before' },
        { value: 'between', label: 'Between' },
        ...commonOperators,
      ];
    case 'select':
    case 'boolean':
      return [
        { value: 'equals', label: 'Equals' },
        { value: 'notEquals', label: 'Not equals' },
        ...commonOperators,
      ];
    default:
      return [
        { value: 'contains', label: 'Contains' },
        { value: 'equals', label: 'Equals' },
        ...commonOperators,
      ];
  }
}
