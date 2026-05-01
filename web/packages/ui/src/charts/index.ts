// =============================================================================
// ECharts-based Chart Components
// =============================================================================

// Base Chart Component
export { default as EChart } from './EChart.svelte';

// Basic Charts
export { default as LineChart } from './LineChart.svelte';
export { default as BarChart } from './BarChart.svelte';
export { default as PieChart } from './PieChart.svelte';
export { default as ScatterChart } from './ScatterChart.svelte';

// Statistical Charts
export { default as GaugeChart } from './GaugeChart.svelte';
export { default as RadarChart } from './RadarChart.svelte';
export { default as HeatmapChart } from './HeatmapChart.svelte';

// Hierarchical Charts
export { default as TreemapChart } from './TreemapChart.svelte';
export { default as FunnelChart } from './FunnelChart.svelte';
export { default as SankeyChart } from './SankeyChart.svelte';

// Financial Charts
export { default as CandlestickChart } from './CandlestickChart.svelte';
export { default as WaterfallChart } from './WaterfallChart.svelte';

// Geographic Charts
export { default as MapChart } from './MapChart.svelte';

// =============================================================================
// Types - ECharts Core
// =============================================================================
export type { EChartsOption, ECharts } from 'echarts';
export type { EChartsInstance } from './types';

// =============================================================================
// Types - Configuration Interfaces
// =============================================================================
export type {
  ChartAnimationConfig,
  GradientStop,
  LinearGradient,
  RadialGradient,
  ChartGradient,
  DataLabelConfig,
  AxisConfig,
  DataZoomConfig,
  TooltipConfig,
  LegendConfig,
  ChartExportOptions,
  ChartThemeColors,
} from './types';

// =============================================================================
// Types - Series Data Interfaces
// =============================================================================
export type {
  LineSeriesData,
  BarSeriesData,
  PieDataItem,
  ScatterSeriesData,
  RadarIndicator,
  RadarSeriesData,
  TreemapNode,
  HeatmapDataItem,
  FunnelDataItem,
  SankeyNode,
  SankeyLink,
  CandlestickDataItem,
  WaterfallDataItem,
} from './types';

// Types - Map Chart
export type { MapDataItem, MapType } from './MapChart.types';

// =============================================================================
// Utilities & Constants
// =============================================================================
export {
  defaultChartColors,
  createLinearGradient,
  createRadialGradient,
  formatNumber,
  formatPercent,
  formatCurrency,
} from './types';
