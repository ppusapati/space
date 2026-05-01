/**
 * VirtualList component types and logic
 */

import type { BaseProps } from '../types';

/** VirtualList props interface */
export interface VirtualListProps<T = unknown> extends BaseProps {
  /** List items */
  items: T[];
  /** Fixed item height in pixels */
  itemHeight: number;
  /** Container height (CSS value) */
  height?: string;
  /** Number of items to render outside visible area */
  overscan?: number;
  /** Key extractor function */
  getKey?: (item: T, index: number) => string | number;
}

/** Virtual list state */
export interface VirtualListState {
  scrollTop: number;
  startIndex: number;
  endIndex: number;
  offsetTop: number;
}

/** VirtualList classes */
export const virtualListClasses = {
  container: 'overflow-auto relative',
  content: 'relative',
  item: 'absolute left-0 right-0',
};

/**
 * Calculate visible range based on scroll position
 */
export function calculateVisibleRange(
  scrollTop: number,
  containerHeight: number,
  itemHeight: number,
  itemCount: number,
  overscan: number
): { startIndex: number; endIndex: number; offsetTop: number } {
  const visibleCount = Math.ceil(containerHeight / itemHeight);

  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
  const endIndex = Math.min(
    itemCount - 1,
    Math.floor(scrollTop / itemHeight) + visibleCount + overscan
  );

  const offsetTop = startIndex * itemHeight;

  return { startIndex, endIndex, offsetTop };
}

/**
 * Get total content height
 */
export function getTotalHeight(itemCount: number, itemHeight: number): number {
  return itemCount * itemHeight;
}

/**
 * Get item position
 */
export function getItemPosition(index: number, itemHeight: number): number {
  return index * itemHeight;
}
