<script context="module" lang="ts">
  export interface RadarIndicator {
    name: string;
    max: number;
    min?: number;
  }

  export interface RadarSeriesData {
    name: string;
    value: number[];
    color?: string;
    areaStyle?: boolean;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let indicators: RadarIndicator[] = [];
  export let series: RadarSeriesData[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showLegend: boolean = true;
  export let showTooltip: boolean = true;
  export let height: string = '400px';
  export let loading: boolean = false;
  export let shape: 'polygon' | 'circle' = 'polygon';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher();

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
    tooltip: showTooltip ? {
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
    } : undefined,
    legend: showLegend ? {
      bottom: 0,
      textStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    radar: {
      indicator: indicators.map((ind) => ({
        name: ind.name,
        max: ind.max,
        min: ind.min ?? 0,
      })),
      shape: shape,
      axisName: {
        color: 'var(--color-text-secondary)',
      },
      splitArea: {
        areaStyle: {
          color: ['var(--color-surface-secondary)', 'var(--color-surface-primary)'],
        },
      },
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-border-secondary)',
        },
      },
    },
    series: [
      {
        type: 'radar',
        data: series.map((s) => ({
          value: s.value,
          name: s.name,
          itemStyle: s.color ? { color: s.color } : undefined,
          areaStyle: s.areaStyle ? { opacity: 0.3 } : undefined,
        })),
      },
    ],
  } as EChartsOption;

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }
</script>

<div class={cn('radar-chart', className)}>
  <EChart {option} {height} {loading} on:click={handleClick} />
</div>
