/**
 * View Generic Types
 * Views are reusable presentation patterns within pages
 */

import type { Component, Snippet } from 'svelte';
import type { BaseError, Action, ColorVariant } from './common.types.js';

// ============================================================================
// VIEW GENERIC TYPE
// ============================================================================

/**
 * Generic View Type
 * @template TData - The data type for the view
 * @template TContext - Contextual information available to the view
 */
export interface View<
  TData = unknown,
  TContext extends Record<string, unknown> = Record<string, unknown>
> {
  // Data
  data: TData;
  context: TContext;

  // State
  isLoading: boolean;
  isEmpty: boolean;
  hasError: boolean;
  error?: ViewError;

  // Configuration
  config: ViewConfig;

  // Slots
  slots?: ViewSlots<TData, TContext>;

  // Events
  onDataChange?: (data: TData) => void;
  onContextChange?: (context: TContext) => void;
}

/** View error */
export interface ViewError extends BaseError {
  retryable?: boolean;
}

/** View configuration */
export interface ViewConfig {
  id: string;
  title?: string;
  description?: string;
  refreshable?: boolean;
  collapsible?: boolean;
  maximizable?: boolean;
  bordered?: boolean;
  shadow?: boolean;
}

/** View slots */
export interface ViewSlots<TData = unknown, TContext = unknown> {
  header?: Snippet<[TData, TContext]>;
  content?: Snippet<[TData, TContext]>;
  footer?: Snippet<[TData, TContext]>;
  empty?: Snippet<[TContext]>;
  loading?: Snippet;
  error?: Snippet<[ViewError]>;
}

// ============================================================================
// VIEW VARIANTS
// ============================================================================

/**
 * Card View
 * Display content in a card format
 */
export interface CardView<TData> extends View<TData> {
  variant: 'elevated' | 'outlined' | 'filled';
  hoverable?: boolean;
  clickable?: boolean;
  selected?: boolean;
  actions?: ViewAction<TData>[];
}

/**
 * Timeline View
 * Display items in a chronological timeline
 */
export interface TimelineView<TItem extends TimelineItem>
  extends View<TItem[], { expandedIds?: Set<string> }> {
  orientation: 'vertical' | 'horizontal';
  alternating?: boolean;
  showConnectors?: boolean;
  showTimestamps?: boolean;
  groupByDate?: boolean;

  onItemClick?: (item: TItem) => void;
  onItemExpand?: (item: TItem) => void;
}

/** Timeline item */
export interface TimelineItem {
  id: string;
  timestamp: Date;
  title: string;
  description?: string;
  icon?: string;
  color?: ColorVariant;
  type?: string;
  data?: Record<string, unknown>;
}

/**
 * Kanban View
 * Display items in a kanban board layout
 */
export interface KanbanView<TItem extends KanbanItem, TColumn extends KanbanColumn>
  extends View<{ columns: TColumn[]; items: TItem[] }, { draggedItem?: TItem }> {
  allowDragDrop: boolean;
  allowColumnDragDrop?: boolean;
  showColumnCounts?: boolean;
  showColumnLimits?: boolean;

  onItemMove: (itemId: string, fromColumn: string, toColumn: string, position: number) => void;
  onItemClick?: (item: TItem) => void;
  onColumnReorder?: (columns: TColumn[]) => void;
  onColumnCollapse?: (columnId: string, collapsed: boolean) => void;
}

/** Kanban column */
export interface KanbanColumn {
  id: string;
  title: string;
  color?: ColorVariant;
  limit?: number;
  collapsed?: boolean;
  locked?: boolean;
}

/** Kanban item */
export interface KanbanItem {
  id: string;
  columnId: string;
  title: string;
  position: number;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  assignee?: { id: string; name: string; avatar?: string };
  dueDate?: Date;
  labels?: Array<{ id: string; name: string; color: string }>;
  data?: Record<string, unknown>;
}

/**
 * Calendar View
 * Display events in a calendar format
 */
export interface CalendarView<TEvent extends CalendarEvent>
  extends View<TEvent[], { currentDate: Date; viewMode: CalendarViewMode }> {
  viewMode: CalendarViewMode;
  firstDayOfWeek?: 0 | 1 | 2 | 3 | 4 | 5 | 6;
  showWeekNumbers?: boolean;
  showNavigator?: boolean;
  minDate?: Date;
  maxDate?: Date;

  onDateSelect: (date: Date) => void;
  onEventClick: (event: TEvent) => void;
  onEventCreate?: (start: Date, end: Date) => void;
  onEventMove?: (eventId: string, newStart: Date, newEnd: Date) => void;
  onEventResize?: (eventId: string, newStart: Date, newEnd: Date) => void;
  onViewChange?: (mode: CalendarViewMode, date: Date) => void;
}

/** Calendar view mode */
export type CalendarViewMode = 'day' | 'week' | 'month' | 'year' | 'agenda';

/** Calendar event */
export interface CalendarEvent {
  id: string;
  title: string;
  start: Date;
  end: Date;
  allDay?: boolean;
  color?: ColorVariant | string;
  editable?: boolean;
  draggable?: boolean;
  resizable?: boolean;
  recurring?: RecurrenceRule;
  data?: Record<string, unknown>;
}

/** Recurrence rule */
export interface RecurrenceRule {
  frequency: 'daily' | 'weekly' | 'monthly' | 'yearly';
  interval: number;
  endDate?: Date;
  count?: number;
  byDay?: number[];
  byMonth?: number[];
  byMonthDay?: number[];
}

/**
 * Tree View
 * Display hierarchical data in a tree structure
 */
export interface TreeView<TNode extends TreeNode>
  extends View<TNode[], { expandedKeys: Set<string>; selectedKeys: Set<string> }> {
  expandable?: boolean;
  selectable?: boolean;
  checkable?: boolean;
  draggable?: boolean;
  multiSelect?: boolean;
  showLines?: boolean;
  showIcons?: boolean;

  onExpand: (nodeId: string, expanded: boolean) => void;
  onSelect: (nodeId: string, selected: boolean) => void;
  onCheck?: (nodeId: string, checked: boolean) => void;
  onDrop?: (nodeId: string, parentId: string | null, position: number) => void;
  onNodeClick?: (node: TNode) => void;
  onNodeDoubleClick?: (node: TNode) => void;
}

/** Tree node */
export interface TreeNode {
  id: string;
  label: string;
  parentId?: string | null;
  children?: TreeNode[];
  expanded?: boolean;
  selected?: boolean;
  checked?: boolean;
  disabled?: boolean;
  icon?: string;
  isLeaf?: boolean;
  data?: Record<string, unknown>;
}

/**
 * Gallery View
 * Display items in a gallery/grid layout
 */
export interface GalleryView<TItem extends GalleryItem>
  extends View<TItem[], { selectedId?: string; viewMode: 'grid' | 'masonry' | 'list' }> {
  columns: number | 'auto';
  gap: number;
  aspectRatio?: number;
  showCaptions?: boolean;
  lightbox?: boolean;

  onItemSelect: (item: TItem) => void;
  onItemDelete?: (itemId: string) => void;
  onItemDownload?: (item: TItem) => void;
}

/** Gallery item */
export interface GalleryItem {
  id: string;
  src: string;
  thumbnail?: string;
  title?: string;
  description?: string;
  type: 'image' | 'video' | 'document' | 'other';
  metadata?: Record<string, unknown>;
}

/**
 * Activity Feed View
 * Display activity/audit log
 */
export interface ActivityFeedView<TActivity extends ActivityItem>
  extends View<TActivity[], { filters?: ActivityFilters }> {
  groupByDate?: boolean;
  showAvatars?: boolean;
  showTimestamps?: boolean;
  maxItems?: number;
  loadMore?: boolean;

  onItemClick?: (item: TActivity) => void;
  onLoadMore?: () => Promise<void>;
  onFilterChange?: (filters: ActivityFilters) => void;
}

/** Activity item */
export interface ActivityItem {
  id: string;
  type: string;
  action: string;
  actor: {
    id: string;
    name: string;
    avatar?: string;
  };
  target?: {
    type: string;
    id: string;
    name: string;
  };
  timestamp: Date;
  description?: string;
  metadata?: Record<string, unknown>;
}

/** Activity filters */
export interface ActivityFilters {
  types?: string[];
  actors?: string[];
  startDate?: Date;
  endDate?: Date;
  search?: string;
}

/**
 * Statistics View
 * Display metrics/KPIs
 */
export interface StatisticsView<TMetric extends MetricItem>
  extends View<TMetric[], { dateRange?: { start: Date; end: Date } }> {
  layout: 'grid' | 'list' | 'compact';
  columns?: number;
  showTrends?: boolean;
  showComparison?: boolean;

  onMetricClick?: (metric: TMetric) => void;
  onMetricRefresh?: (metricId: string) => Promise<void>;
}

/** Metric item */
export interface MetricItem {
  id: string;
  label: string;
  value: number;
  previousValue?: number;
  change?: number;
  changePercent?: number;
  trend?: 'up' | 'down' | 'stable';
  target?: number;
  unit?: string;
  format?: 'number' | 'currency' | 'percent' | 'duration';
  icon?: string;
  color?: ColorVariant;
}

// ============================================================================
// VIEW ACTIONS
// ============================================================================

/** View action */
export interface ViewAction<TData = unknown> extends Action<TData> {
  position?: 'header' | 'footer' | 'inline' | 'menu';
}
