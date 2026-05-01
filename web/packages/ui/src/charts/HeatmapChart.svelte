<script context="module" lang="ts">
  export interface HeatmapData {
    x: string | number;
    y: string | number;
    value: number;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import type { ChartAnimationConfig, TooltipConfig, LegendConfig } from './types';

  // Props
  export let data: HeatmapData[] = [];
  export let xCategories: string[] = [];
  export let yCategories: string[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showTooltip: boolean = true;
  export let showVisualMap: boolean = true;
  export let height: string = '400px';
  export let loading: boolean = false;

  // Enhanced props
  export let animation: ChartAnimationConfig | boolean = true;
  export let tooltip: TooltipConfig = {};
  export let minValue: number | 'auto' = 'auto';
  export let maxValue: number | 'auto' = 'auto';
  export let colorRange: string[] = ['#313695', '#4575b4', '#74add1', '#abd9e9', '#e0f3f8', '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#a50026'];
  export let showLabel: boolean = false;
  export let labelFormatter: string | ((params: unknown) => string) = '{c}';
  export let itemBorderRadius: number = 0;
  export let itemBorderWidth: number = 1;
  export let itemBorderColor: string = '#fff';

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher();

  // Calculate min/max values
  $: computedMin = minValue === 'auto'
    ? Math.min(...data.map(d => d.value))
    : minValue;

  $: computedMax = maxValue === 'auto'
    ? Math.max(...data.map(d => d.value))
    : maxValue;

  // Format data for ECharts
  $: formattedData = data.map(d => {
    const xIndex = typeof d.x === 'string' ? xCategories.indexOf(d.x) : d.x;
    const yIndex = typeof d.y === 'string' ? yCategories.indexOf(d.y) : d.y;
    return [xIndex, yIndex, d.value];
  });

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
      position: 'top',
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      formatter: tooltip.formatter ?? ((params: { data: number[] }) => {
        const [x, y, value] = params.data;
        return `${xCategories[x!]}, ${yCategories[y!]}: ${value}`;
      }),
    } : undefined,
    grid: {
      left: '3%',
      right: showVisualMap ? '15%' : '4%',
      bottom: '3%',
      top: title ? '15%' : '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: xCategories,
      splitArea: {
        show: true,
      },
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      axisLabel: {
        color: 'var(--color-text-secondary)',
      },
    },
    yAxis: {
      type: 'category',
      data: yCategories,
      splitArea: {
        show: true,
      },
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      axisLabel: {
        color: 'var(--color-text-secondary)',
      },
    },
    visualMap: showVisualMap ? {
      min: computedMin,
      max: computedMax,
      calculable: true,
      orient: 'vertical',
      right: '2%',
      top: 'center',
      inRange: {
        color: colorRange,
      },
      textStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    series: [{
      type: 'heatmap',
      data: formattedData,
      label: {
        show: showLabel,
        color: 'var(--color-text-primary)',
        formatter: labelFormatter,
      },
      itemStyle: {
        borderRadius: itemBorderRadius,
        borderWidth: itemBorderWidth,
        borderColor: itemBorderColor,
      },
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowColor: 'rgba(0, 0, 0, 0.5)',
        },
      },
    }],
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

  export function downloadImage(filename = 'heatmap-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('heatmap-chart', className)}>
  <EChart
    bind:this={chartRef}
    {option}
    {height}
    {loading}
    {animation}
    on:click={handleClick}
    on:init={handleInit}
  />
</div>
