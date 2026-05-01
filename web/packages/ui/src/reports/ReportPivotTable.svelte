<script lang="ts">
  import { cn } from '../utils/classnames';
  import type { KPIAggregate, ReportFieldFormat } from './report.types';
  import { pivotClasses } from './report.types';
  import { computePivot, formatValue } from './report.logic';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let rowFields: string[] = [];
  export let colFields: string[] = [];
  export let valueField: string;
  export let aggregate: KPIAggregate = 'sum';
  export let rows: Record<string, unknown>[] = [];
  export let format: ReportFieldFormat | undefined = undefined;
  export let showTotals: boolean = true;
  export let title: string = '';
  export let loading: boolean = false;

  let className: string = '';
  export { className as class };

  // ─── Computed ─────────────────────────────────────────────────────────────

  $: pivot = computePivot(rows, rowFields, colFields, valueField, aggregate);

  function cellValue(rowKey: string, colKey: string): string {
    const val = pivot.cells.get(`${rowKey}|${colKey}`) ?? 0;
    return formatValue(val, format);
  }
</script>

<div class={cn(pivotClasses.root, className)}>
  {#if loading}
    <div class="pivot-skeleton">
      {#each Array(4) as _}
        <div class="skeleton-row">
          {#each Array(4) as _}
            <div class="skeleton-cell"></div>
          {/each}
        </div>
      {/each}
    </div>
  {:else if rows.length === 0}
    <div class="pivot-empty">No data available</div>
  {:else}
    <div class="pivot-scroll">
      <table class={pivotClasses.table}>
        <thead>
          <tr>
            <th class={pivotClasses.headerCell}>
              {rowFields.join(' / ')}
            </th>
            {#each pivot.colKeys as ck}
              <th class={pivotClasses.headerCell}>{ck}</th>
            {/each}
            {#if showTotals}
              <th class={cn(pivotClasses.headerCell, pivotClasses.totalCol)}>Total</th>
            {/if}
          </tr>
        </thead>
        <tbody>
          {#each pivot.rowKeys as rk}
            <tr>
              <td class={pivotClasses.rowHeader}>{rk}</td>
              {#each pivot.colKeys as ck}
                <td class={pivotClasses.cell}>{cellValue(rk, ck)}</td>
              {/each}
              {#if showTotals}
                <td class={cn(pivotClasses.cell, pivotClasses.totalCol)}>
                  {formatValue(pivot.rowTotals.get(rk) ?? 0, format)}
                </td>
              {/if}
            </tr>
          {/each}

          {#if showTotals}
            <tr class={pivotClasses.totalRow}>
              <td class={pivotClasses.rowHeader}>Total</td>
              {#each pivot.colKeys as ck}
                <td class={pivotClasses.cell}>
                  {formatValue(pivot.colTotals.get(ck) ?? 0, format)}
                </td>
              {/each}
              <td class={cn(pivotClasses.cell, pivotClasses.grandTotal)}>
                {formatValue(pivot.grandTotal, format)}
              </td>
            </tr>
          {/if}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style lang="postcss">
  :global(.report-pivot) {
    @apply w-full;
  }

  .pivot-scroll {
    @apply overflow-auto;
  }

  :global(.report-pivot__table) {
    @apply w-full border-collapse text-sm;
  }

  :global(.report-pivot__header-cell) {
    @apply px-3 py-2 text-left font-semibold text-gray-700 bg-gray-100 border border-gray-200 whitespace-nowrap;
  }

  :global(.report-pivot__row-header) {
    @apply px-3 py-2 font-medium text-gray-800 bg-gray-50 border border-gray-200 whitespace-nowrap;
  }

  :global(.report-pivot__cell) {
    @apply px-3 py-2 text-right text-gray-700 border border-gray-200 tabular-nums;
  }

  :global(.report-pivot__total-row) {
    @apply bg-gray-100 font-semibold;
  }

  :global(.report-pivot__total-col) {
    @apply bg-gray-50 font-semibold;
  }

  :global(.report-pivot__grand-total) {
    @apply bg-gray-200 font-bold;
  }

  .pivot-skeleton {
    @apply flex flex-col gap-2 p-4;
  }

  .pivot-skeleton .skeleton-row {
    @apply flex gap-2;
  }

  .pivot-skeleton .skeleton-cell {
    @apply flex-1 h-6 rounded bg-gray-200 animate-pulse;
  }

  .pivot-empty {
    @apply flex items-center justify-center h-24 text-sm text-gray-400;
  }
</style>
