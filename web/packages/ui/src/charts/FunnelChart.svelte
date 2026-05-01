<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import type { ChartAnimationConfig, TooltipConfig, LegendConfig, FunnelDataItem } from './types';

  // Props
  export let data: FunnelDataItem[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showLegend: boolean = true;
  export let showTooltip: boolean = true;
  export let showLabels: boolean = true;
  export let height: string = '400px';
  export let loading: boolean = false;

  // Enhanced props
  export let animation: ChartAnimationConfig | boolean = true;
  export let tooltip: TooltipConfig = {};
  export let legend: LegendConfig = {};
  export let colors: string[] = [];
  export let sort: 'ascending' | 'descending' | 'none' = 'descending';
  export let orient: 'vertical' | 'horizontal' = 'vertical';
  export let align: 'left' | 'center' | 'right' = 'center';
  export let gap: number = 2;
  export let minSize: string = '0%';
  export let maxSize: string = '100%';
  export let labelPosition: 'left' | 'right' | 'inside' | 'insideLeft' | 'insideRight' = 'inside';
  export let funnelAlign: 'left' | 'center' | 'right' = 'center';

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher();

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
      trigger: 'item',
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      formatter: tooltip.formatter ?? '{a} <br/>{b} : {c}',
    } : undefined,
    legend: showLegend ? {
      ...getLegendPosition(),
      type: legend.type ?? 'plain',
      orient: legend.orient,
      align: legend.align,
      itemGap: legend.itemGap ?? 10,
      textStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    ...(colors.length > 0 ? { color: colors } : {}),
    series: [{
      name: title || 'Funnel',
      type: 'funnel',
      left: '10%',
      top: title ? 60 : 20,
      bottom: showLegend ? 60 : 20,
      width: '80%',
      min: 0,
      max: 100,
      minSize,
      maxSize,
      sort,
      orient,
      gap,
      funnelAlign,
      label: {
        show: showLabels,
        position: labelPosition,
        color: labelPosition === 'inside' ? '#fff' : 'var(--color-text-primary)',
        formatter: '{b}: {c}',
      },
      labelLine: {
        length: 10,
        lineStyle: {
          width: 1,
          color: 'var(--color-border-primary)',
        },
      },
      itemStyle: {
        borderColor: '#fff',
        borderWidth: 1,
      },
      emphasis: {
        label: {
          fontSize: 14,
          fontWeight: 'bold',
        },
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.3)',
        },
      },
      data: data.map(d => ({
        name: d.name,
        value: d.value,
        ...(d.color ? { itemStyle: { color: d.color } } : {}),
      })),
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

  export function downloadImage(filename = 'funnel-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('funnel-chart', className)}>
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
