/**
 * Report Component Types
 * Prop interfaces for each report widget component, plus shared enums and classes.
 * Mirrors the pattern of forms/formfield.types.ts and tables/table.types.ts.
 */

import type {
  ReportVisualization,
  ReportWidget,
  ReportWidgetType,
  ReportChartType,
  ReportData,
  ReportFieldFormat,
  WidgetChartConfig,
  WidgetTableConfig,
  WidgetKPIConfig,
  ReportTableColumn,
  ConditionalFormat,
  ConditionalOperator,
  DrilldownConfig,
  DrilldownFilterMapping,
  KPIAggregate,
  KPIThreshold,
} from '@samavāya/core';

// Re-export core types for convenience
export type {
  ReportVisualization,
  ReportWidget,
  ReportWidgetType,
  ReportChartType,
  ReportData,
  ReportFieldFormat,
  WidgetChartConfig,
  WidgetTableConfig,
  WidgetKPIConfig,
  ReportTableColumn,
  ConditionalFormat,
  ConditionalOperator,
  DrilldownConfig,
  DrilldownFilterMapping,
  KPIAggregate,
  KPIThreshold,
};

// ============================================================================
// SHARED
// ============================================================================

/** Trend direction computed from comparison */
export interface TrendInfo {
  value: number;
  direction: 'up' | 'down' | 'flat';
}

/** Style object produced by conditional format evaluation */
export type ConditionalStyleMap = Record<string, string>;

// ============================================================================
// ReportKPICard PROPS
// ============================================================================

export interface ReportKPICardProps {
  /** KPI widget configuration */
  config: WidgetKPIConfig;
  /** Data rows to aggregate */
  rows: Record<string, unknown>[];
  /** Pre-computed aggregates from the server (optional, takes priority) */
  aggregates?: Record<string, number>;
  /** Widget title */
  title?: string;
  /** Loading state */
  loading?: boolean;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** CSS class override */
  class?: string;
}

// ============================================================================
// ReportChart PROPS
// ============================================================================

export interface ReportChartProps {
  /** Chart widget configuration */
  config: WidgetChartConfig;
  /** Data rows */
  rows: Record<string, unknown>[];
  /** Widget title */
  title?: string;
  /** Chart height */
  height?: string;
  /** Loading state */
  loading?: boolean;
  /** ECharts theme name or object */
  theme?: string | object;
  /** CSS class override */
  class?: string;
}

// ============================================================================
// ReportTable PROPS
// ============================================================================

export interface ReportTableProps {
  /** Table widget configuration */
  config: WidgetTableConfig;
  /** Data rows */
  rows: Record<string, unknown>[];
  /** Conditional formats to apply */
  conditionalFormats?: ConditionalFormat[];
  /** Widget title */
  title?: string;
  /** Loading state */
  loading?: boolean;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** CSS class override */
  class?: string;
}

// ============================================================================
// ReportPivotTable PROPS
// ============================================================================

export interface ReportPivotTableProps {
  /** Row field codes to group by */
  rowFields: string[];
  /** Column field codes to pivot */
  colFields: string[];
  /** Value field code to aggregate */
  valueField: string;
  /** Aggregate function */
  aggregate: KPIAggregate;
  /** Data rows */
  rows: Record<string, unknown>[];
  /** Display format for values */
  format?: ReportFieldFormat;
  /** Show row/column totals */
  showTotals?: boolean;
  /** Widget title */
  title?: string;
  /** Loading state */
  loading?: boolean;
  /** CSS class override */
  class?: string;
}

// ============================================================================
// ReportSummary PROPS
// ============================================================================

export interface ReportSummaryProps {
  /** Summary metrics — each entry is a label + field_code + aggregate */
  metrics: SummaryMetric[];
  /** Data rows */
  rows: Record<string, unknown>[];
  /** Pre-computed aggregates (optional) */
  aggregates?: Record<string, number>;
  /** Layout direction */
  direction?: 'horizontal' | 'vertical';
  /** Widget title */
  title?: string;
  /** Loading state */
  loading?: boolean;
  /** CSS class override */
  class?: string;
}

export interface SummaryMetric {
  label: string;
  field_code: string;
  aggregate: KPIAggregate;
  format?: ReportFieldFormat;
  icon?: string;
}

// ============================================================================
// DynamicReportRenderer PROPS
// ============================================================================

export interface DynamicReportRendererProps {
  /** Report visualization schema (24-col grid, widgets, theme, formats, drilldowns) */
  visualization: ReportVisualization;
  /** Report data (columns + rows + aggregates) */
  data: ReportData;
  /** Global loading state */
  loading?: boolean;
  /** ECharts theme name or object */
  theme?: string | object;
  /** CSS class override */
  class?: string;
}

// ============================================================================
// CSS CLASS MAPS (like table.types.ts tableClasses)
// ============================================================================

export const reportClasses = {
  renderer: 'report-renderer',
  widget: 'report-widget',
  widgetHeader: 'report-widget__header',
  widgetTitle: 'report-widget__title',
  widgetActions: 'report-widget__actions',
  widgetBody: 'report-widget__body',
  widgetLoading: 'report-widget__loading',
  widgetEmpty: 'report-widget__empty',
} as const;

export const kpiClasses = {
  root: 'report-kpi',
  icon: 'report-kpi__icon',
  value: 'report-kpi__value',
  label: 'report-kpi__label',
  trend: 'report-kpi__trend',
  trendGood: 'report-kpi__trend--good',
  trendBad: 'report-kpi__trend--bad',
  trendFlat: 'report-kpi__trend--flat',
  trendArrow: 'report-kpi__trend-arrow',
  trendLabel: 'report-kpi__trend-label',
  sparkline: 'report-kpi__sparkline',
} as const;

export const kpiSizeClasses = {
  sm: 'report-kpi--sm',
  md: 'report-kpi--md',
  lg: 'report-kpi--lg',
} as const;

export const summaryClasses = {
  root: 'report-summary',
  horizontal: 'report-summary--horizontal',
  vertical: 'report-summary--vertical',
  metric: 'report-summary__metric',
  metricIcon: 'report-summary__metric-icon',
  metricValue: 'report-summary__metric-value',
  metricLabel: 'report-summary__metric-label',
} as const;

export const pivotClasses = {
  root: 'report-pivot',
  table: 'report-pivot__table',
  headerCell: 'report-pivot__header-cell',
  rowHeader: 'report-pivot__row-header',
  cell: 'report-pivot__cell',
  totalRow: 'report-pivot__total-row',
  totalCol: 'report-pivot__total-col',
  grandTotal: 'report-pivot__grand-total',
} as const;
