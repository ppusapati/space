// =============================================================================
// Report Components
// =============================================================================

// Renderer (maps widget types to components — like DynamicFormRenderer)
export { default as DynamicReportRenderer } from './DynamicReportRenderer.svelte';

// Individual widget components (each standalone & reusable)
export { default as ReportKPICard } from './ReportKPICard.svelte';
export { default as ReportChart } from './ReportChart.svelte';
export { default as ReportTable } from './ReportTable.svelte';
export { default as ReportPivotTable } from './ReportPivotTable.svelte';
export { default as ReportSummary } from './ReportSummary.svelte';

// Export menu component
export { default as ReportExportMenu } from './ReportExportMenu.svelte';

// =============================================================================
// Types
// =============================================================================

export type {
  ReportKPICardProps,
  ReportChartProps,
  ReportTableProps,
  ReportPivotTableProps,
  ReportSummaryProps,
  DynamicReportRendererProps,
  SummaryMetric,
  TrendInfo,
  ConditionalStyleMap,
} from './report.types';

export {
  reportClasses,
  kpiClasses,
  kpiSizeClasses,
  summaryClasses,
  pivotClasses,
} from './report.types';

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
  DrilldownConfig,
  KPIAggregate,
  KPIThreshold,
} from './report.types';

// =============================================================================
// Logic (shared utilities — like table.logic.ts)
// =============================================================================

export {
  computeAggregate,
  formatValue,
  computeTrend,
  isTrendGood,
  evaluateConditionalFormats,
  styleMapToString,
  buildChartOption,
  mapChartType,
  buildTableColumns,
  widgetGridStyle,
  computePivot,
} from './report.logic';

export type { PivotResult } from './report.logic';

// =============================================================================
// Export (CSV, XLSX, PDF, Print — like table.export.ts)
// =============================================================================

export {
  exportReport,
  exportWidget,
  exportReportCSV,
  exportReportXLSX,
  exportReportPDF,
  printReport,
} from './report.export';

export type {
  ReportExportFormat,
  ReportExportOptions,
  WidgetExportOptions,
} from './report.export';
