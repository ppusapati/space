<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import type { WidgetTableConfig, ConditionalFormat } from './report.types';
  import { reportClasses } from './report.types';
  import { buildTableColumns, evaluateConditionalFormats, styleMapToString, computeAggregate } from './report.logic';
  import DataGrid from '../tables/DataGrid.svelte';
  import type { ExportFormat } from '../tables/table.export';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let config: WidgetTableConfig;
  export let rows: Record<string, unknown>[] = [];
  export let conditionalFormats: ConditionalFormat[] = [];
  export let title: string = '';
  export let loading: boolean = false;
  export let size: 'sm' | 'md' | 'lg' = 'md';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    rowClick: { row: Record<string, unknown>; index: number };
    sort: { column: string; direction: 'asc' | 'desc' | null };
    export: { format: ExportFormat };
    pageChange: { page: number; pageSize: number };
  }>();

  // ─── Computed ─────────────────────────────────────────────────────────────

  $: columns = buildTableColumns(config);

  $: displayRows = config.row_limit ? rows.slice(0, config.row_limit) : rows;

  $: pageSize = config.page_size ?? 10;

  // ─── Totals row ───────────────────────────────────────────────────────────

  $: totalsRow = config.show_totals
    ? buildTotalsRow(rows, config)
    : null;

  function buildTotalsRow(
    data: Record<string, unknown>[],
    cfg: WidgetTableConfig
  ): Record<string, unknown> | null {
    if (!cfg.show_totals || data.length === 0) return null;
    const totals: Record<string, unknown> = {};
    for (const col of cfg.columns) {
      const vals = data.map((r) => Number(r[col.field_code])).filter((v) => !isNaN(v));
      if (vals.length > 0 && (col.format?.type === 'number' || col.format?.type === 'currency' || col.format?.type === 'percent')) {
        totals[col.field_code] = vals.reduce((a, b) => a + b, 0);
      } else if (col === cfg.columns[0]) {
        totals[col.field_code] = 'Total';
      } else {
        totals[col.field_code] = '';
      }
    }
    return totals;
  }
</script>

<div class={cn('report-table', className)}>
  {#if loading}
    <div class="table-skeleton">
      {#each Array(5) as _}
        <div class="skeleton-row">
          {#each Array(Math.min(columns.length, 5)) as _}
            <div class="skeleton-cell"></div>
          {/each}
        </div>
      {/each}
    </div>
  {:else if rows.length === 0}
    <div class={reportClasses.widgetEmpty}>No data available</div>
  {:else}
    <DataGrid
      data={displayRows}
      {columns}
      {size}
      sortable={true}
      filterable={false}
      searchable={false}
      paginated={config.paginated ?? false}
      pagination={{ page: 1, pageSize, total: displayRows.length }}
      exportable={config.exportable ?? false}
      hoverable={true}
      on:rowClick={(e) => dispatch('rowClick', e.detail)}
      on:sort={(e) => dispatch('sort', e.detail)}
      on:export={(e) => dispatch('export', e.detail)}
      on:pageChange={(e) => dispatch('pageChange', e.detail)}
    />

    {#if totalsRow}
      <div class="totals-row">
        {#each columns as col}
          <div class="totals-cell" style="text-align: {col.align ?? 'left'}; width: {col.width ?? 'auto'};">
            <strong>{totalsRow[col.key] ?? ''}</strong>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style lang="postcss">
  :global(.report-table) {
    @apply w-full overflow-auto;
  }

  .table-skeleton {
    @apply flex flex-col gap-2 p-4;
  }

  .skeleton-row {
    @apply flex gap-3;
  }

  .skeleton-cell {
    @apply flex-1 h-6 rounded bg-gray-200 animate-pulse;
  }

  .totals-row {
    @apply flex border-t-2 border-gray-300 bg-gray-50 px-3 py-2 font-semibold text-sm;
  }

  .totals-cell {
    @apply flex-1 px-2;
  }
</style>
