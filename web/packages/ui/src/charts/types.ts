import type { EChartsOption, ECharts as EChartsInstance } from 'echarts';

// Re-export ECharts types
export type { EChartsOption, EChartsInstance };

// Animation configuration
export interface ChartAnimationConfig {
  enabled?: boolean;
  duration?: number;
  easing?: 'linear' | 'quadraticIn' | 'quadraticOut' | 'quadraticInOut' |
           'cubicIn' | 'cubicOut' | 'cubicInOut' |
           'quarticIn' | 'quarticOut' | 'quarticInOut' |
           'elasticOut' | 'bounceOut';
  delay?: number | ((idx: number) => number);
  threshold?: number;
}

// Gradient configuration
export interface GradientStop {
  offset: number;
  color: string;
}

export interface LinearGradient {
  type: 'linear';
  x: number;
  y: number;
  x2: number;
  y2: number;
  colorStops: GradientStop[];
}

export interface RadialGradient {
  type: 'radial';
  x: number;
  y: number;
  r: number;
  colorStops: GradientStop[];
}

export type ChartGradient = LinearGradient | RadialGradient;

// Data label configuration
export interface DataLabelConfig {
  show?: boolean;
  position?: 'top' | 'bottom' | 'left' | 'right' | 'inside' | 'insideLeft' |
             'insideRight' | 'insideTop' | 'insideBottom' | 'insideTopLeft' |
             'insideTopRight' | 'insideBottomLeft' | 'insideBottomRight';
  formatter?: string | ((params: unknown) => string);
  fontSize?: number;
  fontWeight?: 'normal' | 'bold' | number;
  color?: string;
  rotate?: number;
  offset?: [number, number];
}

// Axis configuration
export interface AxisConfig {
  type?: 'category' | 'value' | 'time' | 'log';
  name?: string;
  nameLocation?: 'start' | 'middle' | 'end';
  nameGap?: number;
  min?: number | 'dataMin' | 'auto';
  max?: number | 'dataMax' | 'auto';
  inverse?: boolean;
  splitNumber?: number;
  interval?: number;
  logBase?: number;
  axisLabel?: {
    rotate?: number;
    formatter?: string | ((value: unknown, index: number) => string);
    interval?: number | 'auto';
  };
}

// Data zoom configuration
export interface DataZoomConfig {
  type?: 'slider' | 'inside';
  show?: boolean;
  start?: number;
  end?: number;
  orient?: 'horizontal' | 'vertical';
  xAxisIndex?: number | number[];
  yAxisIndex?: number | number[];
  filterMode?: 'filter' | 'weakFilter' | 'empty' | 'none';
  throttle?: number;
}

// Tooltip configuration
export interface TooltipConfig {
  show?: boolean;
  trigger?: 'item' | 'axis' | 'none';
  formatter?: string | ((params: unknown) => string);
  axisPointer?: {
    type?: 'line' | 'shadow' | 'cross' | 'none';
    snap?: boolean;
  };
  confine?: boolean;
  position?: 'inside' | 'top' | 'left' | 'right' | 'bottom' |
             ((point: number[], params: unknown, dom: HTMLElement, rect: unknown, size: unknown) => number[] | string);
}

// Legend configuration
export interface LegendConfig {
  show?: boolean;
  type?: 'plain' | 'scroll';
  orient?: 'horizontal' | 'vertical';
  position?: 'top' | 'bottom' | 'left' | 'right';
  align?: 'auto' | 'left' | 'right';
  itemGap?: number;
  itemWidth?: number;
  itemHeight?: number;
  icon?: 'circle' | 'rect' | 'roundRect' | 'triangle' | 'diamond' | 'pin' | 'arrow' | 'none' | string;
  selectedMode?: boolean | 'single' | 'multiple';
}

// Common series data interfaces
export interface LineSeriesData {
  name: string;
  data: number[];
  smooth?: boolean;
  areaStyle?: boolean | { opacity?: number; color?: string | ChartGradient };
  color?: string | ChartGradient;
  lineStyle?: {
    width?: number;
    type?: 'solid' | 'dashed' | 'dotted';
  };
  symbol?: 'circle' | 'rect' | 'roundRect' | 'triangle' | 'diamond' | 'pin' | 'arrow' | 'none';
  symbolSize?: number;
  showSymbol?: boolean;
  step?: 'start' | 'middle' | 'end' | false;
  connectNulls?: boolean;
}

export interface BarSeriesData {
  name: string;
  data: number[];
  color?: string | ChartGradient;
  borderRadius?: number | [number, number, number, number];
  backgroundStyle?: {
    color?: string;
    borderRadius?: number;
  };
}

export interface PieDataItem {
  name: string;
  value: number;
  color?: string;
  selected?: boolean;
  label?: {
    show?: boolean;
    formatter?: string | ((params: unknown) => string);
  };
}

export interface ScatterSeriesData {
  name: string;
  data: [number, number][] | [number, number, number][];
  color?: string;
  symbolSize?: number | ((value: number[]) => number);
  symbol?: 'circle' | 'rect' | 'roundRect' | 'triangle' | 'diamond' | 'pin' | 'arrow' | 'none';
}

export interface RadarIndicator {
  name: string;
  max: number;
  min?: number;
}

export interface RadarSeriesData {
  name: string;
  value: number[];
  color?: string;
  areaStyle?: boolean | { opacity?: number };
  lineStyle?: {
    width?: number;
    type?: 'solid' | 'dashed' | 'dotted';
  };
}

export interface TreemapNode {
  name: string;
  value: number;
  children?: TreemapNode[];
  itemStyle?: {
    color?: string;
  };
}

// Heatmap data
export interface HeatmapDataItem {
  x: string | number;
  y: string | number;
  value: number;
}

// Funnel data
export interface FunnelDataItem {
  name: string;
  value: number;
  color?: string;
}

// Sankey data
export interface SankeyNode {
  name: string;
  value?: number;
  itemStyle?: { color?: string };
}

export interface SankeyLink {
  source: string;
  target: string;
  value: number;
  lineStyle?: { color?: string; opacity?: number };
}

// Candlestick data
export interface CandlestickDataItem {
  date: string;
  open: number;
  close: number;
  low: number;
  high: number;
  volume?: number;
}

// Waterfall data
export interface WaterfallDataItem {
  name: string;
  value: number;
  isTotal?: boolean;
  color?: string;
}

// Export options
export interface ChartExportOptions {
  type?: 'png' | 'jpeg' | 'svg';
  pixelRatio?: number;
  backgroundColor?: string;
  excludeComponents?: string[];
}

// Theme colors
export interface ChartThemeColors {
  primary: string;
  success: string;
  warning: string;
  error: string;
  info: string;
  series: string[];
}

// Default chart colors
export const defaultChartColors = [
  '#5470c6', // Blue
  '#91cc75', // Green
  '#fac858', // Yellow
  '#ee6666', // Red
  '#73c0de', // Light Blue
  '#3ba272', // Dark Green
  '#fc8452', // Orange
  '#9a60b4', // Purple
  '#ea7ccc', // Pink
  '#48b8d0', // Cyan
];

// Utility: Create linear gradient
export function createLinearGradient(
  direction: 'vertical' | 'horizontal',
  colors: string[]
): LinearGradient {
  const colorStops = colors.map((color, index) => ({
    offset: index / (colors.length - 1),
    color,
  }));

  return direction === 'vertical'
    ? { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops }
    : { type: 'linear', x: 0, y: 0, x2: 1, y2: 0, colorStops };
}

// Utility: Create radial gradient
export function createRadialGradient(colors: string[]): RadialGradient {
  const colorStops = colors.map((color, index) => ({
    offset: index / (colors.length - 1),
    color,
  }));

  return { type: 'radial', x: 0.5, y: 0.5, r: 0.5, colorStops };
}

// Utility: Format number with K/M/B suffix
export function formatNumber(value: number, decimals = 1): string {
  if (value >= 1e9) return (value / 1e9).toFixed(decimals) + 'B';
  if (value >= 1e6) return (value / 1e6).toFixed(decimals) + 'M';
  if (value >= 1e3) return (value / 1e3).toFixed(decimals) + 'K';
  return value.toString();
}

// Utility: Format percentage
export function formatPercent(value: number, decimals = 1): string {
  return (value * 100).toFixed(decimals) + '%';
}

// Utility: Format currency
export function formatCurrency(value: number, currency = '₹', decimals = 0): string {
  return currency + value.toLocaleString('en-IN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  });
}
