/**
 * Query Utilities
 * Helper functions for building API queries
 * @packageDocumentation
 */
import type { ListParams, PaginationMeta } from '../types/index.js';
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
export declare class QueryBuilder<TFilters extends Record<string, unknown> = Record<string, unknown>> {
    private params;
    private options;
    constructor(options?: QueryBuilderOptions);
    /**
     * Sets the page number
     */
    page(page: number): this;
    /**
     * Sets the page size
     */
    pageSize(size: number): this;
    /**
     * Sets the sort configuration
     */
    sort(field: string, order?: 'asc' | 'desc'): this;
    /**
     * Sets the search query
     */
    search(query: string): this;
    /**
     * Sets a single filter
     */
    filter<K extends keyof TFilters>(key: K, value: TFilters[K]): this;
    /**
     * Sets multiple filters
     */
    filters(filters: Partial<TFilters>): this;
    /**
     * Clears a specific filter
     */
    clearFilter(key: keyof TFilters): this;
    /**
     * Clears all filters
     */
    clearFilters(): this;
    /**
     * Sets fields to include in response
     */
    fields(fields: string[]): this;
    /**
     * Sets related entities to include
     */
    include(relations: string[]): this;
    /**
     * Includes soft-deleted items
     */
    withDeleted(include?: boolean): this;
    /**
     * Builds the final query parameters
     */
    build(): ListParams<TFilters>;
    /**
     * Resets the builder to defaults
     */
    reset(): this;
    /**
     * Creates a copy of the builder
     */
    clone(): QueryBuilder<TFilters>;
}
/**
 * Creates a query builder
 */
export declare function createQueryBuilder<TFilters extends Record<string, unknown> = Record<string, unknown>>(options?: QueryBuilderOptions): QueryBuilder<TFilters>;
/**
 * Calculates pagination metadata
 */
export declare function calculatePagination(totalItems: number, page: number, pageSize: number): PaginationMeta;
/**
 * Gets the offset for a page
 */
export declare function getOffset(page: number, pageSize: number): number;
/**
 * Gets the page number for an offset
 */
export declare function getPage(offset: number, pageSize: number): number;
/**
 * Generates page numbers for pagination UI
 */
export declare function generatePageNumbers(currentPage: number, totalPages: number, maxVisible?: number): (number | 'ellipsis')[];
/**
 * Serializes filters for URL query string
 */
export declare function serializeFilters(filters: Record<string, unknown>): Record<string, string>;
/**
 * Deserializes filters from URL query string
 */
export declare function deserializeFilters<TFilters extends Record<string, unknown>>(params: Record<string, string>, schema: Record<keyof TFilters, 'string' | 'number' | 'boolean' | 'array' | 'date' | 'object'>): Partial<TFilters>;
//# sourceMappingURL=query.d.ts.map