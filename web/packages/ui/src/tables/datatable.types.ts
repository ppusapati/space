/**
 * Advanced DataTable Types
 * Types for tree/hierarchical data, row grouping, and expandable rows
 */

import type { TableColumn, TableRow, DataGridProps } from './table.types';
import type { Size } from '../types';

// ============================================================================
// Tree/Hierarchical Data Types
// ============================================================================

/** Tree row with depth tracking */
export interface TreeRow<T = Record<string, unknown>> extends TableRow<T> {
  /** Nesting depth (0 = root level) */
  depth: number;
  /** Parent row ID */
  parentId?: string | number;
  /** Has children */
  hasChildren: boolean;
  /** Loading children (for lazy loading) */
  loadingChildren?: boolean;
}

/** Tree data configuration */
export interface TreeConfig<T = Record<string, unknown>> {
  /** Enable tree view */
  enabled: boolean;
  /** Key for children array in data (default: 'children') */
  childrenKey?: string;
  /** Key for parent ID (for flat data with parent references) */
  parentKey?: string;
  /** Default expanded state */
  defaultExpanded?: boolean | 'first' | 'all';
  /** Max depth to auto-expand */
  expandToDepth?: number;
  /** Lazy load children handler */
  loadChildren?: (row: T) => Promise<T[]>;
  /** Indent width in pixels per level */
  indentWidth?: number;
  /** Show connecting lines */
  showLines?: boolean;
  /** Expand icon position */
  expandIconPosition?: 'start' | 'end';
}

// ============================================================================
// Row Grouping Types
// ============================================================================

/** Row group definition */
export interface RowGroup<T = Record<string, unknown>> {
  /** Group key/value */
  key: string;
  /** Group display label */
  label: string;
  /** Grouped rows */
  rows: T[];
  /** Number of rows in group */
  count: number;
  /** Is group expanded */
  expanded: boolean;
  /** Aggregated values for columns */
  aggregates?: Record<string, unknown>;
  /** Nested groups (for multi-level grouping) */
  subGroups?: RowGroup<T>[];
  /** Grouping depth level */
  depth: number;
}

/** Grouping configuration */
export interface GroupConfig<T = Record<string, unknown>> {
  /** Enable grouping */
  enabled: boolean;
  /** Column key(s) to group by */
  groupBy: string | string[];
  /** Default expanded state for groups */
  defaultExpanded?: boolean;
  /** Custom group key generator */
  getGroupKey?: (row: T, columnKey: string) => string;
  /** Custom group label formatter */
  formatGroupLabel?: (key: string, columnKey: string, rows: T[]) => string;
  /** Enable aggregate functions per column */
  aggregates?: Record<string, GroupAggregate>;
  /** Show group count */
  showCount?: boolean;
  /** Allow collapsing groups */
  collapsible?: boolean;
  /** Sort groups */
  sortGroups?: 'asc' | 'desc' | ((a: RowGroup<T>, b: RowGroup<T>) => number);
}

/** Group aggregate types */
export type GroupAggregate = 'sum' | 'avg' | 'min' | 'max' | 'count' | 'first' | 'last';

// ============================================================================
// Expandable Row Types
// ============================================================================

/** Expandable row configuration */
export interface ExpandConfig<T = Record<string, unknown>> {
  /** Enable row expansion */
  enabled: boolean;
  /** Allow multiple expanded rows */
  multiple?: boolean;
  /** Default expanded row IDs */
  defaultExpanded?: (string | number)[];
  /** Expand trigger: 'icon' (click icon only) or 'row' (click entire row) */
  trigger?: 'icon' | 'row';
  /** Expand icon position */
  iconPosition?: 'start' | 'end';
  /** Loading state per row */
  loadingRows?: (string | number)[];
}

// ============================================================================
// Advanced DataTable Props
// ============================================================================

/** Advanced DataTable props (extends DataGrid) */
export interface AdvancedDataTableProps<T = Record<string, unknown>> extends Omit<DataGridProps<T>, 'expandable'> {
  /** Tree configuration */
  tree?: TreeConfig<T>;
  /** Grouping configuration */
  grouping?: GroupConfig<T>;
  /** Row expansion configuration */
  expandable?: ExpandConfig<T>;
  /** Expanded row IDs */
  expandedRowIds?: (string | number)[];
  /** Expanded group keys */
  expandedGroupKeys?: string[];
  /** Virtual scrolling for large datasets */
  virtualScroll?: boolean;
  /** Virtual scroll row height */
  rowHeight?: number;
  /** Virtual scroll buffer rows */
  bufferRows?: number;
}

// ============================================================================
// CSS Classes
// ============================================================================

export const treeClasses = {
  row: 'tree-row',
  indent: 'tree-indent inline-flex items-center',
  expandIcon: 'tree-expand-icon p-1 rounded hover:bg-neutral-100 cursor-pointer transition-colors',
  expandIconExpanded: 'rotate-90',
  expandIconCollapsed: '',
  expandIconLoading: 'animate-spin',
  line: 'tree-line absolute border-l border-neutral-200',
  lineHorizontal: 'border-t border-neutral-200',
  leaf: 'tree-leaf ml-6',
  nodeIcon: 'tree-node-icon w-4 h-4 mr-2',
};

export const groupClasses = {
  row: 'group-row bg-neutral-50 font-medium border-t border-b border-neutral-200',
  cell: 'group-cell px-4 py-3',
  expandIcon: 'inline-flex items-center justify-center w-5 h-5 mr-2 cursor-pointer',
  label: 'group-label',
  count: 'group-count ml-2 text-sm text-neutral-500 font-normal',
  aggregate: 'group-aggregate text-sm text-neutral-600',
  nested: 'group-nested ml-4',
};

export const expandClasses = {
  icon: 'expand-icon p-1 rounded hover:bg-neutral-100 cursor-pointer transition-transform',
  iconExpanded: 'rotate-90',
  content: 'expand-content',
  contentWrapper: 'bg-neutral-50 border-t border-neutral-100',
  contentPadding: 'p-4',
};

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Flatten tree data for rendering
 */
export function flattenTreeData<T extends Record<string, unknown>>(
  data: T[],
  config: TreeConfig<T>,
  expandedIds: Set<string | number>,
  parentId?: string | number,
  depth: number = 0
): TreeRow<T>[] {
  const childrenKey = config.childrenKey || 'children';
  const result: TreeRow<T>[] = [];

  for (const item of data) {
    const id = (item as { id?: string | number }).id ?? Math.random().toString(36).slice(2);
    const children = (item as Record<string, unknown>)[childrenKey] as T[] | undefined;
    const hasChildren = Array.isArray(children) && children.length > 0;
    const isExpanded = expandedIds.has(id);

    result.push({
      id,
      data: item,
      depth,
      parentId,
      hasChildren,
      expanded: isExpanded,
    });

    // Add children if expanded
    if (hasChildren && isExpanded) {
      const childRows = flattenTreeData(children, config, expandedIds, id, depth + 1);
      result.push(...childRows);
    }
  }

  return result;
}

/**
 * Build tree data from flat data with parent references
 */
export function buildTreeFromFlat<T extends Record<string, unknown>>(
  data: T[],
  config: TreeConfig<T>
): T[] {
  const parentKey = config.parentKey || 'parentId';
  const childrenKey = config.childrenKey || 'children';
  const idKey = 'id';

  const map = new Map<string | number, T & { [key: string]: T[] }>();
  const roots: T[] = [];

  // First pass: create map of all items
  for (const item of data) {
    const id = (item as Record<string, unknown>)[idKey] as string | number;
    map.set(id, { ...item, [childrenKey]: [] });
  }

  // Second pass: build tree structure
  for (const item of map.values()) {
    const parentId = (item as Record<string, unknown>)[parentKey] as string | number | undefined;
    if (parentId !== undefined && parentId !== null && map.has(parentId)) {
      const parent = map.get(parentId)!;
      (parent[childrenKey] as T[]).push(item as unknown as T);
    } else {
      roots.push(item as unknown as T);
    }
  }

  return roots;
}

/**
 * Group data by column(s)
 */
export function groupData<T extends Record<string, unknown>>(
  data: T[],
  config: GroupConfig<T>,
  expandedGroups: Set<string>
): RowGroup<T>[] {
  const groupByKeys = Array.isArray(config.groupBy) ? config.groupBy : [config.groupBy];

  function createGroups(rows: T[], keys: string[], depth: number): RowGroup<T>[] {
    if (keys.length === 0) return [];

    const currentKey = keys[0]!;
    const remainingKeys = keys.slice(1);
    const groups = new Map<string, T[]>();

    // Group rows by current key
    for (const row of rows) {
      const value = config.getGroupKey
        ? config.getGroupKey(row, currentKey)
        : String((row as Record<string, unknown>)[currentKey] ?? 'Unknown');

      if (!groups.has(value)) {
        groups.set(value, []);
      }
      groups.get(value)!.push(row);
    }

    // Convert to RowGroup array
    const result: RowGroup<T>[] = [];

    for (const [key, groupRows] of groups.entries()) {
      const groupKey = `${currentKey}:${key}`;
      const label = config.formatGroupLabel
        ? config.formatGroupLabel(key, currentKey, groupRows)
        : key;

      const group: RowGroup<T> = {
        key: groupKey,
        label,
        rows: groupRows,
        count: groupRows.length,
        expanded: config.defaultExpanded ?? expandedGroups.has(groupKey),
        depth,
        aggregates: config.aggregates ? calculateAggregates(groupRows, config.aggregates) : undefined,
      };

      // Create nested groups if there are more keys
      if (remainingKeys.length > 0) {
        group.subGroups = createGroups(groupRows, remainingKeys, depth + 1);
      }

      result.push(group);
    }

    // Sort groups if configured
    if (config.sortGroups) {
      if (typeof config.sortGroups === 'function') {
        result.sort(config.sortGroups);
      } else {
        result.sort((a, b) => {
          const comparison = a.label.localeCompare(b.label);
          return config.sortGroups === 'desc' ? -comparison : comparison;
        });
      }
    }

    return result;
  }

  return createGroups(data, groupByKeys, 0);
}

/**
 * Calculate aggregate values for a group
 */
export function calculateAggregates<T extends Record<string, unknown>>(
  rows: T[],
  aggregates: Record<string, GroupAggregate>
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  for (const [key, aggType] of Object.entries(aggregates)) {
    const values = rows.map((row) => (row as Record<string, unknown>)[key]).filter((v) => v != null);
    const numericValues = values.map(Number).filter((n) => !isNaN(n));

    switch (aggType) {
      case 'sum':
        result[key] = numericValues.reduce((a, b) => a + b, 0);
        break;
      case 'avg':
        result[key] = numericValues.length > 0 ? numericValues.reduce((a, b) => a + b, 0) / numericValues.length : 0;
        break;
      case 'min':
        result[key] = numericValues.length > 0 ? Math.min(...numericValues) : null;
        break;
      case 'max':
        result[key] = numericValues.length > 0 ? Math.max(...numericValues) : null;
        break;
      case 'count':
        result[key] = values.length;
        break;
      case 'first':
        result[key] = values[0] ?? null;
        break;
      case 'last':
        result[key] = values[values.length - 1] ?? null;
        break;
    }
  }

  return result;
}

/**
 * Flatten grouped data for rendering
 */
export function flattenGroups<T extends Record<string, unknown>>(
  groups: RowGroup<T>[],
  expandedGroups: Set<string>
): Array<{ type: 'group'; group: RowGroup<T> } | { type: 'row'; row: T; groupKey: string }> {
  const result: Array<{ type: 'group'; group: RowGroup<T> } | { type: 'row'; row: T; groupKey: string }> = [];

  function processGroups(groups: RowGroup<T>[]) {
    for (const group of groups) {
      result.push({ type: 'group', group });

      if (expandedGroups.has(group.key)) {
        if (group.subGroups && group.subGroups.length > 0) {
          processGroups(group.subGroups);
        } else {
          for (const row of group.rows) {
            result.push({ type: 'row', row, groupKey: group.key });
          }
        }
      }
    }
  }

  processGroups(groups);
  return result;
}

/**
 * Get all row IDs in tree (including children)
 */
export function getAllTreeIds<T extends Record<string, unknown>>(
  data: T[],
  childrenKey: string = 'children'
): (string | number)[] {
  const ids: (string | number)[] = [];

  function collect(items: T[]) {
    for (const item of items) {
      const id = (item as { id?: string | number }).id;
      if (id !== undefined) {
        ids.push(id);
      }
      const children = (item as Record<string, unknown>)[childrenKey] as T[] | undefined;
      if (Array.isArray(children)) {
        collect(children);
      }
    }
  }

  collect(data);
  return ids;
}

/**
 * Get default expanded IDs based on config
 */
export function getDefaultExpandedIds<T extends Record<string, unknown>>(
  data: T[],
  config: TreeConfig<T>
): Set<string | number> {
  const expanded = new Set<string | number>();
  const childrenKey = config.childrenKey || 'children';

  if (config.defaultExpanded === true || config.defaultExpanded === 'all') {
    return new Set(getAllTreeIds(data, childrenKey));
  }

  if (config.defaultExpanded === 'first' && data.length > 0) {
    const id = (data[0] as { id?: string | number }).id;
    if (id !== undefined) {
      expanded.add(id);
    }
  }

  if (config.expandToDepth !== undefined) {
    function expandToDepth(items: T[], depth: number) {
      if (depth > config.expandToDepth!) return;

      for (const item of items) {
        const id = (item as { id?: string | number }).id;
        if (id !== undefined) {
          expanded.add(id);
        }
        const children = (item as Record<string, unknown>)[childrenKey] as T[] | undefined;
        if (Array.isArray(children)) {
          expandToDepth(children, depth + 1);
        }
      }
    }

    expandToDepth(data, 0);
  }

  return expanded;
}

/**
 * Search in tree data (including children)
 */
export function searchTreeData<T extends Record<string, unknown>>(
  data: T[],
  query: string,
  columns: TableColumn<T>[],
  childrenKey: string = 'children'
): T[] {
  if (!query.trim()) return data;

  const lowerQuery = query.toLowerCase();

  function matchesQuery(row: T): boolean {
    return columns.some((col) => {
      const value = (row as Record<string, unknown>)[col.key];
      if (value == null) return false;
      return String(value).toLowerCase().includes(lowerQuery);
    });
  }

  function filterTree(items: T[]): T[] {
    const result: T[] = [];
    for (const item of items) {
      const children = (item as Record<string, unknown>)[childrenKey] as T[] | undefined;
      const filteredChildren = children ? filterTree(children) : [];
      const itemMatches = matchesQuery(item);

      if (itemMatches || filteredChildren.length > 0) {
        result.push({
          ...item,
          [childrenKey]: filteredChildren,
        } as T);
      }
    }
    return result;
  }

  return filterTree(data);
}
