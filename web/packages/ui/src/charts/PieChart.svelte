<script context="module" lang="ts">
  export interface PieDataItem {
    name: string;
    value: number;
    color?: string;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let data: PieDataItem[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showLegend: boolean = true;
  export let showTooltip: boolean = true;
  export let showLabels: boolean = true;
  export let donut: boolean = false;
  export let roseType: boolean | 'radius' | 'area' = false;
  export let height: string = '400px';
  export let loading: boolean = false;
  export let innerRadius: string = '40%';
  export let outerRadius: string = '70%';

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
      formatter: '{a} <br/>{b}: {c} ({d}%)',
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
    } : undefined,
    legend: showLegend ? {
      orient: 'horizontal',
      bottom: 0,
      textStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    series: [
      {
        name: title || 'Data',
        type: 'pie',
        radius: donut ? [innerRadius, outerRadius] : outerRadius,
        center: ['50%', '50%'],
        roseType: roseType,
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: donut ? 8 : 4,
          borderColor: 'var(--color-surface-primary)',
          borderWidth: 2,
        },
        label: showLabels ? {
          show: true,
          formatter: '{b}: {d}%',
          color: 'var(--color-text-secondary)',
        } : {
          show: false,
        },
        labelLine: {
          show: showLabels,
          lineStyle: {
            color: 'var(--color-border-primary)',
          },
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold',
          },
        },
        data: data.map((item) => ({
          name: item.name,
          value: item.value,
          itemStyle: item.color ? { color: item.color } : undefined,
        })),
      },
    ],
  } as EChartsOption;

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }
</script>

<div class={cn('pie-chart', className)}>
  <EChart {option} {height} {loading} on:click={handleClick} />
</div>
