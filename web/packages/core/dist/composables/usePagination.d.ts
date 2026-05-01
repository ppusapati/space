import { Writable, Readable } from 'svelte/store';
import { PaginationState } from '../types/index.js';
export interface UsePaginationOptions {
    initialPage?: number;
    initialPageSize?: number;
    total?: number;
    pageSizeOptions?: number[];
    onChange?: (pagination: PaginationState) => void;
}
export interface UsePaginationReturn {
    page: Writable<number>;
    pageSize: Writable<number>;
    total: Writable<number>;
    pagination: Readable<PaginationState>;
    totalPages: Readable<number>;
    hasNext: Readable<boolean>;
    hasPrevious: Readable<boolean>;
    pageRange: Readable<number[]>;
    startIndex: Readable<number>;
    endIndex: Readable<number>;
    setPage: (page: number) => void;
    setPageSize: (size: number) => void;
    setTotal: (total: number) => void;
    nextPage: () => void;
    prevPage: () => void;
    goToFirst: () => void;
    goToLast: () => void;
    reset: () => void;
    getPageItems: <T>(items: T[]) => T[];
    getVisiblePages: (maxVisible?: number) => number[];
}
export declare function usePagination(options?: UsePaginationOptions): UsePaginationReturn;
//# sourceMappingURL=usePagination.d.ts.map