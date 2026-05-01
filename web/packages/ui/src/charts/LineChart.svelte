<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import * as echarts from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import type {
    LineSeriesData,
    DataLabelConfig,
    ChartAnimationConfig,
    DataZoomConfig,
    AxisConfig,
    TooltipConfig,
    LegendConfig,
  } from './types';

  // Props
  export let categories: string[] = [];
  export let series: LineSeriesData[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showLegend: boolean = true;
  export let showTooltip: boolean = true;
  export let showGrid: boolean = true;
  export let xAxisLabel: string = '';
  export let yAxisLabel: string = '';
  export let height: string = '400px';
  export let loading: boolean = false;
  export let stacked: boolean = false;

  // Enhanced props
  export let dataLabels: DataLabelConfig | boolean = false;
  export let animation: ChartAnimationConfig | boolean = true;
  export let dataZoom: DataZoomConfig | DataZoomConfig[] | boolean = false;
  export let colors: string[] = [];
  export let xAxis: AxisConfig = {};
  export let yAxis: AxisConfig = {};
  export let tooltip: TooltipConfig = {};
  export let legend: LegendConfig = {};
  export let markLine: { data: Array<{ type?: 'average' | 'min' | 'max'; yAxis?: number; name?: string }> } | null = null;
  export let markPoint: { data: Array<{ type?: 'max' | 'min'; name?: string }> } | null = null;

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher();

  // Build data label config
  function getDataLabelConfig(): object | undefined {
    if (!dataLabels) return undefined;
    if (dataLabels === true) {
      return {
        show: true,
        position: 'top',
        color: 'var(--color-text-secondary)',
        fontSize: 11,
      };
    }
    return {
      show: dataLabels.show !== false,
      position: dataLabels.position ?? 'top',
      formatter: dataLabels.formatter,
      fontSize: dataLabels.fontSize ?? 11,
      fontWeight: dataLabels.fontWeight,
      color: dataLabels.color ?? 'var(--color-text-secondary)',
      rotate: dataLabels.rotate,
      offset: dataLabels.offset,
    };
  }

  // Build gradient color
  function buildColor(color: string | LineSeriesData['color']) {
    if (!color) return undefined;
    if (typeof color === 'string') return color;
    if (color.type === 'linear') {
      return new echarts.graphic.LinearGradient(
        color.x, color.y, color.x2, color.y2,
        color.colorStops
      );
    }
    if (color.type === 'radial') {
      return new echarts.graphic.RadialGradient(
        color.x, color.y, color.r,
        color.colorStops
      );
    }
    return undefined;
  }

  // Build area style
  function buildAreaStyle(areaStyle: LineSeriesData['areaStyle'], color?: string | LineSeriesData['color']) {
    if (!areaStyle) return undefined;
    if (areaStyle === true) return { opacity: 0.3 };
    return {
      opacity: areaStyle.opacity ?? 0.3,
      color: areaStyle.color ? buildColor(areaStyle.color) : undefined,
    };
  }

  // Get legend position
  function getLegendPosition() {
    const pos = legend.position ?? 'bottom';
    switch (pos) {
      case 'top': return { top: 0 };
      case 'bottom': return { bottom: 0 };
      case 'left': return { left: 0, orient: 'vertical' };
      case 'right': return { right: 0, orient: 'vertical' };
      default: return { bottom: 0 };
    }
  }

  $: option = {
    title: title ? {
      text: title,
      subtext: subtitle,
      left: 'center',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      subtextStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    tooltip: showTooltip || tooltip.show !== false ? {
      trigger: tooltip.trigger ?? 'axis',
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      formatter: tooltip.formatter,
      confine: tooltip.confine ?? true,
      axisPointer: {
        type: tooltip.axisPointer?.type ?? 'line',
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
    } : undefined,
    legend: showLegend ? {
      ...getLegendPosition(),
      type: legend.type ?? 'plain',
      orient: legend.orient,
      align: legend.align,
      itemGap: legend.itemGap ?? 10,
      itemWidth: legend.itemWidth ?? 25,
      itemHeight: legend.itemHeight ?? 14,
      icon: legend.icon,
      selectedMode: legend.selectedMode,
      textStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    grid: {
      left: '3%',
      right: '4%',
      bottom: showLegend ? '15%' : '3%',
      top: title ? '15%' : '10%',
      containLabel: true,
      show: showGrid,
      borderColor: 'var(--color-border-primary)',
    },
    xAxis: {
      type: xAxis.type ?? 'category',
      boundaryGap: false,
      data: categories,
      name: xAxis.name ?? xAxisLabel,
      nameLocation: xAxis.nameLocation ?? 'middle',
      nameGap: xAxis.nameGap ?? 30,
      min: xAxis.min,
      max: xAxis.max,
      inverse: xAxis.inverse,
      splitNumber: xAxis.splitNumber,
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      axisLabel: {
        color: 'var(--color-text-secondary)',
        rotate: xAxis.axisLabel?.rotate,
        formatter: xAxis.axisLabel?.formatter,
        interval: xAxis.axisLabel?.interval,
      },
    },
    yAxis: {
      type: yAxis.type ?? 'value',
      name: yAxis.name ?? yAxisLabel,
      nameLocation: yAxis.nameLocation ?? 'middle',
      nameGap: yAxis.nameGap ?? 50,
      min: yAxis.min,
      max: yAxis.max,
      inverse: yAxis.inverse,
      splitNumber: yAxis.splitNumber,
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      axisLabel: {
        color: 'var(--color-text-secondary)',
        formatter: yAxis.axisLabel?.formatter,
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-border-secondary)',
        },
      },
    },
    series: series.map((s) => ({
      name: s.name,
      type: 'line',
      stack: stacked ? 'Total' : undefined,
      smooth: s.smooth ?? false,
      data: s.data,
      itemStyle: { color: buildColor(s.color) },
      lineStyle: s.lineStyle ? {
        width: s.lineStyle.width ?? 2,
        type: s.lineStyle.type ?? 'solid',
      } : undefined,
      areaStyle: buildAreaStyle(s.areaStyle, s.color),
      symbol: s.symbol ?? 'circle',
      symbolSize: s.symbolSize ?? 4,
      showSymbol: s.showSymbol ?? true,
      step: s.step,
      connectNulls: s.connectNulls,
      label: getDataLabelConfig(),
      markLine: markLine ? {
        silent: true,
        data: markLine.data.map(m => ({
          type: m.type,
          yAxis: m.yAxis,
          name: m.name,
          lineStyle: { color: 'var(--color-text-secondary)' },
          label: { color: 'var(--color-text-secondary)' },
        })),
      } : undefined,
      markPoint: markPoint ? {
        data: markPoint.data.map(m => ({
          type: m.type,
          name: m.name ?? m.type,
        })),
      } : undefined,
    })),
  } as EChartsOption;

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }

  function handleInit(event: CustomEvent) {
    dispatch('init', event.detail);
  }

  // Expose chart methods
  export function getChart() {
    return chartRef?.getChart();
  }

  export function downloadImage(filename = 'line-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('line-chart', className)}>
  <EChart
    bind:this={chartRef}
    {option}
    {height}
    {loading}
    {animation}
    {dataZoom}
    {colors}
    on:click={handleClick}
    on:init={handleInit}
  />
</div>
