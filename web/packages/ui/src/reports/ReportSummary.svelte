<script lang="ts">
  import { cn } from '../utils/classnames';
  import type { SummaryMetric } from './report.types';
  import { summaryClasses } from './report.types';
  import { computeAggregate, formatValue } from './report.logic';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let metrics: SummaryMetric[] = [];
  export let rows: Record<string, unknown>[] = [];
  export let aggregates: Record<string, number> = {};
  export let direction: 'horizontal' | 'vertical' = 'horizontal';
  export let title: string = '';
  export let loading: boolean = false;

  let className: string = '';
  export { className as class };

  // ─── Computed ─────────────────────────────────────────────────────────────

  function getValue(metric: SummaryMetric): string {
    const raw =
      aggregates[metric.field_code] ??
      computeAggregate(rows, metric.field_code, metric.aggregate);
    return formatValue(raw, metric.format);
  }
</script>

<div
  class={cn(
    summaryClasses.root,
    direction === 'horizontal' ? summaryClasses.horizontal : summaryClasses.vertical,
    className
  )}
>
  {#if loading}
    {#each Array(metrics.length || 3) as _}
      <div class={summaryClasses.metric}>
        <div class="skeleton-bar" style="width: 4rem; height: 1.5rem;"></div>
        <div class="skeleton-bar" style="width: 6rem; height: 0.875rem;"></div>
      </div>
    {/each}
  {:else}
    {#each metrics as metric (metric.field_code)}
      <div class={summaryClasses.metric}>
        {#if metric.icon}
          <div class={summaryClasses.metricIcon}>{metric.icon}</div>
        {/if}
        <div class={summaryClasses.metricValue}>{getValue(metric)}</div>
        <div class={summaryClasses.metricLabel}>{metric.label}</div>
      </div>
    {/each}
  {/if}
</div>

<style lang="postcss">
  :global(.report-summary) {
    @apply flex gap-6 p-2;
  }

  :global(.report-summary--horizontal) {
    @apply flex-row flex-wrap;
  }

  :global(.report-summary--vertical) {
    @apply flex-col;
  }

  :global(.report-summary__metric) {
    @apply flex flex-col items-center gap-0.5 min-w-[5rem];
  }

  :global(.report-summary--vertical .report-summary__metric) {
    @apply flex-row items-center gap-3;
  }

  :global(.report-summary__metric-icon) {
    @apply text-lg text-gray-400;
  }

  :global(.report-summary__metric-value) {
    @apply text-xl font-bold text-gray-900 tabular-nums;
  }

  :global(.report-summary__metric-label) {
    @apply text-xs text-gray-500;
  }

  .skeleton-bar {
    @apply rounded bg-gray-200 animate-pulse;
  }
</style>
