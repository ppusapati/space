<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import type { ChartAnimationConfig, TooltipConfig, DataLabelConfig, WaterfallDataItem } from './types';

  // Props
  export let data: WaterfallDataItem[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showTooltip: boolean = true;
  export let height: string = '400px';
  export let loading: boolean = false;

  // Enhanced props
  export let animation: ChartAnimationConfig | boolean = true;
  export let tooltip: TooltipConfig = {};
  export let dataLabels: DataLabelConfig | boolean = true;
  export let positiveColor: string = '#91cc75';
  export let negativeColor: string = '#ee6666';
  export let totalColor: string = '#5470c6';
  export let barWidth: string | number = '40%';

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher();

  // Calculate waterfall data
  function calculateWaterfallData(items: WaterfallDataItem[]) {
    const placeholder: (number | '-')[] = [];
    const positive: (number | '-')[] = [];
    const negative: (number | '-')[] = [];
    const total: (number | '-')[] = [];

    let runningTotal = 0;

    items.forEach((item, index) => {
      if (item.isTotal) {
        placeholder.push('-');
        positive.push('-');
        negative.push('-');
        total.push(runningTotal);
      } else {
        const value = item.value;
        if (value >= 0) {
          placeholder.push(runningTotal);
          positive.push(value);
          negative.push('-');
          total.push('-');
        } else {
          placeholder.push(runningTotal + value);
          positive.push('-');
          negative.push(Math.abs(value));
          total.push('-');
        }
        runningTotal += value;
      }
    });

    return { placeholder, positive, negative, total };
  }

  $: categories = data.map(d => d.name);
  $: waterfallData = calculateWaterfallData(data);
  $: placeholder = waterfallData.placeholder;
  $: positive = waterfallData.positive;
  $: negative = waterfallData.negative;
  $: total = waterfallData.total;

  // Build data label config
  function getDataLabelConfig(): object | undefined {
    if (!dataLabels) return undefined;
    if (dataLabels === true) {
      return {
        show: true,
        position: 'top',
        color: 'var(--color-text-secondary)',
        fontSize: 11,
        formatter: (params: { value: number | string }) => {
          if (params.value === '-' || params.value === 0) return '';
          return typeof params.value === 'number' ? params.value.toLocaleString() : params.value;
        },
      };
    }
    return {
      show: dataLabels.show !== false,
      position: dataLabels.position ?? 'top',
      formatter: dataLabels.formatter,
      fontSize: dataLabels.fontSize ?? 11,
      fontWeight: dataLabels.fontWeight,
      color: dataLabels.color ?? 'var(--color-text-secondary)',
    };
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
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      formatter: tooltip.formatter ?? ((params: Array<{ seriesName: string; value: number | string; name: string }>) => {
        const item = data.find(d => d.name === params[0]!.name);
        if (!item) return '';
        if (item.isTotal) {
          return `${item.name}: ${item.value.toLocaleString()} (Total)`;
        }
        const sign = item.value >= 0 ? '+' : '';
        return `${item.name}: ${sign}${item.value.toLocaleString()}`;
      }),
    } : undefined,
    legend: { show: false },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: title ? '15%' : '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: categories,
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
      type: 'value',
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
    series: [
      {
        name: 'Placeholder',
        type: 'bar',
        stack: 'Total',
        barWidth,
        itemStyle: {
          borderColor: 'transparent',
          color: 'transparent',
        },
        emphasis: {
          itemStyle: {
            borderColor: 'transparent',
            color: 'transparent',
          },
        },
        data: placeholder,
      },
      {
        name: 'Positive',
        type: 'bar',
        stack: 'Total',
        barWidth,
        itemStyle: {
          color: positiveColor,
          borderRadius: [4, 4, 0, 0],
        },
        label: getDataLabelConfig(),
        data: positive,
      },
      {
        name: 'Negative',
        type: 'bar',
        stack: 'Total',
        barWidth,
        itemStyle: {
          color: negativeColor,
          borderRadius: [4, 4, 0, 0],
        },
        label: getDataLabelConfig(),
        data: negative,
      },
      {
        name: 'Total',
        type: 'bar',
        stack: 'Total',
        barWidth,
        itemStyle: {
          color: totalColor,
          borderRadius: [4, 4, 0, 0],
        },
        label: getDataLabelConfig(),
        data: total,
      },
    ],
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

  export function downloadImage(filename = 'waterfall-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('waterfall-chart', className)}>
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
