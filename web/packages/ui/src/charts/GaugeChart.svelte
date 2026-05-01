<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let value: number = 0;
  export let min: number = 0;
  export let max: number = 100;
  export let title: string = '';
  export let unit: string = '%';
  export let showTooltip: boolean = true;
  export let height: string = '300px';
  export let loading: boolean = false;
  export let colors: [number, string][] = [
    [0.3, 'var(--color-success)'],
    [0.7, 'var(--color-warning)'],
    [1, 'var(--color-error)'],
  ];
  export let splitNumber: number = 10;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher();

  $: option = {
    tooltip: showTooltip ? {
      formatter: '{a} <br/>{b} : {c}' + unit,
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
    } : undefined,
    series: [
      {
        name: title || 'Gauge',
        type: 'gauge',
        min: min,
        max: max,
        splitNumber: splitNumber,
        axisLine: {
          lineStyle: {
            width: 15,
            color: colors,
          },
        },
        pointer: {
          itemStyle: {
            color: 'auto',
          },
          width: 5,
        },
        axisTick: {
          distance: -15,
          length: 8,
          lineStyle: {
            color: '#fff',
            width: 2,
          },
        },
        splitLine: {
          distance: -15,
          length: 15,
          lineStyle: {
            color: '#fff',
            width: 3,
          },
        },
        axisLabel: {
          color: 'var(--color-text-secondary)',
          distance: 25,
          fontSize: 12,
        },
        detail: {
          valueAnimation: true,
          formatter: `{value}${unit}`,
          color: 'var(--color-text-primary)',
          fontSize: 24,
          fontWeight: 'bold',
          offsetCenter: [0, '70%'],
        },
        title: {
          offsetCenter: [0, '90%'],
          fontSize: 14,
          color: 'var(--color-text-secondary)',
        },
        data: [
          {
            value: value,
            name: title,
          },
        ],
      },
    ],
  } as EChartsOption;

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }
</script>

<div class={cn('gauge-chart', className)}>
  <EChart {option} {height} {loading} on:click={handleClick} />
</div>
