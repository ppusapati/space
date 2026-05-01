/**
 * Report Logic — shared utility functions for report widget components.
 * Mirrors the pattern of tables/table.logic.ts.
 */

import type {
  ReportChartType,
  ReportFieldFormat,
  ConditionalFormat,
  KPIAggregate,
  TrendInfo,
  ConditionalStyleMap,
  WidgetChartConfig,
} from './report.types';

import type { EChartsOption } from 'echarts';
import { defaultChartColors, formatNumber, formatCurrency, formatPercent } from '../charts/types';

// ============================================================================
// AGGREGATION
// ============================================================================

/**
 * Compute an aggregate value from a set of rows.
 * Used by KPI cards, summary widgets, and pivot tables.
 */
export function computeAggregate(
  rows: Record<string, unknown>[],
  fieldCode: string,
  aggregate: KPIAggregate
): number {
  const values = rows
    .map((r) => Number(r[fieldCode]))
    .filter((v) => !isNaN(v));

  if (values.length === 0) return 0;

  switch (aggregate) {
    case 'sum':
      return values.reduce((a, b) => a + b, 0);
    case 'avg':
      return values.reduce((a, b) => a + b, 0) / values.length;
    case 'min':
      return Math.min(...values);
    case 'max':
      return Math.max(...values);
    case 'count':
      return values.length;
    case 'last':
      return values[values.length - 1];
    default:
      return values.reduce((a, b) => a + b, 0);
  }
}

// ============================================================================
// FORMATTING
// ============================================================================

/**
 * Format a numeric value using a ReportFieldFormat descriptor.
 */
export function formatValue(
  value: number,
  format?: ReportFieldFormat | null
): string {
  if (format == null) return formatNumber(value);

  const dp = format.decimal_places ?? 0;
  const prefix = format.prefix ?? '';
  const suffix = format.suffix ?? '';

  switch (format.type) {
    case 'currency':
      return prefix + formatCurrency(value, format.currency_code ?? '₹', dp) + suffix;
    case 'percent':
      return prefix + formatPercent(value / 100, dp) + suffix;
    case 'number':
      return (
        prefix +
        value.toLocaleString('en-IN', {
          minimumFractionDigits: dp,
          maximumFractionDigits: dp,
        }) +
        suffix
      );
    case 'date':
    case 'datetime':
    case 'text':
    case 'boolean':
      return prefix + String(value) + suffix;
    default:
      return prefix + formatNumber(value) + suffix;
  }
}

// ============================================================================
// TREND
// ============================================================================

/**
 * Compute trend info between a current and comparison value.
 * Returns direction (up/down/flat) and absolute percentage.
 */
export function computeTrend(current: number, comparison: number): TrendInfo {
  if (comparison === 0) return { value: 0, direction: 'flat' };
  const pct = ((current - comparison) / Math.abs(comparison)) * 100;
  return {
    value: Math.abs(pct),
    direction: pct > 0.5 ? 'up' : pct < -0.5 ? 'down' : 'flat',
  };
}

/**
 * Whether the trend is "good" based on trend_direction config.
 */
export function isTrendGood(
  trend: TrendInfo,
  trendDirection?: 'up_is_good' | 'down_is_good'
): boolean {
  if (trend.direction === 'flat') return true;
  if (trendDirection === 'down_is_good') return trend.direction === 'down';
  return trend.direction === 'up';
}

// ============================================================================
// CONDITIONAL FORMATTING
// ============================================================================

/**
 * Evaluate conditional format rules against a row and return inline styles.
 */
export function evaluateConditionalFormats(
  row: Record<string, unknown>,
  formats: ConditionalFormat[] | undefined
): ConditionalStyleMap {
  if (!formats?.length) return {};
  const style: ConditionalStyleMap = {};

  const sorted = [...formats].sort((a, b) => (a.priority ?? 0) - (b.priority ?? 0));

  for (const fmt of sorted) {
    const val = row[fmt.field_code];
    let match = false;

    switch (fmt.operator) {
      case 'eq':
        match = String(val) === fmt.value;
        break;
      case 'neq':
        match = String(val) !== fmt.value;
        break;
      case 'gt':
        match = Number(val) > Number(fmt.value);
        break;
      case 'gte':
        match = Number(val) >= Number(fmt.value);
        break;
      case 'lt':
        match = Number(val) < Number(fmt.value);
        break;
      case 'lte':
        match = Number(val) <= Number(fmt.value);
        break;
      case 'between':
        match =
          Number(val) >= Number(fmt.value) &&
          Number(val) <= Number(fmt.value2 ?? fmt.value);
        break;
      case 'contains':
        match = String(val).includes(fmt.value);
        break;
      case 'not_contains':
        match = !String(val).includes(fmt.value);
        break;
      case 'is_null':
        match = val == null;
        break;
      case 'is_not_null':
        match = val != null;
        break;
    }

    if (match) {
      if (fmt.style.background_color) style['background-color'] = fmt.style.background_color;
      if (fmt.style.text_color) style['color'] = fmt.style.text_color;
      if (fmt.style.font_weight) style['font-weight'] = fmt.style.font_weight;
    }
  }

  return style;
}

/**
 * Convert a ConditionalStyleMap to an inline style string.
 */
export function styleMapToString(styleMap: ConditionalStyleMap): string {
  return Object.entries(styleMap)
    .map(([k, v]) => `${k}: ${v}`)
    .join('; ');
}

// ============================================================================
// CHART OPTION BUILDER
// ============================================================================

/** Map proto ReportChartType to ECharts series type */
export function mapChartType(type: ReportChartType): string {
  const map: Record<string, string> = {
    bar: 'bar',
    stacked_bar: 'bar',
    horizontal_bar: 'bar',
    line: 'line',
    area: 'line',
    pie: 'pie',
    doughnut: 'pie',
    scatter: 'scatter',
    radar: 'radar',
    gauge: 'gauge',
    funnel: 'funnel',
    heatmap: 'heatmap',
    treemap: 'treemap',
    sankey: 'sankey',
    waterfall: 'bar',
  };
  return map[type] ?? 'bar';
}

/**
 * Build a full EChartsOption from a WidgetChartConfig + data rows.
 * Handles all 15 chart types, grouped series, stacking, area fills,
 * horizontal bars, and ECharts overrides.
 */
export function buildChartOption(
  config: WidgetChartConfig,
  rows: Record<string, unknown>[]
): EChartsOption {
  const xField = config.x_axis_field_code;
  const yFields = config.y_axis_field_codes ?? [];
  const seriesField = config.series_field_code;
  const colors = config.color_palette?.length ? config.color_palette : defaultChartColors;
  const isPie = ['pie', 'doughnut', 'funnel'].includes(config.chart_type);
  const isHorizontal = config.chart_type === 'horizontal_bar';
  const isArea = config.chart_type === 'area';

  // Extract unique categories from x-axis field
  const categories = [...new Set(rows.map((r) => String(r[xField] ?? '')))];

  let seriesList: Record<string, unknown>[] = [];

  if (isPie && yFields.length > 0) {
    // Pie / Doughnut / Funnel → name-value pairs
    const pieData = categories.map((cat) => {
      const val = rows
        .filter((r) => String(r[xField]) === cat)
        .reduce((sum, r) => sum + Number(r[yFields[0]] ?? 0), 0);
      return { name: cat, value: val };
    });
    seriesList = [
      {
        type: config.chart_type === 'funnel' ? 'funnel' : 'pie',
        data: pieData,
        ...(config.chart_type === 'doughnut' ? { radius: ['40%', '70%'] } : { radius: '70%' }),
        label: config.show_data_labels !== false ? { show: true } : { show: false },
      },
    ];
  } else if (seriesField) {
    // Grouped series — one series per unique value in series_field_code
    const groups = new Map<string, Map<string, number>>();
    for (const row of rows) {
      const groupKey = String(row[seriesField] ?? '');
      const cat = String(row[xField] ?? '');
      if (!groups.has(groupKey)) groups.set(groupKey, new Map());
      const yVal = Number(row[yFields[0]] ?? 0);
      const existing = groups.get(groupKey)!.get(cat) ?? 0;
      groups.get(groupKey)!.set(cat, existing + yVal);
    }
    seriesList = [...groups.entries()].map(([name, catMap]) => ({
      name,
      type: mapChartType(config.chart_type),
      data: categories.map((c) => catMap.get(c) ?? 0),
      ...(config.stacking && config.stacking !== 'none' ? { stack: 'stack' } : {}),
      ...(isArea ? { areaStyle: { opacity: 0.3 } } : {}),
      label: config.show_data_labels ? { show: true } : undefined,
    }));
  } else {
    // One series per y-axis field code
    for (const yField of yFields) {
      const valMap = new Map<string, number>();
      for (const row of rows) {
        const cat = String(row[xField] ?? '');
        valMap.set(cat, (valMap.get(cat) ?? 0) + Number(row[yField] ?? 0));
      }
      seriesList.push({
        name: yField,
        type: mapChartType(config.chart_type),
        data: categories.map((c) => valMap.get(c) ?? 0),
        ...(config.stacking && config.stacking !== 'none' ? { stack: 'stack' } : {}),
        ...(isArea ? { areaStyle: { opacity: 0.3 } } : {}),
        label: config.show_data_labels ? { show: true } : undefined,
      });
    }
  }

  const option: EChartsOption = {
    color: colors,
    tooltip: { trigger: isPie ? 'item' : 'axis' },
    legend: { show: config.show_legend !== false },
    ...(isPie
      ? {}
      : {
          grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
          xAxis: isHorizontal
            ? { type: 'value' as const }
            : { type: 'category' as const, data: categories },
          yAxis: isHorizontal
            ? { type: 'category' as const, data: categories }
            : { type: 'value' as const },
        }),
    series: seriesList as EChartsOption['series'],
  };

  // Merge raw ECharts overrides last
  if (config.echarts_override) {
    Object.assign(option, config.echarts_override);
  }

  return option;
}

// ============================================================================
// TABLE COLUMN MAPPING
// ============================================================================

import type { TableColumn } from '../tables/table.types';
import type { WidgetTableConfig, ReportTableColumn as ProtoColumn } from './report.types';

/**
 * Convert proto ReportTableColumn[] to DataGrid TableColumn[].
 */
export function buildTableColumns(config: WidgetTableConfig): TableColumn[] {
  return config.columns.map((col) => ({
    key: col.field_code,
    header: col.header,
    width: col.width ? `${col.width}px` : undefined,
    minWidth: col.min_width ? `${col.min_width}px` : undefined,
    align: col.align ?? 'left',
    sortable: col.sortable !== false,
    filterable: col.filterable ?? false,
    sticky: col.pinned === 'left' ? 'left' as const : col.pinned === 'right' ? 'right' as const : undefined,
    format: col.format
      ? (value: unknown) => formatValue(Number(value), col.format)
      : undefined,
  }));
}

// ============================================================================
// GRID LAYOUT
// ============================================================================

/**
 * Compute CSS grid placement for a ReportWidget.
 */
export function widgetGridStyle(widget: {
  grid_col: number;
  grid_row: number;
  grid_col_span: number;
  grid_row_span: number;
}): string {
  return [
    `grid-column: ${widget.grid_col} / span ${widget.grid_col_span}`,
    `grid-row: ${widget.grid_row} / span ${widget.grid_row_span}`,
  ].join('; ');
}

// ============================================================================
// PIVOT TABLE
// ============================================================================

export interface PivotResult {
  rowKeys: string[];
  colKeys: string[];
  cells: Map<string, number>; // "rowKey|colKey" → value
  rowTotals: Map<string, number>;
  colTotals: Map<string, number>;
  grandTotal: number;
}

/**
 * Compute a pivot table from flat rows.
 */
export function computePivot(
  rows: Record<string, unknown>[],
  rowFields: string[],
  colFields: string[],
  valueField: string,
  aggregate: KPIAggregate
): PivotResult {
  // Group rows by composite row key and col key
  const groups = new Map<string, number[]>();
  const rowKeySet = new Set<string>();
  const colKeySet = new Set<string>();

  for (const row of rows) {
    const rk = rowFields.map((f) => String(row[f] ?? '')).join(' | ');
    const ck = colFields.map((f) => String(row[f] ?? '')).join(' | ');
    rowKeySet.add(rk);
    colKeySet.add(ck);

    const key = `${rk}|${ck}`;
    if (!groups.has(key)) groups.set(key, []);
    const val = Number(row[valueField] ?? 0);
    if (!isNaN(val)) groups.get(key)!.push(val);
  }

  const rowKeys = [...rowKeySet].sort();
  const colKeys = [...colKeySet].sort();

  // Compute cell values
  const cells = new Map<string, number>();
  for (const [key, values] of groups) {
    cells.set(key, computeAggregate(
      values.map((v) => ({ [valueField]: v })),
      valueField,
      aggregate
    ));
  }

  // Row totals
  const rowTotals = new Map<string, number>();
  for (const rk of rowKeys) {
    let total = 0;
    for (const ck of colKeys) {
      total += cells.get(`${rk}|${ck}`) ?? 0;
    }
    rowTotals.set(rk, total);
  }

  // Col totals
  const colTotals = new Map<string, number>();
  for (const ck of colKeys) {
    let total = 0;
    for (const rk of rowKeys) {
      total += cells.get(`${rk}|${ck}`) ?? 0;
    }
    colTotals.set(ck, total);
  }

  // Grand total
  let grandTotal = 0;
  for (const v of rowTotals.values()) grandTotal += v;

  return { rowKeys, colKeys, cells, rowTotals, colTotals, grandTotal };
}
