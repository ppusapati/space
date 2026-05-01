/**
 * usePagination Composable
 * Creates a reactive pagination state with page navigation
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type { PaginationState } from '../types/index.js';

// ============================================================================
// TYPES
// ============================================================================

export interface UsePaginationOptions {
  initialPage?: number;
  initialPageSize?: number;
  total?: number;
  pageSizeOptions?: number[];
  onChange?: (pagination: PaginationState) => void;
}

export interface UsePaginationReturn {
  // State
  page: Writable<number>;
  pageSize: Writable<number>;
  total: Writable<number>;

  // Derived
  pagination: Readable<PaginationState>;
  totalPages: Readable<number>;
  hasNext: Readable<boolean>;
  hasPrevious: Readable<boolean>;
  pageRange: Readable<number[]>;
  startIndex: Readable<number>;
  endIndex: Readable<number>;

  // Methods
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setTotal: (total: number) => void;
  nextPage: () => void;
  prevPage: () => void;
  goToFirst: () => void;
  goToLast: () => void;
  reset: () => void;

  // Utility
  getPageItems: <T>(items: T[]) => T[];
  getVisiblePages: (maxVisible?: number) => number[];
}

// ============================================================================
// IMPLEMENTATION
// ============================================================================

export function usePagination(options: UsePaginationOptions = {}): UsePaginationReturn {
  const {
    initialPage = 1,
    initialPageSize = 10,
    total: initialTotal = 0,
    pageSizeOptions = [10, 25, 50, 100],
    onChange,
  } = options;

  // ============================================================================
  // STORES
  // ============================================================================

  const page = writable<number>(initialPage);
  const pageSize = writable<number>(initialPageSize);
  const total = writable<number>(initialTotal);

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const totalPages = derived([total, pageSize], ([$total, $pageSize]) =>
    Math.max(1, Math.ceil($total / $pageSize))
  );

  const hasNext = derived([page, totalPages], ([$page, $totalPages]) => $page < $totalPages);

  const hasPrevious = derived(page, ($page) => $page > 1);

  const startIndex = derived([page, pageSize], ([$page, $pageSize]) => ($page - 1) * $pageSize);

  const endIndex = derived([page, pageSize, total], ([$page, $pageSize, $total]) =>
    Math.min($page * $pageSize, $total)
  );

  const pageRange = derived([page, totalPages], ([$page, $totalPages]) => {
    const range: number[] = [];
    for (let i = 1; i <= $totalPages; i++) {
      range.push(i);
    }
    return range;
  });

  const pagination = derived(
    [page, pageSize, total, totalPages, hasNext, hasPrevious],
    ([$page, $pageSize, $total, $totalPages, $hasNext, $hasPrevious]) => ({
      page: $page,
      pageSize: $pageSize,
      total: $total,
      totalPages: $totalPages,
      hasNext: $hasNext,
      hasPrevious: $hasPrevious,
    })
  );

  // ============================================================================
  // SUBSCRIPTIONS
  // ============================================================================

  // Notify on change
  let isInitial = true;
  pagination.subscribe(($pagination) => {
    if (!isInitial) {
      onChange?.($pagination);
    }
    isInitial = false;
  });

  // Ensure page is within bounds when total changes
  total.subscribe(($total) => {
    const $page = get(page);
    const $pageSize = get(pageSize);
    const maxPage = Math.max(1, Math.ceil($total / $pageSize));
    if ($page > maxPage) {
      page.set(maxPage);
    }
  });

  // Reset to page 1 when page size changes
  pageSize.subscribe(() => {
    if (!isInitial) {
      page.set(1);
    }
  });

  // ============================================================================
  // METHODS
  // ============================================================================

  function setPage(newPage: number): void {
    const $totalPages = get(totalPages);
    page.set(Math.max(1, Math.min(newPage, $totalPages)));
  }

  function setPageSize(size: number): void {
    if (pageSizeOptions.includes(size) || pageSizeOptions.length === 0) {
      pageSize.set(size);
    }
  }

  function setTotal(newTotal: number): void {
    total.set(Math.max(0, newTotal));
  }

  function nextPage(): void {
    if (get(hasNext)) {
      page.update(($p) => $p + 1);
    }
  }

  function prevPage(): void {
    if (get(hasPrevious)) {
      page.update(($p) => $p - 1);
    }
  }

  function goToFirst(): void {
    page.set(1);
  }

  function goToLast(): void {
    page.set(get(totalPages));
  }

  function reset(): void {
    page.set(initialPage);
    pageSize.set(initialPageSize);
    total.set(initialTotal);
  }

  // ============================================================================
  // UTILITY FUNCTIONS
  // ============================================================================

  function getPageItems<T>(items: T[]): T[] {
    const $startIndex = get(startIndex);
    const $pageSize = get(pageSize);
    return items.slice($startIndex, $startIndex + $pageSize);
  }

  function getVisiblePages(maxVisible = 7): number[] {
    const $page = get(page);
    const $totalPages = get(totalPages);

    if ($totalPages <= maxVisible) {
      return Array.from({ length: $totalPages }, (_, i) => i + 1);
    }

    const halfVisible = Math.floor(maxVisible / 2);
    let startPage = Math.max(1, $page - halfVisible);
    let endPage = Math.min($totalPages, $page + halfVisible);

    // Adjust if we're near the start or end
    if ($page <= halfVisible) {
      endPage = maxVisible;
    } else if ($page >= $totalPages - halfVisible) {
      startPage = $totalPages - maxVisible + 1;
    }

    const pages: number[] = [];

    // Always show first page
    if (startPage > 1) {
      pages.push(1);
      if (startPage > 2) {
        pages.push(-1); // Ellipsis marker
      }
    }

    // Middle pages
    for (let i = startPage; i <= endPage; i++) {
      pages.push(i);
    }

    // Always show last page
    if (endPage < $totalPages) {
      if (endPage < $totalPages - 1) {
        pages.push(-1); // Ellipsis marker
      }
      pages.push($totalPages);
    }

    return pages;
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    // State
    page,
    pageSize,
    total,

    // Derived
    pagination,
    totalPages,
    hasNext,
    hasPrevious,
    pageRange,
    startIndex,
    endIndex,

    // Methods
    setPage,
    setPageSize,
    setTotal,
    nextPage,
    prevPage,
    goToFirst,
    goToLast,
    reset,

    // Utility
    getPageItems,
    getVisiblePages,
  };
}
