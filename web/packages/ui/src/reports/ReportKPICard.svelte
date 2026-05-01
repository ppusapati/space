<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import type { WidgetKPIConfig } from './report.types';
  import { kpiClasses, kpiSizeClasses } from './report.types';
  import { computeAggregate, computeTrend, isTrendGood, formatValue } from './report.logic';
  import EChart from '../charts/EChart.svelte';
  import type { EChartsOption } from 'echarts';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let config: WidgetKPIConfig;
  export let rows: Record<string, unknown>[] = [];
  export let aggregates: Record<string, number> = {};
  export let title: string = '';
  export let loading: boolean = false;
  export let size: 'sm' | 'md' | 'lg' = 'md';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    click: { field_code: string; value: number };
  }>();

  // ─── Computed ─────────────────────────────────────────────────────────────

  $: mainValue =
    aggregates[config.value_field_code] ??
    computeAggregate(rows, config.value_field_code, config.aggregate);

  $: compValue = config.comparison_field_code
    ? (aggregates[config.comparison_field_code] ??
       computeAggregate(rows, config.comparison_field_code, config.aggregate))
    : undefined;

  $: trend = compValue != null ? computeTrend(mainValue, compValue) : undefined;

  $: trendGood = trend ? isTrendGood(trend, config.trend_direction) : true;

  $: formattedValue = formatValue(mainValue, config.format);

  // Sparkline ECharts option
  $: sparklineOption = config.sparkline_field_code
    ? buildSparkline(rows, config.sparkline_field_code, config.sparkline_type ?? 'line')
    : null;

  // Threshold color
  $: thresholdColor = getThresholdColor(mainValue, config.thresholds);

  function getThresholdColor(
    val: number,
    thresholds?: { value: number; color: string }[]
  ): string | undefined {
    if (!thresholds?.length) return undefined;
    const sorted = [...thresholds].sort((a, b) => b.value - a.value);
    for (const t of sorted) {
      if (val >= t.value) return t.color;
    }
    return undefined;
  }

  function buildSparkline(
    data: Record<string, unknown>[],
    fieldCode: string,
    type: 'line' | 'bar'
  ): EChartsOption {
    const values = data.map((r) => Number(r[fieldCode] ?? 0)).filter((v) => !isNaN(v));
    return {
      grid: { left: 0, right: 0, top: 2, bottom: 2 },
      xAxis: { show: false, type: 'category', data: values.map((_, i) => i) },
      yAxis: { show: false, type: 'value' },
      series: [
        {
          type: type === 'bar' ? 'bar' : 'line',
          data: values,
          showSymbol: false,
          smooth: true,
          lineStyle: { width: 2, color: config.color ?? '#3b82f6' },
          areaStyle: type === 'line' ? { opacity: 0.1, color: config.color ?? '#3b82f6' } : undefined,
          itemStyle: type === 'bar' ? { color: config.color ?? '#3b82f6' } : undefined,
        },
      ],
      tooltip: { show: false },
      animation: false,
    };
  }
</script>

<div
  class={cn(kpiClasses.root, kpiSizeClasses[size], className)}
  style={thresholdColor ? `border-left: 4px solid ${thresholdColor}` : config.color ? `border-left: 4px solid ${config.color}` : ''}
  role="figure"
  aria-label={config.label}
  on:click={() => dispatch('click', { field_code: config.value_field_code, value: mainValue })}
  on:keydown={(e) => { if (e.key === 'Enter') dispatch('click', { field_code: config.value_field_code, value: mainValue }); }}
  tabindex="0"
>
  {#if loading}
    <div class="kpi-skeleton">
      <div class="skeleton-bar skeleton-value"></div>
      <div class="skeleton-bar skeleton-label"></div>
    </div>
  {:else}
    {#if config.icon}
      <div
        class={kpiClasses.icon}
        style={thresholdColor ? `color: ${thresholdColor}` : config.color ? `color: ${config.color}` : ''}
      >
        {config.icon}
      </div>
    {/if}

    <div
      class={kpiClasses.value}
      style={thresholdColor ? `color: ${thresholdColor}` : ''}
    >
      {formattedValue}
    </div>

    <div class={kpiClasses.label}>{config.label}</div>

    {#if trend && trend.direction !== 'flat'}
      <div
        class={cn(
          kpiClasses.trend,
          trendGood ? kpiClasses.trendGood : kpiClasses.trendBad
        )}
      >
        <span class={kpiClasses.trendArrow}>
          {trend.direction === 'up' ? '↑' : '↓'}
        </span>
        <span>{trend.value.toFixed(1)}%</span>
        {#if config.comparison_label}
          <span class={kpiClasses.trendLabel}>{config.comparison_label}</span>
        {/if}
      </div>
    {/if}

    {#if sparklineOption}
      <div class={kpiClasses.sparkline}>
        <EChart option={sparklineOption} height="40px" width="100%" />
      </div>
    {/if}
  {/if}
</div>

<style lang="postcss">
  :global(.report-kpi) {
    @apply flex flex-col items-start gap-1 p-3 cursor-pointer transition-shadow;
  }

  :global(.report-kpi:hover) {
    @apply shadow-sm;
  }

  :global(.report-kpi:focus-visible) {
    @apply outline-2 outline-blue-500 outline-offset-2;
  }

  /* Size variants */
  :global(.report-kpi--sm .report-kpi__value) {
    @apply text-xl;
  }

  :global(.report-kpi--md .report-kpi__value) {
    @apply text-2xl;
  }

  :global(.report-kpi--lg .report-kpi__value) {
    @apply text-4xl;
  }

  :global(.report-kpi__icon) {
    @apply text-2xl leading-none;
  }

  :global(.report-kpi__value) {
    @apply text-2xl font-bold text-gray-900 leading-tight;
  }

  :global(.report-kpi__label) {
    @apply text-sm text-gray-500;
  }

  :global(.report-kpi__trend) {
    @apply flex items-center gap-1 text-sm font-medium text-gray-500;
  }

  :global(.report-kpi__trend--good) {
    @apply text-emerald-600;
  }

  :global(.report-kpi__trend--bad) {
    @apply text-red-500;
  }

  :global(.report-kpi__trend--flat) {
    @apply text-gray-400;
  }

  :global(.report-kpi__trend-label) {
    @apply font-normal text-gray-400;
  }

  :global(.report-kpi__sparkline) {
    @apply mt-1 w-full;
  }

  /* Skeleton loading */
  .kpi-skeleton {
    @apply flex flex-col gap-2 w-full;
  }

  .skeleton-bar {
    @apply rounded bg-gray-200 animate-pulse;
  }

  .skeleton-value {
    @apply h-8 w-24;
  }

  .skeleton-label {
    @apply h-4 w-16;
  }
</style>
