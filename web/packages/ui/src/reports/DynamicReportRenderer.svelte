<script lang="ts">
  import { createEventDispatcher, tick } from 'svelte';
  import type {
    ReportVisualization,
    ReportData,
    ReportWidgetType,
  } from './report.types';
  import type { ReportWidget } from './report.types';
  import { cn } from '../utils/classnames';
  import { reportClasses } from './report.types';
  import { widgetGridStyle } from './report.logic';
  import { exportReport, exportWidget } from './report.export';
  import type { ReportExportFormat } from './report.export';

  // Import all widget components
  import ReportKPICard from './ReportKPICard.svelte';
  import ReportChart from './ReportChart.svelte';
  import ReportTable from './ReportTable.svelte';
  import ReportPivotTable from './ReportPivotTable.svelte';
  import ReportSummary from './ReportSummary.svelte';
  import ReportExportMenu from './ReportExportMenu.svelte';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let visualization: ReportVisualization;
  export let data: ReportData;
  export let loading: boolean = false;
  export let theme: string | object = '';

  /** Report title — used in export headers */
  export let title: string = 'Report';
  /** Filename for exports (no extension) */
  export let exportFilename: string = 'report';
  /** Company name for PDF/print header */
  export let companyName: string = '';
  /** Show the report-level export toolbar */
  export let showExportToolbar: boolean = true;
  /** Show per-widget export buttons */
  export let showWidgetExport: boolean = true;
  /** Available export formats */
  export let exportFormats: ReportExportFormat[] = ['csv', 'xlsx', 'pdf', 'print'];

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    drilldown: { widget_id: string; field_code: string; value: unknown };
    widgetClick: { widget_id: string; params: unknown };
    widgetExport: { widget_id: string; format: ReportExportFormat };
    reportExport: { format: ReportExportFormat };
  }>();

  // ─── Chart refs for image capture ─────────────────────────────────────────

  let chartRefMap: Record<string, ReportChart> = {};
  let isExporting = false;

  /**
   * Capture all chart widgets as PNG data URLs for PDF/print embedding.
   */
  async function captureChartImages(): Promise<Map<string, string>> {
    const images = new Map<string, string>();
    await tick(); // ensure charts are rendered
    for (const [widgetId, chartRef] of Object.entries(chartRefMap)) {
      try {
        const img = chartRef?.exportImage('png');
        if (img) images.set(widgetId, img);
      } catch {
        // chart may not be available
      }
    }
    return images;
  }

  // ─── Report-Level Export ───────────────────────────────────────────────────

  async function handleReportExport(e: CustomEvent<{ format: ReportExportFormat }>) {
    const format = e.detail.format;
    isExporting = true;
    dispatch('reportExport', { format });

    try {
      const chartImages = (format === 'pdf' || format === 'print')
        ? await captureChartImages()
        : undefined;

      await exportReport(format, {
        visualization,
        data,
        filename: exportFilename,
        title,
        companyName,
        chartImages,
      });
    } catch (err) {
      console.error('Report export failed:', err);
    } finally {
      isExporting = false;
    }
  }

  // ─── Per-Widget Export ────────────────────────────────────────────────────

  async function handleWidgetExport(widget: ReportWidget, format: ReportExportFormat) {
    dispatch('widgetExport', { widget_id: widget.widget_id, format });

    try {
      await exportWidget({
        widget,
        rows: getWidgetRows(),
        filename: `${exportFilename}_${widget.widget_id}`,
        format,
        title: widget.title,
      });
    } catch (err) {
      console.error(`Widget export failed (${widget.widget_id}):`, err);
    }
  }

  // ─── Component Map (mirrors DynamicFormRenderer.getFieldComponent) ────────

  function getWidgetComponent(widgetType: ReportWidgetType): any {
    const componentMap: Record<ReportWidgetType, any> = {
      kpi_card: ReportKPICard,
      chart: ReportChart,
      table: ReportTable,
      pivot_table: ReportPivotTable,
      summary: ReportSummary,
    };
    return componentMap[widgetType] ?? null;
  }

  // ─── Data Access ──────────────────────────────────────────────────────────

  function getWidgetRows(): Record<string, unknown>[] {
    return data?.rows ?? [];
  }

  function getAggregates(): Record<string, number> {
    return data?.aggregates ?? {};
  }

  // ─── Drilldown ────────────────────────────────────────────────────────────

  function handleWidgetClick(widgetId: string, detail: Record<string, unknown>) {
    dispatch('widgetClick', { widget_id: widgetId, params: detail });

    const fieldCode = detail.field_code as string | undefined;
    if (fieldCode && visualization.drilldowns?.length) {
      const drilldown = visualization.drilldowns.find(
        (d: { source_widget_id: string; source_field_code: string }) => d.source_widget_id === widgetId && d.source_field_code === fieldCode
      );
      if (drilldown) {
        dispatch('drilldown', {
          widget_id: widgetId,
          field_code: fieldCode,
          value: detail.value,
        });
      }
    }
  }

  // ─── Computed ─────────────────────────────────────────────────────────────

  $: visibleWidgets = (visualization?.widgets ?? []).filter((w: ReportWidget) => !w.hidden);

  $: layoutClass = cn(
    reportClasses.renderer,
    visualization?.layout_mode === 'flow' && 'report-renderer--flow',
    visualization?.layout_mode === 'tabs' && 'report-renderer--tabs',
    className
  );
</script>

<!-- ─── Report Toolbar ──────────────────────────────────────────────────── -->

{#if showExportToolbar}
  <div class="report-toolbar">
    <div class="report-toolbar__left">
      <slot name="toolbar-left" />
    </div>
    <div class="report-toolbar__right">
      <slot name="toolbar-right" />
      <ReportExportMenu
        formats={exportFormats}
        disabled={loading}
        loading={isExporting}
        on:export={handleReportExport}
      />
    </div>
  </div>
{/if}

<!-- ─── Widget Grid ─────────────────────────────────────────────────────── -->

<div class={layoutClass}>
  {#each visibleWidgets as widget (widget.widget_id)}
    {@const Component = getWidgetComponent(widget.widget_type)}

    <div
      class={cn(reportClasses.widget, widget.css_class)}
      style={widgetGridStyle(widget)}
    >
      <!-- Widget Header -->
      <div class={reportClasses.widgetHeader}>
        <h3 class={reportClasses.widgetTitle}>{widget.title}</h3>
        <div class={reportClasses.widgetActions}>
          <slot name="widget-actions" {widget} />
          {#if showWidgetExport && (widget.widget_type === 'table' || widget.widget_type === 'chart' || widget.widget_type === 'kpi_card')}
            <ReportExportMenu
              formats={widget.widget_type === 'chart' ? ['pdf', 'print'] : exportFormats}
              size="sm"
              label=""
              disabled={loading}
              on:export={(e) => handleWidgetExport(widget, e.detail.format)}
            />
          {/if}
        </div>
      </div>

      <!-- Widget Body -->
      <div class={reportClasses.widgetBody}>
        {#if Component == null}
          <div class={reportClasses.widgetEmpty}>
            Unsupported widget type: {widget.widget_type}
          </div>

        {:else if widget.widget_type === 'kpi_card' && widget.kpi_config}
          <ReportKPICard
            config={widget.kpi_config}
            rows={getWidgetRows()}
            aggregates={getAggregates()}
            title={widget.title}
            {loading}
            on:click={(e) => handleWidgetClick(widget.widget_id, e.detail)}
          />

        {:else if widget.widget_type === 'chart' && widget.chart_config}
          <ReportChart
            bind:this={chartRefMap[widget.widget_id]}
            config={widget.chart_config}
            rows={getWidgetRows()}
            title={widget.title}
            {loading}
            {theme}
            on:click={(e) => handleWidgetClick(widget.widget_id, e.detail)}
          />

        {:else if widget.widget_type === 'table' && widget.table_config}
          <ReportTable
            config={widget.table_config}
            rows={getWidgetRows()}
            conditionalFormats={visualization.conditional_formats}
            title={widget.title}
            {loading}
            on:rowClick={(e) => handleWidgetClick(widget.widget_id, e.detail)}
            on:export={(e) => handleWidgetExport(widget, e.detail.format)}
          />

        {:else if widget.widget_type === 'pivot_table'}
          <ReportPivotTable
            rowFields={widget.field_codes?.slice(0, 1) ?? []}
            colFields={widget.field_codes?.slice(1, 2) ?? []}
            valueField={widget.field_codes?.[2] ?? ''}
            rows={getWidgetRows()}
            title={widget.title}
            {loading}
          />

        {:else if widget.widget_type === 'summary'}
          <ReportSummary
            metrics={[]}
            rows={getWidgetRows()}
            aggregates={getAggregates()}
            title={widget.title}
            {loading}
          />

        {:else}
          <div class={reportClasses.widgetEmpty}>
            No configuration provided for {widget.widget_type}
          </div>
        {/if}
      </div>
    </div>
  {/each}
</div>

<style lang="postcss">
  /* ─── Toolbar ────────────────────────────────────────────────────────── */
  .report-toolbar {
    @apply flex items-center justify-between mb-4 px-1;
  }

  .report-toolbar__left {
    @apply flex items-center gap-2;
  }

  .report-toolbar__right {
    @apply flex items-center gap-2;
  }

  /* ─── 24-column grid layout ──────────────────────────────────────────── */
  :global(.report-renderer) {
    @apply grid w-full gap-4;
    grid-template-columns: repeat(24, 1fr);
    min-height: 200px;
  }

  :global(.report-renderer--flow) {
    @apply flex flex-col;
  }

  :global(.report-renderer--tabs) {
    @apply block;
  }

  /* ─── Widget card ────────────────────────────────────────────────────── */
  :global(.report-widget) {
    @apply flex flex-col overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm;
    min-height: 0;
  }

  :global(.report-widget__header) {
    @apply flex items-center justify-between px-4 py-3 border-b border-gray-100;
  }

  :global(.report-widget__title) {
    @apply text-sm font-semibold text-gray-800 m-0;
  }

  :global(.report-widget__actions) {
    @apply flex items-center gap-1;
  }

  :global(.report-widget__body) {
    @apply flex-1 p-4 overflow-auto;
    min-height: 0;
  }

  :global(.report-widget__loading) {
    @apply flex items-center justify-center min-h-[120px];
  }

  :global(.report-widget__empty) {
    @apply flex items-center justify-center min-h-[80px] text-sm text-gray-400;
  }
</style>
