/**
 * Query Utilities
 * Helper functions for building API queries
 * @packageDocumentation
 */
/**
 * Query builder for list requests
 */
export class QueryBuilder {
    params;
    options;
    constructor(options = {}) {
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
    page(page) {
        this.params.page = Math.max(1, page);
        return this;
    }
    /**
     * Sets the page size
     */
    pageSize(size) {
        this.params.pageSize = Math.min(Math.max(1, size), this.options.maxPageSize);
        return this;
    }
    /**
     * Sets the sort configuration
     */
    sort(field, order = 'asc') {
        this.params.sortBy = field;
        this.params.sortOrder = order;
        return this;
    }
    /**
     * Sets the search query
     */
    search(query) {
        this.params.search = query || undefined;
        return this;
    }
    /**
     * Sets a single filter
     */
    filter(key, value) {
        if (!this.params.filters) {
            this.params.filters = {};
        }
        this.params.filters[key] = value;
        return this;
    }
    /**
     * Sets multiple filters
     */
    filters(filters) {
        this.params.filters = {
            ...this.params.filters,
            ...filters,
        };
        return this;
    }
    /**
     * Clears a specific filter
     */
    clearFilter(key) {
        if (this.params.filters) {
            delete this.params.filters[key];
        }
        return this;
    }
    /**
     * Clears all filters
     */
    clearFilters() {
        this.params.filters = undefined;
        return this;
    }
    /**
     * Sets fields to include in response
     */
    fields(fields) {
        this.params.fields = fields;
        return this;
    }
    /**
     * Sets related entities to include
     */
    include(relations) {
        this.params.include = relations;
        return this;
    }
    /**
     * Includes soft-deleted items
     */
    withDeleted(include = true) {
        this.params.includeDeleted = include;
        return this;
    }
    /**
     * Builds the final query parameters
     */
    build() {
        // Clean up undefined values
        const result = {};
        if (this.params.page !== undefined)
            result.page = this.params.page;
        if (this.params.pageSize !== undefined)
            result.pageSize = this.params.pageSize;
        if (this.params.sortBy !== undefined)
            result.sortBy = this.params.sortBy;
        if (this.params.sortOrder !== undefined)
            result.sortOrder = this.params.sortOrder;
        if (this.params.search !== undefined)
            result.search = this.params.search;
        if (this.params.filters !== undefined && Object.keys(this.params.filters).length > 0) {
            result.filters = this.params.filters;
        }
        if (this.params.fields !== undefined && this.params.fields.length > 0) {
            result.fields = this.params.fields;
        }
        if (this.params.include !== undefined && this.params.include.length > 0) {
            result.include = this.params.include;
        }
        if (this.params.includeDeleted)
            result.includeDeleted = true;
        return result;
    }
    /**
     * Resets the builder to defaults
     */
    reset() {
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
    clone() {
        const clone = new QueryBuilder(this.options);
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
export function createQueryBuilder(options) {
    return new QueryBuilder(options);
}
// ============================================================================
// PAGINATION HELPERS
// ============================================================================
/**
 * Calculates pagination metadata
 */
export function calculatePagination(totalItems, page, pageSize) {
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
export function getOffset(page, pageSize) {
    return (Math.max(1, page) - 1) * pageSize;
}
/**
 * Gets the page number for an offset
 */
export function getPage(offset, pageSize) {
    return Math.floor(offset / pageSize) + 1;
}
/**
 * Generates page numbers for pagination UI
 */
export function generatePageNumbers(currentPage, totalPages, maxVisible = 7) {
    if (totalPages <= maxVisible) {
        return Array.from({ length: totalPages }, (_, i) => i + 1);
    }
    const pages = [];
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
export function serializeFilters(filters) {
    const result = {};
    for (const [key, value] of Object.entries(filters)) {
        if (value === undefined || value === null || value === '') {
            continue;
        }
        if (Array.isArray(value)) {
            result[key] = value.join(',');
        }
        else if (value instanceof Date) {
            result[key] = value.toISOString();
        }
        else if (typeof value === 'object') {
            result[key] = JSON.stringify(value);
        }
        else {
            result[key] = String(value);
        }
    }
    return result;
}
/**
 * Deserializes filters from URL query string
 */
export function deserializeFilters(params, schema) {
    const result = {};
    for (const [key, type] of Object.entries(schema)) {
        const value = params[key];
        if (value === undefined || value === '')
            continue;
        switch (type) {
            case 'string':
                result[key] = value;
                break;
            case 'number':
                result[key] = Number(value);
                break;
            case 'boolean':
                result[key] = value === 'true';
                break;
            case 'array':
                result[key] = value.split(',');
                break;
            case 'date':
                result[key] = new Date(value);
                break;
            case 'object':
                try {
                    result[key] = JSON.parse(value);
                }
                catch {
                    // Ignore invalid JSON
                }
                break;
        }
    }
    return result;
}
//# sourceMappingURL=query.js.map