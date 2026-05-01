/**
 * Report Types — matches insighthub.proto ReportVisualization schema
 * Used by DynamicReportRenderer to render report widgets in a 24-column grid.
 */

import type { ChartType } from './page.types.js';

// ============================================================================
// CHART TYPE (extended to match proto's 15 chart types)
// ============================================================================

/**
 * All chart types supported by the report renderer.
 * Maps 1:1 with insighthub.proto ChartDisplayType enum.
 */
export type ReportChartType =
  | 'bar'
  | 'stacked_bar'
  | 'horizontal_bar'
  | 'line'
  | 'area'
  | 'pie'
  | 'doughnut'
  | 'scatter'
  | 'radar'
  | 'gauge'
  | 'funnel'
  | 'heatmap'
  | 'treemap'
  | 'sankey'
  | 'waterfall';

// ============================================================================
// WIDGET TYPES
// ============================================================================

export type ReportWidgetType = 'chart' | 'table' | 'kpi_card' | 'pivot_table' | 'summary';

/** Layout mode for the report visualization */
export type ReportLayoutMode = 'grid' | 'flow' | 'tabs';

// ============================================================================
// REPORT VISUALIZATION (top-level)
// ============================================================================

/** Top-level report visualization definition — stored as JSONB in insights.reports */
export interface ReportVisualization {
  layout_mode: ReportLayoutMode;
  widgets: ReportWidget[];
  theme?: string;
  conditional_formats?: ConditionalFormat[];
  drilldowns?: DrilldownConfig[];
}

// ============================================================================
// REPORT WIDGET
// ============================================================================

/** A single widget in the report layout (24-column grid) */
export interface ReportWidget {
  widget_id: string;
  title: string;
  widget_type: ReportWidgetType;

  /** 24-column grid positioning */
  grid_col: number;
  grid_row: number;
  grid_col_span: number;
  grid_row_span: number;

  /** Type-specific config — exactly one should be set */
  chart_config?: WidgetChartConfig;
  table_config?: WidgetTableConfig;
  kpi_config?: WidgetKPIConfig;

  /** Dataset field codes used by this widget (for dependency tracking) */
  field_codes?: string[];

  /** CSS class override */
  css_class?: string;

  /** Whether the widget is hidden */
  hidden?: boolean;
}

// ============================================================================
// CHART CONFIG
// ============================================================================

/** Chart widget configuration */
export interface WidgetChartConfig {
  chart_type: ReportChartType;
  x_axis_field_code: string;
  y_axis_field_codes: string[];
  series_field_code?: string;
  color_palette?: string[];
  show_legend?: boolean;
  show_data_labels?: boolean;
  stacking?: 'none' | 'normal' | 'percent';
  /** Raw ECharts option overrides (merged last) */
  echarts_override?: Record<string, unknown>;
}

// ============================================================================
// TABLE CONFIG
// ============================================================================

/** Table widget configuration */
export interface WidgetTableConfig {
  columns: ReportTableColumn[];
  default_sort_field?: string;
  default_sort_direction?: 'asc' | 'desc';
  group_by_field?: string;
  show_totals?: boolean;
  row_limit?: number;
  exportable?: boolean;
  paginated?: boolean;
  page_size?: number;
}

/** Column definition for table widgets */
export interface ReportTableColumn {
  field_code: string;
  header: string;
  width?: number;
  min_width?: number;
  align?: 'left' | 'center' | 'right';
  pinned?: 'left' | 'right' | 'none';
  format?: ReportFieldFormat;
  sortable?: boolean;
  filterable?: boolean;
}

/** Field display format */
export interface ReportFieldFormat {
  type: 'number' | 'currency' | 'percent' | 'date' | 'datetime' | 'text' | 'boolean';
  currency_code?: string;
  decimal_places?: number;
  date_format?: string;
  prefix?: string;
  suffix?: string;
}

// ============================================================================
// KPI CARD CONFIG
// ============================================================================

/** KPI card widget configuration */
export interface WidgetKPIConfig {
  value_field_code: string;
  aggregate: KPIAggregate;
  label: string;
  format?: ReportFieldFormat;
  icon?: string;
  color?: string;

  /** Comparison / trend */
  comparison_field_code?: string;
  comparison_label?: string;
  trend_direction?: 'up_is_good' | 'down_is_good';

  /** Sparkline (mini chart) */
  sparkline_field_code?: string;
  sparkline_type?: 'line' | 'bar';

  /** Thresholds for color coding */
  thresholds?: KPIThreshold[];
}

export type KPIAggregate = 'sum' | 'avg' | 'min' | 'max' | 'count' | 'last';

export interface KPIThreshold {
  value: number;
  color: string;
  label?: string;
}

// ============================================================================
// CONDITIONAL FORMAT
// ============================================================================

/** Conditional formatting rule applied to report data */
export interface ConditionalFormat {
  id: string;
  field_code: string;
  operator: ConditionalOperator;
  value: string;
  value2?: string;
  style: ConditionalStyle;
  priority?: number;
}

export type ConditionalOperator =
  | 'eq' | 'neq'
  | 'gt' | 'gte' | 'lt' | 'lte'
  | 'between'
  | 'contains' | 'not_contains'
  | 'is_null' | 'is_not_null';

export interface ConditionalStyle {
  background_color?: string;
  text_color?: string;
  font_weight?: 'normal' | 'bold';
  icon?: string;
}

// ============================================================================
// DRILLDOWN
// ============================================================================

/** Drilldown configuration — click a cell to navigate to another report */
export interface DrilldownConfig {
  source_widget_id: string;
  source_field_code: string;
  target_report_id: string;
  target_dataset_id?: string;
  filter_mappings: DrilldownFilterMapping[];
}

export interface DrilldownFilterMapping {
  source_field_code: string;
  target_parameter: string;
}

// ============================================================================
// REPORT DATA (runtime — what the API returns)
// ============================================================================

/** Report data payload returned by the API */
export interface ReportData {
  /** Column metadata */
  columns: ReportDataColumn[];
  /** Row data — array of field_code → value maps */
  rows: Record<string, unknown>[];
  /** Total row count (before pagination) */
  total_rows: number;
  /** Aggregated values (for KPI widgets) */
  aggregates?: Record<string, number>;
}

export interface ReportDataColumn {
  field_code: string;
  label: string;
  data_type: 'string' | 'number' | 'boolean' | 'date' | 'datetime';
}
