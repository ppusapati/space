<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import type { ChartAnimationConfig, TooltipConfig, DataZoomConfig, CandlestickDataItem } from './types';

  // Props
  export let data: CandlestickDataItem[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showTooltip: boolean = true;
  export let showVolume: boolean = true;
  export let height: string = '500px';
  export let loading: boolean = false;

  // Enhanced props
  export let animation: ChartAnimationConfig | boolean = true;
  export let tooltip: TooltipConfig = {};
  export let dataZoom: DataZoomConfig | DataZoomConfig[] | boolean = true;
  export let upColor: string = '#00da3c';
  export let downColor: string = '#ec0000';
  export let upBorderColor: string = '#008f28';
  export let downBorderColor: string = '#8a0000';
  export let showMA: boolean = true;
  export let maPeriods: number[] = [5, 10, 20];
  export let maColors: string[] = ['#ff6600', '#0066ff', '#00cc00'];

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher();

  // Calculate moving averages
  function calculateMA(dayCount: number, data: number[][]) {
    const result: (number | '-')[] = [];
    for (let i = 0; i < data.length; i++) {
      if (i < dayCount - 1) {
        result.push('-');
        continue;
      }
      let sum = 0;
      for (let j = 0; j < dayCount; j++) {
        sum += (data[i - j]![1])!; // Close price
      }
      result.push(sum / dayCount);
    }
    return result;
  }

  // Transform data for ECharts
  $: dates = data.map(d => d.date);
  $: ohlcData = data.map(d => [d.open, d.close, d.low, d.high]);
  $: volumeData = data.map((d, i) => ({
    value: d.volume ?? 0,
    itemStyle: {
      color: d.close >= d.open ? upColor : downColor,
    },
  }));

  // Calculate MAs
  $: maData = showMA
    ? maPeriods.map(period => calculateMA(period, ohlcData))
    : [];

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
      axisPointer: {
        type: 'cross',
        crossStyle: {
          color: 'var(--color-text-secondary)',
        },
      },
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      formatter: tooltip.formatter,
    } : undefined,
    legend: {
      top: title ? 40 : 10,
      left: 'center',
      data: showMA ? maPeriods.map(p => `MA${p}`) : [],
      textStyle: {
        color: 'var(--color-text-secondary)',
      },
    },
    grid: showVolume ? [
      {
        left: '10%',
        right: '8%',
        top: title ? 80 : 60,
        height: '50%',
      },
      {
        left: '10%',
        right: '8%',
        top: '72%',
        height: '16%',
      },
    ] : [{
      left: '10%',
      right: '8%',
      top: title ? 80 : 60,
      bottom: 80,
    }],
    xAxis: showVolume ? [
      {
        type: 'category',
        data: dates,
        boundaryGap: false,
        axisLine: { lineStyle: { color: 'var(--color-border-primary)' } },
        axisLabel: { color: 'var(--color-text-secondary)' },
        min: 'dataMin',
        max: 'dataMax',
        axisPointer: {
          z: 100,
        },
      },
      {
        type: 'category',
        gridIndex: 1,
        data: dates,
        boundaryGap: false,
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { show: false },
        min: 'dataMin',
        max: 'dataMax',
      },
    ] : [{
      type: 'category',
      data: dates,
      boundaryGap: false,
      axisLine: { lineStyle: { color: 'var(--color-border-primary)' } },
      axisLabel: { color: 'var(--color-text-secondary)' },
      min: 'dataMin',
      max: 'dataMax',
    }],
    yAxis: showVolume ? [
      {
        scale: true,
        splitArea: { show: false },
        axisLine: { lineStyle: { color: 'var(--color-border-primary)' } },
        axisLabel: { color: 'var(--color-text-secondary)' },
        splitLine: { lineStyle: { color: 'var(--color-border-secondary)' } },
      },
      {
        scale: true,
        gridIndex: 1,
        splitNumber: 2,
        axisLabel: { show: false },
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { show: false },
      },
    ] : [{
      scale: true,
      splitArea: { show: false },
      axisLine: { lineStyle: { color: 'var(--color-border-primary)' } },
      axisLabel: { color: 'var(--color-text-secondary)' },
      splitLine: { lineStyle: { color: 'var(--color-border-secondary)' } },
    }],
    dataZoom: dataZoom === true ? [
      { type: 'inside', xAxisIndex: showVolume ? [0, 1] : [0], start: 80, end: 100 },
      { type: 'slider', xAxisIndex: showVolume ? [0, 1] : [0], start: 80, end: 100, top: showVolume ? '92%' : undefined },
    ] : dataZoom === false ? undefined : Array.isArray(dataZoom) ? dataZoom : [dataZoom],
    series: [
      {
        name: 'Candlestick',
        type: 'candlestick',
        data: ohlcData,
        itemStyle: {
          color: upColor,
          color0: downColor,
          borderColor: upBorderColor,
          borderColor0: downBorderColor,
        },
      },
      ...maData.map((ma, i) => ({
        name: `MA${maPeriods[i]}`,
        type: 'line',
        data: ma,
        smooth: true,
        showSymbol: false,
        lineStyle: {
          width: 1,
          color: maColors[i] ?? maColors[0],
        },
      })),
      ...(showVolume ? [{
        name: 'Volume',
        type: 'bar',
        xAxisIndex: 1,
        yAxisIndex: 1,
        data: volumeData,
      }] : []),
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

  export function downloadImage(filename = 'candlestick-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('candlestick-chart', className)}>
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
