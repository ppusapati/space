/**
 * Query Utilities
 * Helper functions for building API queries
 * @packageDocumentation
 */

import type { ListParams, PaginationMeta } from '../types/index.js';

// ============================================================================
// QUERY BUILDER
// ============================================================================

/** Query builder options */
export interface QueryBuilderOptions {
  /** Default page size */
  defaultPageSize?: number;

  /** Maximum page size */
  maxPageSize?: number;

  /** Default sort field */
  defaultSortField?: string;

  /** Default sort order */
  defaultSortOrder?: 'asc' | 'desc';
}

/**
 * Query builder for list requests
 */
export class QueryBuilder<TFilters extends Record<string, unknown> = Record<string, unknown>> {
  private params: ListParams<TFilters>;
  private options: Required<QueryBuilderOptions>;

  constructor(options: QueryBuilderOptions = {}) {
    this.options = {
      defaultPageSize: options.defaultPageSize ?? 20,
      maxPageSize: options.maxPageSize ?? 100,
      defaultSortField: options.defaultSortField ?? 'createdAt',
      defaultSortOrder: options.defaultSortOrder ?? 'desc',
    };

    this.params = {
      page: 1,
      pageSize: this.options.defaultPageSize,
      sortBy: this.options.defaultSortField,
      sortOrder: this.options.defaultSortOrder,
    };
  }

  /**
   * Sets the page number
   */
  page(page: number): this {
    this.params.page = Math.max(1, page);
    return this;
  }

  /**
   * Sets the page size
   */
  pageSize(size: number): this {
    this.params.pageSize = Math.min(Math.max(1, size), this.options.maxPageSize);
    return this;
  }

  /**
   * Sets the sort configuration
   */
  sort(field: string, order: 'asc' | 'desc' = 'asc'): this {
    this.params.sortBy = field;
    this.params.sortOrder = order;
    return this;
  }

  /**
   * Sets the search query
   */
  search(query: string): this {
    this.params.search = query || undefined;
    return this;
  }

  /**
   * Sets a single filter
   */
  filter<K extends keyof TFilters>(key: K, value: TFilters[K]): this {
    if (!this.params.filters) {
      this.params.filters = {} as TFilters;
    }
    this.params.filters[key] = value;
    return this;
  }

  /**
   * Sets multiple filters
   */
  filters(filters: Partial<TFilters>): this {
    this.params.filters = {
      ...this.params.filters,
      ...filters,
    } as TFilters;
    return this;
  }

  /**
   * Clears a specific filter
   */
  clearFilter(key: keyof TFilters): this {
    if (this.params.filters) {
      delete this.params.filters[key];
    }
    return this;
  }

  /**
   * Clears all filters
   */
  clearFilters(): this {
    this.params.filters = undefined;
    return this;
  }

  /**
   * Sets fields to include in response
   */
  fields(fields: string[]): this {
    this.params.fields = fields;
    return this;
  }

  /**
   * Sets related entities to include
   */
  include(relations: string[]): this {
    this.params.include = relations;
    return this;
  }

  /**
   * Includes soft-deleted items
   */
  withDeleted(include = true): this {
    this.params.includeDeleted = include;
    return this;
  }

  /**
   * Builds the final query parameters
   */
  build(): ListParams<TFilters> {
    // Clean up undefined values
    const result: ListParams<TFilters> = {};

    if (this.params.page !== undefined) result.page = this.params.page;
    if (this.params.pageSize !== undefined) result.pageSize = this.params.pageSize;
    if (this.params.sortBy !== undefined) result.sortBy = this.params.sortBy;
    if (this.params.sortOrder !== undefined) result.sortOrder = this.params.sortOrder;
    if (this.params.search !== undefined) result.search = this.params.search;
    if (this.params.filters !== undefined && Object.keys(this.params.filters).length > 0) {
      result.filters = this.params.filters;
    }
    if (this.params.fields !== undefined && this.params.fields.length > 0) {
      result.fields = this.params.fields;
    }
    if (this.params.include !== undefined && this.params.include.length > 0) {
      result.include = this.params.include;
    }
    if (this.params.includeDeleted) result.includeDeleted = true;

    return result;
  }

  /**
   * Resets the builder to defaults
   */
  reset(): this {
    this.params = {
      page: 1,
      pageSize: this.options.defaultPageSize,
      sortBy: this.options.defaultSortField,
      sortOrder: this.options.defaultSortOrder,
    };
    return this;
  }

  /**
   * Creates a copy of the builder
   */
  clone(): QueryBuilder<TFilters> {
    const clone = new QueryBuilder<TFilters>(this.options);
    clone.params = { ...this.params };
    if (this.params.filters) {
      clone.params.filters = { ...this.params.filters };
    }
    if (this.params.fields) {
      clone.params.fields = [...this.params.fields];
    }
    if (this.params.include) {
      clone.params.include = [...this.params.include];
    }
    return clone;
  }
}

/**
 * Creates a query builder
 */
export function createQueryBuilder<TFilters extends Record<string, unknown> = Record<string, unknown>>(
  options?: QueryBuilderOptions
): QueryBuilder<TFilters> {
  return new QueryBuilder<TFilters>(options);
}

// ============================================================================
// PAGINATION HELPERS
// ============================================================================

/**
 * Calculates pagination metadata
 */
export function calculatePagination(
  totalItems: number,
  page: number,
  pageSize: number
): PaginationMeta {
  const totalPages = Math.ceil(totalItems / pageSize);
  const currentPage = Math.min(Math.max(1, page), totalPages || 1);

  return {
    page: currentPage,
    pageSize,
    totalItems,
    totalPages,
    hasNextPage: currentPage < totalPages,
    hasPreviousPage: currentPage > 1,
  };
}

/**
 * Gets the offset for a page
 */
export function getOffset(page: number, pageSize: number): number {
  return (Math.max(1, page) - 1) * pageSize;
}

/**
 * Gets the page number for an offset
 */
export function getPage(offset: number, pageSize: number): number {
  return Math.floor(offset / pageSize) + 1;
}

/**
 * Generates page numbers for pagination UI
 */
export function generatePageNumbers(
  currentPage: number,
  totalPages: number,
  maxVisible = 7
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

  // Adjust if at the start
  if (currentPage <= sidePages + 2) {
    end = maxVisible - 2;
  }

  // Adjust if at the end
  if (currentPage >= totalPages - sidePages - 1) {
    start = totalPages - maxVisible + 3;
  }

  // Add ellipsis before range if needed
  if (start > 2) {
    pages.push('ellipsis');
  }

  // Add range
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

// ============================================================================
// FILTER HELPERS
// ============================================================================

/**
 * Serializes filters for URL query string
 */
export function serializeFilters(
  filters: Record<string, unknown>
): Record<string, string> {
  const result: Record<string, string> = {};

  for (const [key, value] of Object.entries(filters)) {
    if (value === undefined || value === null || value === '') {
      continue;
    }

    if (Array.isArray(value)) {
      result[key] = value.join(',');
    } else if (value instanceof Date) {
      result[key] = value.toISOString();
    } else if (typeof value === 'object') {
      result[key] = JSON.stringify(value);
    } else {
      result[key] = String(value);
    }
  }

  return result;
}

/**
 * Deserializes filters from URL query string
 */
export function deserializeFilters<TFilters extends Record<string, unknown>>(
  params: Record<string, string>,
  schema: Record<keyof TFilters, 'string' | 'number' | 'boolean' | 'array' | 'date' | 'object'>
): Partial<TFilters> {
  const result: Partial<TFilters> = {};

  for (const [key, type] of Object.entries(schema)) {
    const value = params[key];
    if (value === undefined || value === '') continue;

    switch (type) {
      case 'string':
        (result as Record<string, unknown>)[key] = value;
        break;
      case 'number':
        (result as Record<string, unknown>)[key] = Number(value);
        break;
      case 'boolean':
        (result as Record<string, unknown>)[key] = value === 'true';
        break;
      case 'array':
        (result as Record<string, unknown>)[key] = value.split(',');
        break;
      case 'date':
        (result as Record<string, unknown>)[key] = new Date(value);
        break;
      case 'object':
        try {
          (result as Record<string, unknown>)[key] = JSON.parse(value);
        } catch {
          // Ignore invalid JSON
        }
        break;
    }
  }

  return result;
}
