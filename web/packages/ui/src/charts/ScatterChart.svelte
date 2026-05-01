<script context="module" lang="ts">
  export interface ScatterSeriesData {
    name: string;
    data: [number, number][] | [number, number, number][]; // [x, y] or [x, y, size]
    color?: string;
    symbolSize?: number | ((value: number[]) => number);
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let series: ScatterSeriesData[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showLegend: boolean = true;
  export let showTooltip: boolean = true;
  export let xAxisLabel: string = '';
  export let yAxisLabel: string = '';
  export let height: string = '400px';
  export let loading: boolean = false;
  export let xAxisMin: number | 'dataMin' = 'dataMin';
  export let xAxisMax: number | 'dataMax' = 'dataMax';
  export let yAxisMin: number | 'dataMin' = 'dataMin';
  export let yAxisMax: number | 'dataMax' = 'dataMax';

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
      trigger: 'item',
      formatter: (params: any) => {
        return `${params.seriesName}<br/>X: ${params.value[0]}<br/>Y: ${params.value[1]}${params.value[2] ? '<br/>Size: ' + params.value[2] : ''}`;
      },
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
    grid: {
      left: '3%',
      right: '4%',
      bottom: showLegend ? '15%' : '3%',
      top: title ? '15%' : '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: xAxisLabel,
      nameLocation: 'middle',
      nameGap: 30,
      min: xAxisMin,
      max: xAxisMax,
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      axisLabel: {
        color: 'var(--color-text-secondary)',
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-border-secondary)',
        },
      },
    },
    yAxis: {
      type: 'value',
      name: yAxisLabel,
      nameLocation: 'middle',
      nameGap: 50,
      min: yAxisMin,
      max: yAxisMax,
      axisLine: {
        lineStyle: {
          color: 'var(--color-border-primary)',
        },
      },
      axisLabel: {
        color: 'var(--color-text-secondary)',
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-border-secondary)',
        },
      },
    },
    series: series.map((s) => ({
      name: s.name,
      type: 'scatter',
      data: s.data,
      symbolSize: s.symbolSize ?? ((value: number[]) => {
        return value[2] ? Math.sqrt(value[2]) * 2 : 10;
      }),
      itemStyle: s.color ? { color: s.color } : undefined,
    })),
  } as EChartsOption;

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }
</script>

<div class={cn('scatter-chart', className)}>
  <EChart {option} {height} {loading} on:click={handleClick} />
</div>
