/**
 * Widget Generic Types
 * Widgets are self-contained, configurable UI blocks
 */

import type { Component, Snippet } from 'svelte';
import type { BaseError, ColorVariant, Size } from './common.types.js';

// ============================================================================
// WIDGET GENERIC TYPE
// ============================================================================

/**
 * Generic Widget Type
 * @template TData - Widget data type
 * @template TConfig - Widget configuration type
 */
export interface Widget<
  TData = unknown,
  TConfig extends WidgetConfig = WidgetConfig
> {
  // Identity
  id: string;
  type: WidgetType;

  // Data
  data: TData | null;

  // Configuration
  config: TConfig;

  // State
  state: WidgetState;

  // Methods
  load: () => Promise<void>;
  refresh: () => Promise<void>;
  configure: (config: Partial<TConfig>) => void;
  reset: () => void;

  // Events
  onDataLoad?: (data: TData) => void;
  onError?: (error: WidgetError) => void;
  onRefresh?: () => void;
  onConfigure?: (config: TConfig) => void;
}

/** Widget type */
export type WidgetType =
  | 'metric'
  | 'chart'
  | 'table'
  | 'list'
  | 'calendar'
  | 'timeline'
  | 'map'
  | 'progress'
  | 'activity'
  | 'custom';

/** Widget configuration */
export interface WidgetConfig {
  id: string;
  title: string;
  description?: string;
  icon?: string;
  refreshInterval?: number; // in seconds, 0 = manual only
  cacheDuration?: number; // in seconds
  permissions?: string[];
  showHeader?: boolean;
  showFooter?: boolean;
  actions?: WidgetAction[];
}

/** Widget state */
export interface WidgetState {
  isLoading: boolean;
  isRefreshing: boolean;
  lastUpdated: Date | null;
  error: WidgetError | null;
  isEmpty: boolean;
  isConfiguring: boolean;
}

/** Widget error */
export interface WidgetError extends BaseError {
  retryable: boolean;
}

/** Widget action */
export interface WidgetAction {
  id: string;
  label: string;
  icon?: string;
  handler: () => void | Promise<void>;
  disabled?: boolean;
}

// ============================================================================
// WIDGET VARIANTS
// ============================================================================

/**
 * Metric Widget
 * Display a single metric/KPI
 */
export interface MetricWidget extends Widget<MetricData, MetricWidgetConfig> {
  type: 'metric';
}

/** Metric data */
export interface MetricData {
  value: number;
  previousValue?: number;
  change?: number;
  changePercent?: number;
  trend?: 'up' | 'down' | 'stable';
  target?: number;
  targetPercent?: number;
  unit?: string;
  sparkline?: number[];
}

/** Metric widget config */
export interface MetricWidgetConfig extends WidgetConfig {
  format: 'number' | 'currency' | 'percent' | 'duration' | 'bytes';
  precision?: number;
  locale?: string;
  currency?: string;
  showTrend?: boolean;
  showTarget?: boolean;
  showSparkline?: boolean;
  invertTrend?: boolean; // Lower is better
  thresholds?: {
    warning?: number;
    critical?: number;
    success?: number;
  };
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'compact' | 'card';
}

/**
 * Chart Widget
 * Display data as a chart
 */
export interface ChartWidget extends Widget<ChartData, ChartWidgetConfig> {
  type: 'chart';
}

/** Chart data */
export interface ChartData {
  labels: string[];
  datasets: ChartDataset[];
}

/** Chart dataset */
export interface ChartDataset {
  label: string;
  data: number[];
  color?: string;
  backgroundColor?: string;
  borderColor?: string;
  fill?: boolean;
  tension?: number;
  type?: ChartType; // For mixed charts
}

/** Chart type */
export type ChartType =
  | 'line'
  | 'bar'
  | 'pie'
  | 'doughnut'
  | 'area'
  | 'scatter'
  | 'radar'
  | 'polarArea'
  | 'bubble';

/** Chart widget config */
export interface ChartWidgetConfig extends WidgetConfig {
  chartType: ChartType;
  showLegend?: boolean;
  legendPosition?: 'top' | 'bottom' | 'left' | 'right';
  showGrid?: boolean;
  showXAxis?: boolean;
  showYAxis?: boolean;
  stacked?: boolean;
  horizontal?: boolean;
  aspectRatio?: number;
  animations?: boolean;
  responsive?: boolean;
  maintainAspectRatio?: boolean;
  scales?: {
    x?: AxisConfig;
    y?: AxisConfig;
  };
}

/** Axis config */
export interface AxisConfig {
  display?: boolean;
  title?: string;
  min?: number;
  max?: number;
  beginAtZero?: boolean;
  type?: 'linear' | 'logarithmic' | 'time' | 'category';
}

/**
 * Table Widget
 * Display data in a table
 */
export interface TableWidget<TRow = Record<string, unknown>>
  extends Widget<TRow[], TableWidgetConfig<TRow>> {
  type: 'table';
  pagination?: {
    page: number;
    pageSize: number;
    total: number;
  };
}

/** Table widget config */
export interface TableWidgetConfig<TRow> extends WidgetConfig {
  columns: WidgetTableColumn<TRow>[];
  sortable?: boolean;
  sortColumn?: string;
  sortDirection?: 'asc' | 'desc';
  paginated?: boolean;
  pageSize?: number;
  striped?: boolean;
  compact?: boolean;
  hoverable?: boolean;
  selectable?: boolean;
  onRowClick?: (row: TRow) => void;
}

/** Widget table column */
export interface WidgetTableColumn<TRow> {
  key: keyof TRow | string;
  header: string;
  width?: string;
  align?: 'left' | 'center' | 'right';
  sortable?: boolean;
  format?: (value: unknown, row: TRow) => string;
  component?: Component;
  className?: string;
}

/**
 * List Widget
 * Display data as a list
 */
export interface ListWidget<TItem = Record<string, unknown>>
  extends Widget<TItem[], ListWidgetConfig<TItem>> {
  type: 'list';
}

/** List widget config */
export interface ListWidgetConfig<TItem> extends WidgetConfig {
  keyField: keyof TItem | string;
  primaryField: keyof TItem | string;
  secondaryField?: keyof TItem | string;
  avatarField?: keyof TItem | string;
  imageField?: keyof TItem | string;
  statusField?: keyof TItem | string;
  timestampField?: keyof TItem | string;
  maxItems?: number;
  showMore?: boolean;
  emptyMessage?: string;
  variant?: 'simple' | 'detailed' | 'compact';
  onItemClick?: (item: TItem) => void;
}

/**
 * Calendar Widget
 * Display events in a mini calendar
 */
export interface CalendarWidget<TEvent extends CalendarWidgetEvent>
  extends Widget<TEvent[], CalendarWidgetConfig> {
  type: 'calendar';
  currentDate: Date;
  viewMode: 'month' | 'week' | 'agenda';
}

/** Calendar widget event */
export interface CalendarWidgetEvent {
  id: string;
  title: string;
  start: Date;
  end?: Date;
  allDay?: boolean;
  color?: ColorVariant | string;
}

/** Calendar widget config */
export interface CalendarWidgetConfig extends WidgetConfig {
  defaultView: 'month' | 'week' | 'agenda';
  showNavigation?: boolean;
  showWeekNumbers?: boolean;
  showToday?: boolean;
  firstDayOfWeek?: 0 | 1 | 2 | 3 | 4 | 5 | 6;
  minDate?: Date;
  maxDate?: Date;
  onDateSelect?: (date: Date) => void;
  onEventClick?: (event: CalendarWidgetEvent) => void;
}

/**
 * Timeline Widget
 * Display items in a timeline
 */
export interface TimelineWidget<TItem extends TimelineWidgetItem>
  extends Widget<TItem[], TimelineWidgetConfig> {
  type: 'timeline';
}

/** Timeline widget item */
export interface TimelineWidgetItem {
  id: string;
  timestamp: Date;
  title: string;
  description?: string;
  icon?: string;
  color?: ColorVariant;
  type?: string;
  actor?: {
    name: string;
    avatar?: string;
  };
}

/** Timeline widget config */
export interface TimelineWidgetConfig extends WidgetConfig {
  orientation: 'vertical' | 'horizontal';
  showTimestamps?: boolean;
  showAvatars?: boolean;
  groupByDate?: boolean;
  maxItems?: number;
  compact?: boolean;
  onItemClick?: (item: TimelineWidgetItem) => void;
}

/**
 * Map Widget
 * Display markers on a map
 */
export interface MapWidget<TMarker extends MapMarker>
  extends Widget<TMarker[], MapWidgetConfig> {
  type: 'map';
  center: { lat: number; lng: number };
  zoom: number;
}

/** Map marker */
export interface MapMarker {
  id: string;
  lat: number;
  lng: number;
  label?: string;
  icon?: string;
  color?: ColorVariant;
  data?: Record<string, unknown>;
}

/** Map widget config */
export interface MapWidgetConfig extends WidgetConfig {
  defaultCenter: { lat: number; lng: number };
  defaultZoom: number;
  minZoom?: number;
  maxZoom?: number;
  showControls?: boolean;
  showZoomControls?: boolean;
  clustering?: boolean;
  clusterRadius?: number;
  style?: 'standard' | 'satellite' | 'terrain' | 'dark';
  onMarkerClick?: (marker: MapMarker) => void;
}

/**
 * Progress Widget
 * Display progress/completion
 */
export interface ProgressWidget extends Widget<ProgressData, ProgressWidgetConfig> {
  type: 'progress';
}

/** Progress data */
export interface ProgressData {
  value: number;
  max: number;
  label?: string;
  sublabel?: string;
  segments?: ProgressSegment[];
}

/** Progress segment */
export interface ProgressSegment {
  value: number;
  label: string;
  color: ColorVariant;
}

/** Progress widget config */
export interface ProgressWidgetConfig extends WidgetConfig {
  variant: 'linear' | 'circular' | 'gauge';
  showValue?: boolean;
  showLabel?: boolean;
  size?: Size;
  strokeWidth?: number;
  animated?: boolean;
  thresholds?: {
    warning?: number;
    critical?: number;
    success?: number;
  };
}

/**
 * Activity Widget
 * Display recent activity feed
 */
export interface ActivityWidget<TActivity extends ActivityWidgetItem>
  extends Widget<TActivity[], ActivityWidgetConfig> {
  type: 'activity';
}

/** Activity widget item */
export interface ActivityWidgetItem {
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
}

/** Activity widget config */
export interface ActivityWidgetConfig extends WidgetConfig {
  maxItems?: number;
  showAvatars?: boolean;
  showTimestamps?: boolean;
  groupByDate?: boolean;
  filterTypes?: string[];
  onItemClick?: (item: ActivityWidgetItem) => void;
  onLoadMore?: () => Promise<void>;
}

// ============================================================================
// WIDGET SLOTS
// ============================================================================

/** Widget slots */
export interface WidgetSlots<TData = unknown> {
  header?: Snippet<[TData | null, WidgetState]>;
  content?: Snippet<[TData | null, WidgetState]>;
  footer?: Snippet<[TData | null, WidgetState]>;
  empty?: Snippet;
  loading?: Snippet;
  error?: Snippet<[WidgetError]>;
  configure?: Snippet<[WidgetConfig]>;
}
