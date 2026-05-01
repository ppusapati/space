<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import type { WidgetChartConfig } from './report.types';
  import { reportClasses } from './report.types';
  import { buildChartOption } from './report.logic';
  import EChart from '../charts/EChart.svelte';
  import type { EChartsOption } from 'echarts';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let config: WidgetChartConfig;
  export let rows: Record<string, unknown>[] = [];
  export let title: string = '';
  export let height: string = '360px';
  export let loading: boolean = false;
  export let theme: string | object = '';

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher<{
    click: { params: unknown };
    legendChange: { params: unknown };
  }>();

  // ─── Computed ─────────────────────────────────────────────────────────────

  $: chartOption = buildChartOption(config, rows);

  // ─── Public API ───────────────────────────────────────────────────────────

  /** Export chart as image */
  export function exportImage(type: 'png' | 'jpeg' | 'svg' = 'png'): string | undefined {
    return chartRef?.exportImage({ type });
  }

  /** Download chart as file */
  export function downloadImage(filename: string = 'chart', type: 'png' | 'jpeg' | 'svg' = 'png') {
    chartRef?.downloadImage(filename, { type });
  }

  /** Get the underlying ECharts instance */
  export function getChart() {
    return chartRef?.getChart();
  }
</script>

<div class={cn('report-chart', className)}>
  {#if loading}
    <div class={reportClasses.widgetLoading} style="height: {height};">
      <div class="chart-skeleton">
        <div class="skeleton-bar" style="height: 60%;"></div>
        <div class="skeleton-bar" style="height: 80%;"></div>
        <div class="skeleton-bar" style="height: 45%;"></div>
        <div class="skeleton-bar" style="height: 90%;"></div>
        <div class="skeleton-bar" style="height: 55%;"></div>
        <div class="skeleton-bar" style="height: 70%;"></div>
      </div>
    </div>
  {:else if rows.length === 0}
    <div class={reportClasses.widgetEmpty} style="height: {height};">
      No data available
    </div>
  {:else}
    <EChart
      bind:this={chartRef}
      option={chartOption}
      {height}
      {theme}
      loading={false}
      autoResize={true}
      on:click={(e) => dispatch('click', { params: e.detail.params })}
      on:legendselectchanged={(e) => dispatch('legendChange', { params: e.detail.params })}
    />
  {/if}
</div>

<style lang="postcss">
  :global(.report-chart) {
    @apply w-full;
  }

  .chart-skeleton {
    @apply flex items-end justify-center gap-2 w-full h-full p-8;
  }

  .chart-skeleton .skeleton-bar {
    @apply flex-1 rounded-t bg-gray-200 animate-pulse;
  }

  :global(.report-widget__loading) {
    @apply flex items-center justify-center;
  }

  :global(.report-widget__empty) {
    @apply flex items-center justify-center text-sm text-gray-400;
  }
</style>
